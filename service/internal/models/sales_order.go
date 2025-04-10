package models

import (
	"gorm.io/gorm"
)

type SalesOrder struct {
	gorm.Model
	ID            uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID    *uint          `json:"customer_id" gorm:"column:customer_id"`
	OrderTypeID   *uint          `json:"order_type_id" gorm:"column:order_type_id"`
	CurrencyID    *uint          `json:"currency_id" gorm:"column:currency_id"`
	WarehouseID   *uint          `json:"warehouse_id" gorm:"column:warehouse_id"`
	PaymentID     *uint          `json:"payment_id" gorm:"column:payment_id"`
	VatID         *uint          `json:"vat_id" gorm:"column:vat_id"`
	Pph23ID       *uint          `json:"pph23_id" gorm:"column:pph23_id"`
	BranchID      *uint          `json:"branch_id" gorm:"column:branch_id"`
	RevNo         *int           `json:"rev_no" gorm:"column:rev_no"`
	PoBuyerNo     string         `json:"po_buyer_no" gorm:"column:po_buyer_no"`
	PoBuyerNoOri  *string        `json:"po_buyer_no_ori" gorm:"column:po_buyer_no_ori"`
	SalesOrderNo  *string        `json:"sales_order_no" gorm:"column:sales_order_no"`
	Remark        *string        `json:"remark" gorm:"column:remark"`
	ShipDest      *string        `json:"ship_dest" gorm:"column:ship_dest"`
	Status        string         `json:"status" gorm:"column:status"`
	ExchangeRate  *float64       `json:"exchange_rate" gorm:"column:exchange_rate"`
	VatPerc       *float64       `json:"vat_perc" gorm:"column:vat_perc"`
	Pph23Perc     *float64       `json:"pph23_perc" gorm:"column:pph23_perc"`
	MarkupPerc    *float64       `json:"markup_perc" gorm:"column:markup_perc"`
	IsVat         *int           `json:"is_vat" gorm:"column:is_vat"`
	IsPph23       *int           `json:"is_pph23" gorm:"column:is_pph23"`
	DiscAm        *float64       `json:"disc_am" gorm:"column:disc_am"`
	DiscPerc      *float64       `json:"disc_perc" gorm:"column:disc_perc"`
	DiscPercAm    *float64       `json:"disc_perc_am" gorm:"column:disc_perc_am"`
	DiscFinal     *float64       `json:"disc_final" gorm:"column:disc_final"`
	DiscType      *string        `json:"disc_type" gorm:"column:disc_type"`
	TotalQty      *float64       `json:"total_qty" gorm:"column:total_qty"`
	Subtotal      *float64       `json:"subtotal" gorm:"column:subtotal"`
	TotalDiscount *float64       `json:"total_discount" gorm:"column:total_discount"`
	TotalPph23    *float64       `json:"total_pph23" gorm:"column:total_pph23"`
	TotalVat      *float64       `json:"total_vat" gorm:"column:total_vat"`
	GrandTotal    *float64       `json:"grand_total" gorm:"column:grand_total"`
	QtyOut        *float64       `json:"qty_out" gorm:"column:qty_out"`
	SiTotalAm     *float64       `json:"si_total_am" gorm:"column:si_total_am"`
	SaTotalAm     *float64       `json:"sa_total_am" gorm:"column:sa_total_am"`
	OrderAt       *string        `json:"order_at" gorm:"column:order_at"`
	ShippingAt    *string        `json:"shipping_at" gorm:"column:shipping_at"`
	AgreeAt       *string        `json:"agree_at" gorm:"column:agree_at"`
	DueAt         *string        `json:"due_at" gorm:"column:due_at"`
	ExpiredAt     *string        `json:"expired_at" gorm:"column:expired_at"`
	CreatedByID   *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID   *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID   *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
