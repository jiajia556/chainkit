package main

import (
	"context"
	"flag"
	"time"

	"github.com/jiajia556/chainkit/internal/deposit"
	"github.com/jiajia556/chainkit/internal/deposit/config"
	"github.com/jiajia556/tool-box/log"
	_ "github.com/jiajia556/tool-box/log/std"
	"github.com/jiajia556/tool-box/mysqlx"
	"github.com/jiajia556/tool-box/runner"
)

func main() {
	var configPath string
	var cycle int
	var blockNum uint64
	flag.Uint64Var(&blockNum, "block_num", 0, "Deposit service start block number")
	flag.IntVar(&cycle, "cycle", 30, "Deposit service cycle time in seconds")
	flag.StringVar(&configPath, "config", "E:\\work\\gowork\\chainkit\\chainkit_config.json", "Config json file path")
	flag.Parse()
	err := config.Load(configPath)
	if err != nil {
		panic(err)
	}

	err = mysqlx.InitMysql(config.GetConfig().Mysql)
	if err != nil {
		panic(err)
	}

	logConfig := log.DefaultConfig()
	err = log.Init(logConfig)
	if err != nil {
		panic(err)
	}

	log.Debug("starting deposit service", "123123")
	deposit.BlockNum = blockNum
	err = runner.New(time.Duration(cycle)*time.Second, deposit.Start).Run(context.Background())
	if err != nil {
		panic(err)
	}
}
