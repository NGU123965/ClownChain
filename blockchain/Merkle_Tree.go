package blockchain

import "crypto/sha256"

// MerkleTree merkle树的实现
type MerkleTree struct {
	// 根节点
	RootNode *MerkleNode
}

// MerkleNode merkle节点
type MerkleNode struct {
	// 左子节点
	Left *MerkleNode
	// 右子节点
	Right *MerkleNode
	// 数据
	Data []byte
}

// NewMerkleNode 创建节点
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := &MerkleNode{}
	if left == nil && right == nil {
		// 叶子节点
		hash := sha256.Sum256(data)
		node.Data = hash[:]
	} else {
		// 非叶子节点，保存左子节点哈希和右子节点哈希合到一起之后的再次哈希值
		prevHashes := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHashes)
		node.Data = hash[:]
	}
	node.Left = left
	node.Right = right
	return node
}

// NewMerkleTree 创建Merkle树
// 当前区块中的所有交易
func NewMerkleTree(datas [][]byte) *MerkleTree {
	var nodes []MerkleNode // 保存节点
	// 判断交易数据条数，如果是奇数条，则把最后一条拷贝一份
	if len(datas)%2 != 0 {
		datas = append(datas, datas[len(datas)-1])
	}
	// 创建叶子节点
	for _, data := range datas {
		node := NewMerkleNode(nil, nil, data)
		nodes = append(nodes, *node)
	}
	// 创建非叶子节点(上级节点)
	for i := 0; i < len(datas)/2; i++ {
		var newNodes []MerkleNode // 父节点列表
		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			newNodes = append(newNodes, *node)
		}
		if len(newNodes)%2 != 0 {
			newNodes = append(newNodes, newNodes[len(newNodes)-1])
		}
		nodes = newNodes // 最终，nodes列表中只保存根节点(哈希)
	}
	mtree := MerkleTree{&nodes[0]}
	return &mtree
}
