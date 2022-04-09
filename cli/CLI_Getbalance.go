package cli

import (
	"ClownChain/blockchain"
	"ClownChain/transaction"
	"fmt"
)

// 查询余额
func (cli *CLI) getBalance(from string, nodeID string) {
	// 获取指定地址的余额
	blockchain := blockchain.BlockchainObject(nodeID)
	defer blockchain.DB.Close()
	utxoSet := &transaction.UTXOSet{BlockChain: blockchain}
	amount := utxoSet.GetBalance(from)
	fmt.Printf("\t地址: %s的余额为:%d\n", from, amount)
}
