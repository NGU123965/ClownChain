package transaction

import (
	"ClownChain/blockchain"
	"ClownChain/utils"
	Wallet "ClownChain/wallet"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"time"
)

// 交易相关

// Transaction 交易结构
type Transaction struct {
	// tx hash(交易的唯一标识)
	TxHash []byte
	// 输入
	Vins []*TxInput
	// 输出
	Vouts []*TxOutput
}

// HashTransaction 生成交易哈希
func (tx *Transaction) HashTransaction() {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(tx)
	if nil != err {
		log.Panicf("tx hash generate failed! %v\n", err)
	}
	tm := time.Now().Unix() //添加时间戳标识， 没有时间标识会导致所有的coinbase交易的哈希完全一样
	txHashBytes := bytes.Join([][]byte{result.Bytes(), utils.IntToHex(tm)}, []byte{})
	hash := sha256.Sum256(txHashBytes)
	tx.TxHash = hash[:]
	//hash := sha256.Sum256(result.Bytes())
	//tx.TxHash = hash[:]
}

// NewCoinbaseTransaction 生成coinbase交易
func NewCoinbaseTransaction(address string) *Transaction {
	// 输入
	txInput := &TxInput{[]byte{}, -1, nil, nil}
	// 输出
	txOutput := NewTXOutput(10, address)
	txCoinbase := &Transaction{nil, []*TxInput{txInput}, []*TxOutput{txOutput}}
	// hash
	txCoinbase.HashTransaction()
	return txCoinbase
}

// NewSimpleTransaction 生成转账交易
func NewSimpleTransaction(from string, to string, amount int, blockchain *blockchain.BlockChain, txs []*Transaction, utxoSet *UTXOSet, nodeID string) *Transaction {
	var txInputs []*TxInput   // 输入
	var txOutputs []*TxOutput // 输出
	// 查找指定地址的可用UTXO
	//money, spendableUTXODic := blockchain.FindSpendableUTXO(from, int64(amount), txs)
	money, spendableUTXODic := utxoSet.FindSpendableUTXO(from, int64(amount), txs)

	fmt.Printf("money : %v\n", money)
	// 获取钱包集合
	wallets, _ := Wallet.NewWallets(nodeID)
	wallet := wallets.Wallets[from] // 指定地址对应的钱包结构
	fmt.Printf("spendableUTXODic: %v\n", spendableUTXODic)
	for txHash, indexArray := range spendableUTXODic {
		fmt.Printf("indexArray : %v\n", indexArray)
		txHashBytes, _ := hex.DecodeString(txHash)
		for _, index := range indexArray {
			// 此处的输出是需要消费的，必然会被其它的交易的输入所引用
			txInput := &TxInput{txHashBytes, index, nil, wallet.PublicKey}
			txInputs = append(txInputs, txInput)

		}
	}
	// 转账
	txOutput := NewTXOutput(int64(amount), to)
	txOutputs = append(txOutputs, txOutput)
	// 找零
	txOutput = NewTXOutput(money-int64(amount), from)
	txOutputs = append(txOutputs, txOutput)
	// 生成交易
	tx := &Transaction{nil, txInputs, txOutputs}
	tx.HashTransaction()
	// 对交易进行签名
	// signTransaction()
	// 参数主要为tx, wallet.PrivateKey
	for _, vin := range tx.Vins {
		// 查找所引用的每一个交易
		fmt.Printf("transaction.go  hash : [%x]\n", vin)
	}
	// 交易签名
	blockchain.SignTransaction(tx, wallet.PrivateKey, txs)
	return tx
}

// IsCoinbaseTransaction 判断指定交易是否是一个coinbase交易
func (tx *Transaction) IsCoinbaseTransaction() bool {
	return len(tx.Vins[0].TxHash) == 0 && tx.Vins[0].Vout == -1
}

// Sign 交易签名
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTxs map[string]Transaction) {
	// 判断是否是挖矿交易
	if tx.IsCoinbaseTransaction() {
		return
	}
	for _, vin := range tx.Vins {
		if prevTxs[hex.EncodeToString(vin.TxHash)].TxHash == nil {
			log.Panicf("ERROR:Prev transaction is not correct\n")
		}
	}
	// 提取需要签名的属性
	// 获取copy tx
	txCopy := tx.TrimmedCopy()
	for vin_id, vin := range txCopy.Vins {
		prevTx := prevTxs[hex.EncodeToString(vin.TxHash)] // 获取关联交易
		txCopy.Vins[vin_id].Signature = nil
		txCopy.Vins[vin_id].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Hash
		txCopy.TxHash = txCopy.Hash()
		txCopy.Vins[vin_id].PublicKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.TxHash)
		if nil != err {
			log.Panicf("sign to tx %x failed! %v\n", tx.TxHash, err)
		}
		// ECDSA的签名算法就是一对数字
		// Sig = (R,S)
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vins[vin_id].Signature = signature
	}
}

// Hash 设置一下用于签名的数据哈希
func (tx *Transaction) Hash() []byte {
	txCopy := tx
	txCopy.TxHash = []byte{}
	hash := sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

// Serialize 序列化
func (tx *Transaction) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)         //新建eoncode对象
	if err := encoder.Encode(tx); nil != err { // 编码
		log.Panicf("serialize the tx to byte failed! %v\n", err)
	}
	return result.Bytes()
}

// TrimmedCopy 添加一个交易的拷贝，用于交易签名，返回需要进行签名的交易
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []*TxInput
	var outputs []*TxOutput
	for _, vin := range tx.Vins {
		inputs = append(inputs, &TxInput{vin.TxHash, vin.Vout, nil, nil})
	}
	for _, vout := range tx.Vouts {
		outputs = append(outputs, &TxOutput{vout.Value, vout.Ripemd160Hash})
	}
	txCopy := Transaction{tx.TxHash, inputs, outputs}
	return txCopy
}

// Verify 交易验证
func (tx *Transaction) Verify(prevTxs map[string]Transaction) bool {
	if tx.IsCoinbaseTransaction() {
		return true
	}
	// 检查能否找到交易
	// 查找每个vin所引用的交易hash是否包含在prevTxs
	for _, vin := range tx.Vins {
		if prevTxs[hex.EncodeToString(vin.TxHash)].TxHash == nil {
			log.Panic("ERROR: Tx is Incorrect")
		}
	}

	txCopy := tx.TrimmedCopy()
	// 使用相同的椭圆获取密钥对
	curve := elliptic.P256()

	for vinId, vin := range tx.Vins {
		prevTx := prevTxs[hex.EncodeToString(vin.TxHash)]
		txCopy.Vins[vinId].Signature = nil
		txCopy.Vins[vinId].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Hash
		// 上面是生成哈希的数据，所以要与签名时的数据完全一致
		txCopy.TxHash = txCopy.Hash() // 要验证的数据
		txCopy.Vins[vinId].PublicKey = nil
		// 私钥ID
		// 获取r, s(r和s长度相等，根据椭圆加密计算的结果)
		// r, s代表签名
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])
		// 生成x,y(首先，签名是一个数字对，公钥是X,Y坐标组合，
		// 在生成公钥时，需要将X Y坐标组合到一起，在验证的时候，需要将
		// 公钥的X Y拆开)
		x := big.Int{}
		y := big.Int{}
		pubkeyLen := len(vin.PublicKey)
		x.SetBytes(vin.PublicKey[:(pubkeyLen / 2)])
		y.SetBytes(vin.PublicKey[(pubkeyLen / 2):])
		// 生成验证签名所需的公钥
		rawPublicKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		// 验证签名
		if !ecdsa.Verify(&rawPublicKey, txCopy.TxHash, &r, &s) {
			return false
		}
	}
	return true
}
