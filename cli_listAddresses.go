package main

import (
	"fmt"
	"log"
)

//提取出所有钱包地址
func (cli *CLI) listAddresses() {
	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for i,addr := range addresses {
		fmt.Printf("地址 %d ：%s\n", i, addr)
	}
}