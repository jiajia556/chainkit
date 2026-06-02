package buildcollecttask

import (
	"context"

	"github.com/jiajia556/chainkit/models/chainkitchains"
	"github.com/jiajia556/chainkit/models/chainkitcollectconfig"
	"github.com/jiajia556/chainkit/models/chainkitcollecttasks"
	"github.com/jiajia556/chainkit/models/chainkitcollecttokens"
	"github.com/jiajia556/chainkit/models/chainkituserdepositaddressassetbalance"
	"github.com/jiajia556/chainkit/service"
	"github.com/jiajia556/tool-box/log"
	"github.com/shopspring/decimal"
)

func Start(ctx context.Context) {
	log.Debug("start build collect task")
	defer func() {
		if r := recover(); r != nil {
			log.Error("panic in build collect task", "error", r)
		}
	}()
	chains := chainkitchains.NewList()
	err := chains.FindAll()
	if err != nil {
		log.Error("failed to find chains", "error", err)
		return
	}
	log.Debug("read chains", "count", len(*chains.Records))

	chains.Foreach(func(key int, chain *chainkitchains.Record) bool {
		collectConf := chainkitcollectconfig.NewRecord().GetByChain(chain.Model.Id)
		if !collectConf.Exists() {
			log.Error("failed to read collect config", "chain db id", chain.Model.Id)
			return true
		}

		log.Debug("read collect config", "chain db id", chain.Model.Id, "config id", collectConf.Model.Id)

		srv, err := service.NewChainService(chain.Model.Id)
		if err != nil {
			log.Error("failed to create chain service", "error", err, "chain db id", chain.Model.Id)
			return true
		}
		log.Debug("create chain service", "chain db id", chain.Model.Id)

		gasPrice, err := srv.SuggestGasPrice()
		if err != nil {
			log.Error("failed to suggest gas price", "error", err, "chain db id", chain.Model.Id)
			return true
		}
		log.Debug("suggest gas price", "chain db id", chain.Model.Id, "gas price", gasPrice)

		depTokens := chainkitcollecttokens.NewList().FindAvailableByChainDBID(chain.Model.Id)
		if depTokens.IsEmpty() {
			log.Debug("no available collect token", "chain db id", chain.Model.Id)
			return true
		}
		log.Debug("read collect tokens", "chain db id", chain.Model.Id, "count", len(*depTokens.Records))
		depTokens.Foreach(func(key int, depToken *chainkitcollecttokens.Record) bool {
			handleDepToken(srv, depToken, collectConf, collectConf.Model.DefaultErc20TransferGasLimit, gasPrice)
			return true
		})
		return true
	})
}

func handleDepToken(
	srv *service.ChainService,
	depToken *chainkitcollecttokens.Record,
	collectConf *chainkitcollectconfig.Record,
	defaultGasLimit decimal.Decimal,
	gasPrice decimal.Decimal,
) {
	canTaskList := chainkituserdepositaddressassetbalance.NewList().
		GetCanTaskByTokenBalance(depToken.Model.TokenId, depToken.Model.MinCollectAmount)
	if canTaskList.IsEmpty() {
		return
	}

	gasLimit := depToken.Model.TransferGasLimit
	if gasLimit.IsZero() {
		gasLimit = defaultGasLimit
	}
	// 归集任务创建时先按当前 gasPrice 做快照，并预留 20% 浮动，后续是否补 gas 以这个快照判断。
	gasRequiredAmount := gasPrice.Mul(gasLimit).Mul(decimal.New(12, -1))

	canTaskList.Foreach(func(key int, canTaskBalance *chainkituserdepositaddressassetbalance.Record) bool {
		// 同一地址同一 token 只允许存在一个进行中的归集任务，避免重复归集同一笔余额。
		tasking := chainkitcollecttasks.NewRecord().GetTaskingByAddressAndToken(canTaskBalance.Model.UserDepositAddressId, canTaskBalance.Model.TokenId)
		if tasking.Exists() {
			return true
		}
		// 空地址使用链级默认归集地址；不要回写 depToken.Model，避免影响同一轮循环里的其它判断。
		toAddress := depToken.Model.ToAddress
		if toAddress == "" {
			toAddress = collectConf.Model.DefaultCollectToAddress
		}
		gasBalance, err := srv.BalanceAt(canTaskBalance.Model.Address)
		if err != nil {
			log.Error("failed to get balance at", "error", err, "chain db id", depToken.Model.ChainDbId, "address", canTaskBalance.Model.Address)
			return true
		}
		tasking.Model.ChainDbId = depToken.Model.ChainDbId
		tasking.Model.TokenId = depToken.Model.TokenId
		tasking.Model.UserId = canTaskBalance.Model.UserId
		tasking.Model.UserDepositAddressId = canTaskBalance.Model.UserDepositAddressId
		tasking.Model.FromAddress = canTaskBalance.Model.Address
		tasking.Model.ToAddress = toAddress
		tasking.Model.PlanAmount = canTaskBalance.Model.BalanceAmount
		tasking.Model.GasRequiredAmount = gasRequiredAmount
		tasking.Model.GasBalanceBeforeTx = gasBalance
		tasking.Model.GasLimit = gasLimit
		tasking.Model.GasPrice = gasPrice
		tasking.Model.Status = 0
		if err := tasking.Create(); err != nil {
			log.Error("failed to create collect task", "error", err, "chain db id", depToken.Model.ChainDbId, "user deposit address id", canTaskBalance.Model.UserDepositAddressId, "token id", canTaskBalance.Model.TokenId)
		}
		return true
	})
}
