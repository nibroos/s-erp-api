package models

import (
	"gorm.io/gorm"
)

type QuoDt struct {
	gorm.Model
	ID                *uint          `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	QuotationID       *uint          `json:"quotation_id" gorm:"column:quotation_id"`
	RefID             *uint          `json:"product_item_id" gorm:"column:product_item_id"`
	ItemUnitID        *uint          `json:"item_unit_id" gorm:"column:item_unit_id"`
	VatID             *uint          `json:"vat_id" gorm:"column:vat_id"`
	ProductID         *uint          `json:"product_id" gorm:"column:product_id"`
	ProductItemUnitID *uint          `json:"product_item_unit_id" gorm:"column:product_item_unit_id"`
	QuoDtRefID        *uint          `json:"quo_dt_ref_id" gorm:"column:quo_dt_ref_id"`
	RefJSON           *string        `json:"ref_json" gorm:"column:ref_json"`
	RefType           *string        `json:"ref_type" gorm:"column:ref_type"`
	Remark            *string        `json:"remark" gorm:"column:remark"`
	VatPerc           *float64       `json:"vat_perc" gorm:"column:vat_perc"`
	QtySO             *float64       `json:"qty_so" gorm:"column:qty_so"`
	Qty               *float64       `json:"qty" gorm:"column:qty"`
	PriceSell         *float64       `json:"price_sell" gorm:"column:price_sell"`
	Subtotal          *float64       `json:"subtotal" gorm:"column:subtotal"`
	DiscAm            *float64       `json:"disc_am" gorm:"column:disc_am"`
	DiscPerc          *float64       `json:"disc_perc" gorm:"column:disc_perc"`
	TotalAm           *float64       `json:"total_am" gorm:"column:total_am"`
	PQtySO            *float64       `json:"p_qty_so" gorm:"column:p_qty_so"`
	PQty              *float64       `json:"p_qty" gorm:"column:p_qty"`
	PPriceSell        *float64       `json:"p_price_sell" gorm:"column:p_price_sell"`
	PSubtotal         *float64       `json:"p_subtotal" gorm:"column:p_subtotal"`
	PDiscAm           *float64       `json:"p_disc_am" gorm:"column:p_disc_am"`
	PDiscPerc         *float64       `json:"p_disc_perc" gorm:"column:p_disc_perc"`
	PTotalAm          *float64       `json:"p_total_am" gorm:"column:p_total_am"`
	CreatedByID       *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID       *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID       *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
