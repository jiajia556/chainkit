package collect

import (
	"context"
	"errors"
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
	"github.com/jiajia556/tool-box/mysqlx"
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

	log.Debug("Start collect: found chains", "chain count", len(*chains.Records))

	chains.Foreach(func(key int, chain *chainkitchains.Record) bool {
		log.Debug("Start collect: start to handle chain", "chain db id", chain.Model.Id)
		collectConf := chainkitcollectconfig.NewRecord().GetByChain(chain.Model.Id)
		if !collectConf.Exists() {
			log.Debug("failed to read collect config", "chain db id", chain.Model.Id)
			log.Error("failed to read collect config", "chain db id", chain.Model.Id)
			return true
		}

		srv, err := service.NewChainService(chain.Model.Id)
		if err != nil {
			log.Debug("failed to create chain service", "error", err, "chain db id", chain.Model.Id)
			log.Error("failed to create chain service", "error", err, "chain db id", chain.Model.Id)
			return true
		}
		handleWaiting(srv, chain, collectConf)
		collect(chain)
		log.Debug("Start collect: start checking collect status", "chain db id", chain.Model.Id)
		checkCollectStatus(ctx, srv, chain)
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
		// 地址原生币余额已经足够支付归集手续费时，直接进入可发送队列。
		if collectTask.Model.GasRequiredAmount.LessThanOrEqual(collectTask.Model.GasBalanceBeforeTx) {
			collectTask.SetStatus(chainkitcollecttasks.StatusCanSend)
			return true
		}
		// 同一充值地址的多条归集任务共用一笔补 gas 交易，减少 gas provider 的转账次数。
		_, ok := needGasAddresses[collectTask.Model.UserDepositAddressId]
		if !ok {
			needGasAddresses[collectTask.Model.UserDepositAddressId] = make([]*chainkitcollecttasks.Record, 0)
		}
		needGasAddresses[collectTask.Model.UserDepositAddressId] = append(needGasAddresses[collectTask.Model.UserDepositAddressId], collectTask)
		return true
	})

	for _, tasks := range needGasAddresses {
		if err := createGasTask(srv, tasks, collectConf); err != nil {
			log.Error("failed to create gas task", "error", err, "chain db id", chain.Model.Id)
		}
	}
}

func createGasTask(
	srv *service.ChainService,
	collectTasks []*chainkitcollecttasks.Record,
	collectConf *chainkitcollectconfig.Record,
) error {
	gasProvider := chainkitmnemonicaddresses.NewRecord()
	err := gasProvider.Read(collectConf.Model.GasProviderMnemonicAddressId)
	if err != nil {
		log.Error("failed to read gas provider", "error", err, "mnemonic address id", collectConf.Model.GasProviderMnemonicAddressId)
		return err
	}
	totalGasAmount := decimal.Zero
	collectTaskIds := make([]uint64, len(collectTasks))
	for key, collectTask := range collectTasks {
		// 每条任务只补足“预估所需 gas - 当前余额”的差额；同地址任务在外层已聚合。
		totalGasAmount = totalGasAmount.Add(collectTask.Model.GasRequiredAmount).Sub(collectTask.Model.GasBalanceBeforeTx)
		collectTaskIds[key] = collectTask.Model.Id
	}

	if totalGasAmount.LessThanOrEqual(decimal.Zero) {
		return chainkitcollecttasks.NewRecord().BatchSetCanSend(collectTaskIds)
	}

	currentBalance, err := srv.BalanceAt(gasProvider.Model.Address)
	if err != nil {
		log.Error("failed to get balance at", "error", err, "chain db id", collectTasks[0].Model.ChainDbId, "address", gasProvider.Model.Address)
		return err
	}
	if currentBalance.LessThan(totalGasAmount) {
		// gas provider 余额不足时保持任务 Waiting，下一轮定时任务会再次尝试。
		log.Warn("gas provider balance is not enough", "chain db id", collectTasks[0].Model.ChainDbId, "address", gasProvider.Model.Address, "current balance", currentBalance, "required balance", totalGasAmount)
		return nil
	}

	// 创建补 gas 任务和绑定归集任务必须原子完成，避免出现 gas_task_id=0 或孤儿 gas task。
	session := mysqlx.NewTxSession()
	if err := session.Begin(); err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			_ = session.Rollback()
			panic(r)
		}
	}()
	rollbackWithErr := func(err error) error {
		if rbErr := session.Rollback(); rbErr != nil {
			log.Error("failed to rollback create gas task transaction", "error", rbErr)
		}
		return err
	}

	gasTask := chainkitcollectgasfeetasks.NewRecord(session)
	gasTask.Model.ChainDbId = collectTasks[0].Model.ChainDbId
	gasTask.Model.UserId = collectTasks[0].Model.UserId
	gasTask.Model.UserDepositAddressId = collectTasks[0].Model.UserDepositAddressId
	gasTask.Model.FromAddress = gasProvider.Model.Address
	gasTask.Model.ToAddress = collectTasks[0].Model.FromAddress
	gasTask.Model.RequiredAmount = totalGasAmount
	gasTask.Model.CurrentBalance = currentBalance
	gasTask.Model.GasLimit = decimal.NewFromInt(21000)
	gasTask.Model.Status = chainkitcollectgasfeetasks.StatusWaiting
	if err := gasTask.Create(); err != nil {
		return rollbackWithErr(err)
	}

	affected, err := chainkitcollecttasks.NewRecord(session).BatchSetNeedGas(collectTaskIds, gasTask.Model.Id)
	if err != nil {
		return rollbackWithErr(err)
	}
	if affected != int64(len(collectTaskIds)) {
		// 有任务在事务期间被其它 worker 抢走/改状态，当前 gas task 不再可信，整体回滚后下轮重试。
		return rollbackWithErr(errors.New("collect task status changed while creating gas task"))
	}
	return session.Commit()
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
		// 广播前先原子抢占任务，只有 CanSend -> Sending 成功的 worker 才能真正发交易。
		claimed, err := canSendCollectTask.ClaimCanSend()
		if err != nil {
			log.Error("failed to claim collect task", "error", err, "collect task id", canSendCollectTask.Model.Id)
			return true
		}
		if !claimed {
			return true
		}

		srv, err := service.NewChainService(chain.Model.Id)
		if err != nil {
			log.Error("failed to create chain service", "error", err, "chain db id", chain.Model.Id)
			canSendCollectTask.SetCanSendWithError(err.Error())
			return true
		}
		err = srv.SetFromByDepositAddress(canSendCollectTask.Model.UserDepositAddressId, CollectPassword)
		if err != nil {
			log.Error("failed to set from by deposit address", "error", err, "chain db id", canSendCollectTask.Model.ChainDbId, "user deposit address id", canSendCollectTask.Model.UserDepositAddressId)
			canSendCollectTask.SetCanSendWithError(err.Error())
			return true
		}

		token := chainkittokens.NewRecord()
		err = token.Read(canSendCollectTask.Model.TokenId)
		if err != nil {
			log.Error("failed to read token", "error", err, "token id", canSendCollectTask.Model.TokenId)
			canSendCollectTask.SetCanSendWithError(err.Error())
			return true
		}
		balance, err := srv.BalanceOf(token.Model.ContractAddress, canSendCollectTask.Model.FromAddress)
		if err != nil {
			log.Error("failed to get balance of", "error", err, "chain db id", canSendCollectTask.Model.ChainDbId, "address", canSendCollectTask.Model.FromAddress, "token contract address", token.Model.ContractAddress)
			canSendCollectTask.SetCanSendWithError(err.Error())
			return true
		}
		opts := make([]service.Option, 0)
		if canSendCollectTask.Model.Nonce > 0 {
			// not_found 回退后会保留 nonce，重试时优先复用原 nonce，降低 nonce 空洞风险。
			opts = append(opts, service.Nonce(canSendCollectTask.Model.Nonce))
		}

		hash, nonce, fakeErr, err := srv.TransferERC20(token.Model.ContractAddress, canSendCollectTask.Model.ToAddress, balance, opts...)
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
			canSendCollectTask.SetCanSendWithError(err.Error())
			return true
		}
		if fakeErr != nil {
			// RPC 返回错误但本地已经拿到签名交易 hash，链上可能已经接收；进入 MaybeSent 等状态机确认。
			log.Error(
				"transfer returned uncertain broadcast result",
				"error", fakeErr,
				"chain db id", canSendCollectTask.Model.ChainDbId,
				"from address", canSendCollectTask.Model.FromAddress,
				"to address", canSendCollectTask.Model.ToAddress,
				"amount", balance,
				"token contract address", token.Model.ContractAddress,
				"tx hash", hash,
			)
			canSendCollectTask.SetMaybeSent(balance, hash, nonce, fakeErr.Error())
			return true
		}
		canSendCollectTask.SetSent(balance, hash, nonce)
		return true
	})
}

func checkCollectStatus(ctx context.Context, srv *service.ChainService, chain *chainkitchains.Record) {
	centList := chainkitcollecttasks.NewList().GetCentList(chain.Model.Id)
	log.Debug("checkCollectStatus: start checking collect status", "chain db id", chain.Model.Id, "cent list size", len(*centList.Records))
	centList.Foreach(func(key int, centTask *chainkitcollecttasks.Record) bool {
		status, err := srv.GetTxStatus(centTask.Model.TxHash)
		if err != nil {
			log.Error("checkCollectStatus: failed to get tx status", "error", err, "tx hash", centTask.Model.TxHash)
			return true
		}
		log.Debug("checkCollectStatus: get tx status", "chain db id", chain.Model.Id, "tx hash", centTask.Model.TxHash, "status", status)
		switch status {
		case service.TxStatusNotFound:
			if centTask.SinceSent() > time.Minute*10 {
				occupied, err := srv.IsNonceOccupied(centTask.Model.FromAddress, centTask.Model.Nonce)
				if err != nil {
					log.Error("checkCollectStatus: failed to check nonce occupied", "error", err)
					return true
				}
				if occupied {
					// hash 查不到但 nonce 已被占用，说明可能有同 nonce 交易上链/进池，需要人工核对。
					centTask.SetUnknown()
				} else {
					// nonce 仍未被占用，说明这笔交易大概率没有进入节点池，回到 Waiting 重新建流程。
					centTask.SetWaiting()
				}
			}
		case service.TxStatusPending, service.TxStatusMined:
			return true
		case service.TxStatusUnknown:
			if centTask.SinceSent() > time.Minute*10 {
				log.Debug("checkCollectStatus: tx status unknown, set collect task to unknown", "from address", centTask.Model.FromAddress, "nonce", centTask.Model.Nonce, "tx hash", centTask.Model.TxHash)
				centTask.SetUnknown()
			}
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
			// 以链上余额为准刷新资产余额，并累计本次已归集金额，避免仅依赖任务快照。
			addressBalance := chainkituserdepositaddressassetbalance.NewRecord().ReadByChainAndAddressAndToken(centTask.Model.ChainDbId, centTask.Model.UserDepositAddressId, centTask.Model.TokenId)
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
