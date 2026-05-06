package chainkitscancursor

import (
	"time"
)

type ChainScanCursor struct {
	Id              uint64    `gorm:"column:id;notNull;primaryKey;unsigned;autoIncrement" json:"id"`
	ChainDbId       uint64    `gorm:"column:chain_db_id;unsigned;notNull" json:"chain_db_id"`
	ContractAddress string    `gorm:"column:contract_address;notNull" json:"contract_address"`
	Module          string    `gorm:"column:module;notNull" json:"module"`
	LastestBlock    uint64    `gorm:"column:lastest_block;notNull;unsigned" json:"lastest_block"`
	CreatedAt       time.Time `gorm:"column:created_at;notNull" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;notNull" json:"updated_at"`
}

func (data *ChainScanCursor) ID() uint64 {
	return data.Id
}

func (data *ChainScanCursor) TableName() string {
	return "chain_scan_cursor"
}

func (data *ChainScanCursor) GetCreateDDL() string {
	return "CREATE TABLE `chain_scan_cursor` (\n  `id` bigint unsigned NOT NULL AUTO_INCREMENT,\n  `chain_db_id` bigint unsigned NOT NULL,\n  `contract_address` char(42) COLLATE utf8mb4_general_ci NOT NULL,\n  `module` varchar(128) COLLATE utf8mb4_general_ci NOT NULL,\n  `lastest_block` bigint unsigned NOT NULL,\n  `created_at` datetime NOT NULL,\n  `updated_at` datetime NOT NULL,\n  PRIMARY KEY (`id`),\n  UNIQUE KEY `idx_uniq_contract_chain` (`contract_address`,`module`,`chain_db_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;"
}
