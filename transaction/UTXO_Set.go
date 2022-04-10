package transaction

import (
	"ClownChain/blockchain"
	Wallet "ClownChain/wallet"
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

// 持久化
// utxo表名
const utxoTableName = "utxoTable"

// UTXOSet 结构(保存指定区块链所有的UTXO)
type UTXOSet struct {
	BlockChain *blockchain.BlockChain
}

// Serialize 将utxo集合序列化为字节数组
func (txOutputs *TXOutputs) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result) // 新建encode对象
	if err := encoder.Encode(txOutputs); nil != err {
		log.Panicf("serialize the utxo failed! %v\n", err)
	}
	return result.Bytes()
}

// DeserializeTXOutputs 反序列化
func DeserializeTXOutputs(txOutputBytes []byte) *TXOutputs {
	var txOutputs TXOutputs
	decoder := gob.NewDecoder(bytes.NewReader(txOutputBytes))
	err := decoder.Decode(&txOutputs)
	if nil != err {
		log.Panicf("deserialize the struct of txoutputs failed! %v\n", err)
	}
	return &txOutputs
}

// ResetUTXOSet 重置UTXO
func (utxoSet *UTXOSet) ResetUTXOSet() {
	// 更新utxo table
	// 采用覆盖的方式
	err := utxoSet.BlockChain.DB.Update(func(tx *bolt.Tx) error {
		// 查找utxo 表
		b := tx.Bucket([]byte(utxoTableName))
		if nil != b {
			tx.DeleteBucket([]byte(utxoTableName))
		}
		c, _ := tx.CreateBucket([]byte(utxoTableName))
		if c != nil {
			// 查找所有未花费的输出
			txOutputsMap := utxoSet.BlockChain.FindUTXOMap()
			for keyHash, output := range txOutputsMap {
				txHash, _ := hex.DecodeString(keyHash)
				// 存入表
				err := c.Put(txHash, output.Serialize())
				if nil != err {
					log.Panicf("put the utxo into table failed! %v\n", err)
				}
			}
		}
		return nil
	})

	if nil != err {
		log.Panicf("updata the db of utxoset failed! %v\n", err)
	}
}

// GetBalance 获取余额
func (utxoSet *UTXOSet) GetBalance(address string) int64 {
	// 获取指定地址的UTXO
	UTXOS := utxoSet.FindUTXOWithAddress(address)
	var amount int64 // 余额
	for _, utxo := range UTXOS {
		fmt.Printf("\tutxo-hash : %x\n", utxo.TxHash)
		fmt.Printf("\tutxo-index : %d\n", utxo.Index)
		fmt.Printf("\t\tutxo-Ripemd160Hash : %x\n", utxo.Output.Ripemd160Hash)
		fmt.Printf("\t\tutxo-value : %d\n", utxo.Output.Value)
		amount += utxo.Output.Value
	}
	return amount
}

// FindUTXOWithAddress 查找指定地址的UTXO
func (utxoSet *UTXOSet) FindUTXOWithAddress(address string) []*UTXO {
	var utxos []*UTXO
	// 查找数据库中的utxoTable表
	utxoSet.BlockChain.DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(utxoTableName))
		if nil != b {
			c := b.Cursor() // 游标
			// 遍历每一个UTXO
			for k, v := c.First(); k != nil; k, v = c.Next() {
				// k -> 交易哈希
				// v -> 输出结构的字节数组
				txOutputs := DeserializeTXOutputs(v)
				for _, utxo := range txOutputs.UTXOS {
					if utxo.Output.UnLockScriptPubkeyWithAddress(address) {
						//utxo_single := UTXO{Output:utxo}
						utxos = append(utxos, utxo)
					}
				}
			}
		}

		return nil
	})

	return utxos
}

// FindSpendableUTXO 查找可花费的UTXO
func (utxoSet *UTXOSet) FindSpendableUTXO(from string, amount int64, txs []*Transaction) (int64, map[string][]int) {
	// 查找整个UTXO表中的未花费输出
	// 1. 先从未打包的交易中获取UTXO，如果足够，不再查询UTXO Table
	spendableUTXO := make(map[string][]int)
	// 2. 查找未打包的交易中的UTXO
	unPackagesUTXOs := utxoSet.FindUnPackageSpendableUTXOS(from, txs)
	var value int64 = 0

	for _, utxo := range unPackagesUTXOs {
		value += utxo.Output.Value
		txHash := hex.EncodeToString(utxo.TxHash)
		spendableUTXO[txHash] = append(spendableUTXO[txHash], utxo.Index)
		if value >= amount {
			// 钱够了
			return value, spendableUTXO
		}
	}

	// 2 在获取到未打包交易中的UTXO集合之后，金额仍然不足，从UTXO集合表中获取
	utxoSet.BlockChain.DB.View(func(tx *bolt.Tx) error {
		// 先获取表
		b := tx.Bucket([]byte(utxoTableName))
		if b != nil {
			cursor := b.Cursor() // 有序遍历
		UTXOBREAK:
			for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
				txOutputs := DeserializeTXOutputs(v)
				for _, utxo := range txOutputs.UTXOS {
					// 用户验证判断处理
					if utxo.Output.UnLockScriptPubkeyWithAddress(from) {
						value += utxo.Output.Value
						if value >= amount {
							txHash := hex.EncodeToString(utxo.TxHash)
							spendableUTXO[txHash] = append(spendableUTXO[txHash], utxo.Index)
							break UTXOBREAK
						}
					}
				}
			}
		}
		return nil
	})

	if value < amount {
		log.Panic("余额不足......")
	}
	return value, spendableUTXO
}

// FindUnPackageSpendableUTXOS 从未打包的交易中进行查找
func (utxoSet *UTXOSet) FindUnPackageSpendableUTXOS(from string, txs []*Transaction) []*UTXO {
	var unUTXOs []*UTXO                      // 未打包交易中的UTXO
	spentTXOutputs := make(map[string][]int) // 每个交易中的已花费输出(索引)
	for _, tx := range txs {
		// 排队coinbase交易
		if !tx.IsCoinbaseTransaction() {
			for _, vin := range tx.Vins {
				pubKeyHash := Wallet.Base58Decode([]byte(from))    // 解码，获取公钥哈希
				ripemd160Hash := pubKeyHash[1 : len(pubKeyHash)-4] // 用户名
				if vin.UnLockRipemd160Hash(ripemd160Hash) {        // 解锁
					key := hex.EncodeToString(vin.TxHash)
					spentTXOutputs[key] = append(spentTXOutputs[key], vin.Vout)
				}
			}
		}
	}

	for _, tx := range txs {
	UnUtxoLoop:
		for index, vout := range tx.Vouts {
			// 判断该vout是否属于from
			if vout.UnLockScriptPubkeyWithAddress(from) {
				// 在没有包含已花费输出的情况
				if len(spentTXOutputs) == 0 {
					utxo := &UTXO{tx.TxHash, index, vout}
					unUTXOs = append(unUTXOs, utxo)
				} else {
					for hash, indexArray := range spentTXOutputs {
						txHashStr := hex.EncodeToString(tx.TxHash)
						// 判断当前交易是否包含了已花费输出
						if hash == txHashStr {
							var isUnpkgSpentUTXO bool // 判断该输出是否属于已花费输出
							for _, idx := range indexArray {
								if index == idx {
									isUnpkgSpentUTXO = true
									continue UnUtxoLoop
								}
							}
							if isUnpkgSpentUTXO == false {
								utxo := &UTXO{tx.TxHash, index, vout}
								unUTXOs = append(unUTXOs, utxo)
							}
						} else {
							// 该交易没有包含已花费输出
							utxo := &UTXO{tx.TxHash, index, vout}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}
			}
		}
	}
	return unUTXOs
}

// Update /*
// Utxo update实时更新（简化版）
func (utxoSet *UTXOSet) Update() {
	// 获取最新区块
	latest_block := utxoSet.BlockChain.Iterator().Next()
	db := utxoSet.BlockChain.DB
	err := db.Update(func(tx *bolt.Tx) error {
		// 获取数据表
		b := tx.Bucket([]byte(utxoTableName))
		if nil != b {
			for _, tx := range latest_block.Txs {
				if !tx.IsCoinbaseTransaction() {
					for _, vin := range tx.Vins {
						// 需要更新的输出
						updatedOutputs := TXOutputs{}
						outsBytes := b.Get(vin.TxHash) // 查找交易输入的关联输出
						outs := DeserializeTXOutputs(outsBytes)
						for outIdx, out := range outs.UTXOS {
							if outIdx != vin.Vout {
								updatedOutputs.UTXOS = append(updatedOutputs.UTXOS, out)
							}
						}
						//判断UTXOS长度
						if len(updatedOutputs.UTXOS) == 0 {
							b.Delete(vin.TxHash)
						} else {
							// 存入数据库
							b.Put(vin.TxHash, updatedOutputs.Serialize())
						}
					}
				}
				// 当前区块的最新的输出
				newOutputs := TXOutputs{}
				for i, out := range tx.Vouts {
					newOutputs.UTXOS = append(newOutputs.UTXOS, &UTXO{
						TxHash: tx.TxHash,
						Index:  i,
						Output: out,
					})
				}
				b.Put(tx.TxHash, newOutputs.Serialize())
			}
		}
		return nil
	})
	if nil != err {
		log.Panicf("update the UTXO Table failed! %v\n", err)
	}
}

// Updat1 实现 Utxo table实时更新
func (utxoSet *UTXOSet) Updat1() {
	// 找到需要删除的UTXO
	// 1. 获取最新的区块
	latest_block := utxoSet.BlockChain.Iterator().Next()
	var inputs []*TxInput // 存放最新区块的所有输入
	// 获取需要存入utxo table中的UTXO
	outsMap := make(map[string]*TXOutputs)
	// 2. 查找需要删除的数据
	for _, tx := range latest_block.Txs {
		// 遍历输入
		for _, vin := range tx.Vins {
			inputs = append(inputs, vin)
		}
	}
	// 3. 获取当前最新区块的所有UTXO
	for _, tx := range latest_block.Txs {
		var utxos []*UTXO
		for index, out := range tx.Vouts {
			isSpent := false
			for _, in := range inputs {
				if in.Vout == index && bytes.Compare(tx.TxHash, in.TxHash) == 0 {
					if bytes.Compare(out.Ripemd160Hash, Wallet.Ripemd160Hash(in.PublicKey)) == 0 {
						isSpent = true
						continue
					}
				}
			}
			if isSpent == false {
				utxo := &UTXO{tx.TxHash, index, out}
				utxos = append(utxos, utxo)
			}
		}
		if len(utxos) > 0 {
			txHash := hex.EncodeToString(tx.TxHash)
			outsMap[txHash] = &TXOutputs{utxos}
		}
	}
	// 4. 更新
	err := utxoSet.BlockChain.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoTableName))
		if nil != b {
			// 删除已花费输出
			for _, in := range inputs {
				txOutputsBytes := b.Get(in.TxHash) // 查找当前input所引用的交易哈希
				if len(txOutputsBytes) == 0 {
					continue
				}
				var UTXOS []*UTXO
				// 反序列化
				txOutpus := DeserializeTXOutputs(txOutputsBytes)
				isNeedToDel := false
				for _, utxo := range txOutpus.UTXOS {
					// 判断是哪一个输出被引用
					if in.Vout == utxo.Index {
						if bytes.Compare(utxo.Output.Ripemd160Hash, Wallet.Ripemd160Hash(in.PublicKey)) == 0 {
							isNeedToDel = true // 该输出已经被引用，需要删除
						}
					} else {
						UTXOS = append(UTXOS, utxo)
					}
				}

				if isNeedToDel {
					// 先删除输出
					b.Delete(in.TxHash)
					if len(UTXOS) > 0 {
						preTXOutputs := outsMap[hex.EncodeToString(in.TxHash)]
						preTXOutputs.UTXOS = append(preTXOutputs.UTXOS, UTXOS...)
						// 更新
						outsMap[hex.EncodeToString(in.TxHash)] = preTXOutputs
					}
				}

			}

			for hash, outputs := range outsMap {
				hashBytes, _ := hex.DecodeString(hash)
				err := b.Put(hashBytes, outputs.Serialize())
				if nil != err {
					log.Panicf("put the utxo to table failed! %v\n", err)
				}
			}
		}
		return nil
	})
	if nil != err {
		log.Printf("update the UTXOMap to utxo table failed! %v", err)
	}
}
