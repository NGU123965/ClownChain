package server

import (
	"ClownChain/blockchain"
	"ClownChain/transaction"
	"ClownChain/utils"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
)

// 处理相关请求

//1.version：验证当前节点的最后一个区块是否是最新的区块（区块链是否最新）
func handleVersion(request []byte, bc *blockchain.BlockChain) {
	fmt.Println("handleVersion")
	var buff bytes.Buffer
	var data Version
	//解析request 获取数据
	dataBytes := request[utils.CMDLENGTH:]
	buff.Write(dataBytes)
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&data)
	if nil != err {
		log.Panicf("decode the version cmd failed! %v\n", err)
	}
	// 获取区块高度
	bestHeight := bc.GetBestHeight()
	versionHeight := data.BestHeigth
	fmt.Printf("bestHeight:%v\n", bestHeight)
	fmt.Printf("versionHeight:%v\n", versionHeight)
	// 如果当前节点的区块高度大于versionHeight，将当前节点的版本信息发送给请求节点
	if bestHeight > versionHeight {
		sendVersion(data.AddrFrom, bc)
	} else if bestHeight < versionHeight {
		// 同步，如果当前节点的区块高度小于versionHeight，向请求节点发送同步请求
		sendGetBlocks(data.AddrFrom)
	}
}

//2.getBlocks：从最长的链上获取区块
func handleGetBlocks(request []byte, bc *blockchain.BlockChain) {
	fmt.Println("handleGetBlocks")
	var buff bytes.Buffer
	var data GetBlocks
	//解析request 获取数据
	dataBytes := request[utils.CMDLENGTH:]
	buff.Write(dataBytes)
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&data)
	if nil != err {
		log.Panicf("decode the getBlocks cmd failed! %v\n", err)
	}
	blocks := bc.GetBlockHashes()
	// showblocks
	sendShowBlocks(data.AddrFrom, utils.BLOCK_TYPE, blocks)
}

//3.showBlocks：向其他节点展示当前节点有哪些区块
func handleShowBlocks(request []byte) {
	fmt.Println("handleShowBlocks")
	var buff bytes.Buffer
	var data ShowsBlocks
	//解析request 获取数据
	dataBytes := request[utils.CMDLENGTH:]
	buff.Write(dataBytes)
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&data)
	if nil != err {
		log.Panicf("decode the ShowBlocks cmd failed! %v\n", err)
	}
	// sendGetData
	blockHash := data.Hashes[0]
	sendGetData(data.AddrFrom, utils.BLOCK_TYPE, blockHash)

}

//4.getData：请求一个指定的区块
func handleGetData(request []byte, bc *blockchain.BlockChain) {
	fmt.Println("handleGetData")
	var buff bytes.Buffer
	var data GetData
	//解析request 获取数据
	dataBytes := request[utils.CMDLENGTH:]
	buff.Write(dataBytes)
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&data)
	if nil != err {
		log.Panicf("decode the GetData cmd failed! %v\n", err)
	}
	// 获取指定ID的区块信息 get-block(id)
	block, err := bc.GetBlock(data.ID)
	if nil != err {
		return
	}
	sendBlock(data.AddrFrom, block)
}

//5. block：接收到新区块时，进行处理（存储）
func handleBlock(request []byte, bc *blockchain.BlockChain) {
	fmt.Println("handleBlock")
	var buff bytes.Buffer
	var data BlockData
	//解析request 获取数据
	dataBytes := request[utils.CMDLENGTH:]
	buff.Write(dataBytes)
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&data)
	if nil != err {
		log.Panicf("decode the Block cmd failed! %v\n", err)
	}
	blockBytes := data.Block
	block := blockchain.DeserializeBlock(blockBytes)
	// 添加区块到区块链
	bc.AddBlock(block)
	// 更新UTXO
	utxoSet := &transaction.UTXOSet{BlockChain: bc}
	utxoSet.Update()
}
