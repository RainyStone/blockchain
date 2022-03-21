package main

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"
)

//挖矿原理：通俗解释，计算出来的哈希值符合某个条件即算作成功
func mainx(){
	start := time.Now() //当前时间
	for i:=0;i<100000;i++{ //循环挖矿
		data := sha256.Sum256([]byte(strconv.Itoa(i))) //计算哈希
		fmt.Printf("%10d,%x\n", i, data)
		fmt.Printf("%s\n", string(data[len(data)-2:]))

		if string(data[len(data)-1 : ]) == "0" { //哈希的位数匹配
			usedtime := time.Since(start)
			fmt.Printf("挖矿成功%d Ms", usedtime)
			break
		}
	}
}