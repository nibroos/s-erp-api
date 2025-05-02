package models

import (
	"gorm.io/gorm"
)

type Inventory struct {
	gorm.Model
	ID             uint     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID     *uint    `json:"customer_id" gorm:"column:customer_id"`
	IOTypeID       *uint    `json:"io_type_id" gorm:"column:io_type_id"`
	CurrencyID     *uint    `json:"currency_id" gorm:"column:currency_id"`
	PaymentTermID  *uint    `json:"payment_term_id" gorm:"column:payment_term_id"`
	WarehouseID    uint     `json:"warehouse_id" gorm:"column:warehouse_id"`
	VatID          *uint    `json:"vat_id" gorm:"column:vat_id"`
	Pph23ID        *uint    `json:"pph23_id" gorm:"column:pph23_id"`
	BranchID       *uint    `json:"branch_id" gorm:"column:branch_id"`
	RevNo          *int     `json:"rev_no" gorm:"column:rev_no"`
	InventoryNo    *string  `json:"inventory_no" gorm:"column:inventory_no"`
	InventoryNoOri *string  `json:"inventory_no_ori" gorm:"column:inventory_no_ori"`
	DoNo           *string  `json:"do_no" gorm:"column:do_no"`
	SuratJalanNo   *string  `json:"surat_jalan_no" gorm:"column:surat_jalan_no"`
	InvoiceNo      *string  `json:"invoice_no" gorm:"column:invoice_no"`
	ShipDest       *string  `json:"ship_dest" gorm:"column:ship_dest"`
	Remark         *string  `json:"remark" gorm:"column:remark"`
	Status         string   `json:"status" gorm:"column:status"`
	ExchangeRate   *float64 `json:"exchange_rate" gorm:"column:exchange_rate"`
	IsVat          *int     `json:"is_vat" gorm:"column:is_vat"`
	VatPerc        *float64 `json:"vat_perc" gorm:"column:vat_perc"`
	Pph23Perc      *float64 `json:"pph23_perc" gorm:"column:pph23_perc"`
	TotalQty       *float64 `json:"total_qty" gorm:"column:total_qty"`
	Subtotal       *float64 `json:"subtotal" gorm:"column:subtotal"`
	TotalPph23     *float64 `json:"total_pph23" gorm:"column:total_pph23"`
	TotalVat       *float64 `json:"total_vat" gorm:"column:total_vat"`
	GrandTotal     *float64 `json:"grand_total" gorm:"column:grand_total"`
	DoAt           *string  `json:"do_at" gorm:"column:do_at"`
	IngoingAt      *string  `json:"ingoing_at" gorm:"column:ingoing_at"`
	InvoiceAt      *string  `json:"invoice_at" gorm:"column:invoice_at"`

	CreatedByID *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
