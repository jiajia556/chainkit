package chainkittokens

import (
	"time"
)

type ChainTokens struct {
	Id              uint64    `gorm:"column:id;primaryKey;unsigned;autoIncrement;notNull" json:"id"`
	ChainDbId       uint64    `gorm:"column:chain_db_id;unsigned;notNull;default:1" json:"chain_db_id"`
	ContractAddress string    `gorm:"column:contract_address;notNull" json:"contract_address"`
	Logo            string    `gorm:"column:logo;notNull" json:"logo"`
	Symbol          string    `gorm:"column:symbol;notNull" json:"symbol"`
	Decimals        int8      `gorm:"column:decimals;notNull" json:"decimals"`
	Remark          string    `gorm:"column:remark;notNull" json:"remark"`
	CreatedAt       time.Time `gorm:"column:created_at;notNull" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;notNull" json:"updated_at"`
}

func (data *ChainTokens) ID() uint64 {
	return data.Id
}

func (data *ChainTokens) TableName() string {
	return "chain_tokens"
}

func (data *ChainTokens) GetCreateDDL() string {
	return "CREATE TABLE `chain_tokens` (   `id` bigint unsigned NOT NULL AUTO_INCREMENT,   `chain_db_id` bigint unsigned NOT NULL DEFAULT '1',   `contract_address` char(42) NOT NULL,   `logo` text NOT NULL,   `symbol` varchar(32) NOT NULL,   `decimals` tinyint NOT NULL,   `remark` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,   `created_at` datetime NOT NULL,   `updated_at` datetime NOT NULL,   PRIMARY KEY (`id`),   KEY `idx_contract_address` (`contract_address`),   KEY `idx_symbol` (`symbol`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC;"
}
