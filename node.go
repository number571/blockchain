package main

import (
	bc "./blockchain"
	nt "./network"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func init() {
	if len(os.Args) < 2 {
		panic("failed: len(os.Args) < 2")
	}
	var (
		serveStr     = ""
		addrStr      = ""
		userNewStr   = ""
		userLoadStr  = ""
		chainNewStr  = ""
		chainLoadStr = ""
	)
	var (
		serveExist     = false
		addrExist      = false
		userNewExist   = false
		userLoadExist  = false
		chainNewExist  = false
		chainLoadExist = false
	)
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case strings.HasPrefix(arg, "-serve:"):
			serveStr = strings.Replace(arg, "-serve:", "", 1)
			serveExist = true
		case strings.HasPrefix(arg, "-loadaddr:"):
			addrStr = strings.Replace(arg, "-loadaddr:", "", 1)
			addrExist = true
		case strings.HasPrefix(arg, "-newuser:"):
			userNewStr = strings.Replace(arg, "-newuser:", "", 1)
			userNewExist = true
		case strings.HasPrefix(arg, "-loaduser:"):
			userLoadStr = strings.Replace(arg, "-loaduser:", "", 1)
			userLoadExist = true
		case strings.HasPrefix(arg, "-newchain:"):
			chainNewStr = strings.Replace(arg, "-newchain:", "", 1)
			chainNewExist = true
		case strings.HasPrefix(arg, "-loadchain:"):
			chainLoadStr = strings.Replace(arg, "-loadchain:", "", 1)
			chainLoadExist = true
		}
	}

	if 	!(userNewExist || userLoadExist) || !(chainNewExist || chainLoadExist) || 
		!serveExist || !addrExist {
			panic("failed: !(userNewExist || userLoadExist)"+
				"|| !(chainNewExist || chainLoadExist) || !serveExist || !addrExist")
	}

	Serve = serveStr

	var addresses []string
	err := json.Unmarshal([]byte(readFile(addrStr)), &addresses)
	if err != nil {
		panic("failed: load addresses")
	}

	var mapaddr = make(map[string]bool)
	for _, addr := range addresses {
		if addr == Serve {
			continue
		}
		if _, ok := mapaddr[addr]; ok {
			continue
		}
		mapaddr[addr] = true
		Addresses = append(Addresses, addr)
	}

	if userNewExist {
		User = userNew(userNewStr)
	}
	if userLoadExist {
		User = userLoad(userLoadStr)
	}
	if User == nil {
		panic("failed: load user")
	}

	if chainNewExist {
		Filename = chainNewStr
		Chain = chainNew(chainNewStr)
	}
	if chainLoadExist {
		Filename = chainLoadStr
		Chain = chainLoad(chainLoadStr)
	}
	if Chain == nil {
		panic("failed: load chain")
	}

	Block = bc.NewBlock(User.Address(), Chain.LastHash())
}

func main() {
	nt.Listen(Serve, handleServer)
	for {
		fmt.Scanln()
	}
}

func chainNew(filename string) *bc.BlockChain {
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
