package main

import (
	"bytes"
	"encoding/gob"
	"log"
)

//输出
type TXOutput struct {
	Value int //output保持了“币”，这里的Value
	PubKeyHash []byte //用脚本语言意味着比特币也可以作为智能合约平台，从代码来看，好像也是一个钱包地址
}

//输出锁住的标志，从代码来看，其实就是给这笔输出设置钱包地址，也即设置对应的公钥
func (out *TXOutput) Lock(address []byte) {
	pubkeyhash := Base58Decode(address) //解码
	pubkeyhash = pubkeyhash[1:len(pubkeyhash)-4] //截取有效哈希
	out.PubKeyHash = pubkeyhash
}

//检测是否被锁住，其实就是校验公钥哈希是否一致
func (out *TXOutput) IsLockedWithKey(pubkeyHash []byte) bool {
	return bytes.Compare([]byte(out.PubKeyHash), pubkeyHash) == 0
}

//创建一个输出
func NewTxOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))
	return txo
}

type TXOutputs struct {
	Outputs []TXOutput
}

//对象到二进制
func (outs *TXOutputs) Serialize() []byte {
	var buff bytes.Buffer //开辟内存
	enc := gob.NewEncoder(&buff) //创建编码器
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes() //返回二进制
}

//二进制到对象
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs
	dec := gob.NewDecoder(bytes.NewReader(data)) //解码对象
	err := dec.Decode(&outputs) //解码操作
	if err!=nil {
		log.Panic(err)
	}

	return outputs
}