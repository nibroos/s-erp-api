package models

import (
	"gorm.io/gorm"
)

type Stock struct {
	gorm.Model
	ID          uint     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	WarehouseID uint     `json:"warehouse_id" gorm:"column:warehouse_id"`
	ItemID      uint     `json:"item_id" gorm:"column:item_id"`
	BranchID    uint     `json:"branch_id" gorm:"column:branch_id"`
	Qty         *float64 `json:"qty" gorm:"column:qty"`

	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (Stock) TableName() string {
	return "stocks"
}
