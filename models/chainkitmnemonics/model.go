package chainkitmnemonics

type ChainMnemonics struct {
	Id     uint64 `gorm:"column:id;autoIncrement;notNull;primaryKey;unsigned" json:"id"`
	Words  []byte `gorm:"column:words;notNull" json:"words"`
	Remark string `gorm:"column:remark;default:" json:"remark"`
}

func (data *ChainMnemonics) ID() uint64 {
	return data.Id
}

func (data *ChainMnemonics) TableName() string {
	return "chain_mnemonics"
}

func (data *ChainMnemonics) GetCreateDDL() string {
	return "CREATE TABLE `chain_mnemonics` (   `id` int unsigned NOT NULL AUTO_INCREMENT,   `words` varbinary(255) NOT NULL,   `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '',   PRIMARY KEY (`id`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;"
}
