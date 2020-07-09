package blockchain

import (
	"fmt"
	"bytes"
	"crypto"
	"math"
	mrand"math/rand"
	"math/big"
	"crypto/x509"
	"crypto/rsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
)

// Create private key by size bits.
func GeneratePrivate(bits uint16) *rsa.PrivateKey {
	priv, err := rsa.GenerateKey(rand.Reader, int(bits))
	if err != nil {
		return nil
	}
	return priv
}

// Generate bytes in range [0:256).
func GenerateRandomBytes(max uint) []byte {
	var slice []byte = make([]byte, max)
	_, err := rand.Read(slice)
	if err != nil {
		return nil
	}
	return slice
}

// SHA256(bytes).
func HashSum(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// Sign data by private key.
func Sign(priv *rsa.PrivateKey, data []byte) []byte {
	signature, err := rsa.SignPSS(rand.Reader, priv, crypto.SHA256, data, nil)
	if err != nil {
		return nil
	}
	return signature
}

// Verify data and signature by public key.
func Verify(pub *rsa.PublicKey, data, sign []byte) error {
	return rsa.VerifyPSS(pub, crypto.SHA256, data, sign, nil)
}

// POW for check hash package by Nonce.
func ProofOfWork(blockHash []byte, difficulty uint8) uint64 {
	var (
		Target  = big.NewInt(1)
		intHash = big.NewInt(1)
		nonce   = uint64(mrand.Intn(math.MaxInt32))
		hash    []byte
	)
	Target.Lsh(Target, 256-uint(difficulty))
	for nonce < math.MaxUint64 {
		hash = HashSum(bytes.Join(
			[][]byte{
				blockHash,
				ToBytes(nonce),
			},
			[]byte{},
		))
		if DEBUG {
			fmt.Printf("\rMining: %s", Base64Encode(hash))
		}
		intHash.SetBytes(hash)
		if intHash.Cmp(Target) == -1 {
			if DEBUG {
				fmt.Println()
			}
			break
		}
		nonce++
	}
	return nonce
}

// base64.StdEncoding.EncodeToString
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// base64.StdEncoding.DecodeString
func Base64Decode(data string) []byte {
	result, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil
	}
	return result
}

// Translate uint64 to slice of bytes.
func ToBytes(num uint64) []byte {
	var data = new(bytes.Buffer)
	err := binary.Write(data, binary.BigEndian, num)
	if err != nil {
		return nil
	}
	return data.Bytes()
}

// Translate public key as *rsa.PublicKey to string.
func StringPublic(pub *rsa.PublicKey) string {
	return Base64Encode(x509.MarshalPKCS1PublicKey(pub))
}

// Translate public key as string to *rsa.PublicKey.
func ParsePublic(pubData string) *rsa.PublicKey {
	pub, err := x509.ParsePKCS1PublicKey(Base64Decode(pubData))
	if err != nil {
		return nil
	}
	return pub
}

// Translate public key as *rsa.PublicKey to string.
func StringPrivate(priv *rsa.PrivateKey) string {
	return Base64Encode(x509.MarshalPKCS1PrivateKey(priv))
}

// Translate public key as string to *rsa.PublicKey.
func ParsePrivate(privData string) *rsa.PrivateKey {
	pub, err := x509.ParsePKCS1PrivateKey(Base64Decode(privData))
	if err != nil {
		return nil
	}
	return pub
}
