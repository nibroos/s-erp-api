package models

import (
	"gorm.io/gorm"
)

type PurchaseOrder struct {
	gorm.Model
	ID                       uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID               *uint          `json:"customer_id" gorm:"column:customer_id"`
	PurchaseTypeID           *uint          `json:"purchase_type_id" gorm:"column:purchase_type_id"`
	CurrencyID               *uint          `json:"currency_id" gorm:"column:currency_id"`
	VatID                    *uint          `json:"vat_id" gorm:"column:vat_id"`
	Pph23ID                  *uint          `json:"pph23_id" gorm:"column:pph23_id"`
	PaymentTermID            *uint          `json:"payment_term_id" gorm:"column:payment_term_id"`
	PaymentID                *uint          `json:"payment_id" gorm:"column:payment_id"`
	ShippingTermID           *uint          `json:"shipping_term_id" gorm:"column:shipping_term_id"`
	BranchID                 *uint          `json:"branch_id" gorm:"column:branch_id"`
	IsVat                    *int           `json:"is_vat" gorm:"column:is_vat"`
	RevNo                    *int           `json:"rev_no" gorm:"column:rev_no"`
	PoNo                     *string        `json:"po_no" gorm:"column:po_no"`
	PoNoOri                  *string        `json:"po_no_ori" gorm:"column:po_no_ori"`
	PoDate                   *string        `json:"po_date" gorm:"column:po_date"`
	DeliveryDate             *string        `json:"delivery_date" gorm:"column:delivery_date"`
	ShippingDestination      *string        `json:"shipping_destination" gorm:"column:shipping_destination"`
	Remark                   *string        `json:"remark" gorm:"column:remark"`
	ExchangeRate             *float64       `json:"exchange_rate" gorm:"column:exchange_rate"`
	DiscountPercentage       *float64       `json:"discount_percentage" gorm:"column:discount_percentage"`
	DiscountAmount           *float64       `json:"discount_amount" gorm:"column:discount_amount"`
	DiscountPercentageAmount *float64       `json:"discount_percentage_amount" gorm:"column:discount_percentage_amount"`
	DiscountAmountProduct    *float64       `json:"discount_amount_product" gorm:"column:discount_amount_product"`
	DiscountFinalHeader      *float64       `json:"discount_final_header" gorm:"column:discount_final_header"`
	DiscountType             *string        `json:"discount_type" gorm:"column:discount_type"`
	VatPercentage            *float64       `json:"vat_percentage" gorm:"column:vat_percentage"`
	Pph23Percentage          *float64       `json:"pph23_percentage" gorm:"column:pph23_percentage"`
	TotalAmountProducts      *float64       `json:"total_amount_products" gorm:"column:total_amount_products"`
	Subtotal                 *float64       `json:"subtotal" gorm:"column:subtotal"`
	TotalQty                 *float64       `json:"total_qty" gorm:"column:total_qty"`
	TotalDiscount            *float64       `json:"total_discount" gorm:"column:total_discount"`
	TotalPph23               *float64       `json:"total_pph23" gorm:"column:total_pph23"`
	TotalVat                 *float64       `json:"total_vat" gorm:"column:total_vat"`
	GrandTotal               *float64       `json:"grand_total" gorm:"column:grand_total"`
	Status                   string         `json:"status" gorm:"column:status"`
	CreatedByID              *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID              *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID              *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt                gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
