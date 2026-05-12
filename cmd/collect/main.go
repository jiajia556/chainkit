package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/howeyc/gopass"
	"github.com/jiajia556/chainkit/internal/collect/buildcollecttask"
	"github.com/jiajia556/chainkit/internal/collect/collect"
	"github.com/jiajia556/chainkit/internal/collect/config"
	"github.com/jiajia556/chainkit/internal/collect/providegas"
	"github.com/jiajia556/tool-box/log"
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

	fmt.Println("enter collect password:")
	passwordByte, err := gopass.GetPasswd()
	if err != nil {
		panic(err)
	}
	collect.CollectPassword = string(passwordByte)

	fmt.Println("enter gas password:")
	passwordByte, err = gopass.GetPasswd()
	if err != nil {
		panic(err)
	}
	providegas.Password = string(passwordByte)

	logConfig := log.DefaultConfig()
	logConfig.Output = "file"
	logConfig.File.Path = "collect.log"
	err = log.Init(logConfig)
	if err != nil {
		panic(err)
	}

	err = runner.New(30*time.Minute, buildcollecttask.Start, collect.Start, providegas.Start).Run(context.Background())
	if err != nil {
		panic(err)
	}
}
