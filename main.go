package main

import (
	_ "fmt"
	_ "strconv"
)

func main(){

	cli := CLI{} //创建命令行
	cli.Run() //开启 
}