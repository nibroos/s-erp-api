package models

import (
	"gorm.io/gorm"
)

type SoDt struct {
	gorm.Model
	ID              uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ProductUuid     string         `json:"product_uuid" gorm:"column:product_uuid"`
	SalesOrderID    *uint          `json:"sales_order_id" gorm:"column:sales_order_id"`
	ItemUnitID      *uint          `json:"item_unit_id" gorm:"column:item_unit_id"`
	VatID           *uint          `json:"vat_id" gorm:"column:vat_id"`
	Pph23ID         *uint          `json:"pph23_id" gorm:"column:pph23_id"`
	RefID           *uint          `json:"ref_id" gorm:"column:ref_id"`
	ItemID          uint           `json:"item_id" gorm:"column:item_id"`
	RefType         *string        `json:"ref_type" gorm:"column:ref_type"`
	ItemType        *string        `json:"item_type" gorm:"column:item_type"`
	RefJSON         *string        `json:"ref_json" gorm:"column:ref_json"`
	ItemJSON        *string        `json:"item_json" gorm:"column:item_json"`
	GenCode         *string        `json:"gen_code" gorm:"column:gen_code"`
	Remark          *string        `json:"remark" gorm:"column:remark"`
	VatPerc         *float64       `json:"vat_perc" gorm:"column:vat_perc"`
	VatPercAm       *float64       `json:"vat_perc_am" gorm:"column:vat_perc_am"`
	Pph23Perc       *float64       `json:"pph23_perc" gorm:"column:pph23_perc"`
	Pph23PercAm     *float64       `json:"pph23_perc_am" gorm:"column:pph23_perc_am"`
	MarkupPerc      *float64       `json:"markup_perc" gorm:"column:markup_perc"`
	MarkupPercAm    *float64       `json:"markup_perc_am" gorm:"column:markup_perc_am"`
	IsVat           *int8          `json:"is_vat" gorm:"column:is_vat"`
	IsPph23         *int8          `json:"is_pph23" gorm:"column:is_pph23"`
	IsLockPriceSell *int8          `json:"is_lock_price_sell" gorm:"column:is_lock_price_sell"`
	IsLockMarkup    *int8          `json:"is_lock_markup" gorm:"column:is_lock_markup"`
	QtyOut          *float64       `json:"qty_out" gorm:"column:qty_out"`
	Qty             *float64       `json:"qty" gorm:"column:qty"`
	PriceSell       *float64       `json:"price_sell" gorm:"column:price_sell"`
	PriceBuy        *float64       `json:"price_buy" gorm:"column:price_buy"`
	SubtotalSell    *float64       `json:"subtotal_sell" gorm:"column:subtotal_sell"`
	SubtotalBuy     *float64       `json:"subtotal_buy" gorm:"column:subtotal_buy"`
	DiscAm          *float64       `json:"disc_am" gorm:"column:disc_am"`
	DiscPerc        *float64       `json:"disc_perc" gorm:"column:disc_perc"`
	DiscPercNum     *float64       `json:"disc_perc_num" gorm:"column:disc_perc_num"`
	DiscPercAm      *float64       `json:"disc_perc_am" gorm:"column:disc_perc_am"`
	DiscFinal       *float64       `json:"disc_final" gorm:"column:disc_final"`
	DiscType        *string        `json:"disc_type" gorm:"column:disc_type"`
	TotalAm         *float64       `json:"total_am" gorm:"column:total_am"`
	SiTotalAm       *float64       `json:"si_total_am" gorm:"column:si_total_am"`
	SaTotalAm       *float64       `json:"sa_total_am" gorm:"column:sa_total_am"`
	TotalDp         *float64       `json:"total_dp" gorm:"column:total_dp"`
	HistoryTotalDp  *float64       `json:"history_total_dp" gorm:"column:history_total_dp"`
	CreatedByID     *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID     *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID     *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// SubtotalSell = PriceSell * Qty
// SubtotalBuy = PriceBuy * Qty
// DiscPercNum = PriceSell - (PriceSell * DiscPerc/100)
// DiscPercAm = DiscPercNum * Qty
// DiscType = 'p' OR 'a'
// DiscFinal = DiscPercAm OR DiscAm
// VatPercAm = DiscFinal * VatPerc/100
// TotalAm = DiscFinal + VatPercAm
