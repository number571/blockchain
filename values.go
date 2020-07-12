package main

import (
	bc "./blockchain"
	"io/ioutil"
)

var (
	Addresses []string
	User      *bc.User
)

const (
	SEPARATOR = "_SEPARATOR_"
)

const (
	GET_SIZE        = "[GET_SIZE]"
	ADD_BLOCK       = "[ADD_BLOCK]"
	GET_CHAIN       = "[GET_CHAIN]"
	GET_LASTHASH    = "[GET_LASTHASH]"
	GET_BALANCE     = "[GET_BALANCE]"
	ADD_TRANSACTION = "[ADD_TRANSACTION]"
)

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
