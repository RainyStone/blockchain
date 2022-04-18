package main

import (
	"fmt"
	"log"
)

//获取地址 address 对应的可用于交易的金额
func (cli *CLI) getBalance(address string, nodeID string) {
	if !ValidateAddress(address) {
		log.Panic("钱包地址 address 错误!!!!")
	}

	bc := NewBlockChain(nodeID) //根据地址创建
	UTXOSet := UTXOSet{bc} //创建UTXO
	defer bc.db.Close() //延迟关闭数据库

	balance := 0
	pubkeyhash := Base58Decode([]byte(address)) //提取公钥
	pubkeyhash = pubkeyhash[1:len(pubkeyhash)-4]

	UTXOs := UTXOSet.FindUTXO(pubkeyhash) //根据公钥查询

	for _, out := range UTXOs {
		balance += out.Value //金额叠加
	}

	fmt.Printf("钱包地址：%s，查询出的可用金额：%d", address, balance)
}