package chainkiteventlogs

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type ChainEventLogs struct {
	Id              uint64    `gorm:"column:id;primaryKey;unsigned;autoIncrement;notNull" json:"id"`
	ChainDbId       uint64    `gorm:"column:chain_db_id;unsigned;notNull" json:"chain_db_id"`
	ContractAddress string    `gorm:"column:contract_address;notNull" json:"contract_address"`
	Module          string    `gorm:"column:module;notNull" json:"module"`
	TxHash          string    `gorm:"column:tx_hash;notNull" json:"tx_hash"`
	LogIndex        uint32    `gorm:"column:log_index;unsigned;notNull" json:"log_index"`
	BlockNumber     uint64    `gorm:"column:block_number;unsigned;notNull" json:"block_number"`
	BlockHash       string    `gorm:"column:block_hash;notNull" json:"block_hash"`
	EventSig        []byte    `gorm:"column:event_sig;notNull" json:"event_sig"`
	CreatedAt       time.Time `gorm:"column:created_at;notNull" json:"created_at"`
	RawData         []byte    `gorm:"column:raw_data" json:"raw_data"`
}

func (data *ChainEventLogs) ID() uint64 {
	return data.Id
}

func (data *ChainEventLogs) TableName() string {
	return "chain_event_logs"
}

func (data *ChainEventLogs) GetCreateDDL() string {
	return "CREATE TABLE `chain_event_logs` (\n  `id` bigint unsigned NOT NULL AUTO_INCREMENT,\n  `chain_db_id` bigint unsigned NOT NULL,\n  `contract_address` char(42) COLLATE utf8mb4_general_ci NOT NULL,\n  `module` varchar(128) COLLATE utf8mb4_general_ci NOT NULL,\n  `tx_hash` char(66) COLLATE utf8mb4_general_ci NOT NULL,\n  `log_index` int unsigned NOT NULL,\n  `block_number` bigint unsigned NOT NULL,\n  `block_hash` char(66) COLLATE utf8mb4_general_ci NOT NULL,\n  `event_sig` binary(32) NOT NULL,\n  `created_at` datetime NOT NULL,\n  `raw_data` blob,\n  PRIMARY KEY (`id`),\n  UNIQUE KEY `idx_uniq_chi` (`chain_db_id`,`tx_hash`,`log_index`),\n  KEY `idx_tx_hash` (`tx_hash`),\n  KEY `idx_contract_address` (`contract_address`,`module`,`chain_db_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;"
}

func (data *ChainEventLogs) BeforeCreate(tx *gorm.DB) (err error) {
	data.ContractAddress = strings.ToLower(data.ContractAddress)
	return nil
}
