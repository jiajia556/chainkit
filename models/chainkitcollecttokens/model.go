package chainkitcollecttokens

import (
	"time"

	"github.com/shopspring/decimal"
)

type ChainCollectTokens struct {
	Id               uint64          `gorm:"column:id;primaryKey;unsigned;autoIncrement;notNull" json:"id"`
	ChainDbId        uint64          `gorm:"column:chain_db_id;notNull;unsigned" json:"chain_db_id"`
	TokenId          uint64          `gorm:"column:token_id;unsigned;notNull;default:0" json:"token_id"`
	TokenAddress     string          `gorm:"column:token_address;notNull" json:"token_address"`
	Symbol           string          `gorm:"column:symbol;notNull;default:" json:"symbol"`
	Decimals         uint8           `gorm:"column:decimals;notNull;default:18;unsigned" json:"decimals"`
	MinCollectAmount decimal.Decimal `gorm:"column:min_collect_amount;notNull" json:"min_collect_amount"`
	ToAddress        string          `gorm:"column:to_address;notNull;default:" json:"to_address"`
	Status           int8            `gorm:"column:status;default:1;notNull" json:"status"`
	Remark           string          `gorm:"column:remark;notNull;default:" json:"remark"`
	CreatedAt        time.Time       `gorm:"column:created_at;notNull" json:"created_at"`
	UpdatedAt        time.Time       `gorm:"column:updated_at;notNull" json:"updated_at"`
}

func (data *ChainCollectTokens) ID() uint64 {
	return data.Id
}

func (data *ChainCollectTokens) TableName() string {
	return "chain_collect_tokens"
}

func (data *ChainCollectTokens) GetCreateDDL() string {
	return "CREATE TABLE `chain_collect_tokens` (   `id` bigint unsigned NOT NULL AUTO_INCREMENT,   `chain_db_id` bigint unsigned NOT NULL COMMENT '链配置ID，对应 chain_chains.id',   `token_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT 'token配置ID，对应 chain_tokens.id',   `token_address` char(42) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'token合约地址',   `symbol` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'token符号，如 ETH/USDT',   `decimals` tinyint unsigned NOT NULL DEFAULT '18' COMMENT 'token精度',   `min_collect_amount` decimal(36,0) NOT NULL COMMENT '最低归集数量，使用链上最小单位',   `to_address` char(42) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '归集目标地址，空则使用全局默认地址',   `status` tinyint NOT NULL DEFAULT '1' COMMENT '1-启用 0-禁用',   `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '备注',   `created_at` datetime NOT NULL,   `updated_at` datetime NOT NULL,   PRIMARY KEY (`id`),   UNIQUE KEY `idx_uniq_chain_token` (`chain_db_id`,`token_address`),   KEY `idx_chain_status` (`chain_db_id`,`status`),   KEY `idx_token_id` (`token_id`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='需要归集的token配置表';"
}
