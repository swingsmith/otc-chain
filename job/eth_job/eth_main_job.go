package eth_job

import (
	"fmt"
	"github.com/otc/otc-chain/models"
	"github.com/robfig/cron/v3"
	"time"
)

func newWithSeconds() *cron.Cron {
	secondParser := cron.NewParser(cron.Second | cron.Minute |
		cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)
	return cron.New(cron.WithParser(secondParser), cron.WithChain())
}

func EthMainJob() {
	//****************ETH USDT
	c := newWithSeconds()
	spec := "*/3 * * * * ?" // 每5秒执行一次
	_, err := c.AddFunc(spec, EthUsdtScanJob)
	if err != nil {
		fmt.Println(err)
		//return
	}

	c2 := newWithSeconds()
	spec2 := "0 */5 * * * ?"
	_, err = c2.AddFunc(spec2, EthAddressGenerateJob)
	if err != nil {
		fmt.Println(err)
		//return
	}

	c3 := newWithSeconds()
	spec3 := "0 0 3 */1 * ?"
	entryID3, err := c3.AddFunc(spec3, EthGasJob)
	if err != nil {
		fmt.Println(err)
		//return
	}

	c4 := newWithSeconds()
	spec4 := "0 0 4 */1 * ?" // 每小时执行一次
	entryID4, err := c4.AddFunc(spec4, EthUsdtCollectJob)
	if err != nil {
		fmt.Println(err)
		//return
	}

	c.Start()
	c2.Start()
	c3.Start()
	c4.Start()

	//loc, _ := time.LoadLocation("Asia/Shanghai")
	//t, _ := time.ParseInLocation("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05"), loc)
	//fmt.Println(t)
	//time.Now().After(t)
	// 输出 2021-01-10 17:28:50 +0800 CST
	// time.Local 指定本地时间
	var startTime string = ""
	var config models.Config
	c5 := newWithSeconds()
	spec5 := "*/5 * * * * ?"
	_, err = c5.AddFunc(spec5, func() {
		start, _ := config.GetEthUsdtCollectStartTime()
		flag, _ := config.GetEthUsdtCollectFlag()
		interval, _ := config.GetEthUsdtCollectInterval()
		gap, _ := config.GetEthUsdtTimeGapBetweenGasAndCollect()

		if startTime == "" {
			startTime = start
		} else if start != startTime && flag == "1" {
			fmt.Printf("mainjob开始 当前时间为：%s", start)
			startTime = start
			c3.Stop()
			c4.Stop()
			c3.Remove(entryID3)
			c4.Remove(entryID4)

			loc, _ := time.LoadLocation("Asia/Shanghai")
			t, _ := time.ParseInLocation("2006-01-02 15:04:05", start, loc)

			spec3 = fmt.Sprintf("%d %d %d */%s * ?", t.Second(), t.Minute(), t.Hour(), interval) // 每小时执行一次
			entryID3, err = c3.AddFunc(spec3, EthGasJob)
			if err != nil {
				fmt.Println(err)
				//return
			}
			c3.Start()

			d, _ := time.ParseDuration(gap + "m")
			t = t.Add(d)
			spec4 = fmt.Sprintf("%d %d %d */%s * ?", t.Second(), t.Minute(), t.Hour(), interval) // 每小时执行一次
			entryID4, err = c4.AddFunc(spec4, EthUsdtCollectJob)
			if err != nil {
				fmt.Println(err)
				//return
			}
			c4.Start()

			fmt.Printf("mainjob结束 当前时间为：%s", start)

		} else if flag == "0" {
			fmt.Println("ETH_USDT归集开关关闭")
			c3.Stop()
			c4.Stop()
		}
	})
	if err != nil {
		fmt.Println(err)
		//return
	}
	c5.Start()

	c6 := newWithSeconds()
	spec6 := "*/5 * * * * ?" // 每小时执行一次
	_, err = c6.AddFunc(spec6, EthPriceJob)
	if err != nil {
		fmt.Println(err)
		//return
	}
	c6.Start()

	c7 := newWithSeconds()
	spec7 := "*/10 * * * * ?" // 每小时执行一次
	_, err = c7.AddFunc(spec7, WithdrawStatusScanJob)
	if err != nil {
		fmt.Println(err)
		//return
	}
	c7.Start()

	//select {}
}
