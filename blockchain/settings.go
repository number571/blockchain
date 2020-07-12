package blockchain

import (
	"database/sql"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	CREATE_TABLE = `
CREATE TABLE BlockChain (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    Hash VARCHAR(44) UNIQUE,
    Block TEXT
);
`
)

const (
	// DEBUG          = false
	DEBUG          = true
	KEY_SIZE       = 512
	STORAGE_CHAIN  = "STORAGE-CHAIN"
	STORAGE_VALUE  = 100
	STORAGE_REWARD = 1
	GENESIS_BLOCK  = "GENESIS-BLOCK"
	GENESIS_REWARD = 100
	DIFFICULTY     = 20              // 15
	TXS_LIMIT      = 2               // 6
	START_PERCENT  = 10
	RAND_BYTES     = 32
)

type BlockChain struct {
	index uint64
	DB    *sql.DB
}

type Block struct {
	Nonce        uint64
	Difficulty   uint8
	CurrHash     []byte
	PrevHash     []byte
	Transactions []Transaction
	Mapping      map[string]uint64
	Miner        string // Hashname
	Signature    []byte // Miner sign
	TimeStamp    string
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
