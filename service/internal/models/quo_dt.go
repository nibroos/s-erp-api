package models

import (
	"gorm.io/gorm"
)

type QuoDt struct {
	gorm.Model
	ID           *uint          `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	QuotationID  *uint          `json:"quotation_id" gorm:"column:quotation_id"`
	ItemUnitID   *uint          `json:"item_unit_id" gorm:"column:item_unit_id"`
	VatID        *uint          `json:"vat_id" gorm:"column:vat_id"`
	RefID        *uint          `json:"ref_id" gorm:"column:ref_id"`
	ItemID       uint           `json:"item_id" gorm:"column:item_id"`
	RefType      *string        `json:"ref_type" gorm:"column:ref_type"`
	ItemType     *string        `json:"item_type" gorm:"column:item_type"`
	RefJSON      *string        `json:"ref_json" gorm:"column:ref_json"`
	ItemJSON     *string        `json:"item_json" gorm:"column:item_json"`
	GenCode      *string        `json:"gen_code" gorm:"column:gen_code"`
	Remark       *string        `json:"remark" gorm:"column:remark"`
	VatPerc      *float64       `json:"vat_perc" gorm:"column:vat_perc"`
	VatPercAm    *float64       `json:"vat_perc_am" gorm:"column:vat_perc_am"`
	QtySO        *float64       `json:"qty_so" gorm:"column:qty_so"`
	Qty          *float64       `json:"qty" gorm:"column:qty"`
	PriceSell    *float64       `json:"price_sell" gorm:"column:price_sell"`
	PriceBuy     *float64       `json:"price_buy" gorm:"column:price_buy"`
	SubtotalSell *float64       `json:"subtotal_sell" gorm:"column:subtotal_sell"`
	SubtotalBuy  *float64       `json:"subtotal_buy" gorm:"column:subtotal_buy"`
	DiscAm       *float64       `json:"disc_am" gorm:"column:disc_am"`
	DiscPerc     *float64       `json:"disc_perc" gorm:"column:disc_perc"`
	DiscPercNum  *float64       `json:"disc_perc_num" gorm:"column:disc_perc_num"`
	DiscPercAm   *float64       `json:"disc_perc_am" gorm:"column:disc_perc_am"`
	DiscFinal    *float64       `json:"disc_final" gorm:"column:disc_final"`
	DiscType     *string        `json:"disc_type" gorm:"column:disc_type"`
	TotalAm      *float64       `json:"total_am" gorm:"column:total_am"`
	CreatedByID  *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID  *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID  *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// SubtotalSell = PriceSell * Qty
// SubtotalBuy = PriceBuy * Qty
// DiscPercNum = PriceSell - (PriceSell * DiscPerc/100)
// DiscPercAm = DiscPercNum * Qty
// DiscType = 'p' OR 'a'
// DiscFinal = DiscPercAm OR DiscAm
// VatPercAm = DiscFinal * VatPerc/100
// TotalAm = DiscFinal + VatPercAm
