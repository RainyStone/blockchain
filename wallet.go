package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const version = byte(0x00)      //钱包版本
const walletfile = "wallet.dat" //钱包文件
const addressChecksumlen = 4    //检测地址长度

type Wallet struct {
	PrivateKey ecdsa.PrivateKey //钱包的权限，相等于银行卡密码
	PublicKey []byte //收款地址，相当于银行卡号
}

//创建一个钱包
func NewWallet() *Wallet {
	private, public := newKeyPair() //创建公钥私钥
	wallet := Wallet{private, public} //创建钱包
	return &wallet
}

//创建公钥和私钥对
func newKeyPair() (ecdsa.PrivateKey, []byte){
	curve := elliptic.P256() //创建加密算法
	private, err := ecdsa.GenerateKey(curve, rand.Reader) //生成私钥
	if err != nil {
		log.Panic(err)
	}

	publickey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...) //生成公钥
	return *private, publickey
}

//公钥的校验，从代码来看，应该是对数据两次加密后，获取前 addressChecksumlen 位
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])
	return secondSHA[:addressChecksumlen]
}

//公钥哈希处理
func HashPubkey(pubkey []byte) []byte {
	publicsha256 := sha256.Sum256(pubkey) //处理公钥
	R160Hash := ripemd160.New() //创建一个哈希算法对象
	_, err := R160Hash.Write(publicsha256[:])
	if err != nil {
		log.Panic(err)
	}
	publicR160Hash := R160Hash.Sum(nil)
	return publicR160Hash
}

//获取钱包的地址
func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubkey(w.PublicKey) //取得钱包公钥哈希值

	versionPayload := append([]byte{version}, pubKeyHash...) //加入版本信息
	checksum := checksum(versionPayload) //检测版本与公钥
	fullpayload := append(versionPayload, checksum...) //叠加校验信息

	address := Base58Encode(fullpayload)
	return address //返回钱包地址

}

//校验钱包地址
func ValidateAddress(address string) bool {
	publicHash := Base58Decode([]byte(address))
	actualchecksum := publicHash[len(publicHash) - addressChecksumlen : ]
	version := publicHash[0] //取得版本
	publicHash = publicHash[1 : len(publicHash) - addressChecksumlen]
	targetCheckSum := checksum(append([]byte{version}, publicHash...))
	return bytes.Compare(actualchecksum, targetCheckSum) == 0
}