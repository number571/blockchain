package blockchain

import (
    "os"
    "fmt"
    "bytes"
    "encoding/json"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

func NewChain(receiver string) *BlockChain {
    err := createFile(DB_FILENAME)
    if err != nil {
        panic("can't create database")
    }
    db, err := sql.Open("sqlite3", DB_FILENAME)
    if err != nil {
        panic("can't open database")
    }
    _, err = db.Exec(`
CREATE TABLE IF NOT EXISTS BlockChain (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    Hash VARCHAR(44) UNIQUE,
    Block TEXT
);
`)
    chain := &BlockChain{
        DB: db,
    }
    if !chain.GenesisExist() {
        genesis := &Block{
            CurrHash: []byte(GENESIS_BLOCK),
            Mapping: make(map[string]uint64),
        }
        genesis.Mapping[STORAGE_CHAIN] = STORAGE_VALUE
        genesis.Mapping[receiver] = GENESIS_REWARD
        chain.AddBlock(genesis)
    }
    chain.Index = chain.Size()
    return chain
}

func LoadChain() *BlockChain {
    db, err := sql.Open("sqlite3", DB_FILENAME)
    if err != nil {
        panic("can't open database")
    }
    chain := &BlockChain{
        DB: db,
    }
    chain.Index = chain.Size()
    return chain
}

func (chain *BlockChain) Size() uint64 {
    var id uint64
    row := chain.DB.QueryRow("SELECT Id FROM BlockChain ORDER BY Id DESC")
    row.Scan(&id)
    return id
}

func (chain *BlockChain) GenesisExist() bool {
    var hash string 
    row := chain.DB.QueryRow("SELECT Hash FROM BlockChain WHERE Hash=$1", Base64Encode([]byte(GENESIS_BLOCK)))
    err := row.Scan(&hash)
    if err != nil {
        return false
    }
    return true
}

func (chain *BlockChain) AddBlock(block *Block) {
    chain.Index += 1
    chain.DB.Exec("INSERT INTO BlockChain (Hash, Block) VALUES ($1, $2)", 
        Base64Encode(block.CurrHash),
        serialize(block),
    )
}

func (chain *BlockChain) PrintChain() error {
    rows, err := chain.DB.Query("SELECT Id, Block FROM BlockChain")
    if err != nil {
        return err
    }
    defer rows.Close()
    var (
        sblock string
        block  *Block
        index  uint64
    )
    for rows.Next() {
        rows.Scan(&index, &sblock)
        block = deserialize(sblock)
        if DEBUG && index == 1 {
            if !block.IsGenesis() {
                fmt.Printf("[%d][FAILED] Genesis block undefined\n", index)
            } else {
                fmt.Printf("[%d][SUCCESS] Genesis block found\n", index)
            }
            continue
        }
        if DEBUG {
            if block.Difficulty != DIFFICULTY {
                fmt.Printf("[%d][FAILED] difficulty in block not equal constant value\n", index)
            } else {
                fmt.Printf("[%d][SUCCESS] difficulty in block equal constant value\n", index)
            }

            if !block.HashIsValid() {
                fmt.Printf("[%d][FAILED] hash block is not valid\n", index)
            } else {
                fmt.Printf("[%d][SUCCESS] hash block is valid\n", index)
            }

            if !block.SignIsValid() {
                fmt.Printf("[%d][FAILED] sign block is not valid\n", index)
            } else {
                fmt.Printf("[%d][SUCCESS] sign block is valid\n", index)
            }

            if !block.ProofIsValid() {
                fmt.Printf("[%d][FAILED] proof block is not valid\n", index)
            } else {
                fmt.Printf("[%d][SUCCESS] proof block is valid\n", index)
            }

            if !block.MappingIsValid() {
                fmt.Printf("[%d][FAILED] mapping block is not valid\n", index)
            } else {
                fmt.Printf("[%d][SUCCESS] mapping block is valid\n", index)
            }

            chain.Index = index - 1
            if !chain.TransactionsIsValid(block) {
                fmt.Printf("[%d][FAILED] transactions block is not valid\n", index)
            } else {
                fmt.Printf("[%d][SUCCESS] transactions block is valid\n", index)
            }
            chain.Index = chain.Size()
        }
        if !DEBUG {
            fmt.Printf("[%d] => %s\n", index, sblock)
        }
    }
    return nil
}

func (chain *BlockChain) TransactionsIsValid(block *Block) bool {
    lentxs := len(block.Transactions)
    plusStorage := 0
    for i := 0; i < lentxs; i++ {
        if block.Transactions[i].Sender == STORAGE_CHAIN {
            plusStorage = 1
            break
        }
    }
    if lentxs == 0 || lentxs > TXS_LIMIT + plusStorage {
        if DEBUG {
            fmt.Println("(TransactionsIsValid) lentxs == 0 || lentxs > TXS_LIMIT + plusStorage")
        }
        return false
    }
    for i := 0; i < lentxs-1; i++ {
        for j := i+1; j < lentxs; j++ {
            // rand bytes not be equal
            if bytes.Equal(block.Transactions[i].RandBytes, block.Transactions[j].RandBytes) {
                if DEBUG {
                    fmt.Println("(TransactionsIsValid) bytes.Equal(block.Transactions[i].RandBytes, block.Transactions[j].RandBytes)")
                }
                return false
            }
            // storage tx only one
            if block.Transactions[i].Sender == STORAGE_CHAIN && block.Transactions[j].Sender == STORAGE_CHAIN {
                if DEBUG {
                    fmt.Println("(TransactionsIsValid) block.Transactions[i].Sender == STORAGE_CHAIN && block.Transactions[j].Sender == STORAGE_CHAIN")
                }
                return false
            }
        }
    }
    for i := 0; i < lentxs; i++ {
        tx := block.Transactions[i]
        // storage tx has no hash and signature
        if tx.Sender == STORAGE_CHAIN {
            if tx.Receiver != block.Miner || tx.Value != STORAGE_REWARD {
                if DEBUG {
                    fmt.Println("(TransactionsIsValid) tx.Receiver != block.Miner || tx.Value != STORAGE_REWARD")
                }
                return false
            }
        } else {
            if !tx.HashIsValid() {
                if DEBUG {
                    fmt.Println("(TransactionsIsValid) !tx.HashIsValid()")
                }
                return false
            }
            if !tx.SignIsValid() {
                if DEBUG {
                    fmt.Println("(TransactionsIsValid) !tx.SignIsValid()")
                }
                return false
            }
        }
        if !chain.BalanceIsValid(block, tx.Sender) {
            if DEBUG {
                fmt.Println("(TransactionsIsValid) !chain.BalanceIsValid(block, tx.Sender)")
            }
            return false
        }
        if !chain.BalanceIsValid(block, tx.Receiver) {
            if DEBUG {
                fmt.Println("(TransactionsIsValid) !chain.BalanceIsValid(block, tx.Receiver)")
            }
            return false
        }
    }
    return true
}

func (chain *BlockChain) BalanceIsValid(block *Block, address string) bool {
    lentxs := len(block.Transactions)
    balanceInChain := chain.Balance(address)
    balanceSubBlock := uint64(0)
    balanceAddBlock := uint64(0)
    for j := 0; j < lentxs; j++ {
        tx := block.Transactions[j]
        if tx.Sender == address {
            balanceSubBlock += tx.Value + tx.ToStorage
        }
        if tx.Receiver == address {
            balanceAddBlock += tx.Value
        }
        if STORAGE_CHAIN == address {
            balanceAddBlock += tx.ToStorage
        }
    }
    if _, ok := block.Mapping[address]; !ok {
        if DEBUG {
            fmt.Println("(BalanceIsValid) _, ok := block.Mapping[address]; !ok")
        }
        return false
    }
    if (balanceInChain + balanceAddBlock - balanceSubBlock) != block.Mapping[address] {
        if DEBUG {
            fmt.Println("(BalanceIsValid) (balanceInChain + balanceAddBlock - balanceSubBlock) != block.Mapping[address]")
        }
        return false
    }
    return true
}

func (chain *BlockChain) Balance(address string) uint64 {
    var (
        sblock string
        block *Block
        balance uint64
    )
    rows, err := chain.DB.Query("SELECT Block FROM BlockChain WHERE Id <= $1 ORDER BY Id DESC", chain.Index)
    if err != nil {
        if DEBUG {
            fmt.Println("(Balance) err != nil")
        }
        return balance
    }
    defer rows.Close()
    for rows.Next() {
        rows.Scan(&sblock)
        block = deserialize(sblock)
        if value, ok := block.Mapping[address]; ok {
            balance = value
            break
        }
    }
    return balance
}

func (chain *BlockChain) LastHash() []byte {
    var hash string
    row := chain.DB.QueryRow("SELECT Hash FROM BlockChain ORDER BY Id DESC")
    row.Scan(&hash)
    return Base64Decode(hash)
}

func (chain *BlockChain) PushBlock(user *User, block *Block) {
    if !chain.TransactionsIsValid(block) {
        return
    }
    block.AddTransaction(chain, &Transaction{
        RandBytes: GenerateRandomBytes(RAND_BYTES),
        Sender: STORAGE_CHAIN,
        Receiver: user.Address(),
        Value: STORAGE_REWARD,
    })
    block.CurrHash  = block.Hash()
    block.Signature = block.Sign(user.Private())
    block.Nonce     = block.Proof()
    chain.AddBlock(block)
}

func serialize(data interface{}) string {
    jsonData, _ := json.MarshalIndent(data, "", "\t")
    return string(jsonData)
}

func deserialize(data string) *Block {
    var block Block
    json.Unmarshal([]byte(data), &block)
    return &block
}

func printJSON(data interface{}) {
    fmt.Println(serialize(data))
}

func fileIsExist(filename string) bool {
    if _, err := os.Stat(filename); os.IsNotExist(err) {
        return false
    }
    return true
}

func createFile(filename string) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    file.Close()
    return nil
}
