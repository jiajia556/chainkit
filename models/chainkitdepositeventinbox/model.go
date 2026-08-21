package chainkitdepositeventinbox

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	SourceWebSocket uint8 = 1 << iota
	SourceGetLogs
)

const (
	StatusOrphaned int8 = -1
	StatusIgnored  int8 = -2

	StatusPending    int8 = 0
	StatusReady      int8 = 1
	StatusProcessing int8 = 2
	StatusCompleted  int8 = 3
	StatusRetry      int8 = 4
)

type ChainDepositEventInbox struct {
	Id              uint64          `gorm:"column:id;autoIncrement;notNull;primaryKey;unsigned" json:"id"`
	ChainDbId       uint64          `gorm:"column:chain_db_id;notNull;unsigned" json:"chain_db_id"`
	ContractAddress string          `gorm:"column:contract_address;notNull" json:"contract_address"`
	TxHash          string          `gorm:"column:tx_hash;notNull" json:"tx_hash"`
	LogIndex        uint32          `gorm:"column:log_index;notNull;unsigned" json:"log_index"`
	BlockNumber     uint64          `gorm:"column:block_number;notNull;unsigned" json:"block_number"`
	BlockHash       string          `gorm:"column:block_hash;notNull" json:"block_hash"`
	FromAddress     string          `gorm:"column:from_address;default:null" json:"from_address"`
	ToAddress       string          `gorm:"column:to_address;default:null" json:"to_address"`
	Amount          decimal.Decimal `gorm:"column:amount;default:null" json:"amount"`
	Source          uint8           `gorm:"column:source;notNull;unsigned" json:"source"`
	Status          int8            `gorm:"column:status;default:0;notNull" json:"status"`
	Removed         bool            `gorm:"column:removed;default:0;notNull" json:"removed"`
	RetryCount      uint32          `gorm:"column:retry_count;default:0;notNull;unsigned" json:"retry_count"`
	NextRetryAt     *time.Time      `gorm:"column:next_retry_at;default:null" json:"next_retry_at"`
	LastError       string          `gorm:"column:last_error;default:null" json:"last_error"`
	CreatedAt       time.Time       `gorm:"column:created_at;notNull" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"column:updated_at;notNull" json:"updated_at"`
	ConfirmedAt     *time.Time      `gorm:"column:confirmed_at;default:null" json:"confirmed_at"`
	ProcessedAt     *time.Time      `gorm:"column:processed_at;default:null" json:"processed_at"`
}

func (data *ChainDepositEventInbox) BeforeSave(_ *gorm.DB) error {
	data.normalize()
	return nil
}

func (data *ChainDepositEventInbox) normalize() {
	data.ContractAddress = strings.ToLower(data.ContractAddress)
	data.TxHash = strings.ToLower(data.TxHash)
	data.BlockHash = strings.ToLower(data.BlockHash)
	data.FromAddress = strings.ToLower(data.FromAddress)
	data.ToAddress = strings.ToLower(data.ToAddress)
}

func (data *ChainDepositEventInbox) ID() uint64 {
	return data.Id
}

func (data *ChainDepositEventInbox) TableName() string {
	return "chain_deposit_event_inbox"
}

func (data *ChainDepositEventInbox) GetCreateDDL() string {
	return "CREATE TABLE `chain_deposit_event_inbox` (   `id` bigint unsigned NOT NULL AUTO_INCREMENT,   `chain_db_id` bigint unsigned NOT NULL,   `contract_address` char(42) COLLATE utf8mb4_general_ci NOT NULL,   `tx_hash` char(66) COLLATE utf8mb4_general_ci NOT NULL,   `log_index` int unsigned NOT NULL,   `block_number` bigint unsigned NOT NULL,   `block_hash` char(66) COLLATE utf8mb4_general_ci NOT NULL,   `from_address` char(42) COLLATE utf8mb4_general_ci DEFAULT NULL,   `to_address` char(42) COLLATE utf8mb4_general_ci DEFAULT NULL,   `amount` decimal(36,0) DEFAULT NULL,   `source` tinyint unsigned NOT NULL COMMENT '1:websocket 2:getlogs 3:both',   `status` tinyint NOT NULL DEFAULT '0' COMMENT '0:待确认 1:待处理 2:处理中 3:完成 4:重试 -1:孤块 -2:忽略',   `removed` tinyint(1) NOT NULL DEFAULT '0',   `retry_count` int unsigned NOT NULL DEFAULT '0',   `next_retry_at` datetime DEFAULT NULL,   `last_error` varchar(1024) COLLATE utf8mb4_general_ci DEFAULT NULL,   `created_at` datetime NOT NULL,   `updated_at` datetime NOT NULL,   `confirmed_at` datetime DEFAULT NULL,   `processed_at` datetime DEFAULT NULL,   PRIMARY KEY (`id`),   UNIQUE KEY `uk_chain_tx_log` (`chain_db_id`,`tx_hash`,`log_index`),   KEY `idx_status_retry` (`status`,`next_retry_at`),   KEY `idx_chain_status_block` (`chain_db_id`,`status`,`block_number`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='充值事件采集与待确认队列';"
}
