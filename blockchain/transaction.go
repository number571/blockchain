package blockchain

import (
    "fmt"
    "bytes"
    "errors"
    "crypto/rsa"
)

func (block *Block) NewTransaction(user *User, to string, value uint64) *Transaction {
    tx := &Transaction{
        RandBytes: GenerateRandomBytes(RAND_BYTES),
        PrevBlock: block.PrevHash,
        Sender: user.Address(),
        Receiver: to,
        Value: value,
    }
    if value > TRANSFER_MAX {
        tx.ToStorage = STORAGE_REWARD
    }
    tx.CurrHash  = tx.Hash()
    tx.Signature = tx.Sign(user.Private())
    return tx
}

func (block *Block) AddTransaction(chain *BlockChain, tx *Transaction) error {
    if len(block.Transactions) == TXS_LIMIT && tx.Sender != STORAGE_CHAIN {
        if DEBUG {
            fmt.Println("(AddTransaction) len tx = limit")
        }
        return errors.New("len tx = limit")
    }
    balanceInChain := chain.Balance(tx.Sender)
    balanceInBlock := tx.Value + tx.ToStorage
    if value, ok := block.Mapping[tx.Sender]; ok {
        balanceInChain = value
    }
    if tx.Value > TRANSFER_MAX && tx.ToStorage != STORAGE_REWARD {
        if DEBUG {
            fmt.Println("(AddTransaction) storage reward pass")
        }
        return errors.New("storage reward pass")
    }
    if balanceInBlock > balanceInChain {
        if DEBUG {
            fmt.Println("(AddTransaction) insufficient funds")
        }
        return errors.New("insufficient funds")
    }
    block.Mapping[tx.Sender] = balanceInChain - balanceInBlock
    chain.addBalance(block, tx.Receiver, tx.Value)
    chain.addBalance(block, STORAGE_CHAIN, tx.ToStorage)
    block.Transactions = append(block.Transactions, *tx)
    return nil
}

func (chain *BlockChain) addBalance(block *Block, receiver string, value uint64) {
    balanceInChain := chain.Balance(receiver)
    if v, ok := block.Mapping[receiver]; ok {
        balanceInChain = v
    }
    block.Mapping[receiver] = balanceInChain + value
}

func (tx *Transaction) Hash() []byte {
    return HashSum(bytes.Join(
        [][]byte{
            tx.RandBytes,
            tx.PrevBlock,
            []byte(tx.Sender),
            []byte(tx.Receiver),
            ToBytes(tx.Value),
            ToBytes(tx.ToStorage),
        },
        []byte{},
    ))
}

func (tx *Transaction) Sign(priv *rsa.PrivateKey) []byte {
    return Sign(priv, tx.CurrHash)
}

func (tx *Transaction) HashIsValid() bool {
    return bytes.Equal(tx.Hash(), tx.CurrHash)
}

func (tx *Transaction) SignIsValid() bool {
    return Verify(ParsePublic(tx.Sender), tx.CurrHash, tx.Signature) == nil
}
