package models

import (
	"github.com/shopspring/decimal"
	"time"
)

// 用户账户表
type OtcBalanceInfo struct {
	Id	int	`json:"id" gorm:"column:id"`	
	UserId	string	`json:"user_id" gorm:"column:user_id"`	
	AvailableBalance	decimal.Decimal	`json:"available_balance" gorm:"column:available_balance"`
	FreezeBalance	decimal.Decimal	`json:"freeze_balance" gorm:"column:freeze_balance"`
	CoinType	int	`json:"coin_type" gorm:"column:coin_type"`	
	Version	int	`json:"version" gorm:"column:version"`	
	UpdateTime	time.Time	`json:"update_time" gorm:"column:update_time"`	
	CreateTime	time.Time	`json:"create_time" gorm:"column:create_time"`	
}
func (e *OtcBalanceInfo) TableName() string { 
    return "otc_balance_info"
}
