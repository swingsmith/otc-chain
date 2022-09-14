package trx_job

import (
	"fmt"
	"github.com/otc/otc-chain/models"
	trxService "github.com/otc/otc-chain/tron/service"
	"github.com/otc/otc-chain/utils/db_util"
	"github.com/shopspring/decimal"
	"time"
)

func TrxUsdtCollectJob() {
	fmt.Printf("TrxUsdtCollectJob归集任务开始")
	var offset = 0
	var limit = 10
	var config models.Config
	usdtContractAddress, _ := config.GetTrxUsdtContractAddress()
	collectAddress, _ := config.GetTrxUsdtCollectAddress()
	threshold, _ := config.GetTrxUsdtCollectThreshold()
	th, _ := decimal.NewFromString(threshold)

	client := trxService.GetTrxClient()

	db := db_util.GetDB()

	for {
		var totals []models.Total
		db.Debug().Model(&models.Total{}).Where("type = ?", "TRX_USDT").Offset(offset).Limit(limit).Find(&totals)

		if len(totals) > 0 {
			for _, total := range totals {
				fmt.Printf("total address:%s\n", total.Address)
				balance := total.TotalRecharge.Sub(total.TotalCollect)
				fmt.Printf("当前地址为:%s USDT余额为：%s", total.Address, balance.String())
				if balance.Cmp(th) >= 0 { //超过8美元
					abalance, _ := client.GetBalanceByAddress(total.Address)
					fee := decimal.New(15, 6)

					if abalance.Cmp(fee) >= 0 { //余额大于等于fee
						tx := db.Begin()
						var t string
						var err error
						var address models.Address
						fmt.Println("tx.Debug().Model(&models.Address{})开始")
						tx.Debug().Model(&models.Address{}).Where("address = ? AND type = ?", total.Address, "TRX").First(&address)
						//fmt.Printf("EthUsdtCollectJob当前地址为：%s  私钥为：%s",address.Address, address.PrivateKey)
						ba := balance.Mul(decimal.New(1, 6))
						feeLimit := decimal.New(15, 6)
						if t, err = client.TransferContract(address.PrivateKey, usdtContractAddress, collectAddress, ba.IntPart(), feeLimit.IntPart()); err != nil {
							fmt.Printf("归集ETH_USDT失败,地址为：%s err:%v", total.Address, err)
							tx.Rollback()
						}
						//插入记录表
						var collect = models.CollectRecord{From: total.Address, To: collectAddress, Value: balance, Tx: t, Type: "TRX_USDT", CreateTime: time.Now()}
						if err := tx.Debug().Model(&models.CollectRecord{}).Create(&collect).Error; err != nil {
							tx.Rollback()
						}

						//更新总表
						if err = tx.Debug().Model(&models.Total{}).Select("total_collect", "version", "update_time").Where("address = ? AND type = ? AND version = ?", total.Address, total.Type, total.Version).Updates(models.Total{TotalCollect: total.TotalCollect.Add(balance), Version: total.Version + 1, UpdateTime: time.Now()}).Error; err != nil {
							tx.Rollback()
						}

						tx.Commit()
					}
				}
			}
		}

		offset += limit
		if len(totals) < limit {
			break
		}
	}

	//更新设置状态
	if err := db.Debug().Model(&models.Config{}).Where("config_key = ?", "TRX_USDT_COLLECT_STATUS").Update("config_value", "0").Error; err != nil {
	}

	fmt.Printf("TrxUsdtCollectJob归集任务结束")
}
