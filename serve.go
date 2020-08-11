package main

import (
	bc "./blockchain"
	nt "./network"
	"bytes"
	"database/sql"
	"encoding/hex"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"sort"
)

var (
	Filename    string
	Serve       string
	Chain       *bc.BlockChain
	Block       *bc.Block
	Mutex       sync.Mutex
)

var (
	IsMining    bool
	BreakMining = make(chan bool)
)

func handleServer(conn nt.Conn, pack *nt.Package) {
	nt.Handle(ADD_BLOCK, conn, pack, addBlock)
	nt.Handle(ADD_TRNSX, conn, pack, addTransaction)
	nt.Handle(GET_BLOCK, conn, pack, getBlock)
	nt.Handle(GET_LHASH, conn, pack, getLastHash)
	nt.Handle(GET_BLNCE, conn, pack, getBalance)
	nt.Handle(GET_CSIZE, conn, pack, getChainSize)
}

func getChainSize(pack *nt.Package) string {
	return fmt.Sprintf("%d", Chain.Size())
}

func addBlock(pack *nt.Package) string {
	splited := strings.Split(pack.Data, SEPARATOR)
	if len(splited) != 3 {
		return "fail"
	}

	block := bc.DeserializeBlock(splited[2])
	if !block.IsValid(Chain, Chain.Size()) {
		currSize := Chain.Size()
		num, err := strconv.Atoi(splited[1])
		if err != nil {
			return "fail"
		}
		if currSize < uint64(num) {
			go compareChains(splited[0], uint64(num))
			return "ok"
		}
		return "fail"
	}

	Mutex.Lock()

	Chain.AddBlock(block)
	Block = bc.NewBlock(User.Address(), Chain.LastHash())

	Mutex.Unlock()

	if IsMining {
		BreakMining <- true
		IsMining = false
	}

	return "ok"
}

func compareChains(address string, num uint64) {
	filename := "temp_" + hex.EncodeToString(bc.GenerateRandomBytes(8))
	file, err := os.Create(filename)
	if err != nil {
		return
	}
	file.Close()
	defer func() {
		os.Remove(filename)
	}()

	res := nt.Send(address, &nt.Package{
		Option: GET_BLOCK,
		Data:   fmt.Sprintf("%d", 0),
	})
	if res == nil {
		return
	}

	genesis := bc.DeserializeBlock(res.Data)
	if genesis == nil {
		return
	}

	if !bytes.Equal(genesis.CurrHash, hashBlock(genesis)) {
		return
	}

	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return
	}
	defer db.Close()

	_, err = db.Exec(bc.CREATE_TABLE)
	chain := &bc.BlockChain{
		DB: db,
	}
	chain.AddBlock(genesis)

	defer func() {
		chain.DB.Close()
	}()

	for i := uint64(1); i < num; i++ {
		res := nt.Send(address, &nt.Package{
			Option: GET_BLOCK,
			Data:   fmt.Sprintf("%d", i),
		})
		if res == nil {
			return
		}
		block := bc.DeserializeBlock(res.Data)
		if block == nil {
			return
		}
		if !block.IsValid(chain, i) {
			return
		}
		chain.AddBlock(block)
	}

	Mutex.Lock()

	Chain.DB.Close()
	os.Remove(Filename)

	copyFile(filename, Filename)
	Chain = bc.LoadChain(Filename)
	Block = bc.NewBlock(User.Address(), Chain.LastHash())

	Mutex.Unlock()

	if IsMining {
		BreakMining <- true
		IsMining = false
	}

	return
}

func hashBlock(block *bc.Block) []byte {
	var tempHash []byte
	for _, tx := range block.Transactions {
		tempHash = bc.HashSum(bytes.Join(
			[][]byte{
				tempHash,
				tx.CurrHash,
			},
			[]byte{},
		))
	}
	var list []string
	for hash := range block.Mapping {
		list = append(list, hash)
	}
	sort.Strings(list)
	for _, hash := range list {
		tempHash = bc.HashSum(bytes.Join(
			[][]byte{
				tempHash,
				[]byte(hash),
				bc.ToBytes(block.Mapping[hash]),
			},
			[]byte{},
		))
	}
	return bc.HashSum(bytes.Join(
		[][]byte{
			tempHash,
			bc.ToBytes(uint64(block.Difficulty)),
			block.PrevHash,
			[]byte(block.Miner),
			[]byte(block.TimeStamp),
		},
		[]byte{},
	))
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func getBlock(pack *nt.Package) string {
	num, err := strconv.Atoi(pack.Data)
	if err != nil {
		return ""
	}
	size := Chain.Size()
	if uint64(num) < size {
		return selectBlock(Chain, num)
	}
	return ""
}

func getLastHash(pack *nt.Package) string {
	return bc.Base64Encode(Chain.LastHash())
}

func getBalance(pack *nt.Package) string {
	return fmt.Sprintf("%d", Chain.Balance(pack.Data, Chain.Size()))
}

func selectBlock(chain *bc.BlockChain, i int) string {
	var block string
	row := chain.DB.QueryRow("SELECT Block FROM BlockChain WHERE Id=$1", i+1)
	row.Scan(&block)
	return block
}

func addTransaction(pack *nt.Package) string {
	var tx = bc.DeserializeTX(pack.Data)
	if tx == nil || len(Block.Transactions) == bc.TXS_LIMIT {
		return "fail"
	}
	Mutex.Lock()
	err := Block.AddTransaction(Chain, tx)
	Mutex.Unlock()
	if err != nil {
		return "fail"
	}
	if len(Block.Transactions) == bc.TXS_LIMIT {
		go func() {
			Mutex.Lock()
			block := *Block
			IsMining = true
			Mutex.Unlock()
			res := (&block).Accept(Chain, User, BreakMining)
			Mutex.Lock()
			IsMining = false
			if res == nil && bytes.Equal(block.PrevHash, Block.PrevHash) {
				Chain.AddBlock(&block)
				pushBlockToNet(&block)
			}
			Block = bc.NewBlock(User.Address(), Chain.LastHash())
			Mutex.Unlock()
		}()
	}
	return "ok"
}

func pushBlockToNet(block *bc.Block) {
	var (
		sblock = bc.SerializeBlock(block)
		msg = Serve + SEPARATOR + fmt.Sprintf("%d", Chain.Size()) + SEPARATOR + sblock
	)
	for _, addr := range Addresses {
		go nt.Send(addr, &nt.Package{
			Option: ADD_BLOCK,
			Data: msg,
		})
	}
}
