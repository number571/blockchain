package blockchain

import (
    "database/sql"
)

const (
    DEBUG          = false
    // DEBUG          = true
    KEY_SIZE       = 512
    STORAGE_CHAIN  = "STORAGE-CHAIN"
    STORAGE_VALUE  = 100
    STORAGE_REWARD = 1
    GENESIS_BLOCK  = "GENESIS-BLOCK"
    GENESIS_REWARD = 100
    DIFFICULTY     = 15
    TXS_LIMIT      = 6
    TRANSFER_MAX   = 10
    RAND_BYTES     = 32
    DB_FILENAME    = "blockchain.db"
)

type BlockChain struct {
    Index uint64
    DB *sql.DB
}
// type BlockChain []Block

type Block struct {
    Nonce        uint64
    Difficulty   uint8
    CurrHash     []byte
    PrevHash     []byte
    Transactions []Transaction
    Mapping      map[string]uint64
    Miner        string // Hashname
    Signature    []byte // Miner sign
}

type Transaction struct {
    RandBytes []byte
    PrevBlock []byte
    Sender    string // Hashname
    Receiver  string // Hashname
    Value     uint64
    ToStorage uint64
    CurrHash  []byte
    Signature []byte 
}
