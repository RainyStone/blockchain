package main

import (
	"fmt"
	"log"
)

//转账
func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !ValidateAddress(from) {
		log.Panic("钱包地址 from 错误!!!!")
	}

	if !ValidateAddress(to) {
		log.Panic("钱包地址 to 错误!!!!")
	}

	bc := NewBlockChain(nodeID)

	UTXOSet := UTXOSet{bc}

	defer bc.db.Close()

	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from) //查找钱包

	tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)

	//mineNow，是否立刻挖矿
	if mineNow {
		cbTx := NewCoinBaseTX(from, "") //挖矿交易
	    txs := []*Transaction{cbTx, tx} //交易
	    newBlock := bc.MineBlock(txs)
	    UTXOSet.Update(newBlock)
	}else{
		sendTx(knowNodes[0], tx) //发送交易等待确认
	}

	fmt.Printf("交易成功，从 %s 转账 %d 给 %s", from, amount, to)
}