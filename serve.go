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
	"time"
)

var (
	Filename    string
	Serve       string
	Chain       *bc.BlockChain
	Block       *bc.Block
	IsMining    bool
	BreakMining = make(chan bool)
)

func handleServer(conn nt.Conn, pack *nt.Package) {
	nt.Handle(GET_SIZE, conn, pack, getSize)
	nt.Handle(ADD_BLOCK, conn, pack, addBlock)
	nt.Handle(GET_CHAIN, conn, pack, getChain)
	nt.Handle(GET_LASTHASH, conn, pack, getLastHash)
	nt.Handle(GET_BALANCE, conn, pack, getBalance)
	nt.Handle(ADD_TRANSACTION, conn, pack, addTransaction)
}

func getSize(pack *nt.Package) string {
	return fmt.Sprintf("%d", Chain.Size())
}

func addBlock(pack *nt.Package) string {
	splited := strings.Split(pack.Data, SEPARATOR)
	if len(splited) != 3 {
		return "fail"
	}
	block := bc.DeserializeBlock(splited[2])
	if !Chain.BlockIsValid(block) {
		currSize := Chain.Size()
		num, err := strconv.Atoi(splited[1])
		if err != nil {
			return "fail"
		}
		if currSize < uint64(num) {
			return compareChains(splited[0], uint64(num))
		}
		return "fail"
	}

	Chain.AddBlock(block)
	Block = bc.NewBlock(User.Address(), Chain.LastHash())

	if IsMining {
		BreakMining <- true
		IsMining = false
	}

	return "ok"
}

func compareChains(address string, num uint64) string {
	filename := "temp_" + hex.EncodeToString(bc.GenerateRandomBytes(8))
	file, err := os.Create(filename)
	if err != nil {
		return "fail"
	}
	file.Close()
	defer func() {
		os.Remove(filename)
	}()

	res := nt.Send(address, &nt.Package{
		Option: GET_CHAIN,
		Data:   fmt.Sprintf("%d", 0),
	})
	if res == nil {
		return "fail"
	}

	genesis := bc.DeserializeBlock(res.Data)
	if genesis == nil {
		return "fail"
	}

	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return "fail"
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
			Option: GET_CHAIN,
			Data:   fmt.Sprintf("%d", i),
		})
		if res == nil {
			return "fail"
		}
		block := bc.DeserializeBlock(res.Data)
		if block == nil {
			return "fail"
		}
		if !chain.BlockIsValid(block) {
			return "fail"
		}
		chain.AddBlock(block)
	}

	Chain.DB.Close()
	os.Remove(Filename)

	copyFile(filename, Filename)
	Chain = bc.LoadChain(Filename)
	Block = bc.NewBlock(User.Address(), Chain.LastHash())

	if IsMining {
		BreakMining <- true
		IsMining = false
	}

	return "ok"
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

func getChain(pack *nt.Package) string {
	num, err := strconv.Atoi(pack.Data)
	if err != nil {
		return ""
	}
	size := Chain.Size()
	if uint64(num) < size {
		return getBlock(Chain, num)
	}
	return ""
}

func getBlock(chain *bc.BlockChain, i int) string {
	var block string
	row := chain.DB.QueryRow("SELECT Block FROM BlockChain WHERE Id=$1", i+1)
	row.Scan(&block)
	return block
}

func getLastHash(pack *nt.Package) string {
	return bc.Base64Encode(Chain.LastHash())
}

func getBalance(pack *nt.Package) string {
	return fmt.Sprintf("%d", Chain.Balance(pack.Data))
}

func addTransaction(pack *nt.Package) string {
	var tx = bc.DeserializeTX(pack.Data)
	if tx == nil || len(Block.Transactions) == bc.TXS_LIMIT {
		return "fail"
	}
	if Block.AddTransaction(Chain, tx) != nil {
		return "fail"
	}
	if len(Block.Transactions) == bc.TXS_LIMIT {
		lastBlock := getLastBlock()
		if lastBlock == nil {
			return "fail"
		}
		btime, err := time.Parse(time.RFC3339, lastBlock.TimeStamp)
		if err != nil {
			return "fail"
		}
		mod := time.Now().Sub(btime)
		diff := mod < bc.TIME_SESSION
		go func() {
			block := *Block
			if diff {
				time.Sleep(bc.TIME_SESSION - mod)
			}
			IsMining = true
			res := Chain.AcceptBlock(User, &block, BreakMining)
			IsMining = false
			if res != nil && bytes.Equal(block.PrevHash, Block.PrevHash) {
				Chain.AddBlock(&block)
				pushBlockToNet(&block)
			}
			Block = bc.NewBlock(User.Address(), Chain.LastHash())
		}()
	}
	return "ok"
}

func getLastBlock() *bc.Block {
	var sblock string
	row := Chain.DB.QueryRow("SELECT Block FROM BlockChain ORDER BY Id DESC")
	row.Scan(&sblock)
	return bc.DeserializeBlock(sblock)
}

func pushBlockToNet(block *bc.Block) {
	var sblock = bc.SerializeBlock(block)
	for _, addr := range Addresses {
		nt.Send(addr, &nt.Package{
			Option: ADD_BLOCK,
			Data:   Serve + SEPARATOR + fmt.Sprintf("%d", Chain.Size()) + SEPARATOR + sblock,
		})
	}
}
