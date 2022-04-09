package cli

import (
	Wallet "ClownChain/wallet"
	"fmt"
)

// CreateWallets 创建钱包集合
func (cli *CLI) CreateWallets(nodeID string) {
	fmt.Printf("nodeID : %v\n", nodeID)
	wallets, _ := Wallet.NewWallets(nodeID) // 创建一个集合对象
	wallets.CreateWallet(nodeID)
	fmt.Printf("wallets : %v\n", wallets)
}

// Wallets_3000.dat
