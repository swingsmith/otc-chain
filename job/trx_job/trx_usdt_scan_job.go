package trx_job

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/otc/otc-chain/models"
	"github.com/otc/otc-chain/tron/api"
	"github.com/otc/otc-chain/tron/core"
	"github.com/otc/otc-chain/tron/hexutil"
	"github.com/otc/otc-chain/tron/log"
	"github.com/otc/otc-chain/tron/service"
	"github.com/otc/otc-chain/utils/db_util"
	"github.com/shopspring/decimal"
	"github.com/smirkcat/hdwallet"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"math/big"
	"time"
)

func TrxUsdtScanJob() {
	var (
		start int64
		cur   int64
		end   int64
		step  int64 = 2
		skip  int64 = 30
	)

	client := service.GetTrxClient()
	db := db_util.GetDB()
	tx := db.Begin()

	curHeight, _ := client.GetNowBlockHeight()

	var height = models.Height{Type: "TRX"}
	h, _ := height.GetHeight()
	if h == 0 {
		start = curHeight - skip - step
		cur = curHeight - skip - step
		end = curHeight - skip
	} else if h < curHeight-skip-step {
		start = h + 1
		cur = h + 1
		end = h + 1 + step
	} else {
		tx.Rollback()
		return
	}
	fmt.Printf("TRX扫块开始，start:%d cur:%d end:%d", start, cur, end)
	blocks, err := client.GetBlockByLimitNext(start, end)
	if err != nil {
		tx.Rollback()
		return
	}

	processBlocks(blocks, client, tx)

	cur += step
	if err = tx.Debug().Model(&models.Height{}).Where("type = ?", "TRX").Update("height", cur).Error; err != nil {
		tx.Rollback()
	}

	// 提交事务
	tx.Commit()
}

func processBlocks(blocks *api.BlockListExtention, client *service.GrpcClient, tx *gorm.DB) {
	for _, v := range blocks.Block {
		processBlock(v, client, tx)
	}
}

// 通过translog判断合约转账 如果有转账有扣除，则需调用此方法更精确
var transferid = "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

func processBlock(block *api.BlockExtention, client *service.GrpcClient, tx *gorm.DB) {
	//height := block.GetBlockHeader().GetRawData().GetNumber()

	for _, v := range block.Transactions {
		// transaction.ret.contractRe
		txid := hexutil.Encode(v.Txid)
		// https://tronscan.org/#/transaction/fede1aa9e5c5d7bd179fd62e23bdd11e3c1edd0ca51e41070e34a026d6a42569
		if v.Result == nil || !v.Result.Result {
			continue
		}
		rets := v.Transaction.Ret
		if len(rets) < 1 || rets[0].ContractRet != core.Transaction_Result_SUCCESS {
			continue
		}

		//log.Debugf("process block height %d txid %s", height, txid)
		var transinfo *core.TransactionInfo
		//var fee int64
		// 这里只能有一个
		for _, v1 := range v.Transaction.RawData.Contract {
			if v1.Type == core.Transaction_Contract_TransferContract { //转账合约
				// trx 转账
				//unObj := &core.TransferContract{}
				//err := proto.Unmarshal(v1.Parameter.GetValue(), unObj)
				//if err != nil {
				//	log.Errorf("parse Contract %v err: %v", v1, err)
				//	continue
				//}
				//form := hdwallet.EncodeCheck(unObj.GetOwnerAddress())
				//to := hdwallet.EncodeCheck(unObj.GetToAddress())
				//processTransaction(node, Trx, txid, form, to, height, unObj.GetAmount(), fee)
			} else if v1.Type == core.Transaction_Contract_TriggerSmartContract { //调用智能合约
				// trc20 转账
				unObj := &core.TriggerSmartContract{}
				err := proto.Unmarshal(v1.Parameter.GetValue(), unObj)
				if err != nil {
					log.Errorf("parse Contract %v err: %v", v1, err)
					continue
				}

				contract := hdwallet.EncodeCheck(unObj.GetContractAddress())
				if !IsUsdtContract(contract) {
					continue
				}
				from := hdwallet.EncodeCheck(unObj.GetOwnerAddress())
				data := unObj.GetData()
				// unObj.Data  https://goethereumbook.org/en/transfer-tokens/ 参考eth 操作
				// 只处理 transfer函数产生的交易
				_, _, flag := processTransferData(data, from)
				if flag { // 只有调用了 transfer(address,uint256) 才处理转账
					// 手续费处理 eth 类似 recipt
					if transinfo == nil {
						transinfo, err = client.GetTransactionInfoById(txid)
					}
					if err != nil {
						fmt.Printf("TrxUsdtScanJobGetTransactionInfoById失败,tx:%s, err:%s", txid, err.Error())
						continue
					}

					//fee = transinfo.GetFee()
					// 处理 evenlog 合约转账，如有些合约发起转账并不是全部到账
					// https://tronscan.org/#/address/TWsZk6fs7UisoJAFXiMDXk9aF4PPRzVywZ/transfers
					// https://tronscan.org/#/transaction/0384391ab3ecdf70ffa6e20244718a06b998b8af8a226cda46871dec60b5f14d
					for _, evenlog := range transinfo.Log {
						_, from, to, amount, flag := processEvenlogData(evenlog)
						if flag {
							address := models.Address{Type: "TRX", Address: to}
							act, err := address.IsOurActivatedAddress()
							if err != nil {
								tx.Rollback()
								continue
							}
							if act { //是我们已激活的地址
								processTrc20Transaction(client, tx, txid, from, to, amount)
							}
						}
					}
				}
			} //else if v1.Type == core.Transaction_Contract_TransferAssetContract { //通证转账合约
			//	// trc10 转账
			//	unObj := &core.TransferAssetContract{}
			//	err := proto.Unmarshal(v1.Parameter.GetValue(), unObj)
			//	if err != nil {
			//		log.Errorf("parse Contract %v err: %v", v1, err)
			//		continue
			//	}
			//	contract := hdwallet.EncodeCheck(unObj.GetAssetName())
			//	form := hdwallet.EncodeCheck(unObj.GetOwnerAddress())
			//	to := hdwallet.EncodeCheck(unObj.GetToAddress())
			//	processTransaction(node, contract, txid, form, to, height, unObj.GetAmount(), fee)
			//}
		}
	}
}

func IsUsdtContract(contract string) bool {
	var config models.Config
	addr, _ := config.GetTrxUsdtContractAddress()
	return contract == addr //"TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
}

// 处理合约事件参数
func processEvenlogData(evenlog *core.TransactionInfo_Log) (contract, from, to string, amount int64, flag bool) {
	tmpaddr := evenlog.GetAddress()
	tmpaddr = append([]byte{0x41}, tmpaddr...)
	contract = hdwallet.EncodeCheck(tmpaddr[:])

	amount = new(big.Int).SetBytes(common.TrimLeftZeroes(evenlog.Data)).Int64()

	if len(evenlog.Topics) != 3 {
		flag = false
		return
	}
	if transferid != hexutil.Encode(evenlog.Topics[0]) {
		flag = false
		return
	}
	fmt.Printf("transferid:%s\n", hexutil.Encode(evenlog.Topics[0]))
	if len(evenlog.Topics[1]) != 32 || len(evenlog.Topics[2]) != 32 {
		flag = false
		return
	}
	evenlog.Topics[1][11] = 0x41
	evenlog.Topics[2][11] = 0x41
	from = hdwallet.EncodeCheck(evenlog.Topics[1][11:])
	to = hdwallet.EncodeCheck(evenlog.Topics[2][11:])

	flag = true
	return
}

// 这个结构目前没有用到 只是记录Trc20合约调用对应转换结果
var mapFunctionTcc20 = map[string]string{
	"a9059cbb": "transfer(address,uint256)",
	"70a08231": "balanceOf(address)",
}

// a9059cbb 4 8
// 00000000000000000000004173d5888eedd05efeda5bca710982d9c13b975f98 32 64
// 0000000000000000000000000000000000000000000000000000000000989680 32 64

// 处理合约参数
func processTransferData(trc20 []byte, from string) (to string, amount int64, flag bool) {
	if len(trc20) >= 68 {
		if hexutil.Encode(trc20[:4]) != "a9059cbb" {
			flag = false
			return
		}
		// 多1位41
		trc20[15] = 65 // 0x41
		to = hdwallet.EncodeCheck(trc20[15:36])
		amount = new(big.Int).SetBytes(common.TrimLeftZeroes(trc20[36:68])).Int64()
		flag = true
	}
	return
}

// 处理合约转账参数
func processTransferParameter(to string, amount int64) (data []byte) {
	methodID, _ := hexutil.Decode("a9059cbb")
	addr, _ := hdwallet.DecodeCheck(to)
	paddedAddress := common.LeftPadBytes(addr[1:], 32)
	amountBig := new(big.Int).SetInt64(amount)
	paddedAmount := common.LeftPadBytes(amountBig.Bytes(), 32)
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)
	return
}

// 处理合约获取余额
func processBalanceOfData(trc20 []byte) (amount int64) {
	if len(trc20) >= 32 {
		amount = new(big.Int).SetBytes(common.TrimLeftZeroes(trc20[0:32])).Int64()
	}
	return
}

// 处理合约获取余额参数
func processBalanceOfParameter(addr string) (data []byte) {
	methodID, _ := hexutil.Decode("70a08231")
	add, _ := hdwallet.DecodeCheck(addr)
	paddedAddress := common.LeftPadBytes(add[1:], 32)
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	return
}

func processTrc20Transaction(client *service.GrpcClient, tx *gorm.DB, txid, from, to string, amount int64) {
	// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
	fmt.Println("进入到Transfer事务中")
	var addr models.Address //找用户ID
	if err := tx.Debug().Model(&models.Address{}).Where("type = ? AND address = ?", "TRX", to).First(&addr).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Println("位置1出错")
		tx.Rollback()
	}
	fmt.Printf("当前地址用户ID为：%s", addr.UserId)

	value := decimal.New(amount, -6)

	//插入RechargeRecord
	var rech = models.RechargeRecord{From: from, To: to, Value: value,
		Tx: txid, Type: "TRX_USDT", CreateTime: time.Now()}
	if err := tx.Debug().Model(&models.RechargeRecord{}).Create(&rech).Error; err != nil {
		fmt.Println("位置2出错")
		tx.Rollback()
	}
	//更新用户余额
	var otc models.OtcBalanceInfo
	if err := tx.Debug().Model(&models.OtcBalanceInfo{}).Where("user_id = ? AND coin_type = ?", addr.UserId, 1).First(&otc).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Println("位置3出错")
		tx.Rollback()
	}
	if err := tx.Debug().Model(&models.OtcBalanceInfo{}).Select("available_balance", "version", "update_time").Where("user_id = ? AND coin_type = ? AND version = ?", otc.UserId, otc.CoinType, otc.Version).Updates(models.OtcBalanceInfo{AvailableBalance: otc.AvailableBalance.Add(value), Version: otc.Version + 1, UpdateTime: time.Now()}).Error; err != nil {
		fmt.Println("位置4出错")
		tx.Rollback()
	}

	//更新总表
	var total models.Total
	err := tx.Debug().Model(&models.Total{}).Where("address = ? AND type = ?", addr.Address, "TRX_USDT").First(&total).Error

	if errors.Is(err, gorm.ErrRecordNotFound) { //插入
		total = models.Total{Address: addr.Address, Type: "TRX_USDT", TotalRecharge: value, TotalCollect: decimal.NewFromInt(0), CreateTime: time.Now(), UpdateTime: time.Now()}
		if err = tx.Debug().Model(&models.Total{}).Create(&total).Error; err != nil {
			fmt.Println("位置7出错")
			tx.Rollback()
		}
	} else { //更新
		if err = tx.Debug().Model(&models.Total{}).Select("total_recharge", "version", "update_time").Where("address = ? AND type = ? AND version = ?", total.Address, total.Type, total.Version).Updates(models.Total{TotalRecharge: total.TotalRecharge.Add(value), Version: total.Version + 1, UpdateTime: time.Now()}).Error; err != nil {
			fmt.Println("位置6出错")
			tx.Rollback()
		}
	}
	fmt.Println("进入到Transfer事务结束")
}
