package eth_job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	token "github.com/otc/otc-chain/contracts_erc20"
	"github.com/otc/otc-chain/models"
	"github.com/otc/otc-chain/utils/db_util"
	"github.com/otc/otc-chain/utils/eth_util"
	"gorm.io/gorm"
	"math/big"
	"strings"
	"time"

	service "github.com/otc/otc-chain/service/eth_service"
	"github.com/shopspring/decimal"
)

// LogTransfer ..
type LogTransfer struct {
	From   common.Address
	To     common.Address
	Tokens *big.Int
}

// LogApproval ..
type LogApproval struct {
	TokenOwner common.Address
	Spender    common.Address
	Tokens     *big.Int
}

func EthUsdtScanJob() {
	fmt.Println("EthJob")
	var start int64
	var end int64
	var current int64
	var skip int64 = 20
	var step int64 = 2
	client := eth_util.GetEthClient()
	db := db_util.GetDB()
	hCur, _ := service.GetNowBlockHeight()
	tx := db.Begin()

	var h = models.Height{Type: "ETH"}
	if height, _ := h.GetHeight(); height == 0 {
		start = hCur - skip - step
		current = hCur - skip - step
		end = hCur - skip
	} else if height < hCur-skip-step {
		start = height + 1
		current = height + 1
		end = height + 1 + step
	} else {
		tx.Rollback()
		return
	}

	var config models.Config
	contractAddr, _ := config.GetEthUsdtContractAddress()
	contractAddress := common.HexToAddress(contractAddr)

	// 0x Protocol (ZRX) token address
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(start),
		ToBlock:   big.NewInt(end),
		Addresses: []common.Address{
			contractAddress,
		},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		//log.Fatal(err)
	}
	//fmt.Println(logs)

	contractAbi, err := abi.JSON(strings.NewReader(string(token.TokenABI)))
	if err != nil {
		//log.Fatal(err)
	}

	logTransferSig := []byte("Transfer(address,address,uint256)")
	//LogApprovalSig := []byte("Approval(address,address,uint256)")
	logTransferSigHash := crypto.Keccak256Hash(logTransferSig)
	//logApprovalSigHash := crypto.Keccak256Hash(LogApprovalSig)

	for _, vLog := range logs {
		fmt.Printf("Log Block Number: %d\n", vLog.BlockNumber)
		fmt.Printf("Log Index: %d\n", vLog.Index)
		fmt.Printf("Log tx: %s\n", vLog.TxHash)

		switch vLog.Topics[0].Hex() {
		case logTransferSigHash.Hex():
			fmt.Printf("进入到Transfer事件\n")
			fmt.Printf("Log Name: Transfer\n")

			var transferEvent LogTransfer

			data, err := contractAbi.Unpack("Transfer", vLog.Data)
			if err != nil {
				//log.Fatal(err)
			}
			//fmt.Println(data)

			transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
			transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())

			vstr, _ := json.Marshal(data[0])
			transferEvent.Tokens, _ = new(big.Int).SetString(string(vstr), 10)

			fmt.Printf("From: %s\n", transferEvent.From.Hex())
			fmt.Printf("To: %s\n", transferEvent.To.Hex())
			fmt.Printf("Tokens: %s\n", transferEvent.Tokens.String())

			var act bool
			var address = models.Address{Type: "ETH", Address: transferEvent.To.Hex(), Status: 1}
			if act, err = address.IsOurActivatedAddress(); err != nil {
				fmt.Printf("不是我们的用户地址：%s", transferEvent.To.Hex())
			}
			if act {
				// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
				fmt.Println("进入到Transfer事务中")
				var addr models.Address //找用户ID
				if err = tx.Debug().Model(&models.Address{}).Where("type = ? AND address = ?", "ETH", transferEvent.To.Hex()).First(&addr).Error; errors.Is(err, gorm.ErrRecordNotFound) {
					fmt.Println("位置1出错")
					tx.Rollback()
				}
				fmt.Printf("当前地址用户ID为：%s", addr.UserId)

				value := decimal.NewFromBigInt(transferEvent.Tokens, 0).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(6)))

				//插入RechargeRecord
				var rech = models.RechargeRecord{From: transferEvent.From.Hex(), To: transferEvent.To.Hex(), Value: value,
					Tx: vLog.TxHash.Hex(), Type: "ETH_USDT", CreateTime: time.Now()}
				if err = tx.Debug().Model(&models.RechargeRecord{}).Create(&rech).Error; err != nil {
					fmt.Println("位置2出错")
					tx.Rollback()
				}
				//更新用户余额
				var otc models.OtcBalanceInfo
				if err = tx.Debug().Model(&models.OtcBalanceInfo{}).Where("user_id = ? AND coin_type = ?", addr.UserId, 1).First(&otc).Error; errors.Is(err, gorm.ErrRecordNotFound) {
					fmt.Println("位置3出错")
					tx.Rollback()
				}
				if err = tx.Debug().Model(&models.OtcBalanceInfo{}).Select("available_balance", "version", "update_time").Where("user_id = ? AND coin_type = ? AND version = ?", otc.UserId, otc.CoinType, otc.Version).Updates(models.OtcBalanceInfo{AvailableBalance: otc.AvailableBalance.Add(value), Version: otc.Version + 1, UpdateTime: time.Now()}).Error; err != nil {
					fmt.Println("位置4出错")
					tx.Rollback()
				}

				//更新总表
				var total models.Total
				err = tx.Debug().Model(&models.Total{}).Where("address = ? AND type = ?", addr.Address, "ETH_USDT").First(&total).Error

				if errors.Is(err, gorm.ErrRecordNotFound) { //插入
					total = models.Total{Address: addr.Address, Type: "ETH_USDT", TotalRecharge: value, TotalCollect: decimal.NewFromInt(0),CreateTime: time.Now(), UpdateTime: time.Now()}
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
		}
		fmt.Printf("\n")
	}

	current += step
	if err = tx.Debug().Model(&models.Height{}).Where("type = ?", "ETH").Update("height", current).Error; err != nil {
		tx.Rollback()
	}

	// 提交事务
	tx.Commit()
}
