package cli

import (
	Wallet "ClownChain/wallet"
	"fmt"
)

func (cli *CLI) getAddressLists(nodeID string) {
	fmt.Println("打印所有钱包地址...")
	wallets, _ := Wallet.NewWallets(nodeID)
	for address := range wallets.Wallets {
		fmt.Printf("address : [%s]\n", address)
	}
}
