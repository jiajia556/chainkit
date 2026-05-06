package chainkittransferdetails

import (
	"time"

	"github.com/shopspring/decimal"
)

type ChainTransferDetails struct {
	Id               uint64          `gorm:"column:id;autoIncrement;notNull;primaryKey;unsigned" json:"id"`
	ChainDbId        uint64          `gorm:"column:chain_db_id;unsigned;notNull" json:"chain_db_id"`
	FromAddressType  string          `gorm:"column:from_address_type;notNull" json:"from_address_type"`
	FromAddressId    uint64          `gorm:"column:from_address_id;unsigned;notNull" json:"from_address_id"`
	TokenId          uint64          `gorm:"column:token_id;unsigned;notNull" json:"token_id"`
	TokenAddress     string          `gorm:"column:token_address;notNull" json:"token_address"`
	To               string          `gorm:"column:to;notNull" json:"to"`
	Amount           decimal.Decimal `gorm:"column:amount;notNull" json:"amount"`
	TransferRecordId uint64          `gorm:"column:transfer_record_id;unsigned;notNull" json:"transfer_record_id"`
	Status           int8            `gorm:"column:status;notNull" json:"status"`
	Remark           string          `gorm:"column:remark;default:null" json:"remark"`
	CreatedAt        time.Time       `gorm:"column:created_at;notNull" json:"created_at"`
	UpdatedAt        time.Time       `gorm:"column:updated_at;notNull" json:"updated_at"`
}

func (data *ChainTransferDetails) ID() uint64 {
	return data.Id
}

func (data *ChainTransferDetails) TableName() string {
	return "chain_transfer_details"
}

func (data *ChainTransferDetails) GetCreateDDL() string {
	return "CREATE TABLE `chain_transfer_details` (   `id` bigint unsigned NOT NULL AUTO_INCREMENT,   `chain_db_id` bigint unsigned NOT NULL,   `from_address_type` varchar(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT 'mnemonic, user_deposit',   `from_address_id` bigint unsigned NOT NULL,   `token_id` bigint unsigned NOT NULL,   `token_address` char(42) COLLATE utf8mb4_general_ci NOT NULL,   `to` char(42) COLLATE utf8mb4_general_ci NOT NULL,   `amount` decimal(32,0) NOT NULL,   `transfer_record_id` bigint unsigned NOT NULL,   `status` tinyint(1) NOT NULL COMMENT '0-waiting 1-pending 2-success -1-failed',   `remark` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL,   `created_at` datetime NOT NULL,   `updated_at` datetime NOT NULL,   PRIMARY KEY (`id`),   KEY `idx_from_address_id_status` (`from_address_id`,`status`),   KEY `idx_transfer_record_id` (`transfer_record_id`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;"
}
