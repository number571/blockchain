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
	ADD_BLOCK = iota + 1
	ADD_TRNSX 
	GET_BLOCK       
	GET_LHASH    
	GET_BLNCE   
	GET_CSIZE
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
