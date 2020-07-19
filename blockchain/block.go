package blockchain

import (
	"time"
	"errors"
	"bytes"
	"crypto/rsa"
	"math/big"
	"sort"
)

func NewBlock(miner string, prevHash []byte) *Block {
	return &Block{
		Difficulty: DIFFICULTY,
		PrevHash:   prevHash,
		Miner:      miner,
		Mapping:    make(map[string]uint64),
	}
}

func (block *Block) Accept(chain *BlockChain, user *User, ch chan bool) error {
	if !block.transactionsIsValid(chain) {
		return errors.New("transactions is not valid")
	}
	block.AddTransaction(chain, &Transaction{
		RandBytes: GenerateRandomBytes(RAND_BYTES),
		Sender:    STORAGE_CHAIN,
		Receiver:  user.Address(),
		Value:     STORAGE_REWARD,
	})
	block.TimeStamp = time.Now().Format(time.RFC3339)
	block.CurrHash = block.hash()
	block.Signature = block.sign(user.Private())
	block.Nonce = block.proof(ch)
	return nil
}

func (block *Block) AddTransaction(chain *BlockChain, tx *Transaction) error {
	if tx == nil {
		return errors.New("tx is null")
	}
	if tx.Value == 0 {
		return errors.New("tx value = 0")
	}
	if len(block.Transactions) == TXS_LIMIT && tx.Sender != STORAGE_CHAIN {
		return errors.New("len tx = limit")
	}
	var balanceInChain uint64
	balanceInTX := tx.Value + tx.ToStorage
	if value, ok := block.Mapping[tx.Sender]; ok {
		balanceInChain = value
	} else {
		balanceInChain = chain.Balance(tx.Sender)
	}
	if tx.Value > START_PERCENT && tx.ToStorage != STORAGE_REWARD {
		return errors.New("storage reward pass")
	}
	if balanceInTX > balanceInChain {
		return errors.New("insufficient funds")
	}
	block.Mapping[tx.Sender] = balanceInChain - balanceInTX
	block.addBalance(chain, tx.Receiver, tx.Value)
	block.addBalance(chain, STORAGE_CHAIN, tx.ToStorage)
	block.Transactions = append(block.Transactions, *tx)
	return nil
}

func (block *Block) IsValid(chain *BlockChain) bool {
	switch {
	case block == nil:
		return false
	case block.Difficulty != DIFFICULTY:
		return false
	case !block.hashIsValid(chain, chain.Size()):
		return false
	case !block.signIsValid():
		return false
	case !block.proofIsValid():
		return false
	case !block.mappingIsValid():
		return false
	case !block.timeIsValid(chain, chain.Size()):
		return false
	case !block.transactionsIsValid(chain):
		return false
	}
	return true
}

func (block *Block) addBalance(chain *BlockChain, receiver string, value uint64) {
	var balanceInChain uint64
	if v, ok := block.Mapping[receiver]; ok {
		balanceInChain = v
	} else {
		balanceInChain = chain.Balance(receiver)
	}
	block.Mapping[receiver] = balanceInChain + value
}

func (block *Block) timeIsValid(chain *BlockChain, index uint64) bool {
	btime, err := time.Parse(time.RFC3339, block.TimeStamp)
	if err != nil {
		return false
	}

	diff := time.Now().Sub(btime)
	if diff < 0 {
		return false
	}

	var sblock string
	row := chain.DB.QueryRow("SELECT Block FROM BlockChain WHERE Hash=$1", 
		Base64Encode(block.PrevHash))
	row.Scan(&sblock)

	lblock := DeserializeBlock(sblock)
	if lblock == nil {
		return false
	}

	ltime, err := time.Parse(time.RFC3339, lblock.TimeStamp)
	if err != nil {
		return false
	}

	result := btime.Sub(ltime)
	return result > 0
}

func (block *Block) transactionsIsValid(chain *BlockChain) bool {
	lentxs := len(block.Transactions)
	plusStorage := 0
	for i := 0; i < lentxs; i++ {
		if block.Transactions[i].Sender == STORAGE_CHAIN {
			plusStorage = 1
			break
		}
	}
	if lentxs == 0 || lentxs > TXS_LIMIT+plusStorage {
		return false
	}
	for i := 0; i < lentxs-1; i++ {
		for j := i + 1; j < lentxs; j++ {
			if bytes.Equal(block.Transactions[i].RandBytes, block.Transactions[j].RandBytes) {
				return false
			}
			if 	block.Transactions[i].Sender == STORAGE_CHAIN && 
				block.Transactions[j].Sender == STORAGE_CHAIN {
					return false
			}
		}
	}
	for i := 0; i < lentxs; i++ {
		tx := block.Transactions[i]
		if tx.Sender == STORAGE_CHAIN {
			if tx.Receiver != block.Miner || tx.Value != STORAGE_REWARD {
				return false
			}
		} else {
			if !tx.hashIsValid() {
				return false
			}
			if !tx.signIsValid() {
				return false
			}
		}
		if !block.balanceIsValid(chain, tx.Sender) {
			return false
		}
		if !block.balanceIsValid(chain, tx.Receiver) {
			return false
		}
	}
	return true
}

func (block *Block) balanceIsValid(chain *BlockChain, address string) bool {
	if _, ok := block.Mapping[address]; !ok {
		return false
	}
	lentxs := len(block.Transactions)
	balanceInChain := chain.Balance(address)
	balanceSubBlock := uint64(0)
	balanceAddBlock := uint64(0)
	for j := 0; j < lentxs; j++ {
		tx := block.Transactions[j]
		if tx.Sender == address {
			balanceSubBlock += tx.Value + tx.ToStorage
		}
		if tx.Receiver == address {
			balanceAddBlock += tx.Value
		}
		if tx.Receiver == address && STORAGE_CHAIN == address {
			balanceAddBlock += tx.ToStorage
		}
	}
	if (balanceInChain + balanceAddBlock - balanceSubBlock) != block.Mapping[address] {
		return false
	}
	return true
}

func (block *Block) hash() []byte {
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
			[]byte(block.TimeStamp),
		},
		[]byte{},
	))
}

func (block *Block) sign(priv *rsa.PrivateKey) []byte {
	return Sign(priv, block.CurrHash)
}

func (block *Block) proof(ch chan bool) uint64 {
	return ProofOfWork(block.CurrHash, block.Difficulty, ch)
}

func (block *Block) hashIsValid(chain *BlockChain, index uint64) bool {
	if !bytes.Equal(block.hash(), block.CurrHash) {
		return false
	}
	var id uint64
	row := chain.DB.QueryRow("SELECT Id FROM BlockChain WHERE Hash=$1", 
		Base64Encode(block.PrevHash))
	row.Scan(&id)
	return id == index
}

func (block *Block) signIsValid() bool {
	return Verify(ParsePublic(block.Miner), block.CurrHash, block.Signature) == nil
}

func (block *Block) proofIsValid() bool {
	intHash := big.NewInt(1)
	Target := big.NewInt(1)
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

func (block *Block) mappingIsValid() bool {
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
