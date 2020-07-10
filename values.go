package main

import (
	bc "./blockchain"
)

var (
	Filename string
	Addresses []string
	Serve string
	Chain *bc.BlockChain
	Block *bc.Block
	User  *bc.User
)

const (
	SEPARATOR = "_SEPARATOR_"
)

const (
	GET_SIZE         = "[GET_SIZE]"
	ADD_BLOCK        = "[ADD_BLOCK]"
	GET_CHAIN        = "[GET_CHAIN]"
	GET_LASTHASH     = "[GET_LASTHASH]"
	GET_BALANCE      = "[GET_BALANCE]"
	ADD_TRANSACTION  = "[ADD_TRANSACTION]"
)
