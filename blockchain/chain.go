package blockchain

import (
    "fmt"
    "bytes"
    "errors"
    "encoding/json"
)

func NewChain(receiver string) *BlockChain {
    genesis := &Block{
        CurrHash: []byte(GENESIS_BLOCK),
        Mapping: make(map[string]uint64),
    }
    genesis.Mapping[STORAGE_CHAIN] = STORAGE_VALUE
    genesis.Mapping[receiver] = GENESIS_REWARD
    return &BlockChain{*genesis}
}

func (chain *BlockChain) CheckBlocks() (int, error) {
    if len(*chain) == 0 || !(*chain)[0].IsGenesis() {
        return 0, errors.New("genesis block undefined")
    }
    for i := 1; i < len(*chain); i++ {
        block := (*chain)[i]
        if block.Difficulty != DIFFICULTY {
            return i, fmt.Errorf("block diff [%d] not supported", i)
        }
        if !block.HashIsValid() {
            return i, fmt.Errorf("hash is not valid [%d]", i)
        }
        if !block.SignIsValid() {
            return i, fmt.Errorf("sign is not valid [%d]", i)
        }
        if !block.ProofIsValid() {
            return i, fmt.Errorf("proof is not valid [%d]", i)
        }
        if !block.MappingIsValid() {
            return i, fmt.Errorf("mapping is not valid [%d]", i)
        }
        lastchain := (*chain)[:i]
        if !lastchain.TransactionsIsValid(&block) {
            return i, fmt.Errorf("transactions is not valid [%d]", i)
        }
    }
    return 0, nil
}

func (block *Block) MappingIsValid() bool {
    for hash := range block.Mapping {
        if hash == STORAGE_CHAIN {
            continue
        }
        flag := false
        for _, tx := range block.Transactions {
            if tx.Sender == hash || tx.Receiver == hash {
                flag = true
                break
            }
        }
        if !flag {
            return false
        }
    }
    return true
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
        return false
    }
    for i := 0; i < lentxs-1; i++ {
        for j := i+1; j < lentxs; j++ {
            // rand bytes not be equal
            if bytes.Equal(block.Transactions[i].RandBytes, block.Transactions[j].RandBytes) {
                return false
            }
            // storage tx only one
            if block.Transactions[i].Sender == STORAGE_CHAIN && block.Transactions[j].Sender == STORAGE_CHAIN {
                return false
            }
        }
    }
    for i := 0; i < lentxs; i++ {
        tx := block.Transactions[i]
        // storage tx has no hash and signature
        if tx.Sender == STORAGE_CHAIN {
            if (tx.Receiver != block.Miner || tx.Value != STORAGE_REWARD) {
                return false
            }
        } else {
            if !tx.HashIsValid() {
                return false
            }
            if !tx.SignIsValid() {
                return false
            }
        }
        if (!chain.balanceIsValid(block, &tx, tx.Sender)) {
            return false
        }
        if (!chain.balanceIsValid(block, &tx, tx.Receiver)) {
            return false
        }
    }
    return true
}

func (chain *BlockChain) balanceIsValid(block *Block, tx *Transaction, address string) bool {
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

func (chain *BlockChain) Balance(address string) uint64 {
    var balance uint64
    for i := len(*chain)-1; i >= 0; i-- {
        if value, ok := (*chain)[i].Mapping[address]; ok {
            balance = value
            break
        }
    }
    return balance
}

func (chain *BlockChain) LastHash() []byte {
    return (*chain)[len(*chain)-1].CurrHash
}

func (chain *BlockChain) PushBlock(user *User, block *Block) {
    if !chain.TransactionsIsValid(block) {
        return
    }

    block.AddTransaction(chain, &Transaction{
        RandBytes: GenerateRandomBytes(32),
        Sender: STORAGE_CHAIN,
        Receiver: user.Address(),
        Value: STORAGE_REWARD,
    })

    block.CurrHash  = block.Hash()
    block.Signature = block.Sign(user.Private())
    block.Nonce     = block.Proof()
    *chain          = append(*chain, *block)
}

func (chain *BlockChain) Print() {
    lenchain := len(*chain)
    for i := 0; i < lenchain; i++ {
        fmt.Printf("[%d] => ", i)
        printJSON((*chain)[i])
    }
}

func printJSON(data interface{}) {
    jsonData, _ := json.MarshalIndent(data, "", "\t")
    fmt.Println(string(jsonData))
}
