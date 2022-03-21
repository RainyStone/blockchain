package main

import (
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db" //数据库文件，在当前目录下
const blockBucket = "blocks"   //名称

type BlockChain struct {
	// blocks []*Block //一个切片，每个元素都是指针，存储block区块的地址
	tip []byte //二进制数据，其实也是一个哈希值
	db  *bolt.DB //数据库
}

// //添加一个区块到区块链中
// func (blockChain *BlockChain) AddBlock(data string) {
// 	//取出当前区块链的最后一个区块
// 	prevBlock := blockChain.blocks[len(blockChain.blocks)-1]
// 	//创建一个新的区块
// 	newBlock := NewBlock(data, prevBlock.Hash)
// 	//当前区块链添加新的区块
// 	blockChain.blocks = append(blockChain.blocks, newBlock)

// }

// //创建一条区块链
// func NewBlockChain() *BlockChain {
// 	return &BlockChain{[]*Block{NewGenesisBlock()}}
// }

type BlockChainIterator struct {
	currentHash []byte //当前的哈希
	db *bolt.DB //数据库
}

//增加一个区块
func (blockChain *BlockChain) AddBlock(data string) {
	var lastHash []byte //上一块哈希
	err := blockChain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket)) //取得数据
		lastHash = bucket.Get([]byte("1")) //取得第一块，即创世块的哈希
		return nil
	})

	if err != nil {
		log.Panic(err) //处理打开错误
	}

	newBlock := NewBlock(data, lastHash) //创建一个新的区块

	err = blockChain.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket)) //取出
		err := bucket.Put(newBlock.Hash, newBlock.Serialize()) //压入数据
		if err != nil {
			log.Panic(err) //处理压入错误
		}

		// "1"是区块链中最新区块的哈希数据的键
		err = bucket.Put([]byte("1"), newBlock.Hash) //压入数据
		if err != nil {
			log.Panic(err) //处理压入错误
		}

		blockChain.tip = newBlock.Hash

		return nil
	})
}

//迭代器
func (blockChain *BlockChain) Iterator() *BlockChainIterator{
	bcit := &BlockChainIterator{blockChain.tip, blockChain.db}
	return bcit //根据区块链创建区块链迭代器
}

//取得下一个区块，从代码来看，应该是取得 currentHash 对应的区块，再把 currentHash 回移到前一区块的 哈希
func (it *BlockChainIterator) next() *Block{
	var block *Block
	err := it.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		encodedBlock := bucket.Get(it.currentHash) //抓取二进制数据
		block = DeserializeBlock(encodedBlock) //解码
		return nil
	})

    if err != nil {
		log.Panic(err)
	}

	it.currentHash = block.PrevBlockHash //哈希赋值
	return block
}

//新建一条区块链
func NewBlockChain() *BlockChain {
	var tip []byte //存储区块链的二进制数据

	db, err := bolt.Open(dbFile, 0600, nil) //打开数据库
	if err != nil {
		log.Panic(err) //处理数据库打开错误
	}

	//处理数据更新
	err = db.Update(func (tx *bolt.Tx) error{
		bucket := tx.Bucket([]byte(blockBucket)) //按照名称打开数据库表格
		if bucket == nil {
			fmt.Println("当前数据库没有区块链，创建一条新的区块链")

			genesis := NewGenesisBlock() //创建创世区块
			bucket, err := tx.CreateBucket([]byte(blockBucket)) //创建一个数据库表格
			if err != nil {
				log.Panic(err) //处理创建错误
			}

			err = bucket.Put(genesis.Hash, genesis.Serialize()) //存入数据
			if err != nil {
				log.Panic(err) //处理存入错误
			}

			err = bucket.Put([]byte("1"), genesis.Hash) //存入数据
			if err != nil {
				log.Panic(err) //处理存入错误
			}

			tip = genesis.Hash //取得哈希
		}else{
			tip = bucket.Get([]byte("1"))
		}
		return nil
	})
	if err != nil {
		log.Panic(err) //处理数据库更新错误
	}

	bc := BlockChain{tip, db} //利用数据库二进制数据，创建一条区块链
	return &bc
}
