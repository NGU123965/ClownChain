package cli

import (
	"ClownChain/blockchain"
	"ClownChain/transaction"
	"fmt"
	"os"
)

// 发送交易
func (cli *CLI) send(from []string, to []string, amount []string, nodeID string) {
	// 检测数据库
	if blockchain.DbExists(nodeID) == false {
		fmt.Println("数据库不存在...")
		os.Exit(1)
	}
	blockchain := blockchain.BlockchainObject(nodeID) // 获取区块链对象
	defer blockchain.DB.Close()
	blockchain.MineNewBlock(from, to, amount, nodeID)
	utxoSet := &transaction.UTXOSet{BlockChain: blockchain}
	utxoSet.Update()
}
