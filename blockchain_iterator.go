package main

import (
	"log"
	"github.com/boltdb/bolt"
)

type BlockChainIterator struct {
	currentHash []byte   //当前的哈希
	db          *bolt.DB //数据库
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