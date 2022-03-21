package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

var (
	maxNonce = math.MaxInt64 //最大的64位整数
)

//从代码来看，应该是targetBits越大，挖矿难度越大
// const targetBits = 24 //对比的位数
const targetBits = 16 //对比的位数

type ProofOfWork struct {
	block *Block //区块
	target *big.Int //存储计算哈希对比的特定整数
}

//创建一个工作量证明的挖矿对象
func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1) //初始化目标整数
	target.Lsh(target, uint(256 - targetBits)) //数据转换，通过位操作实现
	pow := &ProofOfWork{block, target} //创建对象
	return pow
}

//准备数据进行挖矿计算
func (pow *ProofOfWork)prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash, //上一区块的哈希
			pow.block.Data, //当前区块的数据
			IntToHex(pow.block.Timestamp), //时间，十六进制
			IntToHex(int64(targetBits)), //位数，十六进制
			IntToHex(int64(nonce)), //保存工作量的nonce
		},
		[]byte{},
	)

	return data
}

//挖矿执行
func (pow *ProofOfWork) Run() (int , []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0
	fmt.Printf("当前挖矿计算的区块数据：%s\n", pow.block.Data)
	for nonce < maxNonce {
		data := pow.prepareData(nonce) //准备好的数据
		hash = sha256.Sum256(data) //计算出哈希
		fmt.Printf("\r当前区块nonce：%d，当前区块哈希：%x", nonce, hash) //打印显示哈希
		hashInt.SetBytes(hash[:]) //获取要对比的哈希
		//挖矿的校验，可以这样理解，将区块中的部分数据与nonce一起哈希，如果得到的
		//哈希值满足某种条件即视为挖矿成功，这里的条件就是哈希值要小于某个数pow.target
		if hashInt.Cmp(pow.target) == -1{
			break
		}else{
			nonce++
		}
	}
	//nonce 即满足挖矿成功条件的值，即相当于区块链挖矿解释中的解题的答案
	//hash 即当前哈希
	fmt.Print("\n\n")
	return nonce, hash[:] 
}

//校验挖矿是否真的成功
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	data := pow.prepareData(pow.block.Nonce) //准备好的数据
	hash := sha256.Sum256(data) //计算出哈希
	hashInt.SetBytes(hash[:]) //获取要对比的哈希
	isValid := (hashInt.Cmp(pow.target) == -1) //校验数据
	return isValid
}