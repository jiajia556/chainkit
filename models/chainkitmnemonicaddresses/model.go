package chainkitmnemonicaddresses

type ChainMnemonicAddresses struct {
	Id         uint64 `gorm:"column:id;primaryKey;unsigned;autoIncrement;notNull" json:"id"`
	Address    string `gorm:"column:address;notNull;default:" json:"address"`
	MnemonicId uint64 `gorm:"column:mnemonic_id;unsigned;notNull" json:"mnemonic_id"`
	Index      uint32 `gorm:"column:index;unsigned;notNull" json:"index"`
	Remark     string `gorm:"column:remark;notNull;default:" json:"remark"`
}

func (data *ChainMnemonicAddresses) ID() uint64 {
	return data.Id
}

func (data *ChainMnemonicAddresses) TableName() string {
	return "chain_mnemonic_addresses"
}

func (data *ChainMnemonicAddresses) GetCreateDDL() string {
	return "CREATE TABLE `chain_mnemonic_addresses` (   `id` bigint unsigned NOT NULL AUTO_INCREMENT,   `address` char(42) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',   `mnemonic_id` bigint unsigned NOT NULL,   `index` int unsigned NOT NULL,   `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',   PRIMARY KEY (`id`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;"
}
