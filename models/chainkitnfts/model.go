package chainkitnfts

import (
	"time"
)

type ChainNfts struct {
	Id              uint64    `gorm:"column:id;autoIncrement;notNull;primaryKey;unsigned" json:"id"`
	ChainDbId       uint64    `gorm:"column:chain_db_id;default:1;notNull;unsigned" json:"chain_db_id"`
	ContractAddress string    `gorm:"column:contract_address;notNull" json:"contract_address"`
	Name            string    `gorm:"column:name;default:;notNull" json:"name"`
	Symbol          string    `gorm:"column:symbol;default:;notNull" json:"symbol"`
	Standard        string    `gorm:"column:standard;notNull" json:"standard"`
	Logo            string    `gorm:"column:logo;notNull" json:"logo"`
	Remark          string    `gorm:"column:remark;notNull" json:"remark"`
	CreatedAt       time.Time `gorm:"column:created_at;notNull" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;notNull" json:"updated_at"`
}

func (data *ChainNfts) ID() uint64 {
	return data.Id
}

func (data *ChainNfts) TableName() string {
	return "chain_nfts"
}

func (data *ChainNfts) GetCreateDDL() string {
	return "CREATE TABLE `chain_nfts` (   `id` bigint unsigned NOT NULL AUTO_INCREMENT,   `chain_db_id` bigint unsigned NOT NULL DEFAULT '1',   `contract_address` char(42) COLLATE utf8mb4_general_ci NOT NULL,   `name` varchar(128) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',   `symbol` varchar(32) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',   `standard` varchar(16) COLLATE utf8mb4_general_ci NOT NULL COMMENT 'NFT标准，如 ERC721、ERC1155',   `logo` text COLLATE utf8mb4_general_ci NOT NULL,   `remark` text COLLATE utf8mb4_general_ci NOT NULL,   `created_at` datetime NOT NULL,   `updated_at` datetime NOT NULL,   PRIMARY KEY (`id`),   UNIQUE KEY `uk_chain_contract` (`chain_db_id`,`contract_address`),   KEY `idx_symbol` (`symbol`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci ROW_FORMAT=DYNAMIC COMMENT='链上NFT合约配置表';"
}
