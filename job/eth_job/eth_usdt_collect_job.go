package eth_job

import (
	"fmt"
	"github.com/otc/otc-chain/models"
	ethService "github.com/otc/otc-chain/service/eth_service"
	"github.com/otc/otc-chain/utils/db_util"
	"github.com/shopspring/decimal"
	"time"
)

func EthUsdtCollectJob()  {
	fmt.Printf("EthUsdtCollectJob归集任务开始")
	var offset = 0
	var limit = 10
	var config models.Config
	usdtContractAddress, _ := config.GetEthUsdtContractAddress()
	collectAddress, _ := config.GetEthUsdtCollectAddress()
	threshold, _ := config.GetEthUsdtCollectThreshold()
	th, _ := decimal.NewFromString(threshold)

	db := db_util.GetDB()

	for {
		var totals []models.Total
		db.Debug().Model(&models.Total{}).Where("type = ?", "ETH_USDT").Offset(offset).Limit(limit).Find(&totals)

		if len(totals) > 0 {
			for _, total := range totals {
				fmt.Printf("total address:%s\n",total.Address)
				balance := total.TotalRecharge.Sub(total.TotalCollect)
				fmt.Printf("当前地址为:%s USDT余额为：%s", total.Address, balance.String())
				if balance.Cmp(th) >= 0 { //超过100000美元
					abalance, _ := ethService.GetAccountBalance(total.Address)
					fee, _ := ethService.GetTransferGasFee()

					if decimal.NewFromBigInt(abalance,0).Cmp(fee) >= 0{	//余额大于等于fee
						tx := db.Begin()
						var t string
						var err error
						var address models.Address
						fmt.Println("tx.Debug().Model(&models.Address{})开始")
						tx.Debug().Model(&models.Address{}).Where("address = ? AND type = ?", total.Address, "ETH").First(&address)
						//fmt.Printf("EthUsdtCollectJob当前地址为：%s  私钥为：%s",address.Address, address.PrivateKey)
						ba := balance.Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt(6)))
						if t, err = ethService.TransferToken(usdtContractAddress, address.PrivateKey, collectAddress, ba.BigInt()); err != nil{
							fmt.Printf("归集ETH_USDT失败,地址为：%s err:%v", total.Address, err)
							tx.Rollback()
						}
						//插入记录表
						var collect = models.CollectRecord{From: total.Address, To: collectAddress, Value: balance, Tx: t, Type: "ETH_USDT", CreateTime: time.Now()}
						if err := tx.Debug().Model(&models.CollectRecord{}).Create(&collect).Error; err != nil {
							tx.Rollback()
						}

						//更新总表
						if err = tx.Debug().Model(&models.Total{}).Select("total_collect", "version", "update_time").Where("address = ? AND type = ? AND version = ?", total.Address, total.Type, total.Version).Updates(models.Total{TotalCollect: total.TotalCollect.Add(balance), Version: total.Version+1, UpdateTime: time.Now()}).Error; err != nil {
							tx.Rollback()
						}

						tx.Commit()
					}
				}
			}
		}

		offset += limit
		if len(totals) < limit{
			break
		}
	}

	//更新设置状态
	if err := db.Debug().Model(&models.Config{}).Where("config_key = ?", "ETH_USDT_COLLECT_STATUS").Update("config_value","0").Error; err != nil {
	}

	fmt.Printf("EthUsdtCollectJob归集任务结束")
}
