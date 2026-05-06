package chainkitdepositrecord

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ChainDepositRecord struct {
	Id                   uint64          `gorm:"column:id;primaryKey;unsigned;autoIncrement;notNull" json:"id"`
	UserId               uint64          `gorm:"column:user_id;unsigned;notNull" json:"user_id"`
	ChainDbId            uint64          `gorm:"column:chain_db_id;unsigned;notNull" json:"chain_db_id"`
	TokenId              uint64          `gorm:"column:token_id;unsigned;notNull" json:"token_id"`
	EventLogId           uint64          `gorm:"column:event_log_id;unsigned;notNull" json:"event_log_id"`
	UserDepositAddressId uint64          `gorm:"column:user_deposit_address_id;unsigned;notNull" json:"user_deposit_address_id"`
	FromAddress          string          `gorm:"column:from_address;notNull" json:"from_address"`
	ToAddress            string          `gorm:"column:to_address;notNull" json:"to_address"`
	Amount               decimal.Decimal `gorm:"column:amount;notNull" json:"amount"`
	AmountDecimal        decimal.Decimal `gorm:"column:amount_decimal;default:null" json:"amount_decimal"`
	Remark               string          `gorm:"column:remark;default:null" json:"remark"`
	CreatedAt            time.Time       `gorm:"column:created_at;notNull;default:current_timestamp" json:"created_at"`
}

func (data *ChainDepositRecord) ID() uint64 {
	return data.Id
}

func (data *ChainDepositRecord) TableName() string {
	return "chain_deposit_record"
}

func (data *ChainDepositRecord) GetCreateDDL() string {
	return "CREATE TABLE `chain_deposit_record` (   `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',   `user_id` bigint unsigned NOT NULL COMMENT '用户ID',   `chain_db_id` bigint unsigned NOT NULL COMMENT '链配置表ID',   `token_id` bigint unsigned NOT NULL COMMENT '代币ID，0可表示原生币',   `event_log_id` bigint unsigned NOT NULL COMMENT '事件日志表ID',   `user_deposit_address_id` bigint unsigned NOT NULL COMMENT '用户充值地址表ID',   `from_address` char(42) NOT NULL COMMENT '付款地址',   `to_address` varchar(128) NOT NULL COMMENT '收款地址',   `amount` decimal(32,0) NOT NULL COMMENT '原始精度金额，按最小单位保存',   `amount_decimal` decimal(32,18) DEFAULT NULL COMMENT '格式化金额，可选，便于查表',   `remark` varchar(255) DEFAULT NULL COMMENT '备注',   `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',   PRIMARY KEY (`id`),   UNIQUE KEY `uk_chain_token_tx_log` (`chain_db_id`,`token_id`,`event_log_id`),   KEY `idx_user_id` (`user_id`),   KEY `idx_user_address_id` (`user_deposit_address_id`),   KEY `idx_to_address` (`to_address`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='充值记录表';"
}

func (data *ChainDepositRecord) BeforeCreate(tx *gorm.DB) error {
	data.CreatedAt = time.Now()
	data.FromAddress = strings.ToLower(data.FromAddress)
	data.ToAddress = strings.ToLower(data.ToAddress)
	return nil
}
