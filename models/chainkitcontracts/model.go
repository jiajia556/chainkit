package chainkitcontracts

type ChainContracts struct {
	Id        uint64 `gorm:"column:id;unsigned;autoIncrement;notNull;primaryKey" json:"id"`
	ChainDbId uint64 `gorm:"column:chain_db_id;unsigned;notNull" json:"chain_db_id"`
	Name      string `gorm:"column:name;notNull" json:"name"`
	Address   string `gorm:"column:address;notNull" json:"address"`
	Remark    string `gorm:"column:remark;notNull" json:"remark"`
}

func (data *ChainContracts) ID() uint64 {
	return data.Id
}

func (data *ChainContracts) TableName() string {
	return "chain_contracts"
}

func (data *ChainContracts) GetCreateDDL() string {
	return "CREATE TABLE `chain_contracts` (   `id` bigint unsigned NOT NULL AUTO_INCREMENT,   `chain_db_id` bigint unsigned NOT NULL,   `name` varchar(255) NOT NULL,   `address` char(42) NOT NULL,   `remark` varchar(255) NOT NULL,   PRIMARY KEY (`id`),   KEY `idx_address` (`address`),   KEY `idx_name` (`name`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;"
}
