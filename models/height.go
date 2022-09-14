package models

import (
	"errors"
	"github.com/otc/otc-chain/utils/db_util"
	"gorm.io/gorm"
)

type Height struct {
	Id     int    `json:"id" gorm:"column:id"`
	Type   string `json:"type" gorm:"column:type"`
	Height int    `json:"height" gorm:"column:height"`
}

func (e *Height) TableName() string {
	return "height"
}

func (e *Height) GetHeight() (int64, error) {
	db := db_util.GetDB()
	var height Height
	tx := db.Debug().Model(&Height{}).Where("type = ?", e.Type).First(&height)
	//fmt.Printf("当前ETH高度为：%s", strconv.FormatInt(int64(height.Height),10))

	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return 0, tx.Error
	} else {
		return int64(height.Height), nil
	}
}
