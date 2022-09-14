package trx_job

import (
	"encoding/json"
	"fmt"
	"github.com/otc/otc-chain/models"
	trxService "github.com/otc/otc-chain/tron/service"
	"github.com/otc/otc-chain/utils/db_util"
	"github.com/shopspring/decimal"
	"time"
)

func TrxGasJob() {
	fmt.Printf("Trx转gas任务开始")
	now := time.Now()
	t1 := now.Format("2006-01-02 15:04:05")
	fmt.Printf("gasjobstart开始时间为：%s", t1)
	var offset = 0
	var limit = 10
	var config = models.Config{}
	gasPrivateKey, _ := config.GetTrxGasPrivateKey()
	gasAddress, _ := config.GetTrxGasAddress()

	threshold, _ := config.GetTrxUsdtCollectThreshold()
	th, _ := decimal.NewFromString(threshold)

	db := db_util.GetDB()
	client := trxService.GetTrxClient()

	//更新设置状态
	if err := db.Debug().Model(&models.Config{}).Where("config_key = ?", "TRX_USDT_COLLECT_STATUS").Update("config_value", "1").Error; err != nil {
	}

	for {
		var totals []models.Total

		var count int64
		db.Debug().Model(&models.Total{}).Select("COUNT(*)").Where("type = ?", "TRX_USDT").Find(&count)
		fmt.Printf("totals总条数为*************：%d", count)
		db.Debug().Model(&models.Total{}).Where("type = ?", "TRX_USDT").Offset(offset).Limit(limit).Find(&totals)

		b, _ := json.Marshal(totals)
		fmt.Printf("totals***********************\n%s\n", string(b))

		if len(totals) > 0 {
			for _, total := range totals {
				balance := total.TotalRecharge.Sub(total.TotalCollect)
				fmt.Printf("当前地址为:%s USDT余额为：%s\n", total.Address, balance.String())
				if balance.Cmp(th) >= 0 { //超过8美元
					fmt.Printf("进入到判断超过8美元\n")
					abalance, err := client.GetBalanceByAddress(total.Address)
					if err != nil { //未激活
						abalance = decimal.Zero
					}
					fmt.Printf("TrxGasJob当前地址为*********：%s, 余额为 %s", total.Address, abalance.Div(decimal.New(1, 6)))
					fee := decimal.New(15, 6)
					fmt.Printf("当前地址余额为：%s  fee: %s", abalance.String(), fee.String())
					fmt.Printf("")
					if abalance.Cmp(fee) < 0 { //余额小于fee
						fmt.Printf("进入到判断用户地址余额小于fee\n")
						sendFee := decimal.New(2, 6)
						gasBalance, _ := client.GetBalanceByAddress(gasAddress)
						if gasBalance.Cmp(sendFee.Add(fee).Sub(abalance)) >= 0 {
							tx := db.Begin()
							fmt.Printf("进入到判断gas账户余额是否足够\n")
							var t string
							var err error
							fmt.Printf("gasjob client.Transfer开始，地址为：%s\n", total.Address)

							if t, err = client.Transfer(gasPrivateKey, total.Address, fee.Sub(abalance).IntPart()); err != nil { //转gas
								fmt.Printf("转gas失败，转入地址为：%s err:%v \n", total.Address, err)
								tx.Rollback()
							}
							fmt.Printf("gasjob client.Transfer开始,事务为：%s\n", t)
							fmt.Printf("gasjob client.Transfer结束，地址为：%s\n", total.Address)
							//插入记录表
							var gas = models.GasRecord{From: gasAddress, To: total.Address, Value: fee.Sub(abalance).Div(decimal.New(1, 6)), Tx: t, Type: "TRX", CreateTime: time.Now()}
							if err = tx.Debug().Model(&models.GasRecord{}).Create(&gas).Error; err != nil {
								tx.Rollback()
							}
							tx.Commit()
						}
					}
				}
			}
		}

		offset += limit
		if len(totals) < limit {
			break
		}
	}
	now2 := time.Now()
	t2 := now2.Format("2006-01-02 15:04:05")
	fmt.Printf("结束时间为：%s", t2)
	fmt.Printf("Trx转gas任务结束")
}
