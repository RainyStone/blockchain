package main

import "fmt"

func (cli *CLI) reindexUTXO(nodeID string) {
	blockchain := NewBlockChain(nodeID)
	UTXOSet := UTXOSet{blockchain}
	UTXOSet.Reindex()
	count := UTXOSet.CountTransactions()
	fmt.Printf("重建索引成功，已经有 %d 次交易在UTXO集合中\n", count)
}