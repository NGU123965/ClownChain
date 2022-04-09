package server

import (
	"ClownChain/blockchain"
	"ClownChain/utils"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

// 数据同步的函数
func sendVersion(toAddress string, bc *blockchain.BlockChain) {
	fmt.Println("sendVersion")
	// 在比特币中，消息是底层的比特序列，前12个字节指定命令名(verion)
	// 后面的字节包含的是gob编码过的消息结构
	bestHeight := bc.GetBestHeight()
	data := utils.GobEncode(Version{utils.NODE_VERSION, bestHeight, nodeAddress})
	request := append(utils.CommandToBytes(utils.CMD_VERSION), data...)
	sendMessage(toAddress, request)
}

//"客户端(节点)"向服务器发送请求
func sendMessage(to string, msg []byte) {
	fmt.Println("向服务器发送请求...")
	conn, err := net.Dial(utils.PROTOCOL, to)
	if nil != err {
		log.Panicf("connect to server failed! %v", conn)
	}
	defer conn.Close()
	// 要发送的数据添加到请求中
	io.Copy(conn, bytes.NewReader(msg))
	if nil != err {
		log.Panicf("add the data failed! %v\n", err)
	}
}

// 向其他节点展示区块信息
func sendShowBlocks(toAddress string, kind string, hashes [][]byte) {
	data := utils.GobEncode(ShowsBlocks{Hashes: hashes, AddrFrom: nodeAddress, Type: kind})
	request := append(utils.CommandToBytes(utils.CMD_SHOWBLOCKS), data...)
	sendMessage(toAddress, request)
}

// 从指定节点同步数据
func sendGetBlocks(toAddress string) {
	data := utils.GobEncode(GetBlocks{AddrFrom: nodeAddress})
	request := append(utils.CommandToBytes(utils.CMD_GETBLOCKS), data...)
	sendMessage(toAddress, request)
}

// 向其他人展示交易或者区块信息
func sendGetData(toAddress string, kind string, hash []byte) {
	data := utils.GobEncode(GetData{AddrFrom: nodeAddress, ID: hash, Type: kind})
	request := append(utils.CommandToBytes(utils.CMD_GETDATA), data...)
	sendMessage(toAddress, request)
}

// 发送区块信息
func sendBlock(toAddress string, block []byte) {
	data := utils.GobEncode(BlockData{nodeAddress, block})
	request := append(utils.CommandToBytes(utils.CMD_BLOCK), data...)
	sendMessage(toAddress, request)
}
