package main

import (
	"context"
	"flag"
	"os"
	"sync"
	"time"

	"github.com/jiajia556/chainkit/internal/backfillevent"
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
	var backfillLimit int
	var backfillStep uint64
	flag.Uint64Var(&blockNum, "block_num", 0, "Deposit service start block number")
	flag.IntVar(&cycle, "cycle", 30, "Deposit service cycle time in seconds")
	flag.IntVar(&backfillLimit, "backfill_limit", 10, "Backfill task count per cycle")
	flag.Uint64Var(&backfillStep, "backfill_step", 1000, "Backfill block step per RPC query")
	flag.StringVar(&configPath, "config", defaultConfigPath(), "Config json file path")
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

	deposit.BlockNum = blockNum
	backfillevent.TaskLimit = backfillLimit
	backfillevent.BlockStep = backfillStep
	err = runner.New(time.Duration(cycle)*time.Second, func(ctx context.Context) {
		wg := &sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			deposit.Start(ctx)
		}()
		go func() {
			defer wg.Done()
			backfillevent.Start(ctx)
		}()
		wg.Wait()
	}).Run(context.Background())
	if err != nil {
		panic(err)
	}
}

func defaultConfigPath() string {
	if path := os.Getenv("CHAINKIT_CONFIG"); path != "" {
		return path
	}
	return "./config.json"
}
