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
	flag.StringVar(&configPath, "config", "./config.json", "Config json file path")
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
	logConfig.Output = "file"
	logConfig.File.Path = "deposit.log"
	err = log.Init("std", logConfig)
	if err != nil {
		panic(err)
	}

	err = runner.New(30*time.Second, deposit.Start).Run(context.Background())
	if err != nil {
		panic(err)
	}
}
