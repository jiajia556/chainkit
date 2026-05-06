package chainkitdeposittokens

type ChainDepositTokens struct {
	Id             uint64 `gorm:"column:id;autoIncrement;notNull;primaryKey" json:"id"`
	ChainDbId      uint64 `gorm:"column:chain_db_id;notNull" json:"chain_db_id"`
	TokenId        uint64 `gorm:"column:token_id;notNull" json:"token_id"`
	Status         int8   `gorm:"column:status;notNull" json:"status"`
	InitStartBlock uint64 `gorm:"column:init_start_block" json:"init_start_block"`
}

func (data *ChainDepositTokens) ID() uint64 {
	return data.Id
}

func (data *ChainDepositTokens) TableName() string {
	return "chain_deposit_tokens"
}

func (data *ChainDepositTokens) GetCreateDDL() string {
	return "CREATE TABLE `chain_deposit_tokens` (\n  `id` bigint NOT NULL AUTO_INCREMENT,\n  `chain_db_id` bigint unsigned NOT NULL,\n  `token_id` bigint unsigned NOT NULL,\n  `status` tinyint(1) NOT NULL,\n  PRIMARY KEY (`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;"
}
