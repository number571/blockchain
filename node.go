package main

import (
	"os"
	"fmt"
	"strings"
	"io/ioutil"
	nt "./network"
	bc "./blockchain"
)

func init() {
	if len(os.Args) < 2 {
		fmt.Println("failed: len(os.Args) < 2\n")
		os.Exit(1)
	}
	var (
		addrExist  = false
		userExist  = false
		chainExist = false
	)
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case strings.HasPrefix(arg, "-address:"):
			Address = strings.Replace(arg, "-address:", "", 1)
			addrExist = true
		case strings.HasPrefix(arg, "-newuser:"):
			arg = strings.Replace(arg, "-newuser:", "", 1)
			User = userNew(arg)
			if User == nil {
				fmt.Println("failed: generate user\n")
				os.Exit(1)
			}
			userExist = true
		case strings.HasPrefix(arg, "-loaduser:"):
			arg = strings.Replace(arg, "-loaduser:", "", 1)
			User = userLoad(arg)
			if User == nil {
				fmt.Println("failed: load user\n")
				os.Exit(1)
			}
			userExist = true
		case strings.HasPrefix(arg, "-newchain:"):
			arg = strings.Replace(arg, "-newchain:", "", 1)
			Chain = chainNew(arg)
			if Chain == nil {
				fmt.Println("failed: genesis block\n")
				os.Exit(1)
			}
			chainExist = true
		case strings.HasPrefix(arg, "-loadchain:"):
			arg = strings.Replace(arg, "-loadchain:", "", 1)
			Chain = chainLoad(arg)
			if Chain == nil {
				fmt.Println("failed: load chain\n")
				os.Exit(1)
			}
			chainExist = true
		}
	}
	if !userExist || !chainExist || !addrExist {
		fmt.Println("failed: !userExist || !chainExist || !addrExist\n")
		os.Exit(1)
	}
	Block = bc.NewBlock(User.Address(), Chain.LastHash())
}

func main() {
	nt.Listen(Address, handleServer)
	for {
		fmt.Scanln()
	}
}

func chainNew(filename string) *bc.BlockChain {
	if User == nil {
		return nil
	}
	err := bc.NewChain(filename, User.Address())
	if err != nil {
		return nil
	}
	return bc.LoadChain(filename)
}

func chainLoad(filename string) *bc.BlockChain {
	chain := bc.LoadChain(filename)
	if chain == nil {
		return nil 
	}
	return chain
}

func userNew(filename string) *bc.User {
	user := bc.NewUser()
	if user == nil {
		return nil
	}
	err := writeFile(filename, user.Purse())
	if err != nil {
		return nil
	}
	return user
}

func userLoad(filename string) *bc.User {
	priv := readFile(filename)
	if priv == "" {
		return nil 
	}
	user := bc.LoadUser(priv)
	if user == nil {
		return nil 
	}
	return user
}

func writeFile(filename string, data string) error {
	return ioutil.WriteFile(filename, []byte(data), 0644)
}

func readFile(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return ""
	}
	return string(data)
}
