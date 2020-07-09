package main

import (
	"fmt"
	"strconv"
	bc "./blockchain"
	nt "./network" 
)

var (
	Address string
	Chain *bc.BlockChain
	Block *bc.Block
	User  *bc.User
)

const (
	OPT_BLOCKCHAIN  = "[GET-BLOCKCHAIN]"
	OPT_LASTHASH    = "[GET-LASTHASH]"
	OPT_BALANCE     = "[GET-BALANCE]"
	OPT_TRANSACTION = "[ADD-TRANSACTION]"
)

func handleServer(conn nt.Conn, pack *nt.Package) {
	nt.Handle(OPT_BLOCKCHAIN, conn, pack, getBlockChain)
	nt.Handle(OPT_LASTHASH, conn, pack, getLastHash)
	nt.Handle(OPT_BALANCE, conn, pack, getBalance)
	nt.Handle(OPT_TRANSACTION, conn, pack, addTransaction)
}

func getBlockChain(pack *nt.Package) string {
	if Chain == nil {
		return ""
	}
	num, err := strconv.Atoi(pack.Data)
	if err != nil {
		return ""
	}
	if uint64(num) < Chain.Index {
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
	if Chain == nil {
		return ""
	}
	return bc.Base64Encode(Chain.LastHash())
}

func getBalance(pack *nt.Package) string {
	if Chain == nil {
		return "-1"
	}
	return fmt.Sprintf("%d", Chain.Balance(pack.Data))
}

func addTransaction(pack *nt.Package) string {
	var tx = bc.DeserializeTX(pack.Data)
	if tx == nil || len(Block.Transactions) == bc.TXS_LIMIT {
		return "fail"
	}
	Block.AddTransaction(Chain, tx)
	if len(Block.Transactions) == bc.TXS_LIMIT {
		Chain.PushBlock(User, Block)
		Block = bc.NewBlock(User.Address(), Chain.LastHash())
	}
	return "ok"
}
