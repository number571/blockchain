package blockchain

import (
	"database/sql"
	mrand "math/rand"
	"time"
)

func init() {
	mrand.Seed(time.Now().UnixNano())
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
	DEBUG          = true
	KEY_SIZE       = 512
	STORAGE_CHAIN  = "STORAGE-CHAIN"
	STORAGE_VALUE  = 100
	STORAGE_REWARD = 1
	GENESIS_BLOCK  = "GENESIS-BLOCK"
	GENESIS_REWARD = 100
	DIFFICULTY     = 20
	TXS_LIMIT      = 2
	START_PERCENT  = 10
	RAND_BYTES     = 32
)

type BlockChain struct {
	DB    *sql.DB
}

type Block struct {
	Nonce        uint64
	Difficulty   uint8
	CurrHash     []byte
	PrevHash     []byte
	Transactions []Transaction
	Mapping      map[string]uint64
	Miner        string
	Signature    []byte
	TimeStamp    string
}

type Transaction struct {
	RandBytes []byte
	PrevBlock []byte
	Sender    string
	Receiver  string
	Value     uint64
	ToStorage uint64
	CurrHash  []byte
	Signature []byte
}
