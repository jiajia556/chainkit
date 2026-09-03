package chainkiteventbackfilltask

import (
	"time"

	"gorm.io/gorm"
)

type ChainEventBackfillTask struct {
	Id              uint64    `gorm:"column:id;autoIncrement;notNull;primaryKey;unsigned" json:"id"`
	ChainDbId       uint64    `gorm:"column:chain_db_id;notNull;unsigned" json:"chain_db_id"`
	ContractAddress string    `gorm:"column:contract_address;notNull" json:"contract_address"`
	Module          string    `gorm:"column:module;notNull" json:"module"`
	CurrentBlock    uint64    `gorm:"column:current_block;notNull;unsigned" json:"current_block"`
	StartBlock      uint64    `gorm:"column:start_block;notNull;unsigned" json:"start_block"`
	EndBlock        uint64    `gorm:"column:end_block;notNull;unsigned" json:"end_block"`
	Status          int8      `gorm:"column:status;default:0;notNull" json:"status"`
	Remark          string    `gorm:"column:remark;default:null" json:"remark"`
	CreatedAt       time.Time `gorm:"column:created_at;notNull" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;notNull" json:"updated_at"`
}

func (data *ChainEventBackfillTask) ID() uint64 {
	return data.Id
}

func (data *ChainEventBackfillTask) TableName() string {
	return "chain_event_backfill_task"
}

func (data *ChainEventBackfillTask) BeforeCreate(tx *gorm.DB) (err error) {
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	return nil
}

func (data *ChainEventBackfillTask) GetCreateDDL() string {
	return "CREATE TABLE `chain_event_backfill_task` (   `id` bigint unsigned NOT NULL AUTO_INCREMENT,   `chain_db_id` bigint unsigned NOT NULL,   `contract_address` char(42) COLLATE utf8mb4_general_ci NOT NULL,   `module` varchar(128) COLLATE utf8mb4_general_ci NOT NULL,   `current_block` bigint unsigned NOT NULL,   `start_block` bigint unsigned NOT NULL,   `end_block` bigint unsigned NOT NULL,   `status` tinyint NOT NULL DEFAULT '0' COMMENT '0:待扫描,1:扫描中,2:扫描完成,-1:扫描失败',   `remark` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL,   `created_at` datetime NOT NULL,   `updated_at` datetime NOT NULL,   PRIMARY KEY (`id`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;"
}
