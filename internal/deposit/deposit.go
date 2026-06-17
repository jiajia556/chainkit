package deposit

import (
	"context"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jiajia556/chainkit/models/chainkitasset"
	"github.com/jiajia556/chainkit/models/chainkitchains"
	"github.com/jiajia556/chainkit/models/chainkitdepositrecord"
	"github.com/jiajia556/chainkit/models/chainkitdeposittokens"
	"github.com/jiajia556/chainkit/models/chainkittokens"
	"github.com/jiajia556/chainkit/models/chainkituserdepositaddress"
	"github.com/jiajia556/chainkit/models/chainkituserdepositaddressassetbalance"
	"github.com/jiajia556/chainkit/pkg/contracts/erc20"
	"github.com/jiajia556/chainkit/service"
	"github.com/jiajia556/tool-box/log"
	"github.com/shopspring/decimal"
)

var BlockNum uint64

func Start(ctx context.Context) {
	log.Debug("deposit service started")
	chains := chainkitchains.NewList()
	err := chains.FindAll()
	if err != nil {
		log.Error("failed to find chains", "error", err)
		return
	}
	log.Debug("found chains", "count", len(*chains.Records))
	wg := &sync.WaitGroup{}
	chains.Foreach(func(key int, chain *chainkitchains.Record) bool {
		wg.Add(1)
		go handleChain(chain, wg)
		return true
	})
	wg.Wait()
}

func handleChain(chain *chainkitchains.Record, wg *sync.WaitGroup) {
	log.Debug("handling chain")
	defer wg.Done()
	depositTokens := chainkitdeposittokens.NewList().FindAvailableByChainDBID(chain.Model.Id)

	log.Debug("found deposit tokens", "chainDbId", chain.Model.Id, "count", len(*depositTokens.Records))

	cs, err := service.NewChainService(chain.Model.Id)
	if err != nil {
		log.Error("failed to create chain service", "chainDbId", chain.Model.Id, "error", err)
		return
	}

	wg2 := &sync.WaitGroup{}
	depositTokens.Foreach(func(key int, depositToken *chainkitdeposittokens.Record) bool {
		token := chainkittokens.NewRecord()
		_ = token.Read(depositToken.Model.TokenId)
		if !token.Exists() || token.Model.ContractAddress == "0x0000000000000000000000000000000000000000" {
			log.Error("invalid token", "token", token.Model.ContractAddress)
			return true
		}
		log.Debug("scanning deposit token", "chainDbId", chain.Model.Id, "tokenId", token.Model.Id, "contractAddress", token.Model.ContractAddress)
		wg2.Add(1)
		go func(token *chainkittokens.Record, initStartBlock uint64) {
			defer wg2.Done()

			err := cs.ScanBlock(
				context.WithValue(context.Background(), "minDepositAmount", depositToken.Model.MinDepositAmount),
				token.Model.ContractAddress,
				"deposit",
				handleDeposit,
				service.StartBlock(initStartBlock),
				service.Step(BlockNum),
			)
			if err != nil {
				log.Error("failed to scan block", "tokenID", token.Model.Id, "error", err)
			}
		}(token, depositToken.Model.InitStartBlock)
		return true
	})
	wg2.Wait()
}

func handleDeposit(logCtx *service.LogContext, eventLog types.Log) error {
	log.Debug("handling deposit", "log count", len(eventLog.Data), "chainDbId", logCtx.ChainDbId)
	erc20Abi, _ := erc20.NewErc20(common.Address{}, nil)
	transfer, err := erc20Abi.ParseTransfer(eventLog)
	if err != nil {
		return nil
	}

	to := transfer.To.Hex()
	from := transfer.From.Hex()
	amount := decimal.NewFromBigInt(transfer.Value, 0)
	minDepositAmount, ok := logCtx.Ctx.Value("minDepositAmount").(decimal.Decimal)
	if !ok {
		minDepositAmount = decimal.Zero
	}
	if amount.LessThan(minDepositAmount) {
		return nil
	}
	hash := eventLog.TxHash.Hex()

	if amount.Equal(decimal.Zero) {
		return nil
	}

	userDepAddr := chainkituserdepositaddress.NewRecord().ReadByAddress(to)
	if !userDepAddr.Exists() {
		return nil
	}
	token := chainkittokens.NewRecord(logCtx.Session).ReadByChainAndContractAddress(logCtx.ChainDbId, eventLog.Address.Hex())
	if !token.Exists() {
		return nil
	}
	eventLogId, created, err := logCtx.SaveEventLog(eventLog)
	if err != nil {
		return err
	}
	if !created {
		return nil
	}

	userDepAddrBalance := chainkituserdepositaddressassetbalance.NewRecord(logCtx.Session).
		ReadByChainAndAddressAndToken(logCtx.ChainDbId, userDepAddr.Model.Id, token.Model.Id)
	if !userDepAddrBalance.Exists() {
		userDepAddrBalance.Model.UserId = userDepAddr.Model.UserId
		userDepAddrBalance.Model.ChainDbId = logCtx.ChainDbId
		userDepAddrBalance.Model.TokenId = token.Model.Id
		userDepAddrBalance.Model.UserDepositAddressId = userDepAddr.Model.Id
		userDepAddrBalance.Model.Address = userDepAddr.Model.Address
		userDepAddrBalance.Model.ConfirmedInAmount = decimal.Zero
		userDepAddrBalance.Model.BalanceAmount = decimal.Zero
		userDepAddrBalance.Model.LastInTxHash = ""
		err = userDepAddrBalance.Create()
		if err != nil {
			userDepAddrBalance = chainkituserdepositaddressassetbalance.NewRecord(logCtx.Session).
				ReadByChainAndAddressAndToken(logCtx.ChainDbId, userDepAddr.Model.Id, token.Model.Id)
		}
	}
	err = userDepAddrBalance.Deposit(amount, hash)
	if err != nil {
		log.Debug("failed to update user deposit address asset balance", "chainDbId", logCtx.ChainDbId, "userDepositAddressId", userDepAddr.Model.Id, "tokenId", token.Model.Id, "amount", amount.String(), "hash", hash, "error", err)
		return err
	}

	amountDecimal := amount.Div(decimal.New(1, int32(token.Model.Decimals)))

	depRecord := chainkitdepositrecord.NewRecord(logCtx.Session)
	depRecord.Model.UserId = userDepAddr.Model.UserId
	depRecord.Model.ChainDbId = logCtx.ChainDbId
	depRecord.Model.TokenId = token.Model.Id
	depRecord.Model.EventLogId = eventLogId
	depRecord.Model.UserDepositAddressId = userDepAddr.Model.Id
	depRecord.Model.FromAddress = from
	depRecord.Model.ToAddress = to
	depRecord.Model.Amount = amount
	depRecord.Model.AmountDecimal = amountDecimal
	depRecord.Model.Remark = hash
	_ = depRecord.Create()

	if userDepAddr.Model.UserId == 0 {
		return nil
	}

	userAsset := chainkitasset.NewRecord(logCtx.Session).GetByUserAndTokenGroup(userDepAddr.Model.UserId, token.Model.TokenGroupId)
	err = userAsset.IncreaseBalance(amount, "deposit", depRecord.Model.Id, fmt.Sprintf("deposit_%d", depRecord.Model.Id), "")
	if err != nil {
		log.Debug("failed to increase user asset balance", "error", err)
		return err
	}

	return nil
}
