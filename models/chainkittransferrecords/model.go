package chainkittransferrecords

import (
	"time"
)

type ChainTransferRecords struct {
	Id              uint64    `gorm:"column:id;primaryKey;unsigned;autoIncrement;notNull" json:"id"`
	ChainDbId       uint64    `gorm:"column:chain_db_id;unsigned;notNull" json:"chain_db_id"`
	FromAddressType string    `gorm:"column:from_address_type;notNull" json:"from_address_type"`
	FromAddressId   uint64    `gorm:"column:from_address_id;unsigned;notNull" json:"from_address_id"`
	Hash            string    `gorm:"column:hash;notNull" json:"hash"`
	Nonce           uint64    `gorm:"column:nonce;unsigned;notNull" json:"nonce"`
	Status          int8      `gorm:"column:status;default:null" json:"status"`
	CreatedAt       time.Time `gorm:"column:created_at;notNull" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;notNull" json:"updated_at"`
}

func (data *ChainTransferRecords) ID() uint64 {
	return data.Id
}

func (data *ChainTransferRecords) TableName() string {
	return "chain_transfer_records"
}

func (data *ChainTransferRecords) GetCreateDDL() string {
	return "CREATE TABLE `chain_transfer_records` (   `id` bigint unsigned NOT NULL AUTO_INCREMENT,   `chain_db_id` bigint unsigned NOT NULL,   `from_address_type` varchar(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT 'mnemonic, user_deposit',   `from_address_id` bigint unsigned NOT NULL,   `hash` char(66) COLLATE utf8mb4_general_ci NOT NULL,   `nonce` bigint unsigned NOT NULL,   `status` tinyint(1) DEFAULT NULL COMMENT '1-pending 2-success -1-failed',   `created_at` datetime NOT NULL,   `updated_at` datetime NOT NULL,   PRIMARY KEY (`id`),   KEY `idx_address_id_status` (`from_address_id`,`status`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;"
}
