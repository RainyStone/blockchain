package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

//命令行接口
type CLI struct {
	blockChain *BlockChain
}

//用法
func (cli *CLI) printUsage() {
	fmt.Println("用法如下")
	fmt.Println("addblock 向区块链增加块")
	fmt.Println("showchain 显示区块链")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage() //显示用法
		os.Exit(1)
	}
}

func (cli *CLI) addBlock(data string) {
	cli.blockChain.AddBlock(data) //增加区块
	fmt.Printf("区块增加成功，区块数据：%s\n", data)
}

func (cli *CLI) showBlockChain() {
	bci := cli.blockChain.Iterator() //创建循环迭代器
	for {
		block := bci.next() //取得下一个区块
		fmt.Printf("上一块哈希：%x\n", block.PrevBlockHash)
		fmt.Printf("当前块数据：%s\n", block.Data)
		fmt.Printf("当前块哈希：%x\n", block.Hash)
		pow := NewProofOfWork(block) //校验工作量
		fmt.Printf("pow ： %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		//遇到创世区块，跳出循环
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) Run() {
	cli.validateArgs() //校验参数

	//处理命令行参数
	addblockcmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	showchaincmd := flag.NewFlagSet("showchain", flag.ExitOnError)

	addBlockData := addblockcmd.String("data", "", "Block data")

	switch os.Args[1] {
	case "addblock" :
		err := addblockcmd.Parse(os.Args[2:]) //解析参数
		if err != nil {
			log.Panic(err) //处理错误
		}
	case "showchain":
		err := showchaincmd.Parse(os.Args[2:]) //解析参数
		if err != nil {
			log.Panic(err) //处理错误
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if addblockcmd.Parsed() {
		if *addBlockData == "" {
			addblockcmd.Usage()
			os.Exit(1)
		}else{
			cli.addBlock(*addBlockData) //增加区块
		}
	}

	if showchaincmd.Parsed() {
		cli.showBlockChain() //显示区块链
	}

}