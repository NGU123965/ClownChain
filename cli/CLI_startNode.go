package cli

import "ClownChain/server"

// 启动网络服务
func (cli *CLI) startNode(nodeID string) {
	server.StartServer(nodeID)
}
