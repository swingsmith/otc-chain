package models

import (
	"github.com/shopspring/decimal"
	"time"
)

type RechargeRecord struct {
	Id	int	`json:"id" gorm:"column:id"`	
	From	string	`json:"from" gorm:"column:from"`	
	To	string	`json:"to" gorm:"column:to"`	
	Value	decimal.Decimal	`json:"value" gorm:"column:value"`
	Tx	string	`json:"tx" gorm:"column:tx"`	
	Type	string	`json:"type" gorm:"column:type"`	
	CreateTime	time.Time	`json:"create_time" gorm:"column:create_time"`	
}
func (e *RechargeRecord) TableName() string { 
    return "recharge_record"
}
