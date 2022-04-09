package server

type BlockData struct {
	AddrFrom string // 节点地址
	Block    []byte // 序列化的区块结构
}
