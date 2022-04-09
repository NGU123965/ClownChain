package server

// ShowsBlocks 向其他节点展示当前节点有哪些区块
type ShowsBlocks struct {
	Hashes   [][]byte // 哈希
	AddrFrom string   // 自己这个节点的地址
	Type     string   // 数据类型（交易或区块），这里没有交易信息的展示
}
