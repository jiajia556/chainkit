package chainkitasset

import (
	"time"

	"github.com/shopspring/decimal"
)

type ChainAsset struct {
	Id        uint64          `gorm:"column:id;primaryKey;unsigned;autoIncrement;notNull" json:"id"`
	UserId    uint64          `gorm:"column:user_id;notNull;unsigned" json:"user_id"`
	TokenId   uint64          `gorm:"column:token_id;unsigned;notNull" json:"token_id"`
	Available decimal.Decimal `gorm:"column:available;notNull;default:0" json:"available"`
	Frozen    decimal.Decimal `gorm:"column:frozen;notNull;default:0" json:"frozen"`
	Total     decimal.Decimal `gorm:"column:total" json:"total"`
	Version   uint64          `gorm:"column:version;default:0;unsigned;notNull" json:"version"`
	CreatedAt time.Time       `gorm:"column:created_at;notNull;default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at;notNull;default:current_timestamp" json:"updated_at"`
}

func (data *ChainAsset) ID() uint64 {
	return data.Id
}

func (data *ChainAsset) TableName() string {
	return "chain_asset"
}

func (data *ChainAsset) GetCreateDDL() string {
	return "CREATE TABLE `chain_asset` (   `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',   `user_id` bigint unsigned NOT NULL COMMENT '用户ID',   `token_id` bigint unsigned NOT NULL COMMENT '币种ID',   `available` decimal(36,0) NOT NULL DEFAULT '0' COMMENT '可用余额，最小单位整数',   `frozen` decimal(36,0) NOT NULL DEFAULT '0' COMMENT '冻结余额，最小单位整数',   `total` decimal(36,0) GENERATED ALWAYS AS ((`available` + `frozen`)) STORED COMMENT '总余额，最小单位整数',   `version` bigint unsigned NOT NULL DEFAULT '0' COMMENT '乐观锁版本号',   `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',   `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',   PRIMARY KEY (`id`),   UNIQUE KEY `uk_user_token` (`user_id`,`token_id`),   KEY `idx_token_total` (`token_id`,`total`),   KEY `idx_updated_at` (`updated_at`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户资产表';"
}
