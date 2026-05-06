package chainkituserdepositaddressassetbalance

import (
	"fmt"
	"time"

	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/chainkit/models/chainkittokens"
	"github.com/jiajia556/chainkit/service"
	"github.com/jiajia556/tool-box/mysqlx"
	"github.com/shopspring/decimal"
)

type Record struct {
	*models.BaseRecord[*ChainUserDepositAddressAssetBalance]
}

func NewRecord(session ...mysqlx.Session) *Record {
	var dbSession mysqlx.Session
	if len(session) > 0 {
		dbSession = session[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbSession.CreateTableIfNotExists(new(ChainUserDepositAddressAssetBalance))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainUserDepositAddressAssetBalance]{
			Session: dbSession,
			Model:   new(ChainUserDepositAddressAssetBalance),
		},
	}
	return r
}

func (r *Record) ReadByChainAndAddressAndToken(chainDbId, userDepAddrId, tokenId uint64) *Record {
	r.DB().Where("chain_db_id = ? AND user_deposit_address_id = ? AND token_id = ?", chainDbId, userDepAddrId, tokenId).Take(&r.Model)
	return r
}

func (r *Record) Deposit(amount decimal.Decimal, hash string) error {
	chainBalance, err := r.GetBalanceFromChain()
	if err != nil {
		return fmt.Errorf("failed to get balance from chain, error: %w", err)
	}
	sql := fmt.Sprintf(
		"UPDATE `chain_user_deposit_address_asset_balance` SET `balance_amount` = %s, `last_in_tx_hash` = %s, `confirmed_in_amount` = `confirmed_in_amount` + %s, `updated_at` = %s WHERE `id` = %d;",
		chainBalance.String(), hash, amount.String(), time.Now().Format("2006-01-02 15:04:05"), r.Model.Id,
	)
	return r.DB().Exec(sql).Error
}

func (r *Record) GetBalanceFromChain() (decimal.Decimal, error) {
	token := chainkittokens.NewRecord()
	_ = token.Read(r.Model.TokenId)
	if !token.Exists() {
		return decimal.Zero, fmt.Errorf("token not found, tokenId: %d", r.Model.TokenId)
	}
	cs, err := service.NewChainService(r.Model.ChainDbId)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to create chain service, chainDbId: %d, error: %w", r.Model.ChainDbId, err)
	}
	chainBalance, err := cs.BalanceOf(token.Model.ContractAddress, r.Model.Address)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get chain balance, chainDbId: %d, error: %w", r.Model.ChainDbId, err)
	}
	return chainBalance, nil
}
