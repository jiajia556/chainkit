package chainkitcollectconfig

import (
	"github.com/shopspring/decimal"
)

type ChainCollectConfig struct {
	Id                           uint64          `gorm:"column:id;unsigned;notNull;primaryKey" json:"id"`
	ChainDbId                    uint64          `gorm:"column:chain_db_id;unsigned;notNull" json:"chain_db_id"`
	GasProviderMnemonicAddressId uint64          `gorm:"column:gas_provider_mnemonic_address_id;unsigned;notNull" json:"gas_provider_mnemonic_address_id"`
	DefaultCollectToAddress      string          `gorm:"column:default_collect_to_address;notNull" json:"default_collect_to_address"`
	DefaultErc20TransferGasLimit decimal.Decimal `gorm:"column:default_erc20_transfer_gas_limit;notNull" json:"default_erc20_transfer_gas_limit"`
	EIP7702Enabled               bool            `gorm:"column:eip7702_enabled;notNull;default:false" json:"eip7702_enabled"`
	EIP7702DelegateAddress       string          `gorm:"column:eip7702_delegate_address;notNull;default:''" json:"eip7702_delegate_address"`
	EIP7702ExecutorAddress       string          `gorm:"column:eip7702_executor_address;notNull;default:''" json:"eip7702_executor_address"`
	EIP7702MaxBatchItems         uint16          `gorm:"column:eip7702_max_batch_items;unsigned;notNull;default:50" json:"eip7702_max_batch_items"`
	EIP7702CallGasLimit          uint64          `gorm:"column:eip7702_call_gas_limit;unsigned;notNull;default:100000" json:"eip7702_call_gas_limit"`
}

func (data *ChainCollectConfig) ID() uint64 {
	return data.Id
}

func (data *ChainCollectConfig) TableName() string {
	return "chain_collect_config"
}

func (data *ChainCollectConfig) GetCreateDDL() string {
	return `CREATE TABLE chain_collect_config (
  id BIGINT UNSIGNED NOT NULL,
  chain_db_id BIGINT UNSIGNED NOT NULL,
  gas_provider_mnemonic_address_id BIGINT UNSIGNED NOT NULL,
  default_collect_to_address CHAR(42) NOT NULL,
  default_erc20_transfer_gas_limit DECIMAL(16,0) NOT NULL,
  eip7702_enabled TINYINT(1) NOT NULL DEFAULT 0,
  eip7702_delegate_address CHAR(42) NOT NULL DEFAULT '',
  eip7702_executor_address CHAR(42) NOT NULL DEFAULT '',
  eip7702_max_batch_items SMALLINT UNSIGNED NOT NULL DEFAULT 50,
  eip7702_call_gas_limit BIGINT UNSIGNED NOT NULL DEFAULT 100000,
  PRIMARY KEY (id),
  KEY idx_chain_db_id (chain_db_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;`
}
