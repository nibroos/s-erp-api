package models

import (
	"gorm.io/gorm"
)

type BranchItem struct {
	gorm.Model
	ID            uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	BranchID      uint           `json:"branch_id" gorm:"column:branch_id"`
	MsItemID      uint           `json:"ms_item_id" gorm:"column:ms_item_id"`
	Name          string         `json:"name" gorm:"column:name"`
	Specification *string        `json:"specification" gorm:"column:specification"`
	Description   *string        `json:"description" gorm:"column:description"`
	TpbCode       *string        `json:"tpb_code" gorm:"column:tpb_code"`
	MinimumStock  *float64       `json:"minimum_stock" gorm:"column:minimum_stock"`
	PriceSell     *float64       `json:"price_sell" gorm:"column:price_sell"`
	PriceBuy      *float64       `json:"price_buy" gorm:"column:price_buy"`
	Status        int8           `json:"status" gorm:"column:status"`
	CreatedByID   *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID   *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID   *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
