package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

//奖励，矿工挖矿给予的奖励
const subsidy = 10

//输入
type TXInput struct {
	Txid []byte //Txid存储了引用的交易的id
	Vout int //Vout则保存引用的该交易中的一个output索引
	ScriptSig string //ScriptSig仅只是保存了一个任意的用户定义的钱包地址
}

//检查地址是否启动事务，其实好像就是判断是否可以解锁输入
func (input *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return input.ScriptSig == unlockingData
}

//输出
type TXOutput struct {
	Value int //output保持了“币”，这里的Value
	ScriptPubkey string //用脚本语言意味着比特币也可以作为智能合约平台，从代码来看，好像也是一个钱包地址
}

//是否可以解锁输出
func (output *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return output.ScriptPubkey == unlockingData //判断是否可以解锁
}



//交易，包括编号、输入、输出
type Transaction struct {
	ID []byte
	Vin []TXInput
	Vout []TXOutput
}

//检查交易事务是否为coinbase，挖矿得来的奖励
func (tx *Transaction) IsCoinBase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

//设置交易ID，从二进制数据中
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer //开辟内存
	var hash [32]byte //哈希数组
	enc := gob.NewEncoder(&encoded) //创建编码对象
	err := enc.Encode(tx) //编码
	if err != nil {
		log.Panic(err)
	}

	hash = sha256.Sum256(encoded.Bytes()) //计算哈希
	tx.ID = hash[:] //设置哈希
}

//挖矿交易，从创世区块的创建来看，区块链系统中的金额最先是产生于挖矿交易
func NewCoinBaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("挖矿奖励给：%s\n", to)
	}
	txin := TXInput{[]byte{}, -1, data} //输入奖励
	txout := TXOutput{subsidy, to} //输出奖励
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}} //挖矿产生的交易
	return &tx
}

//转账交易
func NewUTXOTransaction(from, to string, amount int, bc *BlockChain) *Transaction {
	var inputs []TXInput //输入
	var outputs []TXOutput //输出
    acc, validOutputs := bc.FindSpendableOutputs(from, amount)
	if acc < amount {
		log.Panic("交易金额不足")
	}

	for txid, outs := range validOutputs { //循环遍历无效输出
		txID, err := hex.DecodeString(txid) //解码
		if err != nil {
			log.Panic(err) //处理错误
		}
		for _, out := range outs {
			input := TXInput{txID, out, from} //输入的交易
			inputs = append(inputs, input) //输出的交易
		}
	}

	//交易叠加
	outputs = append(outputs, TXOutput{amount, to})

	if acc > amount {
		//记录以后的金额，即多余的金额转回给 from
		outputs = append(outputs, TXOutput{acc-amount, from})
	}

	tx := Transaction{nil, inputs, outputs} //交易
	tx.SetID() //设置ID
	return &tx
}
