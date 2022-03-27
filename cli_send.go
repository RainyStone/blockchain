package main

import (
	"fmt"
	"log"
)

//转账
func (cli *CLI) send(from, to string, amount int) {
	if !ValidateAddress(from) {
		log.Panic("钱包地址 from 错误!!!!")
	}

	if !ValidateAddress(to) {
		log.Panic("钱包地址 to 错误!!!!")
	}

	bc := NewBlockChain(from)
	defer bc.db.Close()
	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*Transaction{tx}) //挖矿记账
	fmt.Printf("交易成功，从 %s 转账 %d 给 %s", from, amount, to)
}