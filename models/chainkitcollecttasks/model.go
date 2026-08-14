package chainkitcollecttasks

import (
	"time"

	"github.com/shopspring/decimal"
)

type ChainCollectTasks struct {
	Id                   uint64          `gorm:"column:id;unsigned;autoIncrement;notNull;primaryKey" json:"id"`
	ChainDbId            uint64          `gorm:"column:chain_db_id;unsigned;notNull" json:"chain_db_id"`
	TokenId              uint64          `gorm:"column:token_id;notNull;unsigned" json:"token_id"`
	UserId               uint64          `gorm:"column:user_id;unsigned;notNull" json:"user_id"`
	UserDepositAddressId uint64          `gorm:"column:user_deposit_address_id;unsigned;notNull" json:"user_deposit_address_id"`
	FromAddress          string          `gorm:"column:from_address;notNull" json:"from_address"`
	ToAddress            string          `gorm:"column:to_address;notNull" json:"to_address"`
	PlanAmount           decimal.Decimal `gorm:"column:plan_amount;default:0;unsigned;notNull" json:"plan_amount"`
	ActualAmount         decimal.Decimal `gorm:"column:actual_amount;unsigned;default:null" json:"actual_amount"`
	GasRequiredAmount    decimal.Decimal `gorm:"column:gas_required_amount;unsigned;default:null" json:"gas_required_amount"`
	GasBalanceBeforeTx   decimal.Decimal `gorm:"column:gas_balance_before_tx;unsigned;default:null" json:"gas_balance_before_tx"`
	TxHash               string          `gorm:"column:tx_hash;default:null" json:"tx_hash"`
	Nonce                uint64          `gorm:"column:nonce;unsigned;default:null" json:"nonce"`
	GasLimit             decimal.Decimal `gorm:"column:gas_limit;unsigned;notNull" json:"gas_limit"`
	GasPrice             decimal.Decimal `gorm:"column:gas_price;unsigned;default:null" json:"gas_price"`
	MaxFeePerGas         decimal.Decimal `gorm:"column:max_fee_per_gas;default:null;unsigned" json:"max_fee_per_gas"`
	MaxPriorityFeePerGas decimal.Decimal `gorm:"column:max_priority_fee_per_gas;unsigned;default:null" json:"max_priority_fee_per_gas"`
	GasUsed              decimal.Decimal `gorm:"column:gas_used;unsigned;default:null" json:"gas_used"`
	TxFee                decimal.Decimal `gorm:"column:tx_fee;unsigned;default:null" json:"tx_fee"`
	Status               uint8           `gorm:"column:status;unsigned;notNull;default:0" json:"status"`
	GasTaskId            uint64          `gorm:"column:gas_task_id;notNull;unsigned" json:"gas_task_id"`
	LastError            string          `gorm:"column:last_error;default:null" json:"last_error"`
	SentAt               time.Time       `gorm:"column:sent_at;default:null" json:"sent_at"`
	ConfirmedAt          time.Time       `gorm:"column:confirmed_at;default:null" json:"confirmed_at"`
	Remark               string          `gorm:"column:remark;default:null" json:"remark"`
	CreatedAt            time.Time       `gorm:"column:created_at;notNull;default:current_timestamp" json:"created_at"`
	UpdatedAt            time.Time       `gorm:"column:updated_at;default:current_timestamp;notNull" json:"updated_at"`
}

func (data *ChainCollectTasks) ID() uint64 {
	return data.Id
}

func (data *ChainCollectTasks) TableName() string {
	return "chain_collect_tasks"
}

func (data *ChainCollectTasks) GetCreateDDL() string {
	return "CREATE TABLE `chain_collect_tasks` (\n  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',\n  `chain_db_id` bigint unsigned NOT NULL COMMENT '链配置表ID',\n  `token_id` bigint unsigned NOT NULL COMMENT '代币ID，0表示原生币',\n  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',\n  `user_deposit_address_id` bigint unsigned NOT NULL COMMENT '用户充值地址ID',\n  `from_address` char(42) NOT NULL COMMENT '归集来源地址',\n  `to_address` char(42) NOT NULL COMMENT '归集目标地址',\n  `plan_amount` decimal(36,0) unsigned NOT NULL DEFAULT '0' COMMENT '创建任务时链上余额快照，最小单位',\n  `actual_amount` decimal(36,0) unsigned DEFAULT NULL COMMENT '实际归集金额，最小单位',\n  `gas_required_amount` decimal(36,0) unsigned DEFAULT NULL COMMENT '预计需要gas数量，原生币最小单位',\n  `gas_balance_before_tx` decimal(36,0) unsigned DEFAULT NULL COMMENT '执行前原生币余额，最小单位',\n  `tx_hash` char(66) DEFAULT NULL COMMENT '归集交易hash',\n  `nonce` bigint unsigned DEFAULT NULL COMMENT '交易nonce',\n  `gas_limit` decimal(36,0) unsigned NOT NULL COMMENT '交易gas limit',\n  `gas_price` decimal(36,0) unsigned DEFAULT NULL COMMENT 'legacy gas price，wei',\n  `max_fee_per_gas` decimal(36,0) unsigned DEFAULT NULL COMMENT 'EIP-1559 maxFeePerGas，wei',\n  `max_priority_fee_per_gas` decimal(36,0) unsigned DEFAULT NULL COMMENT 'EIP-1559 maxPriorityFeePerGas，wei',\n  `gas_used` decimal(36,0) unsigned DEFAULT NULL COMMENT '实际消耗gas',\n  `tx_fee` decimal(36,0) unsigned DEFAULT NULL COMMENT '实际交易手续费，原生币最小单位',\n  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '状态：0待处理 1等待gas 2可执行 3发送中 4已广播 5已确认 6失败 7取消 8跳过',\n  `gas_task_id` bigint unsigned NOT NULL,\n  `last_error` varchar(1024) DEFAULT NULL COMMENT '最后错误信息',\n  `sent_at` datetime DEFAULT NULL COMMENT '广播时间',\n  `confirmed_at` datetime DEFAULT NULL COMMENT '确认时间',\n  `remark` varchar(255) DEFAULT NULL COMMENT '备注',\n  `created_at` datetime NOT NULL COMMENT '创建时间',\n  `updated_at` datetime NOT NULL COMMENT '更新时间',\n  PRIMARY KEY (`id`),\n  KEY `idx_address_token_status` (`user_deposit_address_id`,`token_id`,`status`),\n  KEY `idx_chain_status` (`chain_db_id`,`status`),\n  KEY `idx_address_chain_status` (`user_deposit_address_id`,`chain_db_id`,`status`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='链上资产归集任务表';"
}
