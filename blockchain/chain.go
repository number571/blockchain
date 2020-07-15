package blockchain

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"time"
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
		CurrHash:  []byte(GENESIS_BLOCK),
		Mapping:   make(map[string]uint64),
		Miner:     receiver,
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

		if !block.hashIsValid(chain, index-1) {
			fmt.Printf("[%d][FAILED] hash is not valid\n", index)
		} else {
			fmt.Printf("[%d][SUCCESS] hash is valid\n", index)
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

		if !block.timeIsValid(chain, index-1) {
			fmt.Printf("[%d][FAILED] time is not valid\n", index)
		} else {
			fmt.Printf("[%d][SUCCESS] time is valid\n", index)
		}

		size = chain.index
		chain.index = index - 1
		if !block.transactionsIsValid(chain) {
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
		sblock  string
		block   *Block
		balance uint64
	)
	rows, err := chain.DB.Query("SELECT Block FROM BlockChain WHERE Id <= $1 ORDER BY Id DESC", 
		chain.index)
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

func (chain *BlockChain) AddBlock(block *Block) {
	chain.index += 1
	chain.DB.Exec("INSERT INTO BlockChain (Hash, Block) VALUES ($1, $2)",
		Base64Encode(block.CurrHash),
		SerializeBlock(block),
	)
}
