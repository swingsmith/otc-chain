package models

import (
	"github.com/shopspring/decimal"
	"time"
)

type Total struct {
	Id	int	`json:"id" gorm:"column:id"`	
	Address	string	`json:"address" gorm:"column:address"`	
	Type	string	`json:"type" gorm:"column:type"`	// 哪条链的哪个币种
	TotalRecharge	decimal.Decimal	`json:"total_recharge" gorm:"column:total_recharge"`
	TotalCollect	decimal.Decimal	`json:"total_collect" gorm:"column:total_collect"`
	Version	int64	`json:"version" gorm:"column:version"`
	CreateTime	time.Time	`json:"create_time" gorm:"column:create_time"`
	UpdateTime	time.Time	`json:"update_time" gorm:"column:update_time"`
}
func (e *Total) TableName() string { 
    return "total"
}