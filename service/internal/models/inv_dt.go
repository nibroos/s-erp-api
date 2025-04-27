package models

import (
	"gorm.io/gorm"
)

type InvDt struct {
	gorm.Model
	ID           uint     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	InventoryID  uint     `json:"inventory_id" gorm:"column:inventory_id"`
	ItemUnitID   uint     `json:"item_unit_id" gorm:"column:item_unit_id"`
	VatID        *uint    `json:"vat_id" gorm:"column:vat_id"`
	Pph23ID      *uint    `json:"pph23_id" gorm:"column:pph23_id"`
	RefSoDtID    *uint    `json:"ref_so_dt_id" gorm:"column:ref_so_dt_id"`
	RefSoDtBomID *uint    `json:"ref_so_dt_bom_id" gorm:"column:ref_so_dt_bom_id"`
	RefPoDtID    *uint    `json:"ref_po_dt_id" gorm:"column:ref_po_dt_id"`
	RefPoDtBomID *uint    `json:"ref_po_dt_bom_id" gorm:"column:ref_po_dt_bom_id"`
	RefInvDtID   *uint    `json:"ref_inv_dt_id" gorm:"column:ref_inv_dt_id"`
	RefProductID *uint    `json:"ref_product_id" gorm:"column:ref_product_id"`
	ItemID       uint     `json:"item_id" gorm:"column:item_id"`
	ProductUuid  string   `json:"product_uuid" gorm:"column:product_uuid"`
	RefType      string   `json:"ref_type" gorm:"column:ref_type"`
	ItemType     string   `json:"item_type" gorm:"column:item_type"`
	GenCode      *string  `json:"gen_code" gorm:"column:gen_code"`
	Remark       *string  `json:"remark" gorm:"column:remark"`
	VatPerc      *float64 `json:"vat_perc" gorm:"column:vat_perc"`
	VatPercAm    *float64 `json:"vat_perc_am" gorm:"column:vat_perc_am"`
	Pph23Perc    *float64 `json:"pph23_perc" gorm:"column:pph23_perc"`
	Pph23PercAm  *float64 `json:"pph23_perc_am" gorm:"column:pph23_perc_am"`
	IsVat        *int     `json:"is_vat" gorm:"column:is_vat"`
	IsPph23      *int     `json:"is_pph23" gorm:"column:is_pph23"`
	QtyInvoice   *float64 `json:"qty_invoice" gorm:"column:qty_invoice"`
	QtyOut       *float64 `json:"qty_out" gorm:"column:qty_out"`
	Qty          *float64 `json:"qty" gorm:"column:qty"`
	PriceSell    *float64 `json:"price_sell" gorm:"column:price_sell"`
	PriceBuy     *float64 `json:"price_buy" gorm:"column:price_buy"`
	SubtotalSell *float64 `json:"subtotal_sell" gorm:"column:subtotal_sell"`
	SubtotalBuy  *float64 `json:"subtotal_buy" gorm:"column:subtotal_buy"`
	TotalAm      *float64 `json:"total_am" gorm:"column:total_am"`
	ExpiredAt    *string  `json:"expired_at" gorm:"column:expired_at"`
	ItemJSON     string   `json:"item_json" gorm:"column:item_json"`
	CreatedByID  *uint    `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID  *uint    `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID  *uint    `json:"deleted_by_id" gorm:"column:deleted_by_id"`
}

func (InvDt) TableName() string {
	return "inv_dts"
}
