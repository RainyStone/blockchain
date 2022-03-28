package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
)

//奖励，矿工挖矿给予的奖励
const subsidy = 1000

//交易，包括编号、输入、输出
type Transaction struct {
	ID []byte
	Vin []TXInput
	Vout []TXOutput
}

//序列化，对象到二进制
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded) //编码器
	err := enc.Encode(tx) //编码
	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes() //返回二进制数据
}

//反序列化，二进制到对象
func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction
	decoder := gob.NewDecoder(bytes.NewReader(data)) //解码器
	err := decoder.Decode(&transaction) //解码
	if err != nil {
		log.Panic(err)
	}
	return transaction
}

//对交易进行哈希
func (tx *Transaction) Hash() []byte {
	var hash [32]byte
	txCopy := *tx
	txCopy.ID = []byte{}
	hash = sha256.Sum256(txCopy.Serialize()) //取得二进制进行哈希
	return hash[:]
}

//签名，实际上就是付款方验证自己对输入的所有权，相当于输入银行卡密码来授权进行转账交易
func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTXs map[string]Transaction){
	if tx.IsCoinBase() {
		return //如果是挖矿交易则直接返回，无需签名
	}
	for _,vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("以前的交易不正确!!!!")
		}
	}
	txCopy := tx.TrimmedCopy() //拷贝输入没有私钥的交易副本
	for  inID, vin := range txCopy.Vin {
		//设置签名为空与公钥
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		// dataToSign := fmt.Sprintf("%x\n", txCopy) //要签名的数据

		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[inID].Signature = signature
	}

}

//对用于签名的交易进行裁剪得到的副本
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	for _,vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _,vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}
	txCopy := Transaction{tx.ID, inputs, outputs}
	return txCopy
}

//将对象作为字符串展示出来
func (tx Transaction) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Transaction ID：%x", tx.ID))
	for i, input := range tx.Vin {
		lines = append(lines, fmt.Sprintf("----input index：%d", i))
		lines = append(lines, fmt.Sprintf("----input TXID：%x", input.Txid))
		lines = append(lines, fmt.Sprintf("----input OUT：%d", input.Vout))
		lines = append(lines, fmt.Sprintf("----input Signature：%x", input.Signature))
		lines = append(lines, fmt.Sprintf("----input Pubkey：%x", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("----out index：%d", i))
		lines = append(lines, fmt.Sprintf("----out Value：%d", output.Value))
		lines = append(lines, fmt.Sprintf("----out PubKeyHash：%x", output.PubKeyHash))
	}
	return strings.Join(lines, "\n")
}

//签名验证
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinBase() {
		return true //如果是挖矿交易，直接返回，无需验证签名
	}
	for _,vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("以前的交易不正确!!!!")
		}
	}
	txCopy := tx.TrimmedCopy() //拷贝
	curve := elliptic.P256() //加密算法
	for inID, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil


		r := big.Int{}
		s := big.Int{}
		siglen := len(vin.Signature) //统计签名长度
		r.SetBytes(vin.Signature[:(siglen/2)])
		s.SetBytes(vin.Signature[(siglen/2):])

		x := big.Int{}
		y := big.Int{}
		keylen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keylen/2)])
		y.SetBytes(vin.PubKey[(keylen/2):])

		// dataToVerify := fmt.Sprintf("%x\n", txCopy)
		rawPubkey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubkey, txCopy.ID, &r, &s) == false {
			return false
		}
		// txCopy.Vin[inID].PubKey = nil
	}

	return true
	
}

//检查交易事务是否为coinbase，挖矿得来的奖励，即这个交易是创建该区块时的第一笔交易
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
		// randData := make([]byte, 20)
		// _, err := rand.Read(randData)
		// if err != nil {
		// 	log.Panic(err)
		// }

		// data = fmt.Sprintf("%x", randData)

		data = fmt.Sprintf("奖励给：%s", to)
	}
	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTxOutput(subsidy, to) //输出奖励
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}} //挖矿产生的交易
	tx.ID = tx.Hash() //哈希计算设置交易编号
	return &tx
}

//创建基于钱包地址的转账交易
func NewUTXOTransaction(from, to string, amount int, UTXOSet *UTXOSet) *Transaction {
	var inputs []TXInput //输入
	var outputs []TXOutput //输出

	wallets, err := NewWallets() //打开钱包集合
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from) //通过钱包地址获取钱包
	pubkeyhash := HashPubkey(wallet.PublicKey) //获取公钥哈希
    acc, validOutputs := UTXOSet.FindSpendableOutputs(pubkeyhash, amount)
	if acc < amount {
		log.Panic("交易金额不足!!!!")
	}

	for txid, outs := range validOutputs { //循环遍历无效输出
		txID, err := hex.DecodeString(txid) //解码
		if err != nil {
			log.Panic(err) //处理错误
		}
		for _, out := range outs {
			input := TXInput{txID, out, nil, wallet.PublicKey} //输入
			inputs = append(inputs, input)
		}
	}

	//输出
	outputs = append(outputs, *NewTxOutput(amount, to))

	if acc > amount {
		//记录以后的金额，即多余的金额转回给 from，找零
		outputs = append(outputs, *NewTxOutput(acc-amount, from))
	}

	tx := Transaction{nil, inputs, outputs} //交易
	tx.ID = tx.Hash()
	UTXOSet.blockchain.SignTransaction(&tx, wallet.PrivateKey)
	return &tx
}
