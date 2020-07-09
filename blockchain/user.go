package blockchain

import (
	"crypto/rsa"
)

type User struct {
	PrivateKey *rsa.PrivateKey
}

func NewUser() *User {
	return &User{
		PrivateKey: GeneratePrivate(KEY_SIZE),
	}
}

func LoadUser(purse string) *User {
	priv := ParsePrivate(purse)
	if priv == nil {
		return nil
	}
	return &User{
		PrivateKey: priv,
	}
}

func (user *User) Address() string {
	return StringPublic(user.Public())
}

func (user *User) Purse() string {
	return StringPrivate(user.Private())
}

func (user *User) Private() *rsa.PrivateKey {
	return user.PrivateKey
}

func (user *User) Public() *rsa.PublicKey {
	return &(user.PrivateKey).PublicKey
}
