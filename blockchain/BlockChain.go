package blockchain

import (
	"ClownChain/transaction"
	"ClownChain/wallet"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"math/big"
	"os"
	"strconv"
)

const dbName = "bc_%s.db"       //存储区块数据的数据库文件
const blockTableName = "blocks" //表(桶)名称

// BlockChain 区块链基本结构
type BlockChain struct {
	DB  *bolt.DB // 数据库
	Tip []byte   //最新区块的哈希值
}

// DbExists 判断数据库是否存在
func DbExists(nodeID string) bool {
	dbName := fmt.Sprintf(dbName, nodeID)
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		return false
	}
	return true
}

// CreateBlockChainWithGenesisBlock 初始化区块链
func CreateBlockChainWithGenesisBlock(address string, nodeID string) *BlockChain {
	if DbExists(nodeID) {
		fmt.Println("创世区块已存在...")
		os.Exit(1) // 退出
	}
	dbName := fmt.Sprintf(dbName, nodeID)
	// 创建或者打开数据
	db, err := bolt.Open(dbName, 0600, nil)
	if nil != err {
		log.Panicf("open the db failed! %v\n", err)
	}
	//defer db.Close()
	var blockHash []byte // 需要存储到数据库中的区块哈希
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if nil == b {
			// 添加创世区块
			b, err = tx.CreateBucket([]byte(blockTableName))
			if nil != err {
				log.Panicf("create the bucket [%s] failed! %v\n", blockTableName, err)
			}
		}
		if nil != b {
			// 生成交易
			txCoinbase := transaction.NewCoinbaseTransaction(address)
			// 生成创世区块
			genesisBlock := CreateGenesisBlock([]*transaction.Transaction{txCoinbase})
			err = b.Put(genesisBlock.Hash, genesisBlock.Serialize())
			if nil != err {
				log.Panicf("put the data of genesisBlock to db failed! %v\n", err)
			}
			// 存储最新区块的哈希
			err = b.Put([]byte("l"), genesisBlock.Hash)
			if nil != err {
				log.Panicf("put the hash of latest block to db failed! %v\n", err)
			}
			blockHash = genesisBlock.Hash
		}
		return nil
	})
	if nil != err {
		log.Panicf("update the data of genesis block failed! %v\n", err)
	}
	return &BlockChain{db, blockHash}
}

// 添加新的区块到区块链中
//func (bc *BlockChain) AddBlock(txs []*Transaction) {
//
//	// 更新数据
//	err := bc.DB.Update(func(tx *bolt.Tx) error {
//		// 1 获取数据表
//		b := tx.Bucket([]byte(blockTableName))
//		if nil != b { // 2. 确保表存在
//			// 3. 获取最新区块的哈希
//			//	newEstHash := b.Get([]byte("l"))
//			blockBytes := b.Get(bc.Tip)
//			latest_block := DeserializeBlock(blockBytes)
//			// 4. 创建新区块
//			newBlock := NewBlock(latest_block.Heigth+1, latest_block.Hash, txs) // 创建一个新的区块
//			// 5. 存入数据库
//			err := b.Put(newBlock.Hash, newBlock.Serialize())
//			if nil != err {
//				log.Panicf("put the data of new block into db failed! %v\n", err)
//			}
//			// 6. 更新最新区块的哈希
//			err = b.Put([]byte("l"), newBlock.Hash)
//			if nil != err {
//				log.Panicf("put the hash of the newest block into db failed! %v\n", err)
//			}
//			bc.Tip = newBlock.Hash
//		}
//		return nil
//	})
//
//	if nil != err {
//		log.Panicf("update the db of block failed! %v\n", err)
//	}
//}

// PrintChain 遍历输出区块链所有区块的信息
func (bc *BlockChain) PrintChain() {
	fmt.Println("区块链完整信息...")
	var curBlock *Block
	///var currentHash []byte = bc.Tip // 获取最新区块的哈希
	// 创建一个迭代器对象
	bcit := bc.Iterator()
	for {
		fmt.Printf("----------------------------------------\n")
		curBlock = bcit.Next()
		fmt.Printf("\tHeigth : %d\n", curBlock.Heigth)
		fmt.Printf("\tTimeStamp : %d\n", curBlock.TimeStamp)
		fmt.Printf("\tPrevBlockHash : %x\n", curBlock.PrevBlockHash)
		fmt.Printf("\tHash : %x\n", curBlock.Hash)
		fmt.Printf("\tTransaction : %v\n", curBlock.Txs)
		for _, tx := range curBlock.Txs {
			fmt.Printf("\t\t tx-hash: %x\n", tx.TxHash)
			fmt.Println("\t\t输入...")
			for _, vin := range tx.Vins {
				fmt.Printf("\t\t\tvin-txhash:%x\n", vin.TxHash)
				fmt.Printf("\t\t\tvin-vout:%v\n", vin.Vout)
				//	fmt.Printf("\t\t\tvin-scriptsig:%v\n", vin.ScriptSig)
			}
			fmt.Println("\t\t输出...")
			for _, vout := range tx.Vouts {
				fmt.Printf("\t\t\tvout-value:%d\n", vout.Value)
				fmt.Printf("\t\t\tvout-ScriptPubkey:%x\n", vout.Ripemd160Hash)
			}
		}
		fmt.Printf("\tNonce : %d\n", curBlock.Nonce)

		// 判断是否已经遍历到创世区块
		var hashInt big.Int
		hashInt.SetBytes(curBlock.PrevBlockHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break // 跳出循环
		}
		//currentHash = curBlock.PrevBlockHash
	}

}

// BlockchainObject 返回Blockchain 对象
func BlockchainObject(nodeID string) *BlockChain {
	dbName := fmt.Sprintf(dbName, nodeID)
	// 读取数据库
	db, err := bolt.Open(dbName, 0600, nil)
	if nil != err {
		log.Panicf("get the object of blockchain failed! %v\n", err)
	}
	var tip []byte // 最新区块的哈希值
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if nil != b {
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	return &BlockChain{db, tip}
}

// MineNewBlock wallet_3000.dat
// 挖矿(生成新的区块)
// 通过接收交易，进行打包确认，最终生成新的区块
func (bc *BlockChain) MineNewBlock(from []string, to []string, amount []string, nodeID string) {
	fmt.Printf("\tFROM:[%s]\n", from)
	fmt.Printf("\tTO:[%s]\n", to)
	fmt.Printf("\tAMOUNT:[%s]\n", amount)
	// 接收交易
	var txs []*transaction.Transaction // 要打包的交易列表
	for index, address := range from {
		fmt.Printf("\tfrom:[%s], to[%s], amount:[%s]\n", address, to[index], amount[index])
		value, _ := strconv.Atoi(amount[index])
		utxoSet := &transaction.UTXOSet{BlockChain: bc}
		tx := transaction.NewSimpleTransaction(address, to[index], value, bc, txs, utxoSet, nodeID)
		txs = append(txs, tx)
		fmt.Printf("\ttx-hash:%x, tx-vouts:%v, tx-vins:%v\n", tx.TxHash, tx.Vouts, tx.Vins)
	}
	// 给矿工一定的奖励
	// 默认情况下，设置地址列表中的第一个地址为矿工奖励地址
	tx := transaction.NewCoinbaseTransaction(from[0])
	txs = append(txs, tx)
	// 打包交易
	// 生成新的区块
	var block *Block
	// 从数据库中获取最新区块
	bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if nil != b {
			hash := b.Get([]byte("l"))           // 获取最新区块哈希值(当作新生区块的prevHash)
			blockBytes := b.Get(hash)            // 得到最新区块(为了获取区块高度)
			block = DeserializeBlock(blockBytes) // 反序列化
		}
		return nil
	})
	// 在生成新区块之前，对交易签名进行验证
	// 在这里验证一下交易签名
	var _txs []*transaction.Transaction // 未打包的关联交易
	for _, tx := range txs {
		// 验证每一笔交易
		// 第二笔交易引用了第一笔交易的UTXO作为输入
		// 第一笔交易还没有被打包到区块中，所以添加到缓存列表中
		fmt.Printf("txHash : %v\n", tx.TxHash)
		if !bc.VerifyTransaction(tx, _txs) {
			log.Panicf("ERROR : tx [%x] verify failed!", tx)
		}
		_txs = append(_txs, tx)
	}
	// 生成新的区块
	block = NewBlock(block.Heigth+1, block.Hash, txs)
	// 持久化新区块
	bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if nil != b {
			err := b.Put(block.Hash, block.Serialize())
			if nil != err {
				log.Panicf("update the new block to db failed! %v\n", err)
			}
			_ = b.Put([]byte("l"), block.Hash) // 更新数据库中的最新哈希值
			bc.Tip = block.Hash
		}
		return nil
	})
}

// UnUTXOS 返回指定地址的UTXO(未花费输出)
func (bc *BlockChain) UnUTXOS(address string, txs []*transaction.Transaction) []*transaction.UTXO {
	var unUTXOS []*transaction.UTXO
	// 存储所有已花费的输出
	// key:每个input所引用的交易的哈希
	// value:output 索引列表
	spentTXOutputs := make(map[string][]int)
	// 查找缓存(未打包)的交易中是否有该地址的UTXO
	// 先查找输入
	for _, tx := range txs {
		if !tx.IsCoinbaseTransaction() {
			for _, in := range tx.Vins {
				// 验证公钥哈希
				publicKeyHash := Wallet.Base58Decode([]byte(address))
				// version+pubkeyhash+checksum
				ripemd160Hash := publicKeyHash[1 : len(publicKeyHash)-Wallet.AddressChecksumLen]
				if in.UnLockRipemd160Hash(ripemd160Hash) {
					//添加到已花费输入map中
					key := hex.EncodeToString(in.TxHash)
					spentTXOutputs[key] = append(spentTXOutputs[key], in.Vout)
				}
			}
		}
		fmt.Printf("all senntTxOutputs : %v\n", spentTXOutputs)
		// 查找缓存输出与数据库中的输出
	WorkCacheTx:
		for index, vout := range tx.Vouts {
			if vout.UnLockScriptPubkeyWithAddress(address) {
				if len(spentTXOutputs) != 0 {
					var isUtxoTx bool // 判断指定交易是否被其他交易引用
					for txHash, indeArray := range spentTXOutputs {
						txHashStr := hex.EncodeToString(tx.TxHash)
						if txHashStr == txHash { // 此处相等说明input引用了哈希txHashStr交易中的输出
							isUtxoTx = true
							var isSpentUTXO bool
							for _, voutIndex := range indeArray {
								if index == voutIndex {
									isSpentUTXO = true
									continue WorkCacheTx
								}
							}
							if isSpentUTXO == false {
								fmt.Printf("5 db : hash : %v\n", tx.TxHash)
								utxo := &transaction.UTXO{TxHash: tx.TxHash, Index: index, Output: vout}
								unUTXOS = append(unUTXOS, utxo)
							}
						}
					}
					if isUtxoTx == false {
						fmt.Printf("4 db : hash : %v\n", tx.TxHash)
						utxo := &transaction.UTXO{TxHash: tx.TxHash, Index: index, Output: vout}
						unUTXOS = append(unUTXOS, utxo)
					}
				} else {
					fmt.Printf("3 db : hash : %v\n", tx.TxHash)
					utxo := &transaction.UTXO{TxHash: tx.TxHash, Index: index, Output: vout}
					unUTXOS = append(unUTXOS, utxo)
				}
			}
		}
	}
	// 1. 先把所有已花费的输出全部取出
	blockIterator := bc.Iterator()
	for {
		block := blockIterator.Next() // 获取每一个区块信息
		for _, tx := range block.Txs {
			// 查找与address相关的所有交易
			if !tx.IsCoinbaseTransaction() {
				for _, in := range tx.Vins {
					// 判断地址
					// 验证公钥哈希
					publicKeyHash := Wallet.Base58Decode([]byte(address))
					// version+pubkeyhash+checksum
					ripemd160Hash := publicKeyHash[1 : len(publicKeyHash)-Wallet.AddressChecksumLen]
					if in.UnLockRipemd160Hash(ripemd160Hash) {
						key := hex.EncodeToString(in.TxHash)
						// 添加到已花费输出中
						spentTXOutputs[key] = append(spentTXOutputs[key], in.Vout)
					}
				}
			}
		}
		// 退出循环条件
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	// 通过判断查找UTXO
	blockIterator1 := bc.Iterator()
	for {
		block := blockIterator1.Next() // 获取每一个区块信息
		for _, tx := range block.Txs { // 遍历每个区块中的交易
			fmt.Printf("blockhash : %x, tx of block : %v\n", block.Hash, tx.TxHash)

		workDbTx:
			// 再查找输出
			for index, vout := range tx.Vouts {
				// 地址验证(检查输出是否属于传入地址)
				if vout.UnLockScriptPubkeyWithAddress(address) {
					// 判断output是否是一个未花费的输出
					// 判断已花费输出中是否为空
					if spentTXOutputs != nil {
						if len(spentTXOutputs) != 0 {
							var isSpentTXOutput bool
							for txHash, indexArray := range spentTXOutputs {
								for _, i := range indexArray {
									if txHash == hex.EncodeToString(tx.TxHash) && i == index {
										isSpentTXOutput = true
										continue workDbTx
									}
								}
							}
							// 只有遍历完整个spentTXOutputs都没有找到该VOUT，才能说明此vout是一个未花费输出
							if isSpentTXOutput == false {
								utxo := &transaction.UTXO{TxHash: tx.TxHash, Index: index, Output: vout}
								unUTXOS = append(unUTXOS, utxo)
							}
						} else {
							utxo := &transaction.UTXO{TxHash: tx.TxHash, Index: index, Output: vout}
							unUTXOS = append(unUTXOS, utxo)
						}
					}

				}
			}
		}

		// 退出循环条件
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}

	return unUTXOS
}

// 查询指定地址的余额
func (bc *BlockChain) getBalance(address string) int64 {
	utxos := bc.UnUTXOS(address, []*transaction.Transaction{})
	var amount int64
	for _, utxo := range utxos {
		amount += utxo.Output.Value
	}
	return amount
}

// FindSpendableUTXO 转账
// 通过查找可用的UTXO(遍历)，超过需要的资金即可中断
func (bc *BlockChain) FindSpendableUTXO(from string, amout int64, txs []*transaction.Transaction) (int64, map[string][]int) {
	// 查找出来的UTXO的值总和
	var value int64
	// 可用的UTXO
	spendableUTXO := make(map[string][]int)
	// 获取所有UTXO
	utxos := bc.UnUTXOS(from, txs)
	fmt.Printf("len:%d, all utxos : %v\n", len(utxos), utxos)
	// 遍历
	for _, utxo := range utxos {
		// utxo.Output.Value >= amout
		fmt.Printf("utxo.TxHash : %x\n", utxo.TxHash)
		value += utxo.Output.Value
		hash := hex.EncodeToString(utxo.TxHash)
		spendableUTXO[hash] = append(spendableUTXO[hash], utxo.Index)
		if value >= amout {
			break
		}
	}
	if value < amout {
		fmt.Printf("%s 余额不足\n", from)
		os.Exit(1)
	}
	return value, spendableUTXO
}

// VerifyTransaction 验证签名
func (bc *BlockChain) VerifyTransaction(tx *transaction.Transaction, txs []*transaction.Transaction) bool {
	// 查找指定交易的关联交易
	prevTxs := make(map[string]transaction.Transaction)
	for _, vin := range tx.Vins {
		prevTx := bc.FindTransaction(vin.TxHash, txs)
		prevTxs[hex.EncodeToString(prevTx.TxHash)] = prevTx
	}
	// tx.verify()
	return tx.Verify(prevTxs)
}

// FindTransaction 查找指定的交易
// ID代表input所引用的交易哈希
func (bc *BlockChain) FindTransaction(ID []byte, txs []*transaction.Transaction) transaction.Transaction {
	// 查找缓存中是否有符合条件的关联交易
	for _, tx := range txs {
		if bytes.Compare(tx.TxHash, ID) == 0 {

			return *tx
		}
	}

	bcit := bc.Iterator()
	for {
		block := bcit.Next()
		for _, tx := range block.Txs {
			// 判断交易哈希是否相等
			if bytes.Compare(tx.TxHash, ID) == 0 {
				return *tx
			}
		}

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
	return transaction.Transaction{}
}

// SignTransaction 交易签名(考虑到输入所引用的交易中，有可能有未打包的交易)
func (bc *BlockChain) SignTransaction(tx *transaction.Transaction, privateKey ecdsa.PrivateKey, txs []*transaction.Transaction) {
	// coinbase交易不需要签名
	if tx.IsCoinbaseTransaction() {
		return
	}
	// 处理input,查找交易tx的input所引用的vout所属的交易(用于确定交易的发送者)
	prevTXs := make(map[string]transaction.Transaction)
	for _, vin := range tx.Vins {
		// 查找所引用的每一个交易
		fmt.Printf("hash : [%x]\n", vin.TxHash)
		prevTX := bc.FindTransaction(vin.TxHash, txs)
		prevTXs[hex.EncodeToString(prevTX.TxHash)] = prevTX
	}
	// 实现签名函数
	tx.Sign(privateKey, prevTXs)
}

// FindUTXOMap 查找所有UTXO
func (bc *BlockChain) FindUTXOMap() map[string]*transaction.TXOutputs {
	bcit := bc.Iterator()

	// 存储已花费的UTXO的信息
	// key -> value
	// key : 代表指定交易哈希
	// value : 代表所有引用了该交易output的输入
	spentUTXOMap := make(map[string][]*transaction.TxInput)

	// UTXO集合
	// key->指定交易哈希
	// value->该交易中所有的未花费输出
	utxoMaps := make(map[string]*transaction.TXOutputs)

	for {
		block := bcit.Next()
		// 遍历每个区块中的交易
		for i := len(block.Txs) - 1; i >= 0; i-- {
			// 保存输出的列表
			txOutputs := &transaction.TXOutputs{UTXOS: []*transaction.UTXO{}}
			// 获取每一个交易
			tx := block.Txs[i]
			// 判断是否是一个coinbase交易
			if tx.IsCoinbaseTransaction() == false {
				fmt.Printf("tx-hash : %x\n", tx.TxHash)
				// 遍历交易中的每个输入
				for _, txInput := range tx.Vins {
					txHash := hex.EncodeToString(txInput.TxHash) // 当前输入所引用的输出所在的交易哈希
					spentUTXOMap[txHash] = append(spentUTXOMap[txHash], txInput)
				}
			} else {
				fmt.Printf("coinbase tx-hash : %x\n", tx.TxHash)
			}

			// 遍历输出
			txHash := hex.EncodeToString(tx.TxHash)
		WorkOutLoop:
			for index, out := range tx.Vouts {
				txInputs := spentUTXOMap[txHash] // 查找指定哈希的关联输入
				if len(txInputs) > 0 {
					isSpent := false //  判断output是否已经被花费
					for _, in := range txInputs {
						outPublicKey := out.Ripemd160Hash
						inPublicKey := in.PublicKey
						// 检查input和output中的用户是否是同一个
						if bytes.Compare(outPublicKey, Wallet.Ripemd160Hash(inPublicKey)) == 0 {
							if index == in.Vout {
								isSpent = true // 该输出已花费
								continue WorkOutLoop
							}
						}
					}

					if isSpent == false {
						// isSpent为假，说明该交易相关的输入中没输入能够与当前判断的out相匹配
						utxo := transaction.UTXO{TxHash: tx.TxHash, Index: index, Output: out}
						txOutputs.UTXOS = append(txOutputs.UTXOS, &utxo)
						//txOutputs.TxOutputs = append(txOutputs.TxOutputs, out)
					}
				} else {
					// 如果没有input，都是未花费的输出
					utxo := transaction.UTXO{TxHash: tx.TxHash, Index: index, Output: out}
					txOutputs.UTXOS = append(txOutputs.UTXOS, &utxo)
					//txOutputs.TxOutputs = append(txOutputs.TxOutputs, out)

				}
			}

			utxoMaps[txHash] = txOutputs // 该交易所有的UTXO
		}

		// 退出条件
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}

	return utxoMaps
}

// GetBlockHashes 获取区块哈希列表
func (bc *BlockChain) GetBlockHashes() [][]byte {
	var blockHashes [][]byte
	// 遍历区块然后进行存储
	blockIterator := bc.Iterator()
	for {
		block := blockIterator.Next()
		blockHashes = append(blockHashes, block.Hash)

		// 判断是否遍历到创世区块
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return blockHashes
	//return blockHashes
}

// GetBestHeight 获取指定区块链中的区块高度
func (bc *BlockChain) GetBestHeight() int64 {
	return bc.Iterator().Next().Heigth
}

// GetBlock 获取指定区块信息
func (bc *BlockChain) GetBlock(blockHash []byte) ([]byte, error) {
	var blockBytes []byte
	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if nil != b {
			blockBytes = b.Get(blockHash)
		}
		return nil
	})
	if nil != err {
		log.Panicf("view the table blocktable failed! %v\n", err)
		return nil, err
	}
	return blockBytes, nil
}

func (bc *BlockChain) AddBlock(block *Block) {
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if nil != b {
			// 判断传入的区块是否已存在
			blockBytes := b.Get(block.Hash)
			if nil != blockBytes { // 说明当前区块存在
				// 已存在，不需要同步
				return nil
			}
			err := b.Put(block.Hash, block.Serialize())
			if nil != err {
				log.Panicf("sync block failed! %v\n", err)
			}

			blockHash := b.Get([]byte("l"))
			latestBlock := b.Get(blockHash)
			blockDb := DeserializeBlock(latestBlock)

			if blockDb.Heigth < block.Heigth {
				_ = b.Put([]byte("l"), block.Hash)
				bc.Tip = block.Hash
			}
		}

		return nil
	})
	if nil != err {
		log.Panicf("add block failed! %v\n", err)
	}
	fmt.Println("the new block added!")
}
