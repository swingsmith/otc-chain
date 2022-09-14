package eth_job

import (
	"fmt"
	"github.com/otc/otc-chain/models"
	service "github.com/otc/otc-chain/service/eth_service"
)

func EthAddressGenerateJob()  {
	var address = models.Address{Type: "ETH"}
	var unUsedSize int
	var err error
	if unUsedSize, err = address.GetUnusedAddressSize(); err != nil{
		fmt.Println("address.GetUnusedAddressSize()出错")
	}
	
	if unUsedSize < 100 {
		for i := 0; i < 100; i++ {
			addr, prvKey, err := service.GetNewAccount()
			if err != nil {
				fmt.Printf("service.GetNewAccount()出错")
			}

			address = models.Address{Type: "ETH",Address: addr, PrivateKey: prvKey, Status: 0}
			if _, err := address.InsertAddress(); err != nil{
				fmt.Println("address.InsertAddress()失败")
			}

			fmt.Printf("插入ETH地址成功：Address: %s , PrivateKey: %s ",addr, prvKey)
		}
	}
}
