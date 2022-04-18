package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const walletFile = "wallet_%s.dat" //钱包文件

type Wallets struct {
	Wallets map[string]*Wallet //一个字符串对应一个钱包

}

//创建 Wallets ，从文件中获取已经存在的 Wallets
func NewWallets(nodeID string) (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)
	err := wallets.LoadFromFile(nodeID)
	return &wallets, err
}

//创建一个钱包
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet() //创建钱包
	address := fmt.Sprintf("%s", wallet.GetAddress())
	ws.Wallets[address] = wallet //保存钱包地址
	return address
}

//获取所有钱包地址
func (ws *Wallets) GetAddresses() []string {
	var addresses []string
	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}
	return addresses //返回所有钱包地址
}

//获取一个钱包
func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

//从文件中获取 Wallets
func (ws *Wallets) LoadFromFile(nodeID string) error {
	mywalletfile := fmt.Sprintf(walletFile, nodeID) //获取文件名
	if _, err := os.Stat(mywalletfile); os.IsNotExist(err){
		return err
	}
	fileContent, err := ioutil.ReadFile(mywalletfile) //读取文件内容
	if err != nil {
		log.Panic(err)
	}
	//读取文件二进制内容并解密
	var wallets Wallets
	gob.Register(elliptic.P256()) //注册解密算法
	decoder := gob.NewDecoder(bytes.NewReader(fileContent)) //解码
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}
	ws.Wallets = wallets.Wallets
	return nil
}

//Wallets 保存到文件
func (ws *Wallets) SaveToFile(nodeID string) {
	var content bytes.Buffer
	mywalletfile := fmt.Sprintf(walletFile, nodeID) //获取文件名
	gob.Register(elliptic.P256()) //注册加密算法
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}
	err = ioutil.WriteFile(mywalletfile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
