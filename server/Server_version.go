package server

// Version 代表当前的区块版本信息(决定是否需要进行同步 )
type Version struct {
	Version    int    // 版本
	BestHeigth int64  // 当前节点区块的高度
	AddrFrom   string // 当前节点的地址
}
