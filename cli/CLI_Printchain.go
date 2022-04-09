package cli

import (
	"ClownChain/blockchain"
	"fmt"
	"os"
)

// 输出区块链信息
func (cli *CLI) printchain(nodeID string) {
	if blockchain.DbExists(nodeID) == false {
		fmt.Println("数据库不存在...")
		os.Exit(1)
	}
	blockchain := blockchain.BlockchainObject(nodeID) // 获取区块链对象
	defer blockchain.DB.Close()
	blockchain.PrintChain()
}
