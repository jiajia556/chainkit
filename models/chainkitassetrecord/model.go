package chainkitassetrecord

import (
	"time"

	"github.com/shopspring/decimal"
)

type ChainAssetRecord struct {
	Id              uint64          `gorm:"column:id;primaryKey;unsigned;autoIncrement;notNull" json:"id"`
	UserId          uint64          `gorm:"column:user_id;unsigned;notNull" json:"user_id"`
	TokenGroupId    uint64          `gorm:"column:token_group_id;unsigned;notNull" json:"token_group_id"`
	Symbol          string          `gorm:"column:symbol;notNull" json:"symbol"`
	BizType         string          `gorm:"column:biz_type;notNull" json:"biz_type"`
	BizId           uint64          `gorm:"column:biz_id;unsigned;notNull" json:"biz_id"`
	RequestId       string          `gorm:"column:request_id;notNull" json:"request_id"`
	AvailableChange decimal.Decimal `gorm:"column:available_change;notNull;default:0" json:"available_change"`
	FrozenChange    decimal.Decimal `gorm:"column:frozen_change;default:0;notNull" json:"frozen_change"`
	AvailableBefore decimal.Decimal `gorm:"column:available_before;notNull" json:"available_before"`
	AvailableAfter  decimal.Decimal `gorm:"column:available_after;notNull" json:"available_after"`
	FrozenBefore    decimal.Decimal `gorm:"column:frozen_before;notNull" json:"frozen_before"`
	FrozenAfter     decimal.Decimal `gorm:"column:frozen_after;notNull" json:"frozen_after"`
	Remark          string          `gorm:"column:remark;notNull;default:" json:"remark"`
	CreatedAt       time.Time       `gorm:"column:created_at;notNull;default:current_timestamp" json:"created_at"`
}

func (data *ChainAssetRecord) ID() uint64 {
	return data.Id
}

func (data *ChainAssetRecord) TableName() string {
	return "chain_asset_record"
}

func (data *ChainAssetRecord) GetCreateDDL() string {
	return "CREATE TABLE `chain_asset_record` (\n  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',\n  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',\n  `token_group_id` bigint unsigned NOT NULL COMMENT '币种ID',\n  `symbol` varchar(32) NOT NULL,\n  `biz_type` varchar(128) NOT NULL COMMENT '业务类型',\n  `biz_id` bigint unsigned NOT NULL COMMENT '业务ID，例如订单ID、充值ID、提现ID、成交ID',\n  `request_id` varchar(128) NOT NULL COMMENT '幂等ID，防止重复记账',\n  `available_change` decimal(36,0) NOT NULL DEFAULT '0' COMMENT '可用余额变化，最小单位整数，可正可负',\n  `frozen_change` decimal(36,0) NOT NULL DEFAULT '0' COMMENT '冻结余额变化，最小单位整数，可正可负',\n  `available_before` decimal(36,0) NOT NULL COMMENT '变动前可用余额',\n  `available_after` decimal(36,0) NOT NULL COMMENT '变动后可用余额',\n  `frozen_before` decimal(36,0) NOT NULL COMMENT '变动前冻结余额',\n  `frozen_after` decimal(36,0) NOT NULL COMMENT '变动后冻结余额',\n  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',\n  `created_at` datetime NOT NULL COMMENT '创建时间',\n  PRIMARY KEY (`id`),\n  UNIQUE KEY `uk_request_id` (`request_id`),\n  KEY `idx_user_token_id` (`user_id`,`token_group_id`,`id`),\n  KEY `idx_user_token_time` (`user_id`,`token_group_id`,`created_at`),\n  KEY `idx_biz` (`biz_type`,`biz_id`),\n  KEY `idx_created_at` (`created_at`)\n) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户资产流水表';"
}
