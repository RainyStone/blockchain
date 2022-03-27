package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

//命令行接口
type CLI struct {
	blockChain *BlockChain
}

//用法
func (cli *CLI) printUsage() {
	fmt.Println("用法如下：")
	fmt.Println("createwallet 创建钱包")
	fmt.Println("listaddresses 显示所有钱包地址(账户)")
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

func (cli *CLI) Run() {
	cli.validateArgs() //校验参数

	//处理命令行参数
	createwalletcmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listaddressescmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
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
	case "createwallet":
		err := createwalletcmd.Parse(os.Args[2:]) //解析参数
		if err != nil {
			log.Panic(err) //处理错误
		}
	case "listaddresses":
		err := listaddressescmd.Parse(os.Args[2:]) //解析参数
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

	if createwalletcmd.Parsed() {
		cli.createWallet() //创建钱包
	}

	if listaddressescmd.Parsed() {
		cli.listAddresses() //显示所有钱包地址
	}

}