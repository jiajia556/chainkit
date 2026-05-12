package collect

import (
	"context"
	"time"

	"github.com/jiajia556/chainkit/models/chainkitchains"
	"github.com/jiajia556/chainkit/models/chainkitcollectconfig"
	"github.com/jiajia556/chainkit/models/chainkitcollectgasfeetasks"
	"github.com/jiajia556/chainkit/models/chainkitcollecttasks"
	"github.com/jiajia556/chainkit/models/chainkitmnemonicaddresses"
	"github.com/jiajia556/chainkit/models/chainkittokens"
	"github.com/jiajia556/chainkit/models/chainkituserdepositaddressassetbalance"
	"github.com/jiajia556/chainkit/service"
	"github.com/jiajia556/tool-box/log"
	"github.com/jiajia556/tool-box/runner"
	"github.com/shopspring/decimal"
)

var CollectPassword string

func Start(ctx context.Context) {
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
		handleWaiting(srv, chain, collectConf)
		collect(chain)
		runner.TrackAdd(ctx, 1)
		go checkCollectStatus(ctx, srv, chain)
		return true
	})
}

func handleWaiting(srv *service.ChainService, chain *chainkitchains.Record, collectConf *chainkitcollectconfig.Record) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("handleWaiting recover err", err)
		}
	}()

	waitingList := chainkitcollecttasks.NewList().GetWaitingList(chain.Model.Id)
	if waitingList.IsEmpty() {
		return
	}

	needGasAddresses := make(map[uint64][]*chainkitcollecttasks.Record)
	waitingList.Foreach(func(key int, collectTask *chainkitcollecttasks.Record) bool {
		if collectTask.Model.GasRequiredAmount.LessThanOrEqual(collectTask.Model.GasBalanceBeforeTx) {
			collectTask.SetStatus(chainkitcollecttasks.StatusCanSend)
			return true
		}
		_, ok := needGasAddresses[collectTask.Model.UserDepositAddressId]
		if !ok {
			needGasAddresses[collectTask.Model.UserDepositAddressId] = make([]*chainkitcollecttasks.Record, 0)
		}
		needGasAddresses[collectTask.Model.UserDepositAddressId] = append(needGasAddresses[collectTask.Model.UserDepositAddressId], collectTask)
		return true
	})

	for _, tasks := range needGasAddresses {
		createGasTask(srv, tasks, collectConf)
	}
}

func createGasTask(
	srv *service.ChainService,
	collectTasks []*chainkitcollecttasks.Record,
	collectConf *chainkitcollectconfig.Record,
) {
	gasProvider := chainkitmnemonicaddresses.NewRecord()
	err := gasProvider.Read(collectConf.Model.GasProviderMnemonicAddressId)
	if err != nil {
		log.Error("failed to read gas provider", "error", err, "mnemonic address id", collectConf.Model.GasProviderMnemonicAddressId)
		return
	}
	totalGasAmount := decimal.Zero
	collectTaskIds := make([]uint64, len(collectTasks))
	for key, collectTask := range collectTasks {
		totalGasAmount = totalGasAmount.Add(collectTask.Model.GasRequiredAmount).Sub(collectTask.Model.GasBalanceBeforeTx)
		collectTaskIds[key] = collectTask.Model.Id
	}

	if totalGasAmount.LessThanOrEqual(decimal.Zero) {
		chainkitcollecttasks.NewRecord().BatchSetCanSend(collectTaskIds)
		return
	}

	currentBalance, err := srv.BalanceAt(gasProvider.Model.Address)
	if err != nil {
		log.Error("failed to get balance at", "error", err, "chain db id", collectTasks[0].Model.ChainDbId, "address", gasProvider.Model.Address)
		return
	}
	if currentBalance.LessThan(totalGasAmount) {
		log.Warn("gas provider balance is not enough", "chain db id", collectTasks[0].Model.ChainDbId, "address", gasProvider.Model.Address, "current balance", currentBalance, "required balance", totalGasAmount)
		return
	}

	gasTask := chainkitcollectgasfeetasks.NewRecord()
	gasTask.Model.ChainDbId = collectTasks[0].Model.ChainDbId
	gasTask.Model.UserId = collectTasks[0].Model.UserId
	gasTask.Model.UserDepositAddressId = collectTasks[0].Model.UserDepositAddressId
	gasTask.Model.FromAddress = gasProvider.Model.Address
	gasTask.Model.ToAddress = collectTasks[0].Model.FromAddress
	gasTask.Model.RequiredAmount = totalGasAmount
	gasTask.Model.CurrentBalance = currentBalance
	gasTask.Model.GasLimit = decimal.NewFromInt(21000)
	gasTask.Model.Status = chainkitcollectgasfeetasks.StatusWaiting
	_ = gasTask.Create()

	chainkitcollecttasks.NewRecord().BatchSetNeedGas(collectTaskIds, gasTask.Model.Id)
}

func collect(chain *chainkitchains.Record) {
	canSendList := chainkitcollecttasks.NewList().GetCanSendList(chain.Model.Id)
	canSendList.Foreach(func(key int, canSendCollectTask *chainkitcollecttasks.Record) bool {
		pending := chainkitcollecttasks.NewRecord().GetPendingByAddressAndChain(
			canSendCollectTask.Model.UserDepositAddressId,
			canSendCollectTask.Model.ChainDbId,
		)
		if pending.Exists() {
			return true
		}

		srv, err := service.NewChainService(chain.Model.Id)
		if err != nil {
			log.Error("failed to create chain service", "error", err, "chain db id", chain.Model.Id)
			return true
		}
		err = srv.SetFromByDepositAddress(canSendCollectTask.Model.UserDepositAddressId, CollectPassword)
		if err != nil {
			log.Error("failed to set from by deposit address", "error", err, "chain db id", canSendCollectTask.Model.ChainDbId, "user deposit address id", canSendCollectTask.Model.UserDepositAddressId)
			return true
		}

		token := chainkittokens.NewRecord()
		err = token.Read(canSendCollectTask.Model.TokenId)
		if err != nil {
			log.Error("failed to read token", "error", err, "token id", canSendCollectTask.Model.TokenId)
			return true
		}
		balance, err := srv.BalanceOf(token.Model.ContractAddress, canSendCollectTask.Model.FromAddress)
		if err != nil {
			log.Error("failed to get balance of", "error", err, "chain db id", canSendCollectTask.Model.ChainDbId, "address", canSendCollectTask.Model.FromAddress, "token contract address", token.Model.ContractAddress)
			return true
		}
		opts := make([]service.Option, 0)
		if canSendCollectTask.Model.Nonce > 0 {
			opts = append(opts, service.Nonce(canSendCollectTask.Model.Nonce))
		}

		hash, nonce, _, err := srv.TransferERC20(token.Model.ContractAddress, canSendCollectTask.Model.ToAddress, balance, opts...)
		if err != nil {
			log.Error(
				"failed to transfer",
				"error", err,
				"chain db id", canSendCollectTask.Model.ChainDbId,
				"from address", canSendCollectTask.Model.FromAddress,
				"to address", canSendCollectTask.Model.ToAddress,
				"amount", balance,
				"token contract address", token.Model.ContractAddress,
			)
			return true
		}
		canSendCollectTask.SetSent(balance, hash, nonce)
		return true
	})
}

func checkCollectStatus(ctx context.Context, srv *service.ChainService, chain *chainkitchains.Record) {
	defer runner.TrackDone(ctx)

	centList := chainkitcollecttasks.NewList().GetCentList(chain.Model.Id)
	centList.Foreach(func(key int, centTask *chainkitcollecttasks.Record) bool {
		status, err := srv.GetTxStatus(centTask.Model.TxHash)
		if err != nil {
			return false
		}
		switch status {
		case service.TxStatusNotFound:
			if centTask.SinceCreated() > time.Minute*15 {
				occupied, err := srv.IsNonceOccupied(centTask.Model.FromAddress, centTask.Model.Nonce)
				if err != nil {
					log.Error("checkCollectStatus: failed to check nonce occupied", "error", err)
					return true
				}
				if occupied {
					// 特殊情况，人工处理
					centTask.SetUnknown()
				} else {
					centTask.SetWaiting()
				}
			}
		case service.TxStatusPending, service.TxStatusMined:
			return true
		case service.TxStatusFailed:
			centTask.SetFailed()
		case service.TxStatusConfirmed:
			txFee := decimal.Zero
			gasUsed, gasPrice, err := srv.GetGasUsedAndEffectiveGasPrice(centTask.Model.TxHash)
			if err == nil {
				txFee = gasUsed.Mul(gasPrice)
			} else {
				log.Error("failed to get gas used and effective gas price", "error", err, "tx hash", centTask.Model.TxHash)
			}

			centTask.SetConfirmed(gasUsed, txFee)
			addressBalance := chainkituserdepositaddressassetbalance.NewRecord()
			err = addressBalance.Read(centTask.Model.UserDepositAddressId)
			if err != nil {
				log.Error("failed to read address balance", "error", err, "address", centTask.Model.UserDepositAddressId)
				return true
			}
			err = addressBalance.Collected(centTask.Model.ActualAmount, centTask.Model.TxHash)
			if err != nil {
				log.Error("failed to addressBalance.Collected", "error", err, "addressBalance id", addressBalance.Model.Id)
			}
		}
		return true
	})

}
