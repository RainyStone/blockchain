package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	_ "errors"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain_%s.db" //数据库文件，在当前目录下
const blockBucket = "blocks"      //名称
const genesisCoinbaseData = "创世块z0000"

type BlockChain struct {
	// blocks []*Block //一个切片，每个元素都是指针，存储block区块的地址
	tip []byte   //二进制数据，其实也是一个哈希值，保存某个区块对应的哈希，一般是区块链中最新区块对应的哈希
	db  *bolt.DB //数据库
}

//挖矿带来的交易
func (blockchain *BlockChain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte //最后的哈希
	var lastHeight int

	for _, tx := range transactions {
		if !blockchain.VerifyTransaction(tx) {
			log.Panic("交易不正确，存在错误!!!!")
		}
	}

	err := blockchain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket)) //查看数据
		lastHash = bucket.Get([]byte("1"))       //取出最后区块的哈希

		blockData := bucket.Get(lastHash)
		block := DeserializeBlock(blockData)

		lastHeight = block.Height

		return nil
	})

	if err != nil {
		log.Panic(err) //处理错误
	}

	newBlock := NewBlock(transactions, lastHash, lastHeight+1) //创建一个新的区块
	err = blockchain.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))               //取出索引
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

	return newBlock
}

//获取 address 对应的未使用输出的交易列表
func (blockchain *BlockChain) FindUnspentTransactions(pubkeyhash []byte) []Transaction {
	var unspentTXs []Transaction        //交易事务
	spentTXOs := make(map[string][]int) //开辟内存
	bci := blockchain.Iterator()        //迭代器

	for {
		block := bci.next() //循环下一个区块

		//循环区块中的每个交易
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID) //获取交易编号
		Outputs:
			for outindex, out := range tx.Vout { //循环遍历输出
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outindex {
							continue Outputs //循环到不等
						}
					}
				}

				if out.IsLockedWithKey(pubkeyhash) {
					unspentTXs = append(unspentTXs, *tx) //加入列表
				}
			}

			if !tx.IsCoinBase() {
				for _, in := range tx.Vin {
					if in.UsesKey(pubkeyhash) { //判断是否可以锁定
						inTxID := hex.EncodeToString(in.Txid) //编码为字符串
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 { //最后一个区块，跳出
			break
		}
	}

	return unspentTXs
}

//获取 address 对应的所有没有使用的交易输出
func (blockchain *BlockChain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := blockchain.Iterator()
	for {
		block := bci.next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spendoutidx := range spentTXOs[txID] {
						if spendoutidx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if !tx.IsCoinBase() {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return UTXO
}

//获取没有使用的输出以参考输入，从代码来看，实际上就是获取满足金额 amount 要求的 address 对应的交易输出
func (blockchain *BlockChain) FindSpendableOutputs(pubkeyhash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)                     //输出
	unspentTxs := blockchain.FindUnspentTransactions(pubkeyhash) //根据地址查找所有的交易
	accmulated := 0                                              //累计
Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID) //获取编号
		for outindex, out := range tx.Vout {
			if out.IsLockedWithKey(pubkeyhash) && accmulated < amount {
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

//迭代器
func (blockChain *BlockChain) Iterator() *BlockChainIterator {
	bcit := &BlockChainIterator{blockChain.tip, blockChain.db}
	return bcit //根据区块链创建区块链迭代器
}

//判断数据库是否存在
func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

//新建一条区块链
func NewBlockChain(nodeID string) *BlockChain {
	dbFile := fmt.Sprintf(dbFile, nodeID)

	if !dbExists(dbFile) {
		fmt.Println("数据库不存在，创建一个新的数据库")
		os.Exit(1)
	}

	var tip []byte //存储区块链的二进制数据

	db, err := bolt.Open(dbFile, 0600, nil) //打开数据库
	if err != nil {
		log.Panic(err) //处理数据库打开错误
	}

	//处理数据更新
	err = db.Update(func(tx *bolt.Tx) error {
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

func CreateBlockChain(address string, nodeID string) *BlockChain {
	dbFile := fmt.Sprintf(dbFile, nodeID)

	if dbExists(dbFile) {
		fmt.Println("数据库已经存在，无需创建")
		os.Exit(1)
	}

	var tip []byte //存储区块链的二进制数据

	//从创世区块的创建来看，区块链系统中的金额最先是产生于挖矿交易的
	cbtx := NewCoinBaseTX(address, genesisCoinbaseData) //创建创世区块的事务交易
	genesis := NewGenesisBlock(cbtx)                    //创建创世区块的块

	db, err := bolt.Open(dbFile, 0600, nil) //打开数据库
	if err != nil {
		log.Panic(err) //处理数据库打开错误
	}

	err = db.Update(func(tx *bolt.Tx) error {
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

	if err != nil {
		log.Panic(err)
	}

	bc := BlockChain{tip, db} //利用数据库二进制数据，创建一条区块链
	return &bc
}

//交易签名
func (blockchain *BlockChain) SignTransaction(tx *Transaction, privatekey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Vin {
		preTx, err := blockchain.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(preTx.ID)] = preTx
	}

	tx.Sign(privatekey, prevTXs)
}

func (blockchain *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	bci := blockchain.Iterator()
	for {
		block := bci.next()
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return Transaction{}, nil
}

//验证交易
func (blockchain *BlockChain) VerifyTransaction(tx *Transaction) bool {
	prevTxs := make(map[string]Transaction)
	for _, vin := range tx.Vin {
		prevTx, err := blockchain.FindTransaction(vin.Txid) //查找交易
		if err != nil {
			log.Panic(err)
		}
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}

	return tx.Verify(prevTxs)
}

//获取最后一个区块的 Height 参数用于同步
func (blockchain *BlockChain) GetBestHeight() int {
	var lastBlock Block //最后一个区块
	err := blockchain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket)) //取出数据库的数据对象
		lastHash := bucket.Get([]byte("1"))      //取得最后一个区块的哈希
		blockdata := bucket.Get(lastHash)        //取得区块数据
		lastBlock = *DeserializeBlock(blockdata) //解码区块数据
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastBlock.Height
}

//增加区块
func (blockchain *BlockChain) AddBlock(block *Block) {
	err := blockchain.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))

		blockInDb := bucket.Get(block.Hash) //判断区块是否已经存在于数据库中
		if blockInDb != nil {
			fmt.Printf("区块已在数据库中，无需压入，区块哈希: %x \n", block.Hash)
			return nil
		}

		blockData := block.Serialize()           //序列化
		err := bucket.Put(block.Hash, blockData) //压入区块数据
		fmt.Printf("区块压入数据库，区块哈希: %x \n", block.Hash)
		if err != nil {
			log.Panic(err)
		}

		lastHash := bucket.Get([]byte("1"))          //取得数据库中区块链中，之前最后一个区块的哈希
		lastBlockData := bucket.Get(lastHash)        //依据哈希取得该区块的数据
		lastBlock := DeserializeBlock(lastBlockData) //反序列化该区块

		//如果新区块的 Height > 之前最后一个区块的 Heght，则更新数据库中键 "1" 对应的哈希值
		//????暂不清楚为啥要这样做
		if block.Height > lastBlock.Height {
			err = bucket.Put([]byte("1"), block.Hash)
			if err != nil {
				log.Panic(err)
			}
			blockchain.tip = block.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

//区块链中查找区块
func (blockchain *BlockChain) GetBlock(blockhash []byte) (Block, error) {
	var block Block

	err := blockchain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		blockdata := bucket.Get(blockhash)
		if blockdata == nil {
			return errors.New("没有找到哈希对应的区块数据")
		}

		block = *DeserializeBlock(blockdata)

		return nil
	})

	if err != nil {
		return block, err
	}

	return block, nil
}

func (blockchain *BlockChain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := blockchain.Iterator()

	for {
		block := bci.next()
		blocks = append(blocks, block.Hash)
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
}
