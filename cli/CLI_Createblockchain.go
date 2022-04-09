package cli

import (
	"ClownChain/blockchain"
	"ClownChain/transaction"
)

// 创建区块链
func (cli *CLI) createBlockchainWithGenesis(address string, nodeID string) {
	blockchain := blockchain.CreateBlockChainWithGenesisBlock(address, nodeID)
	defer blockchain.DB.Close()

	// 设置utxoSet操作
	utxoSet := &transaction.UTXOSet{BlockChain: blockchain}
	utxoSet.ResetUTXOSet() // 重置数据库，主要是更新UTXO表
}
