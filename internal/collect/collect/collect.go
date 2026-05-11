package collect

import (
	"context"

	"github.com/jiajia556/chainkit/models/chainkitchains"
	"github.com/jiajia556/chainkit/models/chainkitcollectconfig"
	"github.com/jiajia556/chainkit/models/chainkitcollectgasfeetasks"
	"github.com/jiajia556/chainkit/models/chainkitcollecttasks"
	"github.com/jiajia556/chainkit/models/chainkitmnemonicaddresses"
	"github.com/jiajia556/chainkit/service"
	"github.com/jiajia556/tool-box/log"
	"github.com/shopspring/decimal"
)

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
			log.Error("failed to read collect config", "error", err, "chain db id", chain.Model.Id)
			return true
		}

		srv, err := service.NewChainService(chain.Model.Id)
		if err != nil {
			log.Error("failed to create chain service", "error", err, "chain db id", chain.Model.Id)
			return true
		}
		handleWaiting(srv, chain, collectConf)
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
		totalGasAmount = totalGasAmount.Add(collectTask.Model.GasRequiredAmount)
		collectTaskIds[key] = collectTask.Model.Id
	}
	totalGasAmount = totalGasAmount.Sub(collectTasks[0].Model.GasBalanceBeforeTx)

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
}
