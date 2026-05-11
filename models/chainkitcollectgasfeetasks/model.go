package chainkitcollectgasfeetasks

import (
	"time"

	"github.com/shopspring/decimal"
)

type ChainCollectGasFeeTasks struct {
	Id                   uint64          `gorm:"column:id;unsigned;autoIncrement;notNull;primaryKey" json:"id"`
	ChainDbId            uint64          `gorm:"column:chain_db_id;unsigned;notNull" json:"chain_db_id"`
	UserId               uint64          `gorm:"column:user_id;unsigned;notNull" json:"user_id"`
	UserDepositAddressId uint64          `gorm:"column:user_deposit_address_id;unsigned;notNull" json:"user_deposit_address_id"`
	FromAddress          string          `gorm:"column:from_address;notNull" json:"from_address"`
	ToAddress            string          `gorm:"column:to_address;notNull" json:"to_address"`
	RequiredAmount       decimal.Decimal `gorm:"column:required_amount;notNull;default:0" json:"required_amount"`
	CurrentBalance       decimal.Decimal `gorm:"column:current_balance;default:0;notNull" json:"current_balance"`
	SendAmount           decimal.Decimal `gorm:"column:send_amount;notNull;default:0" json:"send_amount"`
	TxHash               string          `gorm:"column:tx_hash;default:null" json:"tx_hash"`
	Nonce                uint64          `gorm:"column:nonce;unsigned;default:null" json:"nonce"`
	GasLimit             decimal.Decimal `gorm:"column:gas_limit;unsigned;notNull" json:"gas_limit"`
	GasPrice             decimal.Decimal `gorm:"column:gas_price;default:null" json:"gas_price"`
	MaxFeePerGas         decimal.Decimal `gorm:"column:max_fee_per_gas;default:null" json:"max_fee_per_gas"`
	MaxPriorityFeePerGas decimal.Decimal `gorm:"column:max_priority_fee_per_gas;default:null" json:"max_priority_fee_per_gas"`
	GasUsed              uint64          `gorm:"column:gas_used;unsigned;default:null" json:"gas_used"`
	TxFee                decimal.Decimal `gorm:"column:tx_fee;default:null" json:"tx_fee"`
	Status               uint8           `gorm:"column:status;default:0;unsigned;notNull" json:"status"`
	LastError            string          `gorm:"column:last_error;default:null" json:"last_error"`
	SentAt               time.Time       `gorm:"column:sent_at;default:null" json:"sent_at"`
	ConfirmedAt          time.Time       `gorm:"column:confirmed_at;default:null" json:"confirmed_at"`
	Remark               string          `gorm:"column:remark;default:null" json:"remark"`
	CreatedAt            time.Time       `gorm:"column:created_at;notNull;default:current_timestamp" json:"created_at"`
	UpdatedAt            time.Time       `gorm:"column:updated_at;notNull;default:current_timestamp" json:"updated_at"`
}

func (data *ChainCollectGasFeeTasks) ID() uint64 {
	return data.Id
}

func (data *ChainCollectGasFeeTasks) TableName() string {
	return "chain_collect_gas_fee_tasks"
}

func (data *ChainCollectGasFeeTasks) GetCreateDDL() string {
	return "CREATE TABLE `chain_collect_gas_fee_tasks` (   `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',   `chain_db_id` bigint unsigned NOT NULL COMMENT '链配置表ID',   `user_id` bigint unsigned NOT NULL COMMENT '用户ID',   `user_deposit_address_id` bigint unsigned NOT NULL COMMENT '用户充值地址ID',   `from_address` char(42) NOT NULL COMMENT 'gas来源地址',   `to_address` char(42) NOT NULL COMMENT 'gas接收地址，即用户充值地址',   `required_amount` decimal(36,0) NOT NULL DEFAULT '0' COMMENT '预计需要gas数量，原生币最小单位',   `current_balance` decimal(36,0) NOT NULL DEFAULT '0' COMMENT '创建任务时接收地址原生币余额，最小单位',   `send_amount` decimal(36,0) NOT NULL DEFAULT '0' COMMENT '实际转gas数量，原生币最小单位',   `tx_hash` char(66) DEFAULT NULL COMMENT '转gas交易hash',   `nonce` bigint unsigned DEFAULT NULL COMMENT '交易nonce',   `gas_limit` decimal(36,0) unsigned NOT NULL COMMENT '交易gas limit，一般原生币转账为21000',   `gas_price` decimal(36,0) DEFAULT NULL COMMENT 'legacy gas price，wei',   `max_fee_per_gas` decimal(36,0) DEFAULT NULL COMMENT 'EIP-1559 maxFeePerGas，wei',   `max_priority_fee_per_gas` decimal(36,0) DEFAULT NULL COMMENT 'EIP-1559 maxPriorityFeePerGas，wei',   `gas_used` bigint unsigned DEFAULT NULL COMMENT '实际消耗gas',   `tx_fee` decimal(36,0) DEFAULT NULL COMMENT '实际手续费，原生币最小单位',   `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '状态：0待处理 1发送中 2已广播 3已确认 4失败 5取消 6跳过',   `last_error` varchar(1024) DEFAULT NULL COMMENT '最后错误信息',   `sent_at` datetime DEFAULT NULL COMMENT '广播时间',   `confirmed_at` datetime DEFAULT NULL COMMENT '确认时间',   `remark` varchar(255) DEFAULT NULL COMMENT '备注',   `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',   `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',   PRIMARY KEY (`id`),   KEY `idx_chain_status` (`chain_db_id`,`status`),   KEY `idx_to_address_status` (`chain_db_id`,`user_deposit_address_id`,`status`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='链上转Gas Fee任务表';"
}
