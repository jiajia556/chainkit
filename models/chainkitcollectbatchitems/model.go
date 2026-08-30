package chainkitcollectbatchitems

import "time"

type ChainCollectBatchItems struct {
	Id                    uint64    `gorm:"column:id;unsigned;autoIncrement;notNull;primaryKey" json:"id"`
	BatchId               uint64    `gorm:"column:batch_id;unsigned;notNull" json:"batch_id"`
	CollectTaskId         uint64    `gorm:"column:collect_task_id;unsigned;notNull" json:"collect_task_id"`
	ItemIndex             uint32    `gorm:"column:item_index;unsigned;notNull" json:"item_index"`
	UserDepositAddressId  uint64    `gorm:"column:user_deposit_address_id;unsigned;notNull" json:"user_deposit_address_id"`
	AuthorityAddress      string    `gorm:"column:authority_address;notNull" json:"authority_address"`
	AuthorizationRequired bool      `gorm:"column:authorization_required;notNull;default:false" json:"authorization_required"`
	AuthorizationNonce    uint64    `gorm:"column:authorization_nonce;unsigned;notNull;default:0" json:"authorization_nonce"`
	DelegateAddress       string    `gorm:"column:delegate_address;notNull" json:"delegate_address"`
	TokenAddress          string    `gorm:"column:token_address;notNull" json:"token_address"`
	RecipientAddress      string    `gorm:"column:recipient_address;notNull" json:"recipient_address"`
	PlannedAmount         string    `gorm:"column:planned_amount;notNull" json:"planned_amount"`
	ActualAmount          string    `gorm:"column:actual_amount;default:null" json:"actual_amount"`
	CallGasLimit          uint64    `gorm:"column:call_gas_limit;unsigned;notNull" json:"call_gas_limit"`
	ResultCode            uint8     `gorm:"column:result_code;unsigned;default:null" json:"result_code"`
	ReturnDataHash        string    `gorm:"column:return_data_hash;default:null" json:"return_data_hash"`
	Status                uint8     `gorm:"column:status;unsigned;notNull;default:0" json:"status"`
	LastError             string    `gorm:"column:last_error;default:null" json:"last_error"`
	CreatedAt             time.Time `gorm:"column:created_at;notNull;default:current_timestamp" json:"created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at;notNull;default:current_timestamp" json:"updated_at"`
}

func (data *ChainCollectBatchItems) ID() uint64 {
	return data.Id
}

func (data *ChainCollectBatchItems) TableName() string {
	return "chain_collect_batch_items"
}

func (data *ChainCollectBatchItems) GetCreateDDL() string {
	return `CREATE TABLE chain_collect_batch_items (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  batch_id BIGINT UNSIGNED NOT NULL,
  collect_task_id BIGINT UNSIGNED NOT NULL,
  item_index INT UNSIGNED NOT NULL,
  user_deposit_address_id BIGINT UNSIGNED NOT NULL,
  authority_address CHAR(42) NOT NULL,
  authorization_required TINYINT(1) NOT NULL DEFAULT 0,
  authorization_nonce BIGINT UNSIGNED NOT NULL DEFAULT 0,
  delegate_address CHAR(42) NOT NULL,
  token_address CHAR(42) NOT NULL COMMENT 'zero address represents native currency',
  recipient_address CHAR(42) NOT NULL,
  planned_amount VARCHAR(78) NOT NULL,
  actual_amount VARCHAR(78) DEFAULT NULL,
  call_gas_limit BIGINT UNSIGNED NOT NULL,
  result_code TINYINT UNSIGNED DEFAULT NULL,
  return_data_hash CHAR(66) DEFAULT NULL,
  status TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0:waiting 1:success 2:failed 3:not_attempted',
  last_error VARCHAR(1024) DEFAULT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_batch_collect_task (batch_id, collect_task_id),
  UNIQUE KEY uk_batch_item_index (batch_id, item_index),
  KEY idx_batch_status (batch_id, status),
  KEY idx_authority (authority_address)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='EIP-7702 sponsored collect batch items';`
}
