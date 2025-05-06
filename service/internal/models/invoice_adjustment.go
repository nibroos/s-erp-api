package models

import (
	"time"

	"gorm.io/gorm"
)

type InvoiceAdjustment struct {
	gorm.Model
	ID              uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID      *uint          `json:"customer_id" gorm:"column:customer_id"`
	CurrencyID      *uint          `json:"currency_id" gorm:"column:currency_id"`
	BranchID        *uint          `json:"branch_id" gorm:"column:branch_id"`
	BankID          *uint          `json:"bank_id" gorm:"column:bank_id"`
	InvoiceNo       *string        `json:"invoice_no" gorm:"column:invoice_no"`
	PaymentDate     *time.Time     `json:"payment_date" gorm:"column:payment_date"`
	PaymentAmount   *float64       `json:"payment_amount" gorm:"column:payment_amount"`
	ExchangeRate    *float64       `json:"exchange_rate" gorm:"column:exchange_rate"`
	Reference       *string        `json:"reference" gorm:"column:reference"`
	RefStartDate    *time.Time     `json:"ref_start_date" gorm:"column:ref_start_date"`
	RefEndDate      *time.Time     `json:"ref_end_date" gorm:"column:ref_end_date"`
	Remark          *string        `json:"remark" gorm:"column:remark"`
	RevNo           *int           `json:"rev_no" gorm:"column:rev_no"`
	TotalInvoice    *float64       `json:"total_invoice" gorm:"column:total_invoice"`
	TotalAdjustment *float64       `json:"total_adjustment" gorm:"column:total_adjustment"`
	TotalBalance    *float64       `json:"total_balance" gorm:"column:total_balance"`
	TotalAdminBank  *float64       `json:"total_admin_bank" gorm:"column:total_admin_bank"`
	GrandTotal      *float64       `json:"grand_total" gorm:"column:grand_total"`
	CreatedByID     *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID     *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID     *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (InvoiceAdjustment) TableName() string {
	return "invoice_adjustments"
}
