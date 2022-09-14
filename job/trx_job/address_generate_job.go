package trx_job

import (
	"fmt"
	"github.com/otc/otc-chain/models"
	"github.com/otc/otc-chain/tron/service"
)

func TrxAddressGenerateJob() {
	var address = models.Address{Type: "TRX"}
	var unUsedSize int
	var err error
	if unUsedSize, err = address.GetUnusedAddressSize(); err != nil {
		fmt.Println("address.GetUnusedAddressSize()出错")
	}

	if unUsedSize < 100 {
		for i := 0; i < 100; i++ {
			addr, prvKey, err := service.GetNewAccount()
			if err != nil {
				fmt.Printf("生成TRX地址出错,err:%s", err.Error())
			}

			address = models.Address{Type: "TRX", Address: addr, PrivateKey: prvKey, Status: 0}
			if _, err := address.InsertAddress(); err != nil {
				fmt.Println("address.InsertAddress()失败")
			}

			fmt.Printf("插入TRX地址成功：Address: %s , PrivateKey: %s ", addr, prvKey)
		}
	}
}
