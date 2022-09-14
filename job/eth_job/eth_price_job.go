package eth_job

import (
	"errors"
	"fmt"
	"github.com/kirinlabs/HttpRequest"
	"github.com/otc/otc-chain/models"
	"github.com/otc/otc-chain/utils/db_util"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
	"log"
)

func EthPriceJob() {
	db := db_util.GetDB()
	var config models.Config
	tx := db.Model(&models.Config{}).Where("config_key = ?", "ETH_PRICE").First(&config)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {

	} else {
		//获取价格插入
		req := HttpRequest.NewRequest().Debug(true).SetTimeout(30)

		// 设置超时时间，不设置时，默认30s
		//req.SetTimeout(30)

		// 设置Headers
		req.SetHeaders(map[string]string{
			"Content-Type": "application/json",
			//"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.99 Safari/537.36",
		})

		// GET 默认调用方法
		url := "https://api.hoolgd.com/open/v1/tickers/market"
		resp, err := req.Get(url)
		defer resp.Close()

		if err != nil {
			log.Printf("req.Get(url) error %s", err.Error())
			return
		}

		if resp.StatusCode() == 200 {
			body, err := resp.Body()

			if err != nil {
				log.Println(err)
				return
			}
			//fmt.Println(string(body))

			json := string(body)
			if !gjson.Valid(json) {
				fmt.Println("error")
			} else {
				fmt.Println("ok")
				price := gjson.Get(json, `data.#(symbol="ETH-USDT").price`)
				fmt.Printf("ETHNowPrice价格为：%s", price.String())

				if err := db.Debug().Model(&models.Config{}).Where("config_key = ?", "ETH_PRICE").Update("config_value", price.String()).Error; err != nil {
				}
			}
		}
	}
}
