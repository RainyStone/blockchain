package main

import (
	"fmt"
	"strconv"
)

//展示区块链数据
func (cli *CLI) showBlockChain(nodeID string) {
	bc := NewBlockChain(nodeID )
	defer bc.db.Close()

	bci := bc.Iterator() //迭代器
	for {
		block := bci.next()
		fmt.Printf("上一块哈希：%x\n", block.PrevBlockHash)
		fmt.Printf("当前区块哈希：%x\n", block.Hash)
		pow := NewProofOfWork(block) //工作量证明
		fmt.Printf("pow：%s\n", strconv.FormatBool(pow.Validate()))
		for i,tx := range block.Transactions {
			fmt.Printf("\t交易 %d ：%v\n", i, tx)
		}
		fmt.Println()
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}