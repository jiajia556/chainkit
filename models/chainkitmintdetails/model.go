package chainkitmintdetails

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ChainMintDetails struct {
	Id              uint64          `gorm:"column:id;autoIncrement;notNull;primaryKey;unsigned" json:"id"`
	ChainDbId       uint64          `gorm:"column:chain_db_id;notNull;unsigned" json:"chain_db_id"`
	FromAddressType string          `gorm:"column:from_address_type;notNull" json:"from_address_type"`
	FromAddressId   uint64          `gorm:"column:from_address_id;notNull;unsigned" json:"from_address_id"`
	TokenId         uint64          `gorm:"column:token_id;notNull;unsigned" json:"token_id"`
	TokenAddress    string          `gorm:"column:token_address;notNull" json:"token_address"`
	To              string          `gorm:"column:to;notNull" json:"to"`
	Amount          decimal.Decimal `gorm:"column:amount;notNull" json:"amount"`
	MintRecordId    uint64          `gorm:"column:mint_record_id;notNull;unsigned" json:"mint_record_id"`
	Status          int8            `gorm:"column:status;notNull" json:"status"`
	Remark          string          `gorm:"column:remark;default:null" json:"remark"`
	CreatedAt       time.Time       `gorm:"column:created_at;notNull" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"column:updated_at;notNull" json:"updated_at"`
}

func (data *ChainMintDetails) ID() uint64 {
	return data.Id
}

func (data *ChainMintDetails) TableName() string {
	return "chain_mint_details"
}

func (data *ChainMintDetails) BeforeCreate(tx *gorm.DB) (err error) {
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	return nil
}

func (data *ChainMintDetails) GetCreateDDL() string {
	return "CREATE TABLE `chain_mint_details` (   `id` bigint unsigned NOT NULL AUTO_INCREMENT,   `chain_db_id` bigint unsigned NOT NULL,   `from_address_type` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'mnemonic, user_deposit',   `from_address_id` bigint unsigned NOT NULL,   `token_id` bigint unsigned NOT NULL,   `token_address` char(42) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,   `to` char(42) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,   `amount` decimal(32,0) NOT NULL,   `mint_record_id` bigint unsigned NOT NULL,   `status` tinyint(1) NOT NULL COMMENT '0-waiting 1-pending 2-success -1-failed',   `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,   `created_at` datetime NOT NULL,   `updated_at` datetime NOT NULL,   PRIMARY KEY (`id`),   KEY `idx_from_address_id_status` (`from_address_id`,`status`),   KEY `idx_mint_record_id` (`mint_record_id`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;"
}
