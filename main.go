package main

import (
	_ "fmt"
	_ "strconv"
)

func main(){
	// fmt.Println("简易区块链演示开始.... ")
	// blockChain := NewBlockChain()
	// blockChain.AddBlock("x 支付 20元给 z")
	// blockChain.AddBlock("b 支付 40元给 a")
	// blockChain.AddBlock("m 支付 30元给 n")

	// for _,block := range blockChain.blocks{
	// 	fmt.Printf("上一块哈希：%x\n", block.PrevBlockHash)
	// 	fmt.Printf("当前块数据：%s\n", block.Data)
	// 	fmt.Printf("当前块哈希：%x\n", block.Hash)
	// 	pow := NewProofOfWork(block) //校验工作量
	// 	fmt.Printf("pow ： %s\n", strconv.FormatBool(pow.Validate()))
	// 	fmt.Println()
	// }

	blockChain := NewBlockChain() //创建区块链
	defer blockChain.db.Close() //延迟关闭数据库
	cli := CLI{blockChain} //创建命令行
	cli.Run() //开启 
}