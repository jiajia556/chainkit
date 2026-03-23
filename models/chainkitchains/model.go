package chainkitchains

type ChainChains struct {
	Id      uint64 `gorm:"column:id;notNull;primaryKey;unsigned;autoIncrement" json:"id"`
	Name    string `gorm:"column:name;notNull" json:"name"`
	Rpc     string `gorm:"column:rpc;notNull" json:"rpc"`
	ChainId int64  `gorm:"column:chain_id;unsigned;notNull" json:"chain_id"`
	ApiHost string `gorm:"column:api_host;notNull;default:" json:"api_host"`
	ApiKey  string `gorm:"column:api_key;notNull;default:" json:"api_key"`
}

func (data *ChainChains) ID() uint64 {
	return data.Id
}

func (data *ChainChains) TableName() string {
	return "chain_chains"
}

func (data *ChainChains) GetCreateDDL() string {
	return "CREATE TABLE `chain_chains` (   `id` int unsigned NOT NULL AUTO_INCREMENT,   `name` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,   `rpc` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,   `chain_id` int unsigned NOT NULL,   `api_host` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',   `api_key` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',   PRIMARY KEY (`id`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC;"
}
