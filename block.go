package main

import (
	"bytes"
	_ "bytes"
	_ "crypto/sha256"
	_ "crypto/sha256"
	"encoding/gob"
	"log"
	_ "strconv"
	"time"
)

type Block struct {
	Timestamp     int64  //时间线，1970/01/01 00.00.00
	// Data          []byte //交易数据
	Transactions []*Transaction //交易的集合
	PrevBlockHash []byte //上一块数据的哈希
	Hash          []byte //当前块数据的哈希
	Nonce int //工作量证明
}

//对于交易实现哈希计算
func (block *Block) HashTransactions() []byte {
	var transactions [][]byte
	for _,tx := range block.Transactions {
		transactions = append(transactions, tx.Serialize())
	}
	mTree := NewMerkleTree(transactions)
	return mTree.RootNode.data
}

//创建一个区块
func NewBlock(transaactions []*Transaction, previousHash []byte) *Block {
	//block是一个指针，取得一个对象初始化之后的地址
	block := &Block{time.Now().Unix(), transaactions, previousHash, []byte{}, 0}
	// block.SetHash() //设置区块的哈希
	//对这个区块进行挖矿
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run() //开始挖矿
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

//创建创世区块
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

//对象转化为二进制字节集，可以写入文件
func (block *Block) Serialize() []byte{
	var result bytes.Buffer //开辟内存，存放字节集合
	encoder := gob.NewEncoder(&result) //编码对象创建
	err := encoder.Encode(block) //编码操作
	if err != nil {
		log.Panic(err) //处理错误
	}

	return result.Bytes() //返回字节
}

//读取文件，读到二进制字节集，将二进制字节集转化为对象
func DeserializeBlock(data []byte) *Block{
	var block Block //对象存储用于字节转化的对象
	decoder := gob.NewDecoder(bytes.NewReader(data)) //解码
	err := decoder.Decode(&block) //尝试解码
	if err != nil {
		log.Panic(err) //处理错误
	}

	return &block
}
