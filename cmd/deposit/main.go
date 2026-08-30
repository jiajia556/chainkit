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
	var catchUpBudget time.Duration
	var rpcRequestInterval time.Duration
	var inboxBatchSize int
	var inboxPollInterval time.Duration
	var enableWebSocket bool
	flag.Uint64Var(&blockNum, "block_num", 0, "Deposit service start block number")
	flag.IntVar(&cycle, "cycle", 30, "Deposit service cycle time in seconds")
	flag.IntVar(&backfillLimit, "backfill_limit", 10, "Backfill task count per cycle")
	flag.Uint64Var(&backfillStep, "backfill_step", 1000, "Backfill block step per RPC query")
	flag.DurationVar(&catchUpBudget, "catch_up_budget", 20*time.Second, "Maximum catch-up time per deposit token in one cycle; 0 scans until caught up")
	flag.DurationVar(&rpcRequestInterval, "rpc_request_interval", time.Second, "Minimum interval between deposit RPC requests on the same chain; 0 disables throttling")
	flag.IntVar(&inboxBatchSize, "inbox_batch_size", 200, "Maximum deposit inbox events processed per batch")
	flag.DurationVar(&inboxPollInterval, "inbox_poll_interval", time.Second, "Deposit inbox polling interval when the queue is not full")
	flag.BoolVar(&enableWebSocket, "enable_websocket", false, "Enable the deposit WebSocket log subscription")
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
	deposit.CatchUpBudget = catchUpBudget
	deposit.RPCRequestInterval = rpcRequestInterval
	deposit.InboxProcessLimit = inboxBatchSize
	deposit.InboxIdleInterval = inboxPollInterval
	backfillevent.TaskLimit = backfillLimit
	backfillevent.BlockStep = backfillStep

	var startLongRunningOnce sync.Once
	startLongRunning := func(ctx context.Context) {
		startLongRunningOnce.Do(func() {
			if enableWebSocket {
				startTracked(ctx, "deposit websocket", deposit.StartWebSocket)
			}
			startTracked(ctx, "deposit inbox processor", deposit.StartInboxProcessor)
		})
	}

	r := runner.New(
		time.Duration(cycle)*time.Second,
		startLongRunning,
		deposit.Start,
		backfillevent.Start,
	)
	runner.WithTrackedWait()(r)
	err = r.Run(context.Background())
	if err != nil {
		panic(err)
	}
}

func startTracked(ctx context.Context, name string, task func(context.Context)) {
	if !runner.SafeTrackAdd(ctx, 1) {
		log.Error("failed to track long-running service", "service", name)
		return
	}
	go func() {
		defer runner.SafeTrackDone(ctx)
		task(ctx)
	}()
}

func defaultConfigPath() string {
	if path := os.Getenv("CHAINKIT_CONFIG"); path != "" {
		return path
	}
	return "./config.json"
}
