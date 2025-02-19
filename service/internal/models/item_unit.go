package models

import (
	"gorm.io/gorm"
)

type ItemUnit struct {
	gorm.Model
	ID          uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	MsItemID    uint           `json:"ms_item_id" gorm:"column:ms_item_id"`
	UnitID      uint           `json:"unit_id" gorm:"column:unit_id"`
	Conversion  *float64       `json:"conversion" gorm:"column:conversion"`
	PriceSell   *float64       `json:"price_sell" gorm:"column:price_sell"`
	PriceBuy    *float64       `json:"price_buy" gorm:"column:price_buy"`
	Status      int8           `json:"status" gorm:"column:status"`
	CreatedByID *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
