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

	bc := NewBlockChain()

	UTXOSet := UTXOSet{bc}

	defer bc.db.Close()


	tx := NewUTXOTransaction(from, to, amount, &UTXOSet)
	cbTx := NewCoinBaseTX(from, "") //挖矿交易
	txs := []*Transaction{cbTx, tx} //交易

	newBlock := bc.MineBlock(txs)
	UTXOSet.Update(newBlock)

	fmt.Printf("交易成功，从 %s 转账 %d 给 %s", from, amount, to)
}