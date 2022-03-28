package main

import "crypto/sha256"

//默克尔树
type MerkleTree struct {
	RootNode *MerkleTreeNode
}

//节点
type MerkleTreeNode struct {
	Left *MerkleTreeNode
	Right *MerkleTreeNode
	data []byte
}

//创建一棵默克尔树
func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleTreeNode

	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _,dataum := range data {
		node := NewMerkleTreeNode(nil, nil, dataum)
		nodes = append(nodes, *node) //节点追加
	}

	for i:=0; i<len(data)/2; i++ {
		var newlevel []MerkleTreeNode
		for j:=0; j<len(nodes); j+=2 {
			node := NewMerkleTreeNode(&nodes[j], &nodes[j+1], nil)
			newlevel = append(newlevel, *node)
		}
		nodes = newlevel
	}
	mTree := MerkleTree{&nodes[0]}
	return &mTree
}

//创建一个默克尔树节点
func NewMerkleTreeNode(left, right *MerkleTreeNode, data []byte) *MerkleTreeNode {
	mNode := MerkleTreeNode{}
	if left==nil && right==nil { //相当于叶子节点
		hash := sha256.Sum256(data) //计算哈希
		mNode.data = hash[:]
	}else{//相当于非叶子节点
		preHashes := append(left.data, right.data...)
		hash := sha256.Sum256(preHashes)
		mNode.data = hash[:]
	}
	mNode.Left = left
	mNode.Right = right
	return &mNode //返回树节点
}