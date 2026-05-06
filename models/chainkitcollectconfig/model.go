package chainkitcollectconfig

import (
	"github.com/shopspring/decimal"
)

type ChainCollectConfig struct {
	Id                            uint64          `gorm:"column:id;unsigned;notNull;primaryKey" json:"id"`
	GasProviderMnemonicAddresseId uint64          `gorm:"column:gas_provider_mnemonic_addresse_id;unsigned;notNull" json:"gas_provider_mnemonic_addresse_id"`
	DefaultCollectToAddress       string          `gorm:"column:default_collect_to_address;notNull" json:"default_collect_to_address"`
	DefaultErc20TransferGaslimit  decimal.Decimal `gorm:"column:default_erc20_transfer_gaslimit;notNull" json:"default_erc20_transfer_gaslimit"`
}

func (data *ChainCollectConfig) ID() uint64 {
	return data.Id
}

func (data *ChainCollectConfig) TableName() string {
	return "chain_collect_config"
}

func (data *ChainCollectConfig) GetCreateDDL() string {
	return "CREATE TABLE `chain_collect_config` (   `id` bigint unsigned NOT NULL,   `gas_provider_mnemonic_addresse_id` bigint unsigned NOT NULL,   `default_collect_to_address` char(42) NOT NULL,   `default_erc20_transfer_gaslimit` decimal(16,0) NOT NULL,   PRIMARY KEY (`id`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;"
}
