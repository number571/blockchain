package main

import (
	"os"
	"fmt"
	"bufio"
	"strconv"
	"strings"
	"io/ioutil"
	"encoding/json"
	nt "./network"
	bc "./blockchain"
)

var (
	Addresses []string
	User  *bc.User
)

const (
	OPT_BLOCKCHAIN  = "[GET-BLOCKCHAIN]"
	OPT_LASTHASH    = "[GET-LASTHASH]"
	OPT_BALANCE     = "[GET-BALANCE]"
	OPT_TRANSACTION = "[ADD-TRANSACTION]"
)

func init() {
	if len(os.Args) < 2 {
		fmt.Println("failed: len(os.Args) < 2\n")
		os.Exit(1)
	}
	var (
		addrExist  = false
		userExist  = false
	)
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
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
		case strings.HasPrefix(arg, "-loadaddr:"):
			arg = strings.Replace(arg, "-loadaddr:", "", 1)
			err := json.Unmarshal([]byte(readFile(arg)), &Addresses)
			if err != nil {
				fmt.Println("failed: load addresses\n")
				os.Exit(1)
			}
			if len(Addresses) == 0 {
				fmt.Println("failed: len(Addresses) == 0\n")
				os.Exit(1)
			}
			addrExist = true
		}
	}
	if !userExist || !addrExist {
		fmt.Println("failed: !userExist || !addrExist\n")
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
			Option: OPT_BLOCKCHAIN,
			Data: fmt.Sprintf("%d", i),
		})
		if res.Data == "" {
			break
		}
		fmt.Printf("[%d] => %s\n", i, res.Data)	
	}
	fmt.Println()
}

func chainTX(splited []string) {
	if len(splited) != 3 {
		fmt.Println("failed: len(splited) != 3\n")
		return
	}
	res := nt.Send(Addresses[0], &nt.Package{
		Option: OPT_LASTHASH,
	})
	num, err := strconv.Atoi(splited[2])
	if err != nil {
		fmt.Println("failed: strconv.Atoi(num)\n")
		return
	}
	tx := bc.NewTransaction(User, bc.Base64Decode(res.Data), splited[1], uint64(num))
	sertx := bc.SerializeTX(tx)
	for _, addr := range Addresses {
		nt.Send(addr, &nt.Package{
			Option: OPT_TRANSACTION,
			Data: sertx,
		})
	}
	fmt.Println("success: transaction sent\n")
}

func chainBalance(splited []string) {
	if len(splited) != 2 {
		fmt.Println("failed: len(splited) != 2\n")
		return
	}
	res := nt.Send(Addresses[0], &nt.Package{
		Option: OPT_BALANCE,
		Data: splited[1],
	})
	fmt.Println("Balance:", res.Data, "coins\n")
}

func userBalance() {
	if User == nil {
		fmt.Println("failed: user == nil\n")
		return
	}
	res := nt.Send(Addresses[0], &nt.Package{
		Option: OPT_BALANCE,
		Data: User.Address(),
	})
	fmt.Println("Balance:", res.Data, "coins\n")
}

func userAddress() {
	if User == nil {
		fmt.Println("failed: user == nil\n")
		return
	}
	fmt.Println("Address:", User.Address(), "\n")
}

func userPurse() {
	if User == nil {
		fmt.Println("failed: user == nil\n")
		return
	}
	fmt.Println("Purse:", User.Purse(), "\n")
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

func inputString(begin string) string {
	fmt.Print(begin)
	msg, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.Replace(msg, "\n", "", 1)
}
