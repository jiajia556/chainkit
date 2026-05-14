package chainkitdeposittokens

type ChainDepositTokens struct {
	Id             uint64 `gorm:"column:id;primaryKey;autoIncrement;notNull" json:"id"`
	ChainDbId      uint64 `gorm:"column:chain_db_id;unsigned;notNull" json:"chain_db_id"`
	TokenId        uint64 `gorm:"column:token_id;unsigned;notNull" json:"token_id"`
	Status         int8   `gorm:"column:status;notNull" json:"status"`
	InitStartBlock uint64 `gorm:"column:init_start_block;unsigned;notNull" json:"init_start_block"`
}

func (data *ChainDepositTokens) ID() uint64 {
	return data.Id
}

func (data *ChainDepositTokens) TableName() string {
	return "chain_deposit_tokens"
}

func (data *ChainDepositTokens) GetCreateDDL() string {
	return "CREATE TABLE `chain_deposit_tokens` (   `id` bigint NOT NULL AUTO_INCREMENT,   `chain_db_id` bigint unsigned NOT NULL,   `token_id` bigint unsigned NOT NULL,   `status` tinyint(1) NOT NULL,   `init_start_block` bigint unsigned NOT NULL,   PRIMARY KEY (`id`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;"
}
