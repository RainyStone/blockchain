package main

import (
	"fmt"
	"log"
)

func (cli *CLI) startNode(nodeID string, minerAddress string) {
	fmt.Printf("----启动一个节点：%s\n", nodeID)
	if len(minerAddress) > 0{
		if ValidateAddress(minerAddress) {
			fmt.Printf("正在挖矿的地址：%s\n", minerAddress)
		}else{
			log.Panic("----挖矿地址错误!!!!")
		}
	}
	StartServer(nodeID, minerAddress) //启动服务器
}