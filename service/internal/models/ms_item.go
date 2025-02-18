package models

import (
	"gorm.io/gorm"
)

type MsItem struct {
	gorm.Model
	ID             uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ItemSubGroupID *uint          `json:"item_sub_group_id" gorm:"column:item_sub_group_id"`
	UnitID         *uint          `json:"unit_id" gorm:"column:unit_id"`
	Name           string         `json:"name" gorm:"column:name"`
	Specification  *string        `json:"specification" gorm:"column:specification"`
	TpbCode        *string        `json:"tpb_code" gorm:"column:tpb_code"`
	PriceSell      *float64       `json:"price_sell" gorm:"column:price_sell"`
	PriceBuy       *float64       `json:"price_buy" gorm:"column:price_buy"`
	MinimumStock   *float64       `json:"minimum_stock" gorm:"column:minimum_stock"`
	Status         int8           `json:"status" gorm:"column:status"`
	CreatedByID    *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID    *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID    *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
