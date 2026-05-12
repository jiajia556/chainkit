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

	chains.Foreach(func(key int, chain *chainkitchains.Record) bool {
		collectConf := chainkitcollectconfig.NewRecord().GetByChain(chain.Model.Id)
		if !collectConf.Exists() {
			log.Error("failed to read collect config", "chain db id", chain.Model.Id)
			return true
		}

		srv, err := service.NewChainService(chain.Model.Id)
		if err != nil {
			log.Error("failed to create chain service", "error", err, "chain db id", chain.Model.Id)
			return true
		}
		gasPrice, err := srv.SuggestGasPrice()
		if err != nil {
			log.Error("failed to suggest gas price", "error", err, "chain db id", chain.Model.Id)
			return true
		}

		depTokens := chainkitcollecttokens.NewList().FindAvailableByChainDBID(chain.Model.Id)
		if depTokens.IsEmpty() {
			return true
		}
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
	gasRequiredAmount := gasPrice.Mul(gasLimit).Mul(decimal.New(12, -1))

	canTaskList.Foreach(func(key int, canTaskBalance *chainkituserdepositaddressassetbalance.Record) bool {
		tasking := chainkitcollecttasks.NewRecord().GetTaskingByAddressAndToken(canTaskBalance.Model.UserDepositAddressId, canTaskBalance.Model.TokenId)
		if tasking.Exists() {
			return true
		}
		if depToken.Model.ToAddress == "" {
			depToken.Model.ToAddress = collectConf.Model.DefaultCollectToAddress
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
		tasking.Model.ToAddress = depToken.Model.ToAddress
		tasking.Model.PlanAmount = canTaskBalance.Model.BalanceAmount
		tasking.Model.GasRequiredAmount = gasRequiredAmount
		tasking.Model.GasBalanceBeforeTx = gasBalance
		tasking.Model.GasLimit = gasLimit
		tasking.Model.GasPrice = gasPrice
		tasking.Model.Status = 0
		_ = tasking.Create()
		return true
	})
}
