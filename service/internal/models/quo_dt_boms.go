package models

import (
	"time"

	"gorm.io/gorm"
)

type QuoDtBom struct {
	gorm.Model
	ID          *uint          `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	QuotationID uint           `json:"quotation_id" gorm:"column:quotation_id"`
	QuoDtID     uint           `json:"quo_dt_id" gorm:"column:quo_dt_id"`
	ProductID   uint           `json:"product_id" gorm:"column:product_id"`
	ItemID      uint           `json:"item_id" gorm:"column:item_id"`
	ItemUnitID  *uint          `json:"item_unit_id" gorm:"column:item_unit_id"`
	RefJSON     *string        `json:"ref_json" gorm:"column:ref_json"`
	Remark      *string        `json:"remark" gorm:"column:remark"`
	Qty         *float64       `json:"qty" gorm:"column:qty"`
	PriceSell   *float64       `json:"price_sell" gorm:"column:price_sell"`
	PriceBuy    *float64       `json:"price_buy" gorm:"column:price_buy"`
	Subtotal    *float64       `json:"subtotal" gorm:"column:subtotal"`
	CreatedByID *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
