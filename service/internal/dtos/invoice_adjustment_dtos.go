package dtos

import (
	"encoding/json"
	"time"
)

type GetInvoiceAdjustmentsRequest struct {
	Global         *string `json:"global"`
	InvoiceNo      *string `json:"invoice_no"`
	Reference      *string `json:"reference"`
	Remark         *string `json:"remark"`
	RevNo          *int    `json:"rev_no"`
	CustomerID     *int    `json:"customer_id"`
	CurrencyID     *int    `json:"currency_id"`
	BranchID       *int    `json:"branch_id"`
	BankID         *int    `json:"bank_id"`
	StartDate      *string `json:"start_date"`
	EndDate        *string `json:"end_date"`
	PerPage        *string `json:"per_page" default:"10"`
	Page           *string `json:"page" default:"1"`
	OrderColumn    *string `json:"order_column" default:"id"`
	OrderDirection *string `json:"order_direction" default:"asc"`
}

type CreateInvoiceAdjustmentDtRequest struct {
	InvoiceUUID      string           `json:"invoice_uuid"`
	RefID            *uint            `json:"ref_id"`
	RefType          *string          `json:"ref_type"`
	RefJSON          *json.RawMessage `json:"ref_json"`
	InvoiceNo        *string          `json:"invoice_no"`
	InvoiceDate      *string          `json:"invoice_date"`
	InvoiceAmount    *float64         `json:"invoice_amount"`
	TotalAdjustment  *float64         `json:"total_adjustment"`
	BalanceAmount    *float64         `json:"balance_amount"`
	AdjustmentAmount *float64         `json:"adjustment_amount"`
	AdminBank        *float64         `json:"admin_bank"`
	TotalAmount      *float64         `json:"total_amount"`
}

type CreateInvoiceAdjustmentRequest struct {
	CustomerID *uint `json:"customer_id"`
	CurrencyID *uint `json:"currency_id"`
	BranchID   *uint `json:"branch_id"`
	BankID     *uint `json:"bank_id"`
	// InvoiceNo       *string                            `json:"invoice_no"`
	PaymentDate     *string                            `json:"payment_date"`
	PaymentAmount   *float64                           `json:"payment_amount"`
	ExchangeRate    *float64                           `json:"exchange_rate"`
	Reference       *string                            `json:"reference"`
	RefStartDate    *string                            `json:"ref_start_date"`
	RefEndDate      *string                            `json:"ref_end_date"`
	Remark          *string                            `json:"remark"`
	RevNo           *int                               `json:"rev_no"`
	TotalInvoice    *float64                           `json:"total_invoice"`
	TotalAdjustment *float64                           `json:"total_adjustment"`
	TotalBalance    *float64                           `json:"total_balance"`
	TotalAdminBank  *float64                           `json:"total_admin_bank"`
	GrandTotal      *float64                           `json:"grand_total"`
	AdjustmentDts   []CreateInvoiceAdjustmentDtRequest `json:"adjustment_dts"`
}

type UpdateInvoiceAdjustmentDtRequest struct {
	ID                  *uint            `json:"id"`
	InvoiceUUID         string           `json:"invoice_uuid"`
	InvoiceAdjustmentID *uint            `json:"invoice_adjustment_id"`
	RefID               *uint            `json:"ref_id"`
	RefType             *string          `json:"ref_type"`
	RefJSON             *json.RawMessage `json:"ref_json"`
	InvoiceNo           *string          `json:"invoice_no"`
	InvoiceDate         *string          `json:"invoice_date"`
	InvoiceAmount       *float64         `json:"invoice_amount"`
	TotalAdjustment     *float64         `json:"total_adjustment"`
	BalanceAmount       *float64         `json:"balance_amount"`
	AdjustmentAmount    *float64         `json:"adjustment_amount"`
	AdminBank           *float64         `json:"admin_bank"`
	TotalAmount         *float64         `json:"total_amount"`
}

type UpdateInvoiceAdjustmentRequest struct {
	ID                  uint                               `json:"id"`
	InvoiceAdjustmentID *uint                              `json:"invoice_adjustment_id"`
	CustomerID          *uint                              `json:"customer_id"`
	CurrencyID          *uint                              `json:"currency_id"`
	BranchID            *uint                              `json:"branch_id"`
	BankID              *uint                              `json:"bank_id"`
	InvoiceNo           *string                            `json:"invoice_no"`
	PaymentDate         *string                            `json:"payment_date"`
	PaymentAmount       *float64                           `json:"payment_amount"`
	ExchangeRate        *float64                           `json:"exchange_rate"`
	Reference           *string                            `json:"reference"`
	RefStartDate        *string                            `json:"ref_start_date"`
	RefEndDate          *string                            `json:"ref_end_date"`
	Remark              *string                            `json:"remark"`
	RevNo               *int                               `json:"rev_no"`
	TotalInvoice        *float64                           `json:"total_invoice"`
	TotalAdjustment     *float64                           `json:"total_adjustment"`
	TotalBalance        *float64                           `json:"total_balance"`
	TotalAdminBank      *float64                           `json:"total_admin_bank"`
	GrandTotal          *float64                           `json:"grand_total"`
	AdjustmentDts       []UpdateInvoiceAdjustmentDtRequest `json:"adjustment_dts"`
}

type GetInvoiceAdjustmentByIDRequest struct {
	ID uint `json:"id"`
}

type GetInvoiceAdjustmentParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetInvoiceAdjustmentParams(id uint) *GetInvoiceAdjustmentParams {
	defaultIsDeleted := 0
	return &GetInvoiceAdjustmentParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type GetInvoiceAdjustmentDtParams struct {
	ID                  uint
	InvoiceAdjustmentID uint
	IsDeleted           *int
}

type DeleteInvoiceAdjustmentRequest struct {
	ID uint `json:"id"`
}

type InvoiceAdjustmentListDTO struct {
	ID              int      `json:"id" db:"id"`
	CustomerID      *uint    `json:"customer_id" db:"customer_id"`
	CurrencyID      *uint    `json:"currency_id" db:"currency_id"`
	BranchID        *uint    `json:"branch_id" db:"branch_id"`
	BankID          *uint    `json:"bank_id" db:"bank_id"`
	InvoiceNo       *string  `json:"invoice_no" db:"invoice_no"`
	PaymentDate     *string  `json:"payment_date" db:"payment_date"`
	PaymentAmount   *float64 `json:"payment_amount" db:"payment_amount"`
	ExchangeRate    *float64 `json:"exchange_rate" db:"exchange_rate"`
	Reference       *string  `json:"reference" db:"reference"`
	RefStartDate    *string  `json:"ref_start_date" db:"ref_start_date"`
	RefEndDate      *string  `json:"ref_end_date" db:"ref_end_date"`
	Remark          *string  `json:"remark" db:"remark"`
	RevNo           *int     `json:"rev_no" db:"rev_no"`
	TotalInvoice    *float64 `json:"total_invoice" db:"total_invoice"`
	TotalAdjustment *float64 `json:"total_adjustment" db:"total_adjustment"`
	TotalBalance    *float64 `json:"total_balance" db:"total_balance"`
	TotalAdminBank  *float64 `json:"total_admin_bank" db:"total_admin_bank"`
	GrandTotal      *float64 `json:"grand_total" db:"grand_total"`
	CreatedByID     *uint    `json:"created_by_id" db:"created_by_id"`
	UpdatedByID     *uint    `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID     *uint    `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName   *string  `json:"created_by_name" db:"created_by_name"`
	UpdatedByName   *string  `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt       *string  `json:"created_at" db:"created_at"`
	UpdatedAt       *string  `json:"updated_at" db:"updated_at"`
	DeletedAt       *string  `json:"deleted_at" db:"deleted_at"`

	CustomerName *string `json:"customer_name" db:"customer_name"`
	CurrencyName *string `json:"currency_name" db:"currency_name"`
	BranchName   *string `json:"branch_name" db:"branch_name"`
	BankName     *string `json:"bank_name" db:"bank_name"`
}

type InvoiceAdjustmentDetailDTO struct {
	ID              uint     `json:"id" db:"id"`
	CustomerID      *uint    `json:"customer_id" db:"customer_id"`
	CurrencyID      *uint    `json:"currency_id" db:"currency_id"`
	BranchID        *uint    `json:"branch_id" db:"branch_id"`
	BankID          *uint    `json:"bank_id" db:"bank_id"`
	InvoiceNo       *string  `json:"invoice_no" db:"invoice_no"`
	PaymentDate     *string  `json:"payment_date" db:"payment_date"`
	PaymentAmount   *float64 `json:"payment_amount" db:"payment_amount"`
	ExchangeRate    *float64 `json:"exchange_rate" db:"exchange_rate"`
	Reference       *string  `json:"reference" db:"reference"`
	RefStartDate    *string  `json:"ref_start_date" db:"ref_start_date"`
	RefEndDate      *string  `json:"ref_end_date" db:"ref_end_date"`
	Remark          *string  `json:"remark" db:"remark"`
	RevNo           *int     `json:"rev_no" db:"rev_no"`
	TotalInvoice    *float64 `json:"total_invoice" db:"total_invoice"`
	TotalAdjustment *float64 `json:"total_adjustment" db:"total_adjustment"`
	TotalBalance    *float64 `json:"total_balance" db:"total_balance"`
	TotalAdminBank  *float64 `json:"total_admin_bank" db:"total_admin_bank"`
	GrandTotal      *float64 `json:"grand_total" db:"grand_total"`

	CreatedByID   *uint   `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint   `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint   `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeletedAt     *string `json:"deleted_at" db:"deleted_at"`

	CustomerName *string `json:"customer_name" db:"customer_name"`
	CurrencyName *string `json:"currency_name" db:"currency_name"`
	BranchName   *string `json:"branch_name" db:"branch_name"`
	BankName     *string `json:"bank_name" db:"bank_name"`

	AdjustmentDts []InvoiceAdjustmentDtListDTO `json:"adjustment_dts"`
}

type InvoiceAdjustmentDtListDTO struct {
	ID                  *uint            `json:"id" db:"id"`
	InvoiceUUID         *string          `json:"invoice_uuid" db:"invoice_uuid"`
	InvoiceAdjustmentID *uint            `json:"invoice_adjustment_id" db:"invoice_adjustment_id"`
	RefID               *uint            `json:"ref_id" db:"ref_id"`
	RefType             *string          `json:"ref_type" db:"ref_type"`
	RefJSON             *json.RawMessage `json:"ref_json" db:"ref_json"`
	InvoiceNo           *string          `json:"invoice_no" db:"invoice_no"`
	InvoiceDate         *string          `json:"invoice_date" db:"invoice_date"`
	InvoiceAmount       *float64         `json:"invoice_amount" db:"invoice_amount"`
	TotalAdjustment     *float64         `json:"total_adjustment" db:"total_adjustment"`
	BalanceAmount       *float64         `json:"balance_amount" db:"balance_amount"`
	AdjustmentAmount    *float64         `json:"adjustment_amount" db:"adjustment_amount"`
	AdminBank           *float64         `json:"admin_bank" db:"admin_bank"`
	TotalAmount         *float64         `json:"total_amount" db:"total_amount"`

	CreatedByID   *uint     `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint     `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint     `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName *string   `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string   `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *string   `json:"updated_at" db:"updated_at"`
	DeletedAt     *string   `json:"deleted_at" db:"deleted_at"`
}

type GetInvoiceAdjustmentsResult struct {
	InvoiceAdjustments []InvoiceAdjustmentListDTO
	Total              int
	Err                error
}

type GetReferenceInvoicesRequest struct {
	RefType        *string `json:"ref_type"` // "sales_invoice", "invoice_dp", "all"
	CustomerID     *int    `json:"customer_id"`
	RefStartDate   *string `json:"ref_start_date"`
	RefEndDate     *string `json:"ref_end_date"`
	PerPage        *string `json:"per_page" default:"10"`
	Page           *string `json:"page" default:"1"`
	OrderColumn    *string `json:"order_column" default:"id"`
	OrderDirection *string `json:"order_direction" default:"asc"`
}

type ReferenceInvoiceListDTO struct {
	ID              *uint    `json:"id" db:"id"`
	InvoiceUUID     *string  `json:"invoice_uuid" db:"invoice_uuid"`
	RefID           *uint    `json:"ref_id" db:"ref_id"`
	RefType         *string  `json:"ref_type" db:"ref_type"` // "sales_invoice" or "invoice_dp"
	CustomerID      *uint    `json:"customer_id" db:"customer_id"`
	CustomerName    *string  `json:"customer_name" db:"customer_name"`
	CurrencyID      *uint    `json:"currency_id" db:"currency_id"`
	CurrencyName    *string  `json:"currency_name" db:"currency_name"`
	BranchID        *uint    `json:"branch_id" db:"branch_id"`
	BranchName      *string  `json:"branch_name" db:"branch_name"`
	BankID          *uint    `json:"bank_id" db:"bank_id"`
	BankName        *string  `json:"bank_name" db:"bank_name"`
	InvoiceNo       *string  `json:"invoice_no" db:"invoice_no"`
	InvoiceDate     *string  `json:"invoice_date" db:"invoice_date"`
	RawInvoiceDate  *string  `json:"raw_invoice_date" db:"raw_invoice_date"`
	SortOrder       *string  `json:"sort_order" db:"sort_order"`
	InvoiceAmount   *float64 `json:"invoice_amount" db:"invoice_amount"`
	TotalAdjustment *float64 `json:"total_adjustment" db:"total_adjustment"`
	BalanceAmount   *float64 `json:"balance_amount" db:"balance_amount"`
	Remark          *string  `json:"remark" db:"remark"`
	Status          *string  `json:"status" db:"status"`
	CreatedAt       *string  `json:"created_at" db:"created_at"`
	UpdatedAt       *string  `json:"updated_at" db:"updated_at"`
}

type GetReferenceInvoicesResult struct {
	ReferenceInvoices []ReferenceInvoiceListDTO
	Total             int
	Err               error
}
