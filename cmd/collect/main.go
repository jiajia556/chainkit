package main

import (
	"flag"

	"github.com/jiajia556/chainkit/internal/collect/config"
	"github.com/jiajia556/tool-box/log"
	"github.com/jiajia556/tool-box/mysqlx"
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
	logConfig.File.Path = "collect.log"
	err = log.Init(logConfig)
	if err != nil {
		panic(err)
	}
}
