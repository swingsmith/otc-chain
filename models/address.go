package models

import (
	"errors"
	"fmt"
	"github.com/otc/otc-chain/utils/db_util"
	"gorm.io/gorm"
	"strconv"
)

type Address struct {
	Id         int    `json:"id" gorm:"column:id"`
	UserId     string `json:"user_id" gorm:"column:user_id"`
	Address    string `json:"address" gorm:"column:address"`
	PrivateKey string `json:"private_key" gorm:"column:private_key"`
	Type       string `json:"type" gorm:"column:type"`
	Status     int    `json:"status" gorm:"column:status"`
}

func (e *Address) TableName() string {
	return "address"
}

func (e *Address) IsOurActivatedAddress() (bool, error) {
	var address Address
	db := db_util.GetDB()
	tx := db.Debug().Model(&Address{}).Where("type = ? AND address = ? AND status = ?", e.Type, e.Address, 1).First(&address)
	fmt.Printf("IsOurActivatedAddress影响行数为：%d", tx.RowsAffected)

	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		fmt.Printf("IsOurActivatedAddressNot当前地址不是我们的用户,地址：%s", e.Address)
		return false, nil
	} else {
		fmt.Printf("IsOurActivatedAddress当前地址是我们的用户，ID为：%s , 地址：%s", address.UserId, address.Address)
		return true, nil
	}
}

func (e *Address) GetUserIdByAddress() (string, error) {
	var address Address
	db := db_util.GetDB()
	tx := db.Debug().Model(&Address{}).Where("type = ? AND address = ?", e.Type, e.Address).First(&address)

	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return "", tx.Error
	} else {
		fmt.Printf("当前地址是我们的用户，ID为：%s , 地址：%s", address.UserId, address.Address)
		return address.UserId, nil
	}
}

func (e *Address) GetUnusedAddressSize() (int, error) {
	var count int
	db := db_util.GetDB()
	err := db.Debug().Model(&Address{}).Select("COUNT(*)").Where("type = ? AND status = ?", e.Type, 0).Find(&count).Error

	//fmt.Printf("当前未使用的ETH 地址总数为：%d", count)
	fmt.Println()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 100, err
	} else {
		return count, nil
	}
}

func (e *Address) InsertAddress() (int, error) {
	db := db_util.GetDB()
	tx := db.Debug().Model(&Address{}).Create(e)
	fmt.Println("InsertEthAddress影响行数" + strconv.FormatInt(tx.RowsAffected, 10))
	if tx.RowsAffected > 0 {
		return int(tx.RowsAffected), nil
	} else {
		return 0, tx.Error
	}
}
