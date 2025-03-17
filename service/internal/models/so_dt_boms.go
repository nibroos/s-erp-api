package models

import (
	"gorm.io/gorm"
)

type SoDtBom struct {
	gorm.Model
	ID           uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	ProductUuid  string         `json:"product_uuid" gorm:"column:product_uuid"`
	SotationID   *uint          `json:"sales_order_id" gorm:"column:sales_order_id"`
	SoDtID       uint           `json:"so_dt_id" gorm:"column:so_dt_id"`
	ProductID    uint           `json:"product_id" gorm:"column:product_id"`
	ItemID       uint           `json:"item_id" gorm:"column:item_id"`
	ItemUnitID   *uint          `json:"item_unit_id" gorm:"column:item_unit_id"`
	ItemJSON     *string        `json:"item_json" gorm:"column:item_json"`
	GenCode      *string        `json:"gen_code" gorm:"column:gen_code"`
	Remark       *string        `json:"remark" gorm:"column:remark"`
	Qty          float64        `json:"qty" gorm:"column:qty"`
	PriceSell    float64        `json:"price_sell" gorm:"column:price_sell"`
	PriceBuy     float64        `json:"price_buy" gorm:"column:price_buy"`
	SubtotalSell float64        `json:"subtotal_sell" gorm:"column:subtotal_sell"`
	SubtotalBuy  float64        `json:"subtotal_buy" gorm:"column:subtotal_buy"`
	CreatedByID  *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID  *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID  *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
