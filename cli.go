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
	fmt.Println("createwallet ----创建钱包")
	fmt.Println("listaddresses ----显示所有钱包地址(账户)")
	fmt.Println("getbalance -address 钱包地址 ----根据地址查询金额")
	fmt.Println("createblockchain -address 钱包地址 ----根据地址创建区块链")
	fmt.Println("send -from From -to To -amount Amount -mine 是否立刻挖矿 ----转账")
	fmt.Println("showchain ----显示区块链")
	fmt.Println("reindexutxo ----重建UTXO(未花费输出)索引，即刷新持久化的UTXO数据")
	fmt.Println("startnode -miner Addr ----启动一个节点并设置挖矿地址")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage() //显示用法
		os.Exit(1)
	}
}

func (cli *CLI) Run() {
	cli.validateArgs() //校验参数

	// nodeID := os.Getenv("NODE_ID")
	// if nodeID == "" {
	// 	fmt.Printf("----必须设置运行端口号\n")
	// 	os.Exit(1)
	// }
	//由于windows系统环境变量设置不便，这里直接设置 nodeID
	// nodeID := "3000"
	// nodeID := "3001"
	nodeID := "3002"

	//处理命令行参数
	createwalletcmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listaddressescmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	getbalancecmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createblockchaincmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendcmd := flag.NewFlagSet("send", flag.ExitOnError)
	showchaincmd := flag.NewFlagSet("showchain", flag.ExitOnError)
	reindexutxocmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	startnodecmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	getbalanceaddress := getbalancecmd.String("address", "", "查询地址")
	createblockchainaddress := createblockchaincmd.String("address", "", "查询地址")
	sendfrom := sendcmd.String("from", "", "付款地址，谁给的")
	sendto := sendcmd.String("to", "", "收款地址，给谁的")
	sendamount := sendcmd.Int("amount", 0, "转账金额")
	sendmine := sendcmd.Bool("mine", false, "是否立刻挖矿")
	startnodeminer := startnodecmd.String("miner", "", "是否启动挖矿功能")

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
	case "reindexutxo":
		err := reindexutxocmd.Parse(os.Args[2:]) //解析参数
		if err != nil {
			log.Panic(err) //处理错误
		}
	case "startnode":
		err := startnodecmd.Parse(os.Args[2:]) //解析参数
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
		cli.getBalance(*getbalanceaddress, nodeID) //获取可用金额
	}

	if createblockchaincmd.Parsed() {
		if *createblockchainaddress == "" {
			createblockchaincmd.Usage()
			os.Exit(1)
		}
		cli.createBlockChain(*createblockchainaddress, nodeID) //创建区块链
	}

	if sendcmd.Parsed() {
		if *sendfrom == "" || *sendto == "" || *sendamount <= 0 {
			sendcmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendfrom, *sendto, *sendamount, nodeID, *sendmine)
	}

	if showchaincmd.Parsed() {
		cli.showBlockChain(nodeID) //显示区块链
	}

	if createwalletcmd.Parsed() {
		cli.createWallet(nodeID) //创建钱包
	}

	if listaddressescmd.Parsed() {
		cli.listAddresses(nodeID) //显示所有钱包地址
	}

	if reindexutxocmd.Parsed() {
		cli.reindexUTXO(nodeID) //重建索引
	}

	if startnodecmd.Parsed() {
		cli.startNode(nodeID, *startnodeminer) //启动服务器
	}

}