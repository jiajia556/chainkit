package chainkituserdepositaddress

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type ChainUserDepositAddress struct {
	Id                  uint64    `gorm:"column:id;autoIncrement;notNull;primaryKey;unsigned" json:"id"`
	UserId              uint64    `gorm:"column:user_id;unsigned;notNull" json:"user_id"`
	ChainDbId           uint64    `gorm:"column:chain_db_id;unsigned;notNull" json:"chain_db_id"`
	TokenId             uint64    `gorm:"column:token_id;unsigned;notNull" json:"token_id"`
	Address             string    `gorm:"column:address;notNull" json:"address"`
	PrivateKeyEncrypted []byte    `gorm:"column:private_key_encrypted;notNull" json:"private_key_encrypted"`
	Remark              string    `gorm:"column:remark;default:null" json:"remark"`
	CreatedAt           time.Time `gorm:"column:created_at;notNull;default:current_timestamp" json:"created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at;notNull;default:current_timestamp" json:"updated_at"`
}

func (data *ChainUserDepositAddress) ID() uint64 {
	return data.Id
}

func (data *ChainUserDepositAddress) TableName() string {
	return "chain_user_deposit_address"
}

func (data *ChainUserDepositAddress) GetCreateDDL() string {
	return "CREATE TABLE `chain_user_deposit_address` (\n  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',\n  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',\n  `token_id` bigint unsigned NOT NULL COMMENT '代币ID，0可表示链原生币',\n  `address` char(42) NOT NULL COMMENT '充值地址',\n  `private_key_encrypted` varbinary(1024) NOT NULL COMMENT '加密后的私钥二进制',\n  `remark` varchar(255) DEFAULT NULL COMMENT '备注',\n  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n  PRIMARY KEY (`id`),\n  KEY `idx_user_id` (`user_id`),\n  KEY `idx_address` (`address`),\n  KEY `idx_user_chain_token` (`user_id`,`token_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户充值地址表';"
}

func (data *ChainUserDepositAddress) BeforeCreate(tx *gorm.DB) error {
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	data.Address = strings.ToLower(data.Address)
	return nil
}
