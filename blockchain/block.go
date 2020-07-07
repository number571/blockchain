package blockchain

import (
    "sort"
    "bytes"
    "math/big"
    "crypto/rsa"
)

func NewBlock(miner string, prevHash []byte) *Block {
    return &Block{
        Difficulty: DIFFICULTY,
        PrevHash: prevHash,
        Miner: miner,
        Mapping: make(map[string]uint64),
    }
}

func (block *Block) IsGenesis() bool {
    if bytes.Equal(block.CurrHash, []byte(GENESIS_BLOCK)) {
        return true
    }
    return false
}

func (block *Block) Hash() []byte {
    var tempHash []byte
    for _, tx := range block.Transactions {
        tempHash = HashSum(bytes.Join(
            [][]byte{
                tempHash,
                tx.CurrHash,
            },
            []byte{},
        ))
    }
    var list []string
    for hash := range block.Mapping {
        list = append(list, hash)
    }
    sort.Strings(list)
    for _, hash := range list {
        tempHash = HashSum(bytes.Join(
            [][]byte{
                tempHash,
                []byte(hash),
                ToBytes(block.Mapping[hash]),
            },
            []byte{},
        ))
    }
    return HashSum(bytes.Join(
        [][]byte{
            tempHash,
            ToBytes(uint64(block.Difficulty)),
            block.PrevHash,
            []byte(block.Miner),
        },
        []byte{},
    ))
}

func (block *Block) Sign(priv *rsa.PrivateKey) []byte {
    return Sign(priv, block.CurrHash)
}

func (block *Block) Proof() uint64 {
    return ProofOfWork(block.CurrHash, block.Difficulty)
}

func (block *Block) HashIsValid() bool {
    return bytes.Equal(block.Hash(), block.CurrHash)
}

func (block *Block) SignIsValid() bool {
    return Verify(ParsePublic(block.Miner), block.CurrHash, block.Signature) == nil
}

func (block *Block) ProofIsValid() bool {
    intHash := big.NewInt(1)
    Target  := big.NewInt(1)
    hash := HashSum(bytes.Join(
        [][]byte{
            block.CurrHash,
            ToBytes(block.Nonce),
        },
        []byte{},
    ))
    intHash.SetBytes(hash)
    Target.Lsh(Target, 256-uint(block.Difficulty))
    if intHash.Cmp(Target) == -1 {
        return true
    }
    return false
}

func (block *Block) MappingIsValid() bool {
    for hash := range block.Mapping {
        if hash == STORAGE_CHAIN {
            continue
        }
        flag := false
        for _, tx := range block.Transactions {
            if tx.Sender == hash || tx.Receiver == hash {
                flag = true
                break
            }
        }
        if !flag {
            return false
        }
    }
    return true
}
