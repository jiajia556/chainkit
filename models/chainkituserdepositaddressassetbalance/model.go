package chainkituserdepositaddressassetbalance

import (
	"time"

	"github.com/shopspring/decimal"
)

type ChainUserDepositAddressAssetBalance struct {
	Id                   uint64          `gorm:"column:id;primaryKey;unsigned;autoIncrement;notNull" json:"id"`
	UserId               uint64          `gorm:"column:user_id;unsigned;notNull" json:"user_id"`
	ChainDbId            uint64          `gorm:"column:chain_db_id;unsigned;notNull" json:"chain_db_id"`
	TokenId              uint64          `gorm:"column:token_id;unsigned;notNull" json:"token_id"`
	UserDepositAddressId uint64          `gorm:"column:user_deposit_address_id;unsigned;notNull" json:"user_deposit_address_id"`
	Address              string          `gorm:"column:address;notNull" json:"address"`
	ConfirmedInAmount    decimal.Decimal `gorm:"column:confirmed_in_amount;notNull;default:0" json:"confirmed_in_amount"`
	CollectedOutAmount   decimal.Decimal `gorm:"column:collected_out_amount;notNull;default:0" json:"collected_out_amount"`
	ManualOutAmount      decimal.Decimal `gorm:"column:manual_out_amount;default:0;notNull" json:"manual_out_amount"`
	BalanceAmount        decimal.Decimal `gorm:"column:balance_amount;notNull;default:0" json:"balance_amount"`
	LastInTxHash         string          `gorm:"column:last_in_tx_hash;default:null" json:"last_in_tx_hash"`
	LastCollectTxHash    string          `gorm:"column:last_collect_tx_hash;default:null" json:"last_collect_tx_hash"`
	CreatedAt            time.Time       `gorm:"column:created_at;notNull;default:current_timestamp" json:"created_at"`
	UpdatedAt            time.Time       `gorm:"column:updated_at;notNull;default:current_timestamp" json:"updated_at"`
}

func (data *ChainUserDepositAddressAssetBalance) ID() uint64 {
	return data.Id
}

func (data *ChainUserDepositAddressAssetBalance) TableName() string {
	return "chain_user_deposit_address_asset_balance"
}

func (data *ChainUserDepositAddressAssetBalance) GetCreateDDL() string {
	return "CREATE TABLE `chain_user_deposit_address_asset_balance` (   `id` bigint unsigned NOT NULL AUTO_INCREMENT,   `user_id` bigint unsigned NOT NULL COMMENT '用户ID',   `chain_db_id` bigint unsigned NOT NULL COMMENT '链ID',   `token_id` bigint unsigned NOT NULL COMMENT '代币ID',   `user_deposit_address_id` bigint unsigned NOT NULL COMMENT '充值地址ID',   `address` char(42) NOT NULL COMMENT '地址',   `confirmed_in_amount` decimal(36,0) NOT NULL DEFAULT '0' COMMENT '已确认累计流入',   `collected_out_amount` decimal(36,0) NOT NULL DEFAULT '0' COMMENT '已归集累计流出',   `manual_out_amount` decimal(36,0) NOT NULL DEFAULT '0' COMMENT '人工转出累计',   `balance_amount` decimal(36,0) NOT NULL DEFAULT '0' COMMENT '当前可归集余额',   `last_in_tx_hash` char(42) DEFAULT NULL COMMENT '最后一笔流入交易hash',   `last_collect_tx_hash` char(42) DEFAULT NULL COMMENT '最后一笔归集交易hash',   `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,   `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,   PRIMARY KEY (`id`),   UNIQUE KEY `uk_addr_token` (`chain_db_id`,`user_deposit_address_id`,`token_id`),   KEY `idx_user_id` (`user_id`),   KEY `idx_user_address_id` (`user_deposit_address_id`),   KEY `idx_tb` (`token_id`,`balance_amount`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='充值地址资产余额表';"
}
