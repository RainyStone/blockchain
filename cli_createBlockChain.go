package main

import (
	"fmt"
	"log"
)

//创建区块链
func (cli *CLI) createBlockChain(address string) {
	if !ValidateAddress(address) {
		log.Panic("钱包地址 address 错误!!!!")
	}
	bc := CreateBlockChain(address) //创建一个区块链
	bc.db.Close()
	fmt.Println("blockchain 创建成功")
}