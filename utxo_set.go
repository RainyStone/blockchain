package main

import (
	"encoding/hex"
	"log"

	"github.com/boltdb/bolt"
)

const utxoBucket = "chainstate" //存储状态，应该是保存区块链所有未花费输出的

//二次封装区块链
type UTXOSet struct {
	blockchain *BlockChain
}

//输出查找并返回未使用的输出
func (utxo UTXOSet) FindSpendableOutputs(publickeyhash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int) //处理输出
	accumulated := 0                         //累计金额
	db := utxo.blockchain.db                 //调用数据库

	//查询数据
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket)) //查询数据
		cur := bucket.Cursor() //当前游标
		for key,value := cur.First(); key!=nil; key,value = cur.Next() {
			txID := hex.EncodeToString(key) //交易编号
			outs := DeserializeOutputs(value) //解码
			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(publickeyhash) && accumulated<amount {
					accumulated += out.Value //叠加金额
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx) //叠加交易输出
				}
			}
		}
		return nil
	})

	if err != nil {
		log.Panic(err) //输出错误
	}

	return accumulated, unspentOutputs //返回数据
}

//查找UTXO，按照公钥查询
func (utxo UTXOSet) FindUTXO(publickeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := utxo.blockchain.db
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket)) //查询数据
		cur := bucket.Cursor() //当前游标
		for key,value := cur.First(); key!=nil; key,value = cur.Next() {
			outs := DeserializeOutputs(value) //反序列化数据库的数据
			for _,out := range outs.Outputs {
				if out.IsLockedWithKey(publickeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	})

	if err!= nil {
		log.Panic(err)
	}

	return UTXOs
}

//统计交易
func (utxo UTXOSet) CountTransactions() int {
	db := utxo.blockchain.db
	counter := 0

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket)) //查询数据
		cur := bucket.Cursor() //当前游标

		for k,_ := cur.First(); k!=nil; k,_=cur.Next() {
			counter++
		}

		return nil
	})

	if err!= nil {
		log.Panic(err)
	}

	return counter
}

//插入数据时，重建索引，其实把UTXO进行持久化了
func (utxo UTXOSet) Reindex() {
	db := utxo.blockchain.db
	bucketname := []byte(utxoBucket)
	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketname) //删除
		if err!=nil && err!=bolt.ErrBucketNotFound {
			log.Panic(err)
		}
		_,err = tx.CreateBucket(bucketname) //新建
		if err!= nil {
			log.Panic(err)
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	UTXO := utxo.blockchain.FindUTXO()
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketname) //取出数据

		for txID,outs := range UTXO {
			key,err := hex.DecodeString(txID)
			if err!= nil {
				log.Panic(err)
			}
			err = bucket.Put(key, outs.Serialize())
			if err!= nil {
				log.Panic(err)
			}
		}

		return nil
	})

}

//刷新数据，更新区块链中当前的未花费输出列表
func (utxo UTXOSet) Update(block *Block) {
	db := utxo.blockchain.db
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		for _,tx := range block.Transactions { //循环遍历所有数据
			if !tx.IsCoinBase() { //取出非挖矿交易
				for _,vin := range tx.Vin {
					updateOuts := TXOutputs{} //创建集合
					outsBytes := bucket.Get(vin.Txid) //取出数据
					outs := DeserializeOutputs(outsBytes) //解码二进制数据
					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updateOuts.Outputs = append(updateOuts.Outputs, out)
						}
					}
					if len(updateOuts.Outputs) == 0 {
						err := bucket.Delete(vin.Txid) //，删除
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := bucket.Put(vin.Txid, updateOuts.Serialize())
						if err != nil {
							log.Panic(err)
						}
					}
				}
			}

			newOutputs := TXOutputs{}
			for _, out := range tx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}
			err := bucket.Put(tx.ID, newOutputs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}


		return nil
	})

	if err!= nil {
		log.Panic(err)
	}
}