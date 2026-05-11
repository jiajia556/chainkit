package chainkitchains

type ChainChains struct {
	Id                uint64 `gorm:"column:id;primaryKey;unsigned;autoIncrement;notNull" json:"id"`
	Name              string `gorm:"column:name;notNull" json:"name"`
	Rpc               string `gorm:"column:rpc;notNull" json:"rpc"`
	ChainId           uint64 `gorm:"column:chain_id;unsigned;notNull" json:"chain_id"`
	SafeConfirmations uint64 `gorm:"column:safe_confirmations;unsigned;notNull" json:"safe_confirmations"`
}

func (data *ChainChains) ID() uint64 {
	return data.Id
}

func (data *ChainChains) TableName() string {
	return "chain_chains"
}

func (data *ChainChains) GetCreateDDL() string {
	return "CREATE TABLE `chain_chains` (   `id` int unsigned NOT NULL AUTO_INCREMENT,   `name` varchar(32) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NOT NULL,   `rpc` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NOT NULL,   `chain_id` bigint unsigned NOT NULL,   `safe_confirmations` bigint unsigned NOT NULL,   PRIMARY KEY (`id`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC;"
}
