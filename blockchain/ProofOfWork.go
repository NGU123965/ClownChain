package blockchain

import (
	"ClownChain/utils"
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

// 目标难度值，代表生成的哈希前targetBit位为0，才能满足条件
const targetBit = 16

// ProofOfWork 工作量证明
type ProofOfWork struct {
	Block  *Block   // 对指定的区块进行验证
	target *big.Int // 大数据存储
}

// NewProofOfWork 创建新的POW对象
func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	// 8
	// 前2位都为0
	// 左移
	// 8-2
	// 1 << 6
	//0010 0000 = 64
	target = target.Lsh(target, 256-targetBit)

	return &ProofOfWork{block, target}
}

// Run 开始工作量证明
func (proofOfWork *ProofOfWork) Run() ([]byte, int64) {

	var nonce = 0     // 碰撞次数
	var hash [32]byte // 生成的哈希值

	var hashInt big.Int // 存储哈希转换之后生成的数据，最终和target数据进行比较
	for {
		// 1. 数据拼接
		dataBytes := proofOfWork.prepareData(nonce) // 得到准备数据
		hash = sha256.Sum256(dataBytes)
		hashInt.SetBytes(hash[:])
		fmt.Printf("hash : \r%x", hash)
		// 难度比较
		if proofOfWork.target.Cmp(&hashInt) == 1 {
			break
		}
		nonce++
	}
	fmt.Printf("\n碰撞次数: %d\n", nonce)
	return hash[:], int64(nonce)
}

// 准备数据，将区块相差属性搭接越来，返回一个字节数组
func (proofOfWork *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join([][]byte{
		proofOfWork.Block.PrevBlockHash,
		proofOfWork.Block.HashTransactions(),
		utils.IntToHex(proofOfWork.Block.TimeStamp),
		utils.IntToHex(proofOfWork.Block.Heigth),
		utils.IntToHex(int64(nonce)),
		utils.IntToHex(targetBit),
	}, []byte{})

	return data
}
