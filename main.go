package main

import (
	"fmt"
	bc "./blockchain" 
)

var (
	_ = fmt.Sprintf("")
)

func main() {
	user1 := bc.NewUser()
	chain := bc.NewChain(user1.Address())

	user2 := bc.NewUser()

	for i := 0; i < 3; i++ {
		block := bc.NewBlock(user2.Address(), chain.LastHash())
		block.AddTransaction(chain, block.NewTransaction(user1, "aaa", 15))
		block.AddTransaction(chain, block.NewTransaction(user1, "bbb", 10))
		chain.PushBlock(user2, block)
	}
	
	chain.Print()

	// fmt.Println(user2.Address(), chain.Balance(user2.Address()))
	fmt.Println(chain.CheckBlocks())
}
