package main

import (
	"os"
	"fmt"
	"bufio"
	"strconv"
	"strings"
	"encoding/json"
	nt "./network"
	bc "./blockchain"
)

func init() {
	if len(os.Args) < 2 {
		fmt.Println("failed: len(os.Args) < 2\n")
		os.Exit(1)
	}
	var (
		addrStr     = ""
		userNewStr  = ""
		userLoadStr = ""
	)
	var (
		addrExist     = false
		userNewExist  = false
		userLoadExist = false
	)
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case strings.HasPrefix(arg, "-loadaddr:"):
			addrStr = strings.Replace(arg, "-loadaddr:", "", 1)
			addrExist = true
		case strings.HasPrefix(arg, "-newuser:"):
			userNewStr = strings.Replace(arg, "-newuser:", "", 1)
			userNewExist = true
		case strings.HasPrefix(arg, "-loaduser:"):
			userLoadStr = strings.Replace(arg, "-loaduser:", "", 1)
			userLoadExist = true
		}
	}

	err := json.Unmarshal([]byte(readFile(addrStr)), &Addresses)
	if err != nil {
		fmt.Println("failed: load addresses\n")
		os.Exit(1)
	}
	if len(Addresses) == 0 {
		fmt.Println("failed: len(Addresses) == 0\n")
		os.Exit(1)
	}

	if !(userNewExist || userLoadExist) || !addrExist {
		fmt.Println("failed: !(userNewExist || userLoadExist) || !addrExist\n")
		os.Exit(1)
	}

	if userNewExist {
		User = userNew(userNewStr)
	}
	if userLoadExist {
		User = userLoad(userLoadStr)
	}
	if User == nil {
		fmt.Println("failed: load user\n")
		os.Exit(1)
	}
}

func main() {
	handleClient()
}

func handleClient() {
	var (
		message string
		splited []string
	)
	for {
		message = inputString("")
		splited = strings.Split(message, " ")
		switch splited[0] {
		case "/exit": os.Exit(0)
		case "/user":
			if len(splited) < 2 {
				fmt.Println("failed: len(user) < 2\n")
				continue
			}
			switch splited[1] {
			case "address": userAddress()
			case "purse": userPurse()
			case "balance": userBalance() // ИЗМЕНИТЬ В КЛИЕНТЕ
			}
		case "/chain": 
			if len(splited) < 2 {
				fmt.Println("failed: len(chain) < 2\n")
				continue
			}
			switch splited[1] {
			case "print": chainPrint()
			case "tx": chainTX(splited[1:])
			case "balance": chainBalance(splited[1:]) // ИЗМЕНИТЬ В КЛИЕНТЕ
			}
		}
	}
}

func chainPrint() {
	for i := 0; ; i++ {
		res := nt.Send(Addresses[0], &nt.Package{
			Option: GET_CHAIN,
			Data: fmt.Sprintf("%d", i),
		})
		if res.Data == "" {
			break
		}
		fmt.Printf("[%d] => %s\n", i+1, res.Data)	
	}
	fmt.Println()
}

func chainTX(splited []string) {
	if len(splited) != 3 {
		fmt.Println("failed: len(splited) != 3\n")
		return
	}
	num, err := strconv.Atoi(splited[2])
	if err != nil {
		fmt.Println("failed: strconv.Atoi(num)\n")
		return
	}
	for _, addr := range Addresses {
		res := nt.Send(addr, &nt.Package{
			Option: GET_LASTHASH,
		})
		if res == nil {
			continue
		}
		tx := bc.NewTransaction(User, bc.Base64Decode(res.Data), splited[1], uint64(num))
		res = nt.Send(addr, &nt.Package{
			Option: ADD_TRANSACTION,
			Data: bc.SerializeTX(tx),
		})
		if res == nil {
			continue
		}
		if res.Data == "ok" {
			fmt.Printf("ok: (%s)\n", addr)
		} else {
			fmt.Printf("fail: (%s)\n", addr)
		}
	}
	fmt.Println()
}

func chainBalance(splited []string) {
	if len(splited) != 2 {
		fmt.Println("fail: len(splited) != 2\n")
		return
	}
	for _, addr := range Addresses {
		res := nt.Send(addr, &nt.Package{
			Option: GET_BALANCE,
			Data: splited[1],
		})
		fmt.Printf("Balance (%s): %s coins\n", addr, res.Data)
	}
	fmt.Println()
}

func userBalance() {
	for _, addr := range Addresses {
		res := nt.Send(addr, &nt.Package{
			Option: GET_BALANCE,
			Data: User.Address(),
		})
		fmt.Printf("Balance (%s): %s coins\n", addr, res.Data)
	}
	fmt.Println()
}

func userAddress() {
	fmt.Println("Address:", User.Address(), "\n")
}

func userPurse() {
	fmt.Println("Purse:", User.Purse(), "\n")
}

func inputString(begin string) string {
	fmt.Print(begin)
	msg, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.Replace(msg, "\n", "", 1)
}
