package main

import (
	"os"
	"io/ioutil"
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/common"
	contract "./contracts"
)

type UserType struct {
	Purse string
	AddressHex string
	AddressEth common.Address
	PublicKey *ecdsa.PublicKey
	PrivateKey *ecdsa.PrivateKey
}

var (
	ClientETH = connectToETH("http://127.0.0.1:5555") 
	User *UserType
)

func init() {
	if len(os.Args) < 2 {
		panic("failed: len(os.Args) < 2")
	}
	var (
		userLoadStr = ""
		userLoadExist = false
	)
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case strings.HasPrefix(arg, "-loaduser:"):
			userLoadStr = strings.Replace(arg, "-loaduser:", "", 1)
			userLoadExist = true
		}
	}
	if !userLoadExist {
		panic("failed: !userLoadExist")
	}
	if ClientETH == nil {
		panic("failed: connect to ETH")
	}
	User = userLoad(userLoadStr)
	if User == nil {
		panic("failed: load user")
	}
}

// Deploy contract and save address in file.
func main() {
	auth := resetAuth(User)
	address, tx, instance, err := contract.DeployContract(auth, ClientETH)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(address.Hex())
	fmt.Println(tx.Hash().Hex())
	_ = instance

	contractFile := "contract.address"
	writeFile(contractFile, address.Hex())
}

func writeFile(filename string, data string) error {
	return ioutil.WriteFile(filename, []byte(data), 0644)
}

func userLoad(purse string) *UserType {
	priv, err := crypto.HexToECDSA(purse)
	if err != nil {
		return nil
	}
	pub, ok := priv.Public().(*ecdsa.PublicKey)
	if !ok {
		return nil
	}
	addressHex := crypto.PubkeyToAddress(*pub).Hex()
	addressEth  := common.HexToAddress(addressHex)
	return &UserType{
		Purse:      purse,
		AddressHex: addressHex,
		AddressEth: addressEth,
		PublicKey:  pub,
		PrivateKey: priv,
	}
}

func connectToETH(address string) *ethclient.Client {
	client, err := ethclient.Dial(address)
	if err != nil {
		return nil
	}
	return client
}

func resetAuth(user *UserType) *bind.TransactOpts {
	nonce, err := ClientETH.PendingNonceAt(context.Background(), User.AddressEth)
	if err != nil {
		return nil
	}

	gasPrice, err := ClientETH.SuggestGasPrice(context.Background())
	if err != nil {
		return nil
	}

	auth := bind.NewKeyedTransactor(user.PrivateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)

	// auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	return auth
}

func fileIsExist(name string) bool {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return false
	}
	return true
}
