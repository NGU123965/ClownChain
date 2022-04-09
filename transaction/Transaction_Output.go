package transaction

import (
	Wallet "ClownChain/wallet"
	"bytes"
)

// TxOutput 交易输出
type TxOutput struct {
	// 1. 有多少钱(金额)
	Value int64
	// 2. 钱是谁的(用户名)
	// ScriptPubkey	string
	Ripemd160Hash []byte // 哈希值
}

// UnLockScriptPubkeyWithAddress output 身份验证
func (txOutput *TxOutput) UnLockScriptPubkeyWithAddress(address string) bool {
	hash160 := Lock(address)
	return bytes.Compare(txOutput.Ripemd160Hash, hash160) == 0
}

// Lock 相当于锁定
func Lock(address string) []byte {
	publicKeyHash := Wallet.Base58Decode([]byte(address))
	hash160 := publicKeyHash[1 : len(publicKeyHash)-Wallet.AddressChecksumLen]
	return hash160
}

// NewTXOutput 创建output对象
func NewTXOutput(value int64, address string) *TxOutput {
	txOutput := &TxOutput{}
	hash160 := Lock(address)
	txOutput.Value = value
	txOutput.Ripemd160Hash = hash160
	return txOutput
}
