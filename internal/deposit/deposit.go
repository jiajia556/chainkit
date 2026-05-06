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
	"github.com/jiajia556/tool-box/mysqlx"
	"github.com/shopspring/decimal"
)

func Start(ctx context.Context) {
	chains := chainkitchains.NewList()
	err := chains.FindAll()
	if err != nil {
		log.Error("failed to find chains", "error", err)
		return
	}
	wg := &sync.WaitGroup{}
	chains.Foreach(func(key int, chain *chainkitchains.Record) bool {
		wg.Add(1)
		go handleChain(chain, wg)
		return true
	})
	wg.Wait()
}

func handleChain(chain *chainkitchains.Record, wg *sync.WaitGroup) {
	defer wg.Done()
	depositTokens := chainkitdeposittokens.NewList().FindAvailableByChainDBID(chain.Model.Id)

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
			return true
		}
		wg2.Add(1)
		go func(token *chainkittokens.Record, initStartBlock uint64) {
			defer wg2.Done()

			err := cs.ScanBlock(
				token.Model.ContractAddress,
				"deposit",
				handleDeposit,
				service.StartBlock(initStartBlock),
			)
			if err != nil {
				log.Error("failed to scan block", "tokenID", token.Model.Id, "error", err)
			}
		}(token, depositToken.Model.InitStartBlock)
		return true
	})
	wg2.Wait()
}

func handleDeposit(mysqlSession *mysqlx.TxSession, log types.Log, eventLogId uint64, chainDbID uint64) error {
	erc20Abi, _ := erc20.NewErc20(common.Address{}, nil)
	transfer, err := erc20Abi.ParseTransfer(log)
	if err != nil {
		return nil
	}

	to := transfer.To.Hex()
	from := transfer.From.Hex()
	amount := decimal.NewFromBigInt(transfer.Value, 0)
	hash := log.TxHash.Hex()

	userDepAddr := chainkituserdepositaddress.NewRecord().ReadByAddress(to)
	if !userDepAddr.Exists() {
		return nil
	}
	token := chainkittokens.NewRecord(mysqlSession).ReadByChainAndContractAddress(chainDbID, log.Address.Hex())
	if !token.Exists() {
		return nil
	}
	userDepAddrBalance := chainkituserdepositaddressassetbalance.NewRecord(mysqlSession).
		ReadByChainAndAddressAndToken(chainDbID, userDepAddr.Model.Id, token.Model.Id)
	if !userDepAddrBalance.Exists() {
		userDepAddrBalance.Model.UserId = userDepAddr.Model.UserId
		userDepAddrBalance.Model.ChainDbId = chainDbID
		userDepAddrBalance.Model.TokenId = token.Model.Id
		userDepAddrBalance.Model.UserDepositAddressId = userDepAddr.Model.Id
		userDepAddrBalance.Model.Address = userDepAddr.Model.Address
		userDepAddrBalance.Model.ConfirmedInAmount = decimal.Zero
		userDepAddrBalance.Model.BalanceAmount = decimal.Zero
		userDepAddrBalance.Model.LastInTxHash = ""
		err = userDepAddrBalance.Create()
		if err != nil {
			userDepAddrBalance = chainkituserdepositaddressassetbalance.NewRecord(mysqlSession).
				ReadByChainAndAddressAndToken(chainDbID, userDepAddr.Model.Id, token.Model.Id)
		}
	}
	_ = userDepAddrBalance.Deposit(amount, hash)

	amountDecimal := amount.Div(decimal.New(1, int32(token.Model.Decimals)))

	depRecord := chainkitdepositrecord.NewRecord(mysqlSession)
	depRecord.Model.UserId = userDepAddr.Model.UserId
	depRecord.Model.ChainDbId = chainDbID
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

	userAsset := chainkitasset.NewRecord(mysqlSession).ReadByUserAndToken(userDepAddr.Model.UserId, token.Model.Id)
	_ = userAsset.IncreaseBalance(amount, "deposit", depRecord.Model.Id, fmt.Sprintf("deposit_%d", depRecord.Model.Id), "")

	return nil
}
