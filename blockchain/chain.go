package blockchain

import (
    "os"
    "fmt"
    "time"
    "bytes"
    "errors"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

func NewChain(filename, receiver string) error {
    file, err := os.Create(filename)
    if err != nil {
        return errors.New("create database")
    }
    file.Close()
    db, err := sql.Open("sqlite3", filename)
    if err != nil {
        return errors.New("open database")
    }
    defer db.Close()
    _, err = db.Exec(CREATE_TABLE)
    chain := &BlockChain{
        DB: db,
    }
    genesis := &Block{
        CurrHash: []byte(GENESIS_BLOCK),
        Mapping: make(map[string]uint64),
        Miner: receiver,
        TimeStamp: time.Now().Format(time.RFC3339),
    }
    genesis.Mapping[STORAGE_CHAIN] = STORAGE_VALUE
    genesis.Mapping[receiver] = GENESIS_REWARD
    chain.AddBlock(genesis)
    return nil
}

func LoadChain(filename string) *BlockChain {
    db, err := sql.Open("sqlite3", filename)
    if err != nil {
        return nil 
    }
    chain := &BlockChain{
        DB: db,
    }
    chain.index = chain.Size()
    return chain
}

func (chain *BlockChain) BlockIsValid(block *Block) bool {
    // size := chain.index
    // chain.index = size - 1
    // defer func() {
    //     chain.index = size
    // }()
    switch {
    case block == nil: 
        // println(1)
        return false
    case block.Difficulty != DIFFICULTY: 
        // println(2)
        return false
    case !block.hashIsValid(): 
        // println(3)
        return false
    case !chain.hashIsValid(block, chain.Size()): 
        // println(4)
        return false
    case !block.signIsValid(): 
        // println(5)
        return false
    case !block.proofIsValid(): 
        // println(6)
        return false
    case !block.mappingIsValid(): 
        // println(7)
        return false
    case !chain.timeIsValid(block, chain.Size()):
        // println(8)
        return false
    case !chain.transactionsIsValid(block): 
        // println(9)
        return false
    }
    return true
}

func (chain *BlockChain) Size() uint64 {
    var index uint64
    row := chain.DB.QueryRow("SELECT Id FROM BlockChain ORDER BY Id DESC")
    row.Scan(&index)
    return index
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
        size   uint64
    )
    for rows.Next() {
        rows.Scan(&index, &sblock)
        block = DeserializeBlock(sblock)

        if index == 1 {
            if !bytes.Equal(block.CurrHash, []byte(GENESIS_BLOCK)) {
                fmt.Printf("[%d][FAILED] Genesis block undefined\n", index)
            } else {
                fmt.Printf("[%d][SUCCESS] Genesis block found\n", index)
            }
            goto print
        }

        if block.Difficulty != DIFFICULTY {
            fmt.Printf("[%d][FAILED] difficulty is not valid\n", index)
        } else {
            fmt.Printf("[%d][SUCCESS] difficulty is valid\n", index)
        }

        if !block.hashIsValid() {
            fmt.Printf("[%d][FAILED] block hash is not valid\n", index)
        } else {
            fmt.Printf("[%d][SUCCESS] block hash is valid\n", index)
        }

        if !chain.hashIsValid(block, index - 1) {
            fmt.Printf("[%d][FAILED] chain hash is not valid\n", index)
        } else {
            fmt.Printf("[%d][SUCCESS] chain hash is valid\n", index)
        }

        if !block.signIsValid() {
            fmt.Printf("[%d][FAILED] sign is not valid\n", index)
        } else {
            fmt.Printf("[%d][SUCCESS] sign is valid\n", index)
        }

        if !block.proofIsValid() {
            fmt.Printf("[%d][FAILED] proof is not valid\n", index)
        } else {
            fmt.Printf("[%d][SUCCESS] proof is valid\n", index)
        }

        if !block.mappingIsValid() {
            fmt.Printf("[%d][FAILED] mapping is not valid\n", index)
        } else {
            fmt.Printf("[%d][SUCCESS] mapping is valid\n", index)
        }

        if !chain.timeIsValid(block, index - 1) {
            fmt.Printf("[%d][FAILED] time is not valid\n", index)
        } else {
            fmt.Printf("[%d][SUCCESS] time is valid\n", index)
        }

        size = chain.index
        chain.index = index - 1
        if !chain.transactionsIsValid(block) {
            fmt.Printf("[%d][FAILED] transactions is not valid\n", index)
        } else {
            fmt.Printf("[%d][SUCCESS] transactions is valid\n", index)
        }
        chain.index = size

print:
        fmt.Printf("[%d] => %s\n\n", index, sblock)
    }
    return nil
}

func (chain *BlockChain) Balance(address string) uint64 {
    var (
        sblock string
        block *Block
        balance uint64
    )
    rows, err := chain.DB.Query("SELECT Block FROM BlockChain WHERE Id <= $1 ORDER BY Id DESC", chain.index)
    if err != nil {
        return balance
    }
    defer rows.Close()
    for rows.Next() {
        rows.Scan(&sblock)
        block = DeserializeBlock(sblock)
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

func (chain *BlockChain) AcceptBlock(user *User, block *Block, ch chan bool) *Block {
    if !chain.transactionsIsValid(block) {
        return nil
    }
    block.AddTransaction(chain, &Transaction{
        RandBytes: GenerateRandomBytes(RAND_BYTES),
        Sender: STORAGE_CHAIN,
        Receiver: user.Address(),
        Value: STORAGE_REWARD,
    })
    block.TimeStamp = time.Now().Format(time.RFC3339)
    block.CurrHash  = block.hash()
    block.Signature = block.sign(user.Private())
    block.Nonce     = block.proof(ch)
    return block
}

func (chain *BlockChain) AddBlock(block *Block) {
    chain.index += 1
    chain.DB.Exec("INSERT INTO BlockChain (Hash, Block) VALUES ($1, $2)", 
        Base64Encode(block.CurrHash),
        SerializeBlock(block),
    )
}

func (chain *BlockChain) timeIsValid(block *Block, index uint64) bool {
    btime, err := time.Parse(time.RFC3339, block.TimeStamp)
    if err != nil {
        // fmt.Println("@", 1)
        return false
    }

    diff := time.Now().Sub(btime)
    if diff < 0 {
        return false
    }

    var sblock string
    row := chain.DB.QueryRow("SELECT Block FROM BlockChain WHERE Hash=$1", Base64Encode(block.PrevHash))
    row.Scan(&sblock)

    lblock := DeserializeBlock(sblock)
    if lblock == nil {
        // fmt.Println("@", 2)
        return false
    }

    ltime, err := time.Parse(time.RFC3339, lblock.TimeStamp)
    if err != nil {
        // fmt.Println("@", 3)
        return false
    }

    result := btime.Sub(ltime)

    // fmt.Println(SerializeBlock(block))
    // fmt.Println()
    // fmt.Println(SerializeBlock(lblock))
    // fmt.Println()

    // fmt.Printf("@:: %v - %v\n", btime, ltime)
    // fmt.Printf("@:: %v - %v\n", result, TIME_SESSION)

    return result >= TIME_SESSION
}

func (chain *BlockChain) hashIsValid(block *Block, index uint64) bool {
    var id uint64
    row := chain.DB.QueryRow("SELECT Id FROM BlockChain WHERE Hash=$1", Base64Encode(block.PrevHash))
    row.Scan(&id)
    return id == index
}

func (chain *BlockChain) transactionsIsValid(block *Block) bool {
    lentxs := len(block.Transactions)
    plusStorage := 0
    for i := 0; i < lentxs; i++ {
        if block.Transactions[i].Sender == STORAGE_CHAIN {
            plusStorage = 1
            break
        }
    }
    if lentxs == 0 || lentxs > TXS_LIMIT + plusStorage {
        // fmt.Println("Q:", 1)
        return false
    }
    for i := 0; i < lentxs-1; i++ {
        for j := i+1; j < lentxs; j++ {
            // rand bytes not be equal
            if bytes.Equal(block.Transactions[i].RandBytes, block.Transactions[j].RandBytes) {
                // fmt.Println("Q:", 2)
                return false
            }
            // storage tx only one
            if block.Transactions[i].Sender == STORAGE_CHAIN && block.Transactions[j].Sender == STORAGE_CHAIN {
                // fmt.Println("Q:", 3)
                return false
            }
        }
    }
    for i := 0; i < lentxs; i++ {
        tx := block.Transactions[i]
        // storage tx has no hash and signature
        if tx.Sender == STORAGE_CHAIN {
            if tx.Receiver != block.Miner || tx.Value != STORAGE_REWARD {
                // fmt.Println("Q:", 4)
                return false
            }
        } else {
            if !tx.hashIsValid() {
                // fmt.Println("Q:", 5)
                return false
            }
            if !tx.signIsValid() {
                // fmt.Println("Q:", 6)
                return false
            }
        }
        if !chain.balanceIsValid(block, tx.Sender) {
            // fmt.Println("Q:", 7)
            return false
        }
        if !chain.balanceIsValid(block, tx.Receiver) {
            // fmt.Println("Q:", 8)
            return false
        }
    }
    return true
}

func (chain *BlockChain) balanceIsValid(block *Block, address string) bool {
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
        return false
    }
    if (balanceInChain + balanceAddBlock - balanceSubBlock) != block.Mapping[address] {
        return false
    }
    return true
}
