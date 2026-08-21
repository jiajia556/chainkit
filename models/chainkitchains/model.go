package chainkitchains

type ChainChains struct {
	Id                uint64 `gorm:"column:id;autoIncrement;notNull;primaryKey;unsigned" json:"id"`
	Name              string `gorm:"column:name;notNull" json:"name"`
	Rpc               string `gorm:"column:rpc;notNull" json:"rpc"`
	WsRpc             string `gorm:"column:ws_rpc;default:null" json:"ws_rpc"`
	ChainId           uint64 `gorm:"column:chain_id;notNull;unsigned" json:"chain_id"`
	SafeConfirmations uint64 `gorm:"column:safe_confirmations;notNull;unsigned" json:"safe_confirmations"`
}

func (data *ChainChains) ID() uint64 {
	return data.Id
}

func (data *ChainChains) TableName() string {
	return "chain_chains"
}

func (data *ChainChains) GetCreateDDL() string {
	return "CREATE TABLE `chain_chains` (   `id` int unsigned NOT NULL AUTO_INCREMENT,   `name` varchar(32) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NOT NULL,   `rpc` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NOT NULL,   `ws_rpc` varchar(512) DEFAULT NULL COMMENT 'WebSocket RPC地址',   `chain_id` bigint unsigned NOT NULL,   `safe_confirmations` bigint unsigned NOT NULL,   PRIMARY KEY (`id`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC;"
}
