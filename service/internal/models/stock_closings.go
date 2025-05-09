package models

import (
	"gorm.io/gorm"
)

type StockClosings struct {
	gorm.Model
	ID             uint     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	WarehouseID    uint     `json:"warehouse_id" gorm:"column:warehouse_id"`
	ItemID         uint     `json:"item_id" gorm:"column:item_id"`
	ClosingAt      *string  `json:"closing_at" gorm:"column:closing_at"`
	LastClosingAt  *string  `json:"last_closing_at" gorm:"column:last_closing_at"`
	BeginQty       *float64 `json:"begin_qty" gorm:"column:begin_qty"`
	InQty          *float64 `json:"in_qty" gorm:"column:in_qty"`
	OutQty         *float64 `json:"out_qty" gorm:"column:out_qty"`
	AdjustmentQty  *float64 `json:"adjustment_qty" gorm:"column:adjustment_qty"`
	EndQty         *float64 `json:"end_qty" gorm:"column:end_qty"`
	PriceSell      *float64 `json:"price_sell" gorm:"column:price_sell"`
	PriceBuy       *float64 `json:"price_buy" gorm:"column:price_buy"`
	TotalValueSell *float64 `json:"total_value_sell" gorm:"column:total_value_sell"`
	TotalValueBuy  *float64 `json:"total_value_buy" gorm:"column:total_value_buy"`
	IsFinalized    *int     `json:"is_finalized" gorm:"column:is_finalized"`
	Remarks        *string  `json:"remarks" gorm:"column:remarks"`
	CreatedByID    *uint    `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID    *uint    `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID    *uint    `json:"deleted_by_id" gorm:"column:deleted_by_id"`

	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (StockClosings) TableName() string {
	return "stock_closings"
}
