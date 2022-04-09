package transaction

import (
	Wallet "ClownChain/wallet"
	"bytes"
)

// TxInput 交易输入
type TxInput struct {
	// 交易哈希(不是当前交易的哈希，而是该输入所引用的交易的哈希)
	TxHash []byte
	// 引用的上一笔交易的output索引
	Vout int

	Signature []byte // 数字签名
	// 公钥
	PublicKey []byte //公钥
}

func (in *TxInput) UnLockRipemd160Hash(ripemd160Hash []byte) bool {
	// 获取input的ripemd160哈希
	inputRipemd160 := Wallet.Ripemd160Hash(in.PublicKey)
	return bytes.Compare(inputRipemd160, ripemd160Hash) == 0
}
