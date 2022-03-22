package main

import (
	"bytes"
	"encoding/binary"
	"log"
)

//整数转化为十六进制
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	//返回字节集
	return buff.Bytes()
}

func ReverseBytes(data []byte) {
	for i,j:=0, len(data)-1; i<j; i,j=i+1,j-1 {
		data[i], data[j] = data[j], data[i]
	}
}