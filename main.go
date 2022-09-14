package main

import (
	"github.com/otc/otc-chain/job/eth_job"
	"github.com/otc/otc-chain/job/trx_job"
)

func main() {
	//****************ETH USDT
	eth_job.EthMainJob()
	trx_job.TrxMainJob()
	//client := service.GetTrxClient()
	//blocks, err := client.GetBlockByLatestNum(1)
	//if err != nil {
	//	fmt.Println("TRX扫块出错")
	//	return
	//}
	//fmt.Println("扫块开始")
	//for i, block := range blocks.Block {
	//	fmt.Printf("第%d个块\n", i)
	//	fmt.Println(block)
	//}
	//trans, err := client.GetTransactionInfoById("3f6a55570c911863df91a42468e21c7b5b01f4e8e998e952d5ccd9b16ed535cc")
	//if err != nil {
	//	fmt.Println("GetTransactionById失败")
	//	return
	//}
	//fmt.Println(trans)

	select {}
}
