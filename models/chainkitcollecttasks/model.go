package chainkitcollecttasks

import (
	"time"

	"github.com/shopspring/decimal"
)

type ChainCollectTasks struct {
	Id                   uint64          `gorm:"column:id;unsigned;autoIncrement;notNull;primaryKey" json:"id"`
	ChainDbId            uint64          `gorm:"column:chain_db_id;unsigned;notNull" json:"chain_db_id"`
	TokenId              uint64          `gorm:"column:token_id;notNull;unsigned" json:"token_id"`
	UserId               uint64          `gorm:"column:user_id;unsigned;notNull" json:"user_id"`
	UserDepositAddressId uint64          `gorm:"column:user_deposit_address_id;unsigned;notNull" json:"user_deposit_address_id"`
	FromAddress          string          `gorm:"column:from_address;notNull" json:"from_address"`
	ToAddress            string          `gorm:"column:to_address;notNull" json:"to_address"`
	PlanAmount           decimal.Decimal `gorm:"column:plan_amount;default:0;unsigned;notNull" json:"plan_amount"`
	ActualAmount         decimal.Decimal `gorm:"column:actual_amount;unsigned;default:null" json:"actual_amount"`
	GasRequiredAmount    decimal.Decimal `gorm:"column:gas_required_amount;unsigned;default:null" json:"gas_required_amount"`
	GasBalanceBeforeTx   decimal.Decimal `gorm:"column:gas_balance_before_tx;unsigned;default:null" json:"gas_balance_before_tx"`
	TxHash               string          `gorm:"column:tx_hash;default:null" json:"tx_hash"`
	Nonce                uint64          `gorm:"column:nonce;unsigned;default:null" json:"nonce"`
	GasLimit             decimal.Decimal `gorm:"column:gas_limit;unsigned;notNull" json:"gas_limit"`
	GasPrice             decimal.Decimal `gorm:"column:gas_price;unsigned;default:null" json:"gas_price"`
	MaxFeePerGas         decimal.Decimal `gorm:"column:max_fee_per_gas;default:null;unsigned" json:"max_fee_per_gas"`
	MaxPriorityFeePerGas decimal.Decimal `gorm:"column:max_priority_fee_per_gas;unsigned;default:null" json:"max_priority_fee_per_gas"`
	GasUsed              decimal.Decimal `gorm:"column:gas_used;unsigned;default:null" json:"gas_used"`
	TxFee                decimal.Decimal `gorm:"column:tx_fee;unsigned;default:null" json:"tx_fee"`
	Status               uint8           `gorm:"column:status;unsigned;notNull;default:0" json:"status"`
	GasTaskId            uint64          `gorm:"column:gas_task_id;notNull;unsigned;default:0" json:"gas_task_id"`
	CollectMethod        uint8           `gorm:"column:collect_method;unsigned;notNull;default:0" json:"collect_method"`
	BatchId              uint64          `gorm:"column:batch_id;unsigned;notNull;default:0" json:"batch_id"`
	LastError            string          `gorm:"column:last_error;default:null" json:"last_error"`
	SentAt               time.Time       `gorm:"column:sent_at;default:null" json:"sent_at"`
	ConfirmedAt          time.Time       `gorm:"column:confirmed_at;default:null" json:"confirmed_at"`
	Remark               string          `gorm:"column:remark;default:null" json:"remark"`
	CreatedAt            time.Time       `gorm:"column:created_at;notNull;default:current_timestamp" json:"created_at"`
	UpdatedAt            time.Time       `gorm:"column:updated_at;default:current_timestamp;notNull" json:"updated_at"`
}

func (data *ChainCollectTasks) ID() uint64 {
	return data.Id
}

func (data *ChainCollectTasks) TableName() string {
	return "chain_collect_tasks"
}

func (data *ChainCollectTasks) GetCreateDDL() string {
	return `CREATE TABLE chain_collect_tasks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  chain_db_id BIGINT UNSIGNED NOT NULL,
  token_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  user_deposit_address_id BIGINT UNSIGNED NOT NULL,
  from_address CHAR(42) NOT NULL,
  to_address CHAR(42) NOT NULL,
  plan_amount DECIMAL(36,0) UNSIGNED NOT NULL DEFAULT 0,
  actual_amount DECIMAL(36,0) UNSIGNED DEFAULT NULL,
  gas_required_amount DECIMAL(36,0) UNSIGNED DEFAULT NULL,
  gas_balance_before_tx DECIMAL(36,0) UNSIGNED DEFAULT NULL,
  tx_hash CHAR(66) DEFAULT NULL,
  nonce BIGINT UNSIGNED DEFAULT NULL COMMENT 'traditional collect transaction nonce',
  gas_limit DECIMAL(36,0) UNSIGNED NOT NULL,
  gas_price DECIMAL(36,0) UNSIGNED DEFAULT NULL,
  max_fee_per_gas DECIMAL(36,0) UNSIGNED DEFAULT NULL,
  max_priority_fee_per_gas DECIMAL(36,0) UNSIGNED DEFAULT NULL,
  gas_used DECIMAL(36,0) UNSIGNED DEFAULT NULL,
  tx_fee DECIMAL(36,0) UNSIGNED DEFAULT NULL,
  status TINYINT UNSIGNED NOT NULL DEFAULT 0,
  gas_task_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  collect_method TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0:undecided 1:traditional 2:eip7702',
  batch_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  last_error VARCHAR(1024) DEFAULT NULL,
  sent_at DATETIME DEFAULT NULL,
  confirmed_at DATETIME DEFAULT NULL,
  remark VARCHAR(255) DEFAULT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  PRIMARY KEY (id),
  KEY idx_address_token_status (user_deposit_address_id, token_id, status),
  KEY idx_chain_status (chain_db_id, status),
  KEY idx_address_chain_status (user_deposit_address_id, chain_db_id, status),
  KEY idx_batch_id (batch_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='chain asset collect tasks';`
}
