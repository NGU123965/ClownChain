package blockchain

import (
	"ClownChain/transaction"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"
)

// Block 区块结构
type Block struct {
	TimeStamp     int64                      // 区块时间戳
	Heigth        int64                      // 区块高度
	PrevBlockHash []byte                     // 父区块哈希
	Hash          []byte                     // 当前区块哈希
	Txs           []*transaction.Transaction // 交易数据
	Nonce         int64                      // 随机值（PoW）
}

// NewBlock 创建新区块
func NewBlock(height int64, prevBlockHash []byte, txs []*transaction.Transaction) *Block {
	fmt.Println("NewBlock...")
	var block Block
	block = Block{Heigth: height, PrevBlockHash: prevBlockHash, Txs: txs, TimeStamp: time.Now().Unix()}
	//block.SetHash() // 生成区块当前哈希
	pow := NewProofOfWork(&block)
	hash, nonce := pow.Run() // 工作量证明
	block.Hash = hash
	block.Nonce = nonce
	return &block
}

// 计算区块哈希
//func (b *Block) SetHash()  {
//	// int64转换成字节数组
//	heightBytes := IntToHex(b.Heigth) // 刻度转换
//	timeStampBytes := IntToHex(b.TimeStamp)
//	// 拼接所有属性，进行哈希
//	blockBytes := bytes.Join([][]byte{heightBytes, timeStampBytes, b.PrevBlockHash, b.Data},[]byte{})
//	hash := sha256.Sum256(blockBytes)
//	b.Hash = hash[:]
//}

// CreateGenesisBlock 生成创世区块
func CreateGenesisBlock(txs []*transaction.Transaction) *Block {
	return NewBlock(1, nil, txs)
}

// Serialize 序列化，将区块结构序列化为[]byte
func (block *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)            //新建encode对象
	if err := encoder.Encode(block); nil != err { // 编码
		log.Panicf("serialize the block to byte failed! %v\n", err)
	}
	return result.Bytes()
}

// DeserializeBlock 反序列化， 将字节数组结构化为区块
func DeserializeBlock(blockBytes []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	if err := decoder.Decode(&block); nil != err {
		log.Panicf("deserialize the []byte to block failed! %v\n", err)
	}
	return &block
}

// HashTransactions 把区块中的所有交易结构转换成[]byte
func (block *Block) HashTransactions() []byte {
	var transactions [][]byte
	for _, tx := range block.Txs {
		transactions = append(transactions, tx.Serialize())
	}
	mTree := NewMerkleTree(transactions)
	//// sha256
	//txHash := sha256.Sum256(bytes.Join(txHashes, []byte{}))
	// 改成Merkle树根哈希
	return mTree.RootNode.Data
}
