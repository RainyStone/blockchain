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