package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

const protocol = "tcp"   //安全保障的网络协议
const nodeVersion = 1    //版本
const commandlength = 12 //命令长度

var nodeAddress string                     //节点地址，应该是本地自己节点的地址
var miningAddress string                   //挖矿地址，启动服务器时设置
var knowNodes = []string{"localhost:3000"} //已知的节点，第一个节点默认是中心节点
var blocksInTransit = [][]byte{}           // transit : 运输、运送
var mempool = make(map[string]Transaction) //内存池，存储交易

type addr struct {
	Addrlist []string //节点地址集合
}

type block struct {
	AddrFrom string //来源地址
	Block    []byte //块
}

type getblocks struct {
	AddrFrom string //来源地址
}

type getdata struct {
	AddrFrom string //来源
	Type     string //类型
	ID       []byte
}

type inv struct {
	AddrFrom string //来源
	Type     string //类型
	Items    [][]byte
}

type tx struct {
	AddrFrom    string
	Transaction []byte
}

type verzion struct {
	Version    int //版本参数
	BestHeight int
	AddrFrom   string
}

//启动服务器
func StartServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddress
	In,err := net.Listen(protocol, nodeAddress) //监听
	if err!= nil {
		log.Panic(err)
	}

	defer In.Close()
	bc := NewBlockChain(nodeID)
	if nodeAddress != knowNodes[0] {
		sendVersion(knowNodes[0], bc)
	}
	for {
		conn,err := In.Accept() //接收请求
		if err!=nil {
			log.Panic(err)
		}
		go handleConnection(conn, bc) //异步处理
	}
}

//字节到命令
func bytesToCommand(bytes []byte) string {
	var command []byte
	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b) //增加命令的字节
		}
	}
	return fmt.Sprintf("%s", command)
}

//命令到字节
func commandToBytes(command string) []byte {
	var bytes [commandlength]byte
	for index, char := range command {
		bytes[index] = byte(char) //char转换成字节
	}
	return bytes[:]
}

//提取命令
func extractCommand(request []byte) []byte {
	return request[:commandlength]
}

//请求块
func requestBlocks() {
	for _, node := range knowNodes { //给所有已知节点发送请求
		sendGetBlocks(node)
	}
}

//发送块
func sendBlock(addr string, bc *Block) {
	// data := block{addr, bc.Serialize()}
	data := block{nodeAddress, bc.Serialize()}             //构造模块
	payload := gobEncode(data)                             //数据编码
	request := append(commandToBytes("block"), payload...) //定制请求
	sendData(addr, request)                                //发送数据
}

//发送地址
func sendaddr(address string) {
	nodes := addr{knowNodes}                              //已知的节点
	nodes.Addrlist = append(nodes.Addrlist, nodeAddress)  //追加当前节点
	payload := gobEncode(nodes)                           //增加解码的节点
	request := append(commandToBytes("addr"), payload...) //创建请求
	sendData(address, request)                            //发送数据
}

//发送数据
func sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr) //建立TCP网络连接对象
	if err != nil {
		fmt.Printf("----网络地址 %s 不可到达!!!!\n", addr)
		var updateNodes []string
		for _, node := range knowNodes {
			if node != addr {
				updateNodes = append(updateNodes, node) //刷新节点
			}
		}
		//更新可达的已知节点
		knowNodes = updateNodes
		return
	}

	defer conn.Close()                            //延迟关闭
	_, err = io.Copy(conn, bytes.NewReader(data)) //拷贝数据，发送
	if err != nil {
		log.Panic(err)
	}
}

//发送请求
func sendInv(address, kind string, items [][]byte) {
	inventory := inv{nodeAddress, kind, items}
	payload := gobEncode(inventory)                      //对历史数据进行编码
	request := append(commandToBytes("inv"), payload...) //网络请求
	sendData(address, request)
}

//发送请求多个模块
func sendGetBlocks(address string) {
	payload := gobEncode(getblocks{nodeAddress})
	request := append(commandToBytes("getblocks"), payload...)
	sendData(address, request)
}

//发送请求数据
func sendGetData(address, kind string, id []byte) {
	// payload := gobEncode(getdata{address, kind, id})
	payload := gobEncode(getdata{nodeAddress, kind, id})
	request := append(commandToBytes("getdata"), payload...)
	sendData(address, request)
}

//发送交易
func sendTx(addr string, tnx *Transaction) {
	data := tx{nodeAddress, tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)
	sendData(addr, request)
}

//发送版本信息
func sendVersion(addr string, bc *BlockChain) {
	bestHeight := bc.GetBestHeight() //最后一个区块的 Height
	payload := gobEncode(verzion{nodeVersion, bestHeight, nodeAddress})
	request := append(commandToBytes("version"), payload...)
	sendData(addr, request)
}

//处理 block 命令请求
func handleBlock(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandlength:]) //取出数据
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block           //区块数据
	block := DeserializeBlock(blockData) //解码
	fmt.Printf("----收到一个新的区块，区块哈希：%x\n", block.Hash)
	bc.AddBlock(block)
	fmt.Printf("区块处理完成，区块哈希：%x \n", block.Hash)
	if len(blocksInTransit) > 0 {
		blockhash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockhash) //发送请求
		blocksInTransit = blocksInTransit[1:]
		fmt.Printf("区块同步中....\n")
	}else{
		UTXOSet := UTXOSet{bc}
		UTXOSet.Reindex() //重建本地索引，可能是增加了区块，所以未花费输出 UTXO 会变化，因此 UTXO 的本地持久化数据需要重建
		fmt.Printf("----区块同步完成，本地索引重建完成\n")
	}
}

//读取网络地址，处理 addr 命令请求
func handleaddr(request []byte) {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[commandlength:]) //取出数据
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	knowNodes = append(knowNodes, payload.Addrlist...) //追加已知节点列表
	fmt.Printf("----已经有了 %d 个节点\n", len(knowNodes))
	requestBlocks() //请求区块数据
}

//处理 inv 命令请求
func handleInv(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandlength:]) //取出数据
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("----收到库存 %d 个，类型：%s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items
		blockhash := payload.Items[0]
		sendGetData(payload.AddrFrom, "block", blockhash)

		newInTransit := [][]byte{}
		for _,b := range blocksInTransit {
			if bytes.Compare(b,blockhash)!=0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit //同步区块
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]
		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddrFrom, "tx", txID) //发起请求的交易
		}
	}
}

//获取多个块，处理 getblocks 命令请求
func handleGetBlocks(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload getblocks

	buff.Write(request[commandlength:]) //取出数据
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := bc.GetBlockHashes()
	sendInv(payload.AddrFrom, "block", blocks)
}

//获取数据，处理 getdata 命令请求
func handleGetData(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload getdata

	buff.Write(request[commandlength:]) //取出数据
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block,err := bc.GetBlock([]byte(payload.ID)) //获取一个区块
		if err!=nil {
			return
		}
		sendBlock(payload.AddrFrom, &block) //发送区块
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID] //内存池
		sendTx(payload.AddrFrom, &tx) //发送交易
	}


}

//获取交易，处理 tx 命令请求
func handleTx(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request[commandlength:]) //取出数据
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction //交易数据
	tx := DeserializeTransaction(txData) //解码交易数据
	mempool[hex.EncodeToString(tx.ID)] = tx //处理交易，将交易放入内存池中
	fmt.Printf("----nodeAddress：%s, knowNodes[0]：%s\n", nodeAddress, knowNodes[0])
	if nodeAddress == knowNodes[0] {
		for _,node := range knowNodes {
			if node != nodeAddress && node != payload.AddrFrom {
				sendInv(node, "tx", [][]byte{tx.ID}) //发送库存，从代码来看，应该是将本地节点的库存交易数据广播给其它节点
			}
		}
	}else{
		if len(mempool)>=2 && len(miningAddress)>0 {
		MineTransactions:
		    var txs []*Transaction
			for id := range mempool {
				tx := mempool[id] //取得交易
				if bc.VerifyTransaction(&tx) { //校验交易合法性
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				fmt.Println("本地交易池中没有任何交易，等待新的交易加入....")
				return
			}

			cbTx := NewCoinBaseTX(miningAddress, "") //创建挖矿交易，为地址 miningAddress 挖矿
			txs = append(txs, cbTx)

			newBlock := bc.MineBlock(txs) //挖矿
			UTXOSet := UTXOSet{bc}
			UTXOSet.Reindex() //重建索引，因为挖掘到新的区块，UTXO 集合有变化，所以需要重建索引 (即更新本地 UTXO 持久化副本)
			fmt.Printf("已经挖掘到新的区块\n")
			for _,tx := range txs {
				txID := hex.EncodeToString(tx.ID) //交易编号
				delete(mempool, txID) //删除交易池中已经打包到挖出来的区块中的交易
			}

			for _,node := range knowNodes {
				if node != nodeAddress {
					sendInv(node, "block", [][]byte{newBlock.Hash}) //本地节点 nodeAddress 挖矿成功后，将新挖出来的区块向其它节点广播
				}
			}

			if len(mempool)>0 {
				goto MineTransactions
			}

		}

	}

}

//处理版本，处理 version 命令请求
func handleVersion(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload verzion

	buff.Write(request[commandlength:]) //取出数据
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	myBestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	//版本同步
	if myBestHeight < foreignerBestHeight{
		sendGetBlocks(payload.AddrFrom)
	}else if myBestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, bc)
	}

	if !nodeIsKonw(payload.AddrFrom) {
		knowNodes = append(knowNodes, payload.AddrFrom) //如果节点地址未知，将未知节点地址添加到已知节点列表中
	}

}

//处理网络连接
func handleConnection(conn net.Conn, bc *BlockChain) {
	request,err := ioutil.ReadAll(conn)
	if err!=nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandlength])
	fmt.Printf("----收到命令：%s\n", command)

	//处理不同类型的命令请求
	switch command {
	case "addr":
		handleaddr(request)
	case "block":
		handleBlock(request, bc)
	case "inv":
		handleInv(request, bc)
	case "getblocks":
		handleGetBlocks(request, bc)
	case "getdata":
		handleGetData(request, bc)
	case "tx":
		handleTx(request, bc)
	case "version":
		handleVersion(request, bc)
	default:
		fmt.Printf("----未知命令 %s，当前仅支持命令请求 addr、block、inv、getblocks、getdata、tx、version\n", command)
	}
}


func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff) //编码器
	err := enc.Encode(data)      //编码
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes() //字节
}

//判断一个节点是不是已知节点
func nodeIsKonw(addr string) bool {
	for _, node := range knowNodes {
		if node == addr {
			return true
		}
	}
	return false
}
