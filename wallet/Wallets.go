package Wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// 钱包集合的文件
const walletFile = "Wallets_%s.dat" // 存储钱包集合的文件
// Wallets 钱包的集合结构
type Wallets struct {
	// key:string->地址
	// value:钱包结构
	Wallets map[string]*Wallet
}

// NewWallets 初始化一个钱包集合
func NewWallets(nodeID string) (*Wallets, error) {
	walletFile := fmt.Sprintf(walletFile, nodeID)
	//fmt.Printf("walletFile : %v\n", walletFile)
	// 1. 判断文件是否存在
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		wallets := &Wallets{}
		wallets.Wallets = make(map[string]*Wallet)
		return wallets, err
	}
	// 2. 文件存在， 读取内容
	fileContent, err := ioutil.ReadFile(walletFile)
	if nil != err {
		log.Panicf("get file content failed! %v\n", err)
	}
	var wallets Wallets
	// register适用于需要解析的参数中包含interface
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if nil != err {
		log.Panicf("decode filecontent failed %v\n", err)
	}
	return &wallets, nil
}

// CreateWallet 创建新的钱包,并且将其添加到集合
func (wallets *Wallets) CreateWallet(nodeID string) {
	wallet := NewWallet() // 新建钱包对象
	wallets.Wallets[string(wallet.GetAddress())] = wallet
	// 把钱包存储到文件中
	wallets.SaveWallets(nodeID)
}

// SaveWallets 持久化钱包信息（写入文件）
func (wallets *Wallets) SaveWallets(nodeID string) {
	var content bytes.Buffer
	// 注册
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	// 序列化钱包数据
	err := encoder.Encode(&wallets)
	if nil != err {
		log.Panicf("encode the struct of wallets failed! %v\n", err)
	}
	// 清空文件再云存储(此处只保存了一条数据，但该条数据会存储到目前为止所有地址的集合)
	walletFile := fmt.Sprintf(walletFile, nodeID)
	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if nil != err {
		log.Panicf("write the content of wallets to file [%s] failed! %v\n", walletFile, err)
	}
}
