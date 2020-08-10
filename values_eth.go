package main

import (
	"context"
	"math/big"
	"io/ioutil"
	"crypto/ecdsa"
	contract "./contracts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type UserType struct {
	Purse string
	AddressHex string
	AddressEth common.Address
	PublicKey *ecdsa.PublicKey
	PrivateKey *ecdsa.PrivateKey
}

type Estate struct{
    Id *big.Int
    Owner common.Address
    Info string
    Squere *big.Int
    UsefulSquere *big.Int
    RenterAddress common.Address
    PresentStatus bool
    SaleStatus bool
    RentStatus bool
}

type Present struct {
	Id *big.Int
	EstateId *big.Int
	AddressFrom common.Address
	AddressTo common.Address
	Finished bool
}

type Sale struct {
	Id *big.Int
	EstateId *big.Int
	Owner common.Address
	Price *big.Int
	Customers []common.Address
	Prices []*big.Int
	Finished bool
}

type Rent struct {
	Id *big.Int
	EstateId *big.Int
	Owner common.Address
	Renter common.Address
	Time *big.Int
	Money *big.Int
	DeadLine *big.Int
	Finished bool
}

var (
	User *UserType
	ClientETH     = connectToETH("http://127.0.0.1:5555") 
	Instance      = newContract(
		common.HexToAddress(readFile("contract.address")), 
		ClientETH,
	)
)

func loadUser(purse string) *UserType {
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

func newContract(contractAddr common.Address, clientEth *ethclient.Client) *contract.Contract {
	instance, err := contract.NewContract(contractAddr, clientEth)
	if err != nil {
		return nil
	}
	return instance
}

func connectToETH(address string) *ethclient.Client {
	client, err := ethclient.Dial(address)
	if err != nil {
		return nil
	}
	return client
}

func readFile(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return ""
	}
	return string(data)
}

func resetAuth(user *UserType) *bind.TransactOpts {
	nonce, err := ClientETH.PendingNonceAt(context.Background(), crypto.PubkeyToAddress(*user.PublicKey))
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

	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	return auth
}

func getEstates(index *big.Int) *Estate {
	// (*big.Int, common.Address, string, *big.Int, *big.Int, common.Address, error)
	id, owner, info, squere, usefulsquere, renteraddress, err := Instance.GetEstates(&bind.CallOpts{From: User.AddressEth}, index)
	if err != nil {
		return nil
	}
	presentS, saleS, rentS, err := Instance.GetEstatesStatuses(&bind.CallOpts{From: User.AddressEth}, index)
	if err != nil {
		return nil
	}
	return &Estate{
		Id: id,
		Owner: owner,
		Info: info,
		Squere: squere,
		UsefulSquere: usefulsquere,
		RenterAddress: renteraddress,
		PresentStatus: presentS,
		SaleStatus: saleS,
		RentStatus: rentS,
	}
}

func getPresents(index *big.Int) *Present {
	// (*big.Int, common.Address, common.Address, bool, error)
	id, from, to, finished, err := Instance.GetPresents(&bind.CallOpts{From: User.AddressEth}, index)
	if err != nil {
		return nil
	}
	return &Present{
		Id: index,
		EstateId: id,
		AddressFrom: from,
		AddressTo: to,
		Finished: finished,
	}
}

func getSales(index *big.Int) *Sale {
	// (*big.Int, common.Address, *big.Int, []common.Address, []*big.Int, bool, error)
	id, owner, price, customers, prices, finished, err := Instance.GetSales(&bind.CallOpts{From: User.AddressEth}, index)
	if err != nil {
		return nil
	}
	return &Sale{
		Id: index,
		EstateId: id,
		Owner: owner,
		Price: price,
		Customers: customers,
		Prices: prices,
		Finished: finished,
	}
}

func getRents(index *big.Int) *Rent {
	// (*big.Int, common.Address, common.Address, *big.Int, *big.Int, *big.Int, bool, error)
	id, owner, renter, time, money, deadline, finished, err := Instance.GetRents(&bind.CallOpts{From: User.AddressEth}, index)
	if err != nil {
		return nil 
	}
	return &Rent{
		Id: index,
		EstateId: id,
		Owner: owner,
		Renter: renter,
		Time: time,
		Money: money,
		DeadLine: deadline,
		Finished: finished,
	}
}

type EstateStr struct {
	Id *big.Int
    Owner string
    Info string
    Squere *big.Int
    UsefulSquere *big.Int
    RenterAddress string
    PresentStatus bool
    SaleStatus bool
    RentStatus bool
}

func estatesToString(estate *Estate) *EstateStr {
	return &EstateStr{
		Id: estate.Id,
		Owner: estate.Owner.Hex(),
		Info: estate.Info,
		Squere: estate.Squere,
		UsefulSquere: estate.UsefulSquere,
		RenterAddress: estate.RenterAddress.Hex(),
		PresentStatus: estate.PresentStatus,
		SaleStatus: estate.SaleStatus,
		RentStatus: estate.RentStatus,
	}
}

type PresentStr struct {
	Id *big.Int
	EstateId *big.Int
	AddressFrom string
	AddressTo string
	Finished bool
}

func presentsToString(present *Present) *PresentStr {
	return &PresentStr{
		Id: present.Id,
		EstateId: present.EstateId,
		AddressFrom: present.AddressFrom.Hex(),
		AddressTo: present.AddressTo.Hex(),
		Finished: present.Finished,
	}
}

type SaleStr struct {
	Id *big.Int
	EstateId *big.Int
	Owner string
	Price *big.Int
	Customers []string
	Prices []*big.Int
	Finished bool
}

func salesToString(sale *Sale) *SaleStr {
	var customers []string
	for _, cust := range sale.Customers {
		customers = append(customers, cust.Hex())
	}
	return &SaleStr{
		Id: sale.Id,
		EstateId: sale.EstateId,
		Owner: sale.Owner.Hex(),
		Price: sale.Price,
		Customers: customers,
		Prices: sale.Prices,
		Finished: sale.Finished,
	}
}

type RentStr struct {
	Id *big.Int
	EstateId *big.Int
	Owner string
	Renter string
	Time *big.Int
	Money *big.Int
	DeadLine *big.Int
	Finished bool
}

func rentsToString(rent *Rent) *RentStr {
	return &RentStr{
		Id: rent.Id,
		EstateId: rent.EstateId,
		Owner: rent.Owner.Hex(),
		Renter: rent.Renter.Hex(),
		Time: rent.Time,
		Money: rent.Money,
		DeadLine: rent.DeadLine,
		Finished: rent.Finished,
	}
}
