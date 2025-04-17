package models

import (
	"time"

	"gorm.io/gorm"
)

type SalesInvoice struct {
	gorm.Model
	ID                       uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID               *uint          `json:"customer_id" gorm:"column:customer_id"`
	CurrencyID               *uint          `json:"currency_id" gorm:"column:currency_id"`
	PaymentTermID            *uint          `json:"payment_term_id" gorm:"column:payment_term_id"`
	VatID                    *uint          `json:"vat_id" gorm:"column:vat_id"`
	Pph23ID                  *uint          `json:"pph23_id" gorm:"column:pph23_id"`
	BranchID                 *uint          `json:"branch_id" gorm:"column:branch_id"`
	BankID                   *uint          `json:"bank_id" gorm:"column:bank_id"`
	InvoiceNo                *string        `json:"invoice_no" gorm:"column:invoice_no"`
	InvoiceDate              *time.Time     `json:"invoice_date" gorm:"column:invoice_date"`
	ExchangeRate             *float64       `json:"exchange_rate" gorm:"column:exchange_rate"`
	Remark                   *string        `json:"remark" gorm:"column:remark"`
	RevNo                    *int           `json:"rev_no" gorm:"column:rev_no"`
	Status                   *string        `json:"status" gorm:"column:status"`
	Pph23Percentage          *float64       `json:"pph23_percentage" gorm:"column:pph23_percentage"`
	VatPercentage            *float64       `json:"vat_percentage" gorm:"column:vat_percentage"`
	DiscountAmount           *float64       `json:"discount_amount" gorm:"column:discount_amount"`
	DiscountPercentage       *float64       `json:"discount_percentage" gorm:"column:discount_percentage"`
	DiscountPercentageAmount *float64       `json:"discount_percentage_amount" gorm:"column:discount_percentage_amount"`
	DiscountFinal            *float64       `json:"discount_final" gorm:"column:discount_final"`
	DiscountType             *string        `json:"discount_type" gorm:"column:discount_type"`
	Subtotal                 *float64       `json:"subtotal" gorm:"column:subtotal"`
	TotalAmountProducts      *float64       `json:"total_amount_products" gorm:"column:total_amount_products"`
	TotalDpProducts          *float64       `json:"total_dp_products" gorm:"column:total_dp_products"`
	TotalBalanceProducts     *float64       `json:"total_balance_products" gorm:"column:total_balance_products"`
	TotalQty                 *float64       `json:"total_qty" gorm:"column:total_qty"`
	TotalDiscount            *float64       `json:"total_discount" gorm:"column:total_discount"`
	TotalPph23               *float64       `json:"total_pph23" gorm:"column:total_pph23"`
	TotalVat                 *float64       `json:"total_vat" gorm:"column:total_vat"`
	GrandTotal               *float64       `json:"grand_total" gorm:"column:grand_total"`
	CreatedByID              *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID              *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID              *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt                gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (SalesInvoice) TableName() string {
	return "sales_invoices"
}
