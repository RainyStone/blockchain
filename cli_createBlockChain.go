package main

import (
	"fmt"
	"log"
)

//创建区块链
func (cli *CLI) createBlockChain(address string, nodeID string) {
	if !ValidateAddress(address) {
		log.Panic("钱包地址 address 错误!!!!")
	}
	bc := CreateBlockChain(address, nodeID) //创建一个区块链
	defer bc.db.Close() //延迟关闭数据库

	UTXOSet := UTXOSet{bc}
	UTXOSet.Reindex()

	fmt.Println("blockchain 创建成功")
}