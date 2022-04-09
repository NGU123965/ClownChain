package server

import (
	"ClownChain/blockchain"
	"ClownChain/utils"
	"fmt"
	"io/ioutil"
	"log"
	"net"
)

// 3000作为主节点地址
var knowNodes = []string{"localhost:3000"}

// 服务处理文件
var nodeAddress string // 节点地址
// StartServer 启动服务器
func StartServer(nodeID string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID) // 服务节点地址
	fmt.Printf("服务节点 [%s] 启动...\n", nodeAddress)
	// 监听节点
	listen, err := net.Listen(utils.PROTOCOL, nodeAddress)
	if nil != err {
		log.Panicf("listen address of %s failed! %v\n", nodeAddress, err)
	}
	defer listen.Close()
	bc := blockchain.BlockchainObject(nodeID)
	if nodeAddress != knowNodes[0] {
		// 非主节点，向主节点发送请求同步数据
		// sendVersion()
		//	sendMessage(knowNodes[0], nodeAddress)
		sendVersion(knowNodes[0], bc)
	}

	// 主节点接收请求
	for {
		conn, err := listen.Accept()
		if nil != err {
			log.Panicf("connect to node failed! %v\n", err)
		}

		// 分出一个单独的goroutine来对请求进行处理
		go handleConnection(conn, bc)

	}
}

// 处理请求函数
func handleConnection(conn net.Conn, bc *blockchain.BlockChain) {
	request, err := ioutil.ReadAll(conn)
	if nil != err {
		log.Panicf("Receive a Message failed! %v\n", err)
	}
	cmd := utils.BytesToCommand(request[:utils.CMDLENGTH])
	fmt.Printf("Receive a Message : %s\n", cmd)

	switch cmd {
	case utils.CMD_VERSION:
		handleVersion(request, bc)
	case utils.CMD_GETDATA:
		handleGetData(request, bc)
	case utils.CMD_BLOCK:
		handleBlock(request, bc)
	case utils.CMD_SHOWBLOCKS:
		handleShowBlocks(request)
	case utils.CMD_GETBLOCKS:
		handleGetBlocks(request, bc)
	default:
		fmt.Println("Unknown command")
	}
	conn.Close()
}
