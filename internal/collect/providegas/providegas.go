package providegas

import (
	"context"
	"time"

	"github.com/jiajia556/chainkit/models/chainkitchains"
	"github.com/jiajia556/chainkit/models/chainkitcollectconfig"
	"github.com/jiajia556/chainkit/models/chainkitcollectgasfeetasks"
	"github.com/jiajia556/chainkit/models/chainkitcollecttasks"
	"github.com/jiajia556/chainkit/service"
	"github.com/jiajia556/tool-box/log"
	"github.com/shopspring/decimal"
)

var Password string

func Start(ctx context.Context) {
	chains := chainkitchains.NewList()
	err := chains.FindAll()
	if err != nil {
		log.Error("failed to find chains", "error", err)
		return
	}

	chains.Foreach(func(key int, chain *chainkitchains.Record) bool {
		handleChain(chain)
		return true
	})
}

func handleChain(chain *chainkitchains.Record) {
	collectConf := chainkitcollectconfig.NewRecord().GetByChain(chain.Model.Id)
	if !collectConf.Exists() {
		log.Error("failed to read collect config", "chain db id", chain.Model.Id)
		return
	}
	srv, err := service.NewChainService(chain.Model.Id)
	if err != nil {
		log.Error("failed to create chain service", "error", err, "chain db id", chain.Model.Id)
		return
	}

	lastSent := chainkitcollectgasfeetasks.NewRecord().GetLastSent(chain.Model.Id)
	if lastSent.Exists() {
		// 每条链同一时间只推进最早的一笔已广播/疑似已广播 gas 任务，避免 gas provider nonce 乱序。
		status, err := srv.GetTxStatus(lastSent.Model.TxHash)
		if err != nil {
			log.Error("failed to get tx status", "error", err, "tx hash", lastSent.Model.TxHash)
			return
		}
		log.Debug("last sent gas task", "chain db id", chain.Model.Id, "tx hash", lastSent.Model.TxHash, "status", status)
		switch status {
		case service.TxStatusNotFound:
			if lastSent.SinceSent() > time.Minute*10 {
				occupied, err := srv.IsNonceOccupied(lastSent.Model.FromAddress, lastSent.Model.Nonce)
				if err != nil {
					log.Error("failed to check if nonce is occupied", "error", err, "from address", lastSent.Model.FromAddress, "nonce", lastSent.Model.Nonce)
					return
				}
				if occupied {
					// hash 查不到但 nonce 已被占用，说明可能被同 nonce 交易替代，需要人工核对。
					log.Debug("nonce is occupied, set last sent gas task to unknown, need manual handling", "from address", lastSent.Model.FromAddress, "nonce", lastSent.Model.Nonce)
					lastSent.SetUnknown()
				} else {
					// nonce 未被占用，说明交易大概率没有进入节点池，回到 Waiting 重新发送。
					lastSent.SetWaiting()
				}
			}
		case service.TxStatusPending, service.TxStatusMined:
			return
		case service.TxStatusUnknown:
			if lastSent.SinceSent() > time.Minute*10 {
				log.Debug("tx status unknown, set last sent gas task to unknown", "from address", lastSent.Model.FromAddress, "nonce", lastSent.Model.Nonce, "tx hash", lastSent.Model.TxHash)
				lastSent.SetUnknown()
			}
			return
		case service.TxStatusFailed:
			// gas 交易链上失败后，对应归集任务回到 Waiting，下一轮重新计算是否需要补 gas。
			lastSent.SetFailed()
			chainkitcollecttasks.NewRecord().SetWaitingByGasTaskId(lastSent.Model.Id)
		case service.TxStatusConfirmed:
			txFee := decimal.Zero
			gasUsed, gasPrice, err := srv.GetGasUsedAndEffectiveGasPrice(lastSent.Model.TxHash)
			if err == nil {
				txFee = gasUsed.Mul(gasPrice)
			} else {
				log.Error("failed to get gas used and effective gas price", "error", err, "tx hash", lastSent.Model.TxHash)
			}

			lastSent.SetConfirmed(gasUsed, txFee)
			// gas 到账确认后，相关归集任务才允许进入 ERC20 发送阶段。
			chainkitcollecttasks.NewRecord().SetCanSendByGasTaskId(lastSent.Model.Id)
		}
		return
	}

	waiting := chainkitcollectgasfeetasks.NewRecord().GetOneWaiting(chain.Model.Id)
	if !waiting.Exists() {
		return
	}
	// 发送前先原子抢占 Waiting 任务，防止多实例同时从 gas provider 发出多笔补 gas。
	claimed, err := waiting.ClaimWaiting()
	if err != nil {
		log.Error("failed to claim gas task", "error", err, "gas task id", waiting.Model.Id)
		return
	}
	if !claimed {
		return
	}

	err = srv.SetFromByMnemonicAddress(collectConf.Model.GasProviderMnemonicAddressId, Password)
	if err != nil {
		log.Error("failed to set from mnemonic address", "error", err)
		waiting.SetWaitingWithError(err.Error())
		return
	}

	amount := waiting.Model.RequiredAmount
	currentBalance, err := srv.GetFromETHBalance()
	if err != nil {
		log.Error("failed to get from address balance", "error", err)
		waiting.SetWaitingWithError(err.Error())
		return
	}
	gasPrice, err := srv.SuggestGasPrice()
	if err != nil {
		log.Error("failed to suggest gas price", "error", err)
		waiting.SetWaitingWithError(err.Error())
		return
	}
	thisGas := waiting.Model.GasLimit.Mul(gasPrice)
	totalNeed := thisGas.Add(amount)
	if currentBalance.LessThan(totalNeed) {
		// gas provider 余额不足时不失败任务，保留 Waiting 以便充值后自动恢复。
		log.Error("from address balance not enough", "current balance", currentBalance, "total need", totalNeed)
		waiting.SetWaitingWithError("from address balance not enough")
		return
	}

	opts := []service.Option{
		service.GasLimit(waiting.Model.GasLimit),
		service.GasPrice(gasPrice),
		service.CheckBalance(false),
	}
	if waiting.Model.Nonce > 0 {
		opts = append(opts, service.Nonce(waiting.Model.Nonce))
	}

	hash, n, fakeErr, err := srv.TransferETH(
		waiting.Model.ToAddress,
		amount,
		opts...,
	)
	if err != nil {
		log.Error("failed to transfer ETH", "error", err)
		waiting.SetWaitingWithError(err.Error())
		return
	}
	if fakeErr != nil {
		// RPC 报错但已拿到签名交易 hash，链上可能已接收，进入 MaybeSent 由状态检查确认。
		log.Error("failed to transfer ETH due to fake error", "error", fakeErr)
		waiting.SetMaybeSent(amount, hash, n, gasPrice, fakeErr.Error())
		return
	}
	waiting.SetSent(amount, hash, n, gasPrice)
}
