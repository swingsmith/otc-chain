module github.com/otc/otc-chain

go 1.17

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/ethereum/go-ethereum v1.10.18
	github.com/gitstliu/go-id-worker v0.0.0-20190725025543-5a5fe074e612
	github.com/mattn/go-sqlite3 v1.14.9
	github.com/sasaxie/go-client-api v0.0.0-20190820063117-f0587df4b72e
	github.com/semrush/zenrpc/v2 v2.1.1
	github.com/shengdoushi/base58 v1.0.0
	github.com/takama/daemon v1.0.0
	github.com/tidwall/gjson v1.14.1
	xorm.io/xorm v1.3.0
)

require (
	github.com/btcsuite/btcd v0.22.1 // indirect
	//github.com/fbsobreira/gotron-sdk v0.0.0-20211102183839-58a64f4da5f4
	github.com/go-sql-driver/mysql v1.6.0
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/kirinlabs/HttpRequest v1.1.1
	github.com/robfig/cron/v3 v3.0.0
	github.com/shopspring/decimal v1.3.1
	//github.com/smirkcat/hdwallet v0.0.0-20220325043815-d462d0223977
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.21.0
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	google.golang.org/genproto v0.0.0-20200925023002-c2d885f95484
	google.golang.org/grpc v1.44.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gorm.io/driver/mysql v1.3.3
	gorm.io/gorm v1.23.4
)

//replace github.com/otc/otc-chain/tron v0.0.0 => ./tron

//github.com/otc/otc-chain/tron v0.0.0
require github.com/smirkcat/hdwallet v0.0.0-20220325043815-d462d0223977

require (
	github.com/StackExchange/wmi v0.0.0-20180116203802-5d049714c4a6 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.0.1 // indirect
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/deckarep/golang-set v1.8.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/go-ole/go-ole v1.2.1 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/goccy/go-json v0.8.1 // indirect
	github.com/gofrs/uuid v4.2.0+incompatible // indirect
	github.com/golang/protobuf v1.5.0 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/miguelmota/go-ethereum-hdwallet v0.1.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/prometheus/client_golang v1.7.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.10.0 // indirect
	github.com/prometheus/procfs v0.1.3 // indirect
	github.com/rjeczalik/notify v0.9.1 // indirect
	github.com/shirou/gopsutil v3.21.4-0.20210419000835-c7a38de76ee5+incompatible // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tklauser/go-sysconf v0.3.5 // indirect
	github.com/tklauser/numcpus v0.2.2 // indirect
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	xorm.io/builder v0.3.9 // indirect
)
