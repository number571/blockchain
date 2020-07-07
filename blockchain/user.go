package blockchain

import (
	"crypto/rsa"
)

type User rsa.PrivateKey

func NewUser() *User {
	user := User(*GeneratePrivate(KEY_SIZE))
	return &user
}

func (user *User) Address() string {
	return StringPublic(user.Public())
}

func (user *User) Private() *rsa.PrivateKey {
	priv := rsa.PrivateKey(*user)
	return &priv
}

func (user *User) Public() *rsa.PublicKey {
	return &user.PublicKey
}
