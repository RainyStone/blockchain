package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db" //数据库文件，在当前目录下
const blockBucket = "blocks"   //名称
const genesisCoinbaseData = "创世块z0000"

type BlockChain struct {
	// blocks []*Block //一个切片，每个元素都是指针，存储block区块的地址
	tip []byte //二进制数据，其实也是一个哈希值，保存某个区块对应的哈希，一般是区块链中最新区块对应的哈希
	db  *bolt.DB //数据库
}

//挖矿带来的交易
func (blockchain *BlockChain) MineBlock(transactions []*Transaction) {
	var lastHash []byte //最后的哈希
	err := blockchain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket)) //查看数据
		lastHash = bucket.Get([]byte("1")) //取出最后区块的哈希
		return nil
	})

	if err != nil {
		log.Panic(err) //处理错误
	}

	newBlock := NewBlock(transactions, lastHash) //创建一个新的区块
	err = blockchain.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket)) //取出索引
		err := bucket.Put(newBlock.Hash, newBlock.Serialize()) //存入数据库
		if err != nil {
			log.Panic(err) //处理错误
		}

		err = bucket.Put([]byte("1"), newBlock.Hash) //更新数据库最新区块的哈希
		if err != nil {
			log.Panic(err) //处理错误
		}

		//更新区块链中最新区块的哈希
		blockchain.tip = newBlock.Hash
		return nil
	})
}

//获取 address 对应的未使用输出的交易列表
func (blockchain *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction //交易事务
	spentTXOs := make(map[string][]int) //开辟内存
	bci := blockchain.Iterator() //迭代器

	for {
		block := bci.next() //循环下一个区块

		//循环区块中的每个交易
		for _, tx := range block.Transactions {
			// txID := hex.EncodedLen(tx.ID) //获取交易编号
			txID := hex.EncodeToString(tx.ID) //获取交易编号
        Outputs:
		    for outindex, out := range tx.Vout { //循环遍历输出
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outindex{
							continue Outputs //循环到不等
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx) //加入列表
				}
			}

			///////////////
			//这一段总感觉有问题，应该要在获取未交易输出前先确定 spentTXOs
			if ! tx.IsCoinBase() {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) { //判断是否可以锁定
						inTxID := hex.EncodeToString(in.Txid) //编码为字符串
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
			///////////////
		}

		if len(block.PrevBlockHash) == 0 { //最后一个区块，跳出
			break
		}
	}

	return unspentTXs
}

//获取 address 对应的所有没有使用的交易输出
func (blockchain *BlockChain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	//查找未使用输出对应的交易
	unspentTransactions := blockchain.FindUnspentTransactions(address)
	//循环所有交易
	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) { //判断是否是当前 address 对应的输出
				UTXOs = append(UTXOs, out) //加入数据
			}
		}
	}
	return UTXOs
}


//获取没有使用的输出以参考输入，从代码来看，实际上就是获取满足金额 amount 要求的 address 对应的交易输出
func (blockchain *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int) //输出
	unspentTxs := blockchain.FindUnspentTransactions(address) //根据地址查找所有的交易
	accmulated := 0 //累计
Work:
    for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID) //获取编号
		for outindex, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accmulated < amount {
				accmulated += out.Value //累加金额
				//这里可能是考虑到同一笔交易的多个输出有可能对应同一个 address
				unspentOutputs[txID] = append(unspentOutputs[txID], outindex)

				if accmulated >= amount {
					break Work
				}
			}
		}
	}
	return accmulated, unspentOutputs
}
type BlockChainIterator struct {
	currentHash []byte //当前的哈希
	db *bolt.DB //数据库
}


// //增加一个区块
// func (blockChain *BlockChain) AddBlock(data string) {
// 	var lastHash []byte //上一块哈希
// 	err := blockChain.db.View(func(tx *bolt.Tx) error {
// 		bucket := tx.Bucket([]byte(blockBucket)) //取得数据
// 		lastHash = bucket.Get([]byte("1")) //取得第一块，即创世块的哈希
// 		return nil
// 	})

// 	if err != nil {
// 		log.Panic(err) //处理打开错误
// 	}

// 	newBlock := NewBlock(data, lastHash) //创建一个新的区块

// 	err = blockChain.db.Update(func(tx *bolt.Tx) error {
// 		bucket := tx.Bucket([]byte(blockBucket)) //取出
// 		err := bucket.Put(newBlock.Hash, newBlock.Serialize()) //压入数据
// 		if err != nil {
// 			log.Panic(err) //处理压入错误
// 		}

// 		// "1"是区块链中最新区块的哈希数据的键
// 		err = bucket.Put([]byte("1"), newBlock.Hash) //压入数据
// 		if err != nil {
// 			log.Panic(err) //处理压入错误
// 		}

// 		blockChain.tip = newBlock.Hash

// 		return nil
// 	})
// }

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

//判断数据库是否存在
func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}


//新建一条区块链
//????函数中貌似没有使用到 address 参数
func NewBlockChain(address string) *BlockChain {
	if ! dbExists() {
		fmt.Println("数据库不存在，创建一个新的数据库")
		os.Exit(1)
	}

	var tip []byte //存储区块链的二进制数据

	db, err := bolt.Open(dbFile, 0600, nil) //打开数据库
	if err != nil {
		log.Panic(err) //处理数据库打开错误
	}

	//处理数据更新
	err = db.Update(func (tx *bolt.Tx) error{
		bucket := tx.Bucket([]byte(blockBucket)) //按照名称打开数据库表格
		tip = bucket.Get([]byte("1"))
		return nil
	})
	if err != nil {
		log.Panic(err) //处理数据库更新错误
	}

	bc := BlockChain{tip, db} //利用数据库二进制数据，创建一条区块链
	return &bc
}

func createBlockChain(address string) *BlockChain {
	if dbExists() {
		fmt.Println("数据库已经存在，无需创建")
		os.Exit(1)
	}

	var tip []byte //存储区块链的二进制数据

	db, err := bolt.Open(dbFile, 0600, nil) //打开数据库
	if err != nil {
		log.Panic(err) //处理数据库打开错误
	}

	err = db.Update(func(tx *bolt.Tx) error {
		//从创世区块的创建来看，区块链系统中的金额最先是产生于挖矿交易的
		cbtx := NewCoinBaseTX(address, genesisCoinbaseData) //创建创世区块的事务交易
		genesis := NewGenesisBlock(cbtx) //创建创世区块的块
		bucket, err := tx.CreateBucket([]byte(blockBucket))
		if err != nil {
			log.Panic(err) //处理数据库创建错误
		}

		//存储创世区块的哈希为键，值为创世区块的数据
		err = bucket.Put(genesis.Hash, genesis.Serialize()) //存储
		if err != nil {
			log.Panic(err)
		}

		//保存最新区块的哈希，这里的最新区块其实就是创世区块
		err = bucket.Put([]byte("1"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}

		tip = genesis.Hash

		return nil
	})

	bc := BlockChain{tip, db} //利用数据库二进制数据，创建一条区块链
	return &bc
}