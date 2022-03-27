package main

import "bytes"

//输入
type TXInput struct {
	Txid []byte //Txid存储了引用的交易的id
	Vout int //Vout则保存引用的该交易中的一个output索引
	Signature []byte //签名，应该就是私钥
	PubKey []byte //公钥
}

//公钥检测一下地址与交易
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockinghash := HashPubkey(in.PubKey)
	return bytes.Compare(lockinghash, pubKeyHash) == 0
}
