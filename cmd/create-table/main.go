package main

import (
	"github.com/jiajia556/chainkit/internal/chainkit/config"
	"github.com/jiajia556/chainkit/models/chainkitasset"
	"github.com/jiajia556/chainkit/models/chainkitassetrecord"
	"github.com/jiajia556/chainkit/models/chainkitchains"
	"github.com/jiajia556/chainkit/models/chainkitcollectconfig"
	"github.com/jiajia556/chainkit/models/chainkitcollectgasfeetasks"
	"github.com/jiajia556/chainkit/models/chainkitcollecttasks"
	"github.com/jiajia556/chainkit/models/chainkitcollecttokens"
	"github.com/jiajia556/chainkit/models/chainkitcontracts"
	"github.com/jiajia556/chainkit/models/chainkitdepositrecord"
	"github.com/jiajia556/chainkit/models/chainkitdeposittokens"
	"github.com/jiajia556/chainkit/models/chainkiteventbackfilltask"
	"github.com/jiajia556/chainkit/models/chainkiteventlogs"
	"github.com/jiajia556/chainkit/models/chainkitmintdetails"
	"github.com/jiajia556/chainkit/models/chainkitmintrecords"
	"github.com/jiajia556/chainkit/models/chainkitmnemonicaddresses"
	"github.com/jiajia556/chainkit/models/chainkitmnemonics"
	"github.com/jiajia556/chainkit/models/chainkitscancursor"
	"github.com/jiajia556/chainkit/models/chainkittokengroups"
	"github.com/jiajia556/chainkit/models/chainkittokens"
	"github.com/jiajia556/chainkit/models/chainkittransferdetails"
	"github.com/jiajia556/chainkit/models/chainkittransferrecords"
	"github.com/jiajia556/chainkit/models/chainkituserdepositaddress"
	"github.com/jiajia556/chainkit/models/chainkituserdepositaddressassetbalance"
	"github.com/jiajia556/tool-box/log"
	"github.com/jiajia556/tool-box/mysqlx"
)

func main() {
	configPath := ""
	err := config.Load(configPath)
	if err != nil {
		log.Fatal("Failed to load config: %v", err)
	}

	err = mysqlx.InitMysql(config.GetConfig().Mysql)
	if err != nil {
		log.Fatal("Failed to initialize MySQL: %v", err)
	}

	chainkitasset.NewRecord()
	chainkitassetrecord.NewRecord()
	chainkitchains.NewRecord()
	chainkitcollectconfig.NewRecord()
	chainkitcollectgasfeetasks.NewRecord()
	chainkitcollecttasks.NewRecord()
	chainkitcollecttokens.NewRecord()
	chainkitcontracts.NewRecord()
	chainkitdeposittokens.NewRecord()
	chainkitdepositrecord.NewRecord()
	chainkiteventbackfilltask.NewRecord()
	chainkiteventlogs.NewRecord()
	chainkitmintdetails.NewRecord()
	chainkitmintrecords.NewRecord()
	chainkitmnemonicaddresses.NewRecord()
	chainkitmnemonics.NewRecord()
	chainkitscancursor.NewRecord()
	chainkittokengroups.NewRecord()
	chainkittokens.NewRecord()
	chainkittransferdetails.NewRecord()
	chainkittransferrecords.NewRecord()
	chainkituserdepositaddress.NewRecord()
	chainkituserdepositaddressassetbalance.NewRecord()

}
