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
		status, err := srv.GetTxStatus(lastSent.Model.TxHash)
		if err != nil {
			log.Error("failed to get tx status", "error", err, "tx hash", lastSent.Model.TxHash)
			return
		}
		log.Debug("last sent gas task", "chain db id", chain.Model.Id, "tx hash", lastSent.Model.TxHash, "status", status)
		switch status {
		case service.TxStatusNotFound:
			if lastSent.SinceCreated() > time.Minute*1 {
				occupied, err := srv.IsNonceOccupied(lastSent.Model.FromAddress, lastSent.Model.Nonce)
				if err != nil {
					log.Error("failed to check if nonce is occupied", "error", err, "from address", lastSent.Model.FromAddress, "nonce", lastSent.Model.Nonce)
					return
				}
				if occupied {
					// 特殊情况，人工处理
					log.Debug("nonce is occupied, set last sent gas task to unknown, need manual handling", "from address", lastSent.Model.FromAddress, "nonce", lastSent.Model.Nonce)
					lastSent.SetUnknown()
				} else {
					lastSent.SetWaiting()
				}
			}
		case service.TxStatusPending, service.TxStatusMined:
			return
		case service.TxStatusFailed:
			// gasTask置为失败，与其对应的collectTask置回等待
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
			chainkitcollecttasks.NewRecord().SetCanSendByGasTaskId(lastSent.Model.Id)
		}
		return
	}

	waiting := chainkitcollectgasfeetasks.NewRecord().GetOneWaiting(chain.Model.Id)
	if !waiting.Exists() {
		return
	}

	err = srv.SetFromByMnemonicAddress(collectConf.Model.GasProviderMnemonicAddressId, Password)
	if err != nil {
		log.Error("failed to set from mnemonic address", "error", err)
		return
	}

	amount := waiting.Model.RequiredAmount
	currentBalance, err := srv.GetFromETHBalance()
	if err != nil {
		log.Error("failed to get from address balance", "error", err)
		return
	}
	gasPrice, err := srv.SuggestGasPrice()
	if err != nil {
		log.Error("failed to suggest gas price", "error", err)
		return
	}
	thisGas := waiting.Model.GasLimit.Mul(gasPrice)
	totalNeed := thisGas.Add(amount)
	if currentBalance.LessThan(totalNeed) {
		log.Error("from address balance not enough", "current balance", currentBalance, "total need", totalNeed)
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
		return
	}
	if fakeErr != nil {
		log.Error("failed to transfer ETH due to fake error", "error", fakeErr)
	}
	waiting.SetSent(amount, hash, n, gasPrice)
}
