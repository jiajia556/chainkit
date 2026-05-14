package chainkittokengroups

type ChainTokenGroups struct {
	Id     uint64 `gorm:"column:id;primaryKey;unsigned;notNull" json:"id"`
	Name   string `gorm:"column:name;notNull" json:"name"`
	Symbol string `gorm:"column:symbol;notNull" json:"symbol"`
	Remark string `gorm:"column:remark;notNull" json:"remark"`
}

func (data *ChainTokenGroups) ID() uint64 {
	return data.Id
}

func (data *ChainTokenGroups) TableName() string {
	return "chain_token_groups"
}

func (data *ChainTokenGroups) GetCreateDDL() string {
	return "CREATE TABLE `chain_token_groups` (   `id` bigint unsigned NOT NULL,   `name` varchar(64) NOT NULL,   `symbol` varchar(32) NOT NULL,   `remark` varchar(255) NOT NULL,   PRIMARY KEY (`id`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;"
}
