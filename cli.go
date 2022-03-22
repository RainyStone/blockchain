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

//创建区块链
func (cli *CLI) createBlockChain(address string) {
	bc := createBlockChain(address) //创建区块链
	bc.db.Close()
	fmt.Println("创建成功，创建者地址：", address)
}

//获取地址 address 对应的可用于交易的金额
func (cli *CLI) getBalance (address string) {
	bc := NewBlockChain(address) //根据地址创建
	defer bc.db.Close() //延迟关闭数据库

	balance := 0
	UTXOs := bc.FindUTXO(address)
	for _, out := range UTXOs {
		balance += out.Value //金额叠加
	}

	fmt.Printf("钱包地址：%s，查询出的可用金额：%d", address, balance)
}















//用法
func (cli *CLI) printUsage() {
	fmt.Println("用法如下")
	fmt.Println("getbalance -address 钱包地址 根据地址查询金额")
	fmt.Println("createblockchain -address 钱包地址 根据地址创建区块链")
	fmt.Println("send -from From -to To -amount Amount 转账")
	fmt.Println("showchain 显示区块链")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage() //显示用法
		os.Exit(1)
	}
}

func (cli *CLI) showBlockChain() {
	bc := NewBlockChain("") //找出所有数据库????但是 NewBlockChain 函数中貌似没有使用到 address 参数
	defer bc.db.Close()

	bci := bc.Iterator() //迭代器
	for {
		block := bci.next()
		fmt.Printf("上一块哈希：%x\n", block.PrevBlockHash)
		fmt.Printf("当前区块哈希：%x\n", block.Hash)
		pow := NewProofOfWork(block) //工作量证明
		fmt.Printf("pow：%s\n", strconv.FormatBool(pow.Validate()))

		if len(block.PrevBlockHash) == 0{
			break
		}
	}
}

func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockChain(from)
	defer bc.db.Close()
	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*Transaction{tx}) //挖矿记账
	fmt.Printf("交易成功，从 %s 转账 %d 给 %s", from, amount, to)
}

func (cli *CLI) Run() {
	cli.validateArgs() //校验参数

	//处理命令行参数
	getbalancecmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createblockchaincmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendcmd := flag.NewFlagSet("send", flag.ExitOnError)
	showchaincmd := flag.NewFlagSet("showchain", flag.ExitOnError)

	getbalanceaddress := getbalancecmd.String("address", "", "查询地址")
	createblockchainaddress := createblockchaincmd.String("address", "", "查询地址")
	sendfrom := sendcmd.String("from", "", "付款地址，谁给的")
	sendto := sendcmd.String("to", "", "收款地址，给谁的")
	sendamount := sendcmd.Int("amount", 0, "转账金额")

	switch os.Args[1] {
	case "getbalance" :
		err := getbalancecmd.Parse(os.Args[2:]) //解析参数
		if err != nil {
			log.Panic(err) //处理错误
		}
	case "createblockchain":
		err := createblockchaincmd.Parse(os.Args[2:]) //解析参数
		if err != nil {
			log.Panic(err) //处理错误
		}
	case "send":
		err := sendcmd.Parse(os.Args[2:]) //解析参数
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

	if getbalancecmd.Parsed() {
		if *getbalanceaddress == "" {
			getbalancecmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getbalanceaddress) //获取可用金额
	}

	if createblockchaincmd.Parsed() {
		if *createblockchainaddress == "" {
			createblockchaincmd.Usage()
			os.Exit(1)
		}
		cli.createBlockChain(*createblockchainaddress) //创建区块链
	}

	if sendcmd.Parsed() {
		if *sendfrom == "" || *sendto == "" || *sendamount <= 0 {
			sendcmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendfrom, *sendto, *sendamount)
	}

	if showchaincmd.Parsed() {
		cli.showBlockChain() //显示区块链
	}

}