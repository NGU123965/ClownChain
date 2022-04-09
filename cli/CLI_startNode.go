package cli

import "ClownChain/server"

// 实现启动服务的功能

func (cli *CLI) startNode(nodeID string) {
	server.StartServer(nodeID)
}
