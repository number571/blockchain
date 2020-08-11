package blockchain

import (
 "time"
 "errors"
 "bytes"
 "sort"
 "database/sql"
 _ "github.com/mattn/go-sqlite3"
 "os"
 "crypto"
 "crypto/rand"
 "crypto/rsa"
 "crypto/sha256"
 "crypto/x509"
 "encoding/base64"
 "encoding/binary"
 "encoding/json"
 "fmt"
 "math"
 "math/big"
 mrand "math/rand"
)

type BlockChain struct {
 DB    *sql.DB
}

type Block struct {
 Nonce         uint64
 Difficulty    uint8
 CurrHash      []byte
 PrevHash      []byte
 Transactions  []Transaction
 Mapping       map[string]uint64
 Miner         string
 Signature     []byte
 TimeStamp     string
}

type Transaction struct {
  RandBytes  []byte
  PrevBlock  []byte
  Sender     string
  Receiver   string
  Value      uint64
  ToStorage  uint64
  CurrHash   []byte
  Signature  []byte
}

func NewChain(filename, receiver string) error {
  file, err := os.Create(filename)
  if err != nil {
    return err
  }
  file.Close()
  db, err := sql.Open("sqlite3", filename)
  if err != nil {
    return err
  }
  defer db.Close()
  _, err = db.Exec(CREATE_TABLE)
  chain := &BlockChain{
    DB: db,
  }
  genesis := &Block{
    PrevHash: []byte(GENESIS_BLOCK),
    Mapping:   make(map[string]uint64),
    Miner:     receiver,
    TimeStamp: time.Now().Format(time.RFC3339),
  }
  genesis.Mapping[STORAGE_CHAIN] = STORAGE_VALUE
  genesis.Mapping[receiver] = GENESIS_REWARD
  genesis.CurrHash = genesis.hash()
  chain.AddBlock(genesis)
  return nil
}

const (
 CREATE_TABLE = `
CREATE TABLE BlockChain (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    Hash VARCHAR(44) UNIQUE,
    Block TEXT
);
`
)

const (
 GENESIS_BLOCK  = "GENESIS-BLOCK"
  STORAGE_VALUE  = 100
  GENESIS_REWARD = 100
  STORAGE_CHAIN  = "STORAGE-CHAIN"
)

func (chain *BlockChain) AddBlock(block *Block) {
  chain.DB.Exec("INSERT INTO BlockChain (Hash, Block) VALUES ($1, $2)",
   Base64Encode(block.CurrHash),
   SerializeBlock(block),
  )
}

func Base64Encode(data []byte) string {
 return base64.StdEncoding.EncodeToString(data)
}

func SerializeBlock(block *Block) string {
 jsonData, err := json.MarshalIndent(*block, "", "\t")
 if err != nil {
  return ""
 }
 return string(jsonData)
}

func LoadChain(filename string) *BlockChain {
 db, err := sql.Open("sqlite3", filename)
 if err != nil {
  return nil
 }
 chain := &BlockChain{
  DB: db,
 }
 return chain
}

func NewBlock(miner string, prevHash []byte) *Block {
 return &Block{
  Difficulty: DIFFICULTY,
  PrevHash:   prevHash,
  Miner:      miner,
  Mapping:    make(map[string]uint64),
 }
}

const (
 DIFFICULTY = 20
)

func NewTransaction(user *User, lasthash []byte, to string, value uint64) *Transaction {
 tx := &Transaction{
  RandBytes: GenerateRandomBytes(RAND_BYTES),
  PrevBlock: lasthash,
  Sender:    user.Address(),
  Receiver:  to,
  Value:     value,
 }
 if value > START_PERCENT {
  tx.ToStorage = STORAGE_REWARD
 }
 tx.CurrHash = tx.hash()
 tx.Signature = tx.sign(user.Private())
 return tx
}

type User struct {
 PrivateKey *rsa.PrivateKey
}

func GenerateRandomBytes(max uint) []byte {
 var slice []byte = make([]byte, max)
 _, err := rand.Read(slice)
 if err != nil {
  return nil
 }
 return slice
}

const (
 RAND_BYTES = 32
 START_PERCENT = 10
 STORAGE_REWARD = 1
)

func (user *User) Address() string {
 return StringPublic(user.Public())
}

func (user *User) Private() *rsa.PrivateKey {
 return user.PrivateKey
}

func (tx *Transaction) hash() []byte {
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

func (tx *Transaction) sign(priv *rsa.PrivateKey) []byte {
 return Sign(priv, tx.CurrHash)
}

func StringPublic(pub *rsa.PublicKey) string {
 return Base64Encode(x509.MarshalPKCS1PublicKey(pub))
}

func (user *User) Public() *rsa.PublicKey {
 return &(user.PrivateKey).PublicKey
}

func HashSum(data []byte) []byte {
 hash := sha256.Sum256(data)
 return hash[:]
}

func ToBytes(num uint64) []byte {
 var data = new(bytes.Buffer)
 err := binary.Write(data, binary.BigEndian, num)
 if err != nil {
  return nil
 }
 return data.Bytes()
}

func Sign(priv *rsa.PrivateKey, data []byte) []byte {
 signature, err := rsa.SignPSS(rand.Reader, priv, crypto.SHA256, data, nil)
 if err != nil {
  return nil
 }
 return signature
}

func (block *Block) AddTransaction(chain *BlockChain, tx *Transaction) error {
 if tx == nil {
  return errors.New("tx is null")
 }
 if tx.Value == 0 {
  return errors.New("tx value = 0")
 }
 if tx.Sender != STORAGE_CHAIN && len(block.Transactions) == TXS_LIMIT {
  return errors.New("len tx = limit")
 }
 if tx.Sender != STORAGE_CHAIN && tx.Value > START_PERCENT && tx.ToStorage != STORAGE_REWARD {
  return errors.New("storage reward pass")
 }
 if !bytes.Equal(tx.PrevBlock, chain.LastHash()) {
  return errors.New("prev block in tx /= last hash in chain")
 }
 var balanceInChain uint64
 balanceInTX := tx.Value + tx.ToStorage
 if value, ok := block.Mapping[tx.Sender]; ok {
  balanceInChain = value
 } else {
  balanceInChain = chain.Balance(tx.Sender, chain.Size())
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

const (
 TXS_LIMIT = 2
)

func (chain *BlockChain) Balance(address string, size uint64) uint64 {
 var (
  sblock  string
  block   *Block
  balance uint64
 )
 rows, err := chain.DB.Query("SELECT Block FROM BlockChain WHERE Id <= $1 ORDER BY Id DESC", size)
 if err != nil {
  return balance
 }
 defer rows.Close()
 for rows.Next() {
  rows.Scan(&sblock)
  block = DeserializeBlock(sblock)
  if value, ok := block.Mapping[address]; ok {
   balance = value
   break
  }
 }
 return balance
}

func (block *Block) addBalance(chain *BlockChain, receiver string, value uint64) {
 var balanceInChain uint64
 if v, ok := block.Mapping[receiver]; ok {
  balanceInChain = v
 } else {
  balanceInChain = chain.Balance(receiver, chain.Size())
 }
 block.Mapping[receiver] = balanceInChain + value
}

func (chain *BlockChain) Size() uint64 {
 var size uint64
 row := chain.DB.QueryRow("SELECT Id FROM BlockChain ORDER BY Id DESC")
 row.Scan(&size)
 return size
}

func DeserializeBlock(data string) *Block {
 var block Block
 err := json.Unmarshal([]byte(data), &block)
 if err != nil {
  return nil
 }
 return &block
}

func (block *Block) Accept(chain *BlockChain, user *User, ch chan bool) error {
 if !block.transactionsIsValid(chain, chain.Size()) {
  return errors.New("transactions is not valid")
 }
 block.AddTransaction(chain, &Transaction{
  RandBytes: GenerateRandomBytes(RAND_BYTES),
  PrevBlock: chain.LastHash(),
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

func (block *Block) transactionsIsValid(chain *BlockChain, size uint64) bool {
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
   if  block.Transactions[i].Sender == STORAGE_CHAIN && 
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
  if !block.balanceIsValid(chain, tx.Sender, size) {
   return false
  }
  if !block.balanceIsValid(chain, tx.Receiver, size) {
   return false
  }
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

func (tx *Transaction) hashIsValid() bool {
 return bytes.Equal(tx.hash(), tx.CurrHash)
}

func (tx *Transaction) signIsValid() bool {
 return Verify(ParsePublic(tx.Sender), tx.CurrHash, tx.Signature) == nil
}

func (block *Block) balanceIsValid(chain *BlockChain, address string, size uint64) bool {
 if _, ok := block.Mapping[address]; !ok {
  return false
 }
 lentxs := len(block.Transactions)
 balanceInChain := chain.Balance(address, size)
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
  if STORAGE_CHAIN == address {
   balanceAddBlock += tx.ToStorage
  }
 }
 if (balanceInChain + balanceAddBlock - balanceSubBlock) != block.Mapping[address] {
  return false
 }
 return true
}

func ProofOfWork(blockHash []byte, difficulty uint8, ch chan bool) uint64 {
 var (
  Target  = big.NewInt(1)
  intHash = big.NewInt(1)
  nonce   = uint64(mrand.Intn(math.MaxUint32))
  hash    []byte
 )
 Target.Lsh(Target, 256 - uint(difficulty))
 for nonce < math.MaxUint64 {
  select {
  case <-ch:
   if DEBUG {
    fmt.Println()
   }
   return nonce
  default:
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
    return nonce
   }
   nonce++
  }
 }
 return nonce
}

func Verify(pub *rsa.PublicKey, data, sign []byte) error {
 return rsa.VerifyPSS(pub, crypto.SHA256, data, sign, nil)
}

func ParsePublic(pubData string) *rsa.PublicKey {
 pub, err := x509.ParsePKCS1PublicKey(Base64Decode(pubData))
 if err != nil {
  return nil
 }
 return pub
}

const (
 DEBUG = true
)

func init() {
  mrand.Seed(time.Now().UnixNano())
}

func Base64Decode(data string) []byte {
 result, err := base64.StdEncoding.DecodeString(data)
 if err != nil {
  return nil
 }
 return result
}

func GeneratePrivate(bits uint) *rsa.PrivateKey {
 priv, err := rsa.GenerateKey(rand.Reader, int(bits))
 if err != nil {
  return nil
 }
 return priv
}

func StringPrivate(priv *rsa.PrivateKey) string {
 return Base64Encode(x509.MarshalPKCS1PrivateKey(priv))
}

func ParsePrivate(privData string) *rsa.PrivateKey {
 priv , err := x509.ParsePKCS1PrivateKey(Base64Decode(privData))
 if err != nil {
  return nil
 }
 return priv 
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

func (user *User) Purse() string {
 return StringPrivate(user.Private())
}

func (chain *BlockChain) LastHash() []byte {
 var hash string
 row := chain.DB.QueryRow("SELECT Hash FROM BlockChain ORDER BY Id DESC")
 row.Scan(&hash)
 return Base64Decode(hash)
}

func (block *Block) IsValid(chain *BlockChain, size uint64) bool {
 switch {
 case block == nil:
  return false
 case block.Difficulty != DIFFICULTY:
  return false
 case !block.hashIsValid(chain, size):
  return false
 case !block.signIsValid():
  return false
 case !block.proofIsValid():
  return false
 case !block.mappingIsValid():
  return false
 case !block.timeIsValid(chain):
  return false
 case !block.transactionsIsValid(chain, size):
  return false
 }
 return true
}

func SerializeTX(tx *Transaction) string {
 jsonData, err := json.MarshalIndent(*tx, "", "\t")
 if err != nil {
  return ""
 }
 return string(jsonData)
}

func DeserializeTX(data string) *Transaction {
 var tx Transaction
 err := json.Unmarshal([]byte(data), &tx)
 if err != nil {
  return nil
 }
 return &tx
}

const (
 KEY_SIZE = 512
)

func (block *Block) hashIsValid(chain *BlockChain, size uint64) bool {
 if !bytes.Equal(block.hash(), block.CurrHash) {
  return false
 }
 var id uint64
 row := chain.DB.QueryRow("SELECT Id FROM BlockChain WHERE Hash=$1", 
  Base64Encode(block.PrevHash))
 row.Scan(&id)
 return id == size 
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
 Target.Lsh(Target, 256 - uint(block.Difficulty))
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

func (block *Block) timeIsValid(chain *BlockChain) bool {
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

