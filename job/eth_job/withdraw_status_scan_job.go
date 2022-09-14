package eth_job

import (
	"fmt"
	"github.com/otc/otc-chain/models"
	ethService "github.com/otc/otc-chain/service/eth_service"
	"github.com/otc/otc-chain/utils/db_util"
	"time"
)

func WithdrawStatusScanJob()  {
	fmt.Printf("EthWithdrawStatusScanJob任务开始")
	var offset = 0
	var limit = 100
	var count = 6*30

	db := db_util.GetDB()

	for {
		var withdraws []models.WithdrawRecord
		db.Debug().Model(&models.WithdrawRecord{}).Where("type = ? AND status = ?", "ETH_USDT", 0).Offset(offset).Limit(limit).Find(&withdraws)

		if len(withdraws) > 0 {
			for _, withdraw := range withdraws {
				tx := db.Begin()
				if withdraw.Count < count {
					b, err := ethService.GetTransactionReceipt(withdraw.Tx)
					//if err != nil {//出错直接失败
					//	fmt.Printf("WithdrawStatusScanJob Error: %s", err.Error())
					//	if err = tx.Debug().Model(&models.WithdrawRecord{}).Select("status", "count", "version", "update_time").Where("tx = ? AND type = ? AND version = ?", withdraw.Tx, withdraw.Type, withdraw.Version).Updates(models.WithdrawRecord{Status: 2, Count: withdraw.Count+1, Version: withdraw.Version+1, UpdateTime: time.Now()}).Error; err != nil {
					//		tx.Rollback()
					//	}
					//}else
					if b {//成功
						if err = tx.Debug().Model(&models.WithdrawRecord{}).Select("status", "count", "version", "update_time").Where("tx = ? AND type = ? AND version = ?", withdraw.Tx, withdraw.Type, withdraw.Version).Updates(models.WithdrawRecord{Status: 1, Count: withdraw.Count+1, Version: withdraw.Version+1, UpdateTime: time.Now()}).Error; err != nil {
							tx.Rollback()
						}
					}else {//+1
						if err = tx.Debug().Model(&models.WithdrawRecord{}).Select("count", "version", "update_time").Where("tx = ? AND type = ? AND version = ?", withdraw.Tx, withdraw.Type, withdraw.Version).Updates(models.WithdrawRecord{Count: withdraw.Count+1, Version: withdraw.Version+1, UpdateTime: time.Now()}).Error; err != nil {
							tx.Rollback()
						}
					}
				}else {//超过次数判定为失败
					if err := tx.Debug().Model(&models.WithdrawRecord{}).Select("status", "count", "version", "update_time").Where("tx = ? AND type = ? AND version = ?", withdraw.Tx, withdraw.Type, withdraw.Version).Updates(models.WithdrawRecord{Status: 2, Count: withdraw.Count+1, Version: withdraw.Version+1, UpdateTime: time.Now()}).Error; err != nil {
						tx.Rollback()
					}
				}
				tx.Commit()
			}
		}

		offset += limit
		if len(withdraws) < limit{
			break
		}
	}

	fmt.Printf("EthWithdrawStatusScanJob任务结束")
}
