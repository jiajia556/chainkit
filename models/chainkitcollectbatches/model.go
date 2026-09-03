package chainkitcollectbatches

import (
	"time"

	"gorm.io/gorm"
)

type ChainCollectBatches struct {
	Id                       uint64    `gorm:"column:id;unsigned;autoIncrement;notNull;primaryKey" json:"id"`
	ChainDbId                uint64    `gorm:"column:chain_db_id;unsigned;notNull" json:"chain_db_id"`
	SponsorMnemonicAddressId uint64    `gorm:"column:sponsor_mnemonic_address_id;unsigned;notNull" json:"sponsor_mnemonic_address_id"`
	SponsorAddress           string    `gorm:"column:sponsor_address;notNull" json:"sponsor_address"`
	SponsorNonce             uint64    `gorm:"column:sponsor_nonce;unsigned;notNull" json:"sponsor_nonce"`
	ExecutorAddress          string    `gorm:"column:executor_address;notNull" json:"executor_address"`
	TxType                   uint8     `gorm:"column:tx_type;unsigned;notNull" json:"tx_type"`
	AuthorizationCount       uint32    `gorm:"column:authorization_count;unsigned;notNull;default:0" json:"authorization_count"`
	TxHash                   string    `gorm:"column:tx_hash;default:null" json:"tx_hash"`
	RawTx                    []byte    `gorm:"column:raw_tx;type:mediumblob;default:null" json:"-"`
	GasLimit                 uint64    `gorm:"column:gas_limit;unsigned;notNull" json:"gas_limit"`
	MaxFeePerGas             string    `gorm:"column:max_fee_per_gas;notNull" json:"max_fee_per_gas"`
	MaxPriorityFeePerGas     string    `gorm:"column:max_priority_fee_per_gas;notNull" json:"max_priority_fee_per_gas"`
	GasUsed                  uint64    `gorm:"column:gas_used;unsigned;default:null" json:"gas_used"`
	TxFee                    string    `gorm:"column:tx_fee;default:null" json:"tx_fee"`
	Status                   uint8     `gorm:"column:status;unsigned;notNull;default:0" json:"status"`
	LastError                string    `gorm:"column:last_error;default:null" json:"last_error"`
	SentAt                   time.Time `gorm:"column:sent_at;default:null" json:"sent_at"`
	ConfirmedAt              time.Time `gorm:"column:confirmed_at;default:null" json:"confirmed_at"`
	CreatedAt                time.Time `gorm:"column:created_at;notNull;default:current_timestamp" json:"created_at"`
	UpdatedAt                time.Time `gorm:"column:updated_at;notNull;default:current_timestamp" json:"updated_at"`
}

func (data *ChainCollectBatches) ID() uint64 {
	return data.Id
}

func (data *ChainCollectBatches) TableName() string {
	return "chain_collect_batches"
}

func (data *ChainCollectBatches) BeforeCreate(tx *gorm.DB) (err error) {
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	return nil

}

func (data *ChainCollectBatches) GetCreateDDL() string {
	return `CREATE TABLE chain_collect_batches (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  chain_db_id BIGINT UNSIGNED NOT NULL,
  sponsor_mnemonic_address_id BIGINT UNSIGNED NOT NULL,
  sponsor_address CHAR(42) NOT NULL,
  sponsor_nonce BIGINT UNSIGNED NOT NULL,
  executor_address CHAR(42) NOT NULL,
  tx_type TINYINT UNSIGNED NOT NULL COMMENT '2:EIP-1559 4:EIP-7702',
  authorization_count INT UNSIGNED NOT NULL DEFAULT 0,
  tx_hash CHAR(66) DEFAULT NULL,
  raw_tx MEDIUMBLOB DEFAULT NULL,
  gas_limit BIGINT UNSIGNED NOT NULL,
  max_fee_per_gas VARCHAR(78) NOT NULL,
  max_priority_fee_per_gas VARCHAR(78) NOT NULL,
  gas_used BIGINT UNSIGNED DEFAULT NULL,
  tx_fee VARCHAR(78) DEFAULT NULL,
  status TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0:waiting 1:building 2:sent 3:confirmed 4:failed 5:unknown 10:maybe_sent',
  last_error VARCHAR(1024) DEFAULT NULL,
  sent_at DATETIME DEFAULT NULL,
  confirmed_at DATETIME DEFAULT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_chain_sponsor_nonce (chain_db_id, sponsor_address, sponsor_nonce),
  KEY idx_chain_status (chain_db_id, status),
  KEY idx_sponsor_status (chain_db_id, sponsor_address, status),
  KEY idx_tx_hash (tx_hash)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='EIP-7702 sponsored collect batches';`
}
