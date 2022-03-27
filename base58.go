package main

import (
	"bytes"
	"fmt"
	"math/big"
)

//字母表格，最终会展示的字符
var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func Base58Encode(input []byte) []byte {
	// var result []byte
    // x := big.NewInt(0).SetBytes(input) //输入的数据转为二进制

	// base := big.NewInt(int64(len(b58Alphabet)))
	// zero := big.NewInt(0)
	// mod := &big.Int{}
	
	// for x.Cmp(zero) != 0{
	// 	x.DivMod(x, base, mod) //mod，余数赋值
	// 	result = append(result, b58Alphabet[mod.Int64()])
	// }

	// //字节集反转
	// ReverseBytes(result)
	// for _, myb := range input {
	// 	if myb == 0x00 {
	// 		result = append([]byte{b58Alphabet[0]}, result...) //使用可变长参数追加
	// 	}else {
	// 		break
	// 	}
	// }

	// return result

	var result []byte

	x := big.NewInt(0).SetBytes(input) //输入的数据转为二进制

	base := big.NewInt(int64(len(b58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod) //mod，余数赋值
		result = append(result, b58Alphabet[mod.Int64()])
	}

	// https://en.bitcoin.it/wiki/Base58Check_encoding#Version_bytes
	if input[0] == 0x00 {
		result = append(result, b58Alphabet[0])
	}
	//字节集反转
	ReverseBytes(result)

	return result
}

func Base58Decode(input []byte) []byte {
	// result := big.NewInt(0)
	// zeroBytes := 0
	// for _, b := range input {
	// 	if b == 0x00 {
	// 		zeroBytes++
	// 	}
	// }
	// payload := input[zeroBytes:] //取出字节
	// for _, b := range payload {
	// 	charIndex := bytes.IndexByte(b58Alphabet, b) //字母表格
	// 	result.Mul(result, big.NewInt(58)) //乘法
	// 	result.Add(result, big.NewInt(int64(charIndex))) //加法
	// }
	// decoded := result.Bytes()
	// decoded = append(bytes.Repeat([]byte{byte(0x00)}, zeroBytes), decoded...)
	// return decoded

	result := big.NewInt(0)

	for _, b := range input {
		charIndex := bytes.IndexByte(b58Alphabet, b) //字母表格
		result.Mul(result, big.NewInt(58)) //乘法
		result.Add(result, big.NewInt(int64(charIndex))) //加法
	}

	decoded := result.Bytes()

	if input[0] == b58Alphabet[0] {
		decoded = append([]byte{0x00}, decoded...)
	}

	return decoded
}

//base58编码与解码测试函数
func mainTestBase58(){
	fmt.Printf("%v\n", Base58Encode([]byte("abcdefg")))
	fmt.Printf("%s\n", Base58Decode(Base58Encode([]byte("abcdefg"))))
}