package blockchain

import (
	"github.com/boltdb/bolt"
	"log"
)

// BlockChainIterator 区块链迭代器结构
type BlockChainIterator struct {
	DB          *bolt.DB // 数据库
	CurrentHash []byte   // 当前区块的哈希值
}

// Iterator 创建迭代器对象
func (bc *BlockChain) Iterator() *BlockChainIterator {
	return &BlockChainIterator{bc.DB, bc.Tip}
}

// Next 遍历
func (bcit *BlockChainIterator) Next() *Block {
	var block *Block
	err := bcit.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if nil != b {
			// 获取指定哈希的区块数据
			currentBlockBytes := b.Get(bcit.CurrentHash)
			block = DeserializeBlock(currentBlockBytes)
			// 更新迭代器中当前区块的哈希值
			bcit.CurrentHash = block.PrevBlockHash
		}
		return nil
	})
	if nil != err {
		log.Panicf("iterator the db of blockchain failed! %v\n", err)
	}
	return block
}
