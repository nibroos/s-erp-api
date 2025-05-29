package dtos

import "time"

type GetInvoiceMaintenancesRequest struct {
	Global         *string `json:"global"`
	Title          *string `json:"title"`
	InvoiceNo      *string `json:"invoice_no"`
	Remark         *string `json:"remark"`
	Status         *string `json:"status"`
	RevNo          *int    `json:"rev_no"`
	CustomerID     *int    `json:"customer_id"`
	CurrencyID     *int    `json:"currency_id"`
	PaymentTermID  *int    `json:"payment_term_id"`
	VatID          *int    `json:"vat_id"`
	Pph23ID        *int    `json:"pph23_id"`
	BranchID       *int    `json:"branch_id"`
	BankID         *int    `json:"bank_id"`
	StartDate      *string `json:"start_date"`
	EndDate        *string `json:"end_date"`
	PerPage        *string `json:"per_page" default:"10"`
	Page           *string `json:"page" default:"1"`
	OrderColumn    *string `json:"order_column" default:"id"`
	OrderDirection *string `json:"order_direction" default:"asc"`
}

type CreateInvoiceMaintenanceDtRequest struct {
	ProductUuid  string   `json:"product_uuid"`
	ItemUnitID   *uint    `json:"item_unit_id"`
	VatID        *uint    `json:"vat_id"`
	Pph23ID      *uint    `json:"pph23_id"`
	RefID        *uint    `json:"ref_id"`
	RefDtID      *uint    `json:"ref_dt_id"`
	ProductID    *uint    `json:"product_id"`
	RefType      *string  `json:"ref_type"`
	ProductType  *string  `json:"product_type"`
	Remark       *string  `json:"remark"`
	IsVat        *uint    `json:"is_vat"`
	IsPph23      *uint    `json:"is_pph23"`
	Qty          *float64 `json:"qty"`
	Price        *float64 `json:"price"`
	Subtotal     *float64 `json:"subtotal"`
	Discount     *float64 `json:"discount"`
	TotalAmount  *float64 `json:"total_amount"`
	TotalDp      *float64 `json:"total_dp"`
	TotalBalance *float64 `json:"total_balance"`
}

type CreateInvoiceMaintenanceRequest struct {
	CustomerID               *uint                               `json:"customer_id"`
	CurrencyID               *uint                               `json:"currency_id"`
	PaymentTermID            *uint                               `json:"payment_term_id"`
	VatID                    *uint                               `json:"vat_id"`
	Pph23ID                  *uint                               `json:"pph23_id"`
	BranchID                 *uint                               `json:"branch_id"`
	BankID                   *uint                               `json:"bank_id"`
	Title                    *string                             `json:"title"`
	InvoiceNo                *string                             `json:"invoice_no"`
	InvoiceDate              *string                             `json:"invoice_date"`
	DueDate                  *string                             `json:"due_date"`
	ExchangeRate             *float64                            `json:"exchange_rate"`
	Remark                   *string                             `json:"remark"`
	Status                   *string                             `json:"status"`
	ApprovedStatus           *string                             `json:"approved_status"`
	RevNo                    *int                                `json:"rev_no"`
	Pph23Percentage          *float64                            `json:"pph23_percentage"`
	VatPercentage            *float64                            `json:"vat_percentage"`
	DiscountAmount           *float64                            `json:"discount_amount"`
	DiscountPercentage       *float64                            `json:"discount_percentage"`
	DiscountPercentageAmount *float64                            `json:"discount_percentage_amount"`
	DiscountFinal            *float64                            `json:"discount_final"`
	DiscountType             *string                             `json:"discount_type"`
	TotalAmountProducts      *float64                            `json:"total_amount_products"`
	TotalDpProducts          *float64                            `json:"total_dp_products"`
	TotalBalanceProducts     *float64                            `json:"total_balance_products"`
	Subtotal                 *float64                            `json:"subtotal"`
	TotalQty                 *float64                            `json:"total_qty"`
	TotalDiscount            *float64                            `json:"total_discount"`
	TotalPph23               *float64                            `json:"total_pph23"`
	TotalVat                 *float64                            `json:"total_vat"`
	GrandTotal               *float64                            `json:"grand_total"`
	InvoiceMaintenanceDts    []CreateInvoiceMaintenanceDtRequest `json:"invoice_maintenance_dts"`
}

type UpdateInvoiceMaintenanceDtRequest struct {
	ID                     *uint    `json:"id"`
	InvoiceMaintenanceDtID *uint    `json:"invoice_maintenance_dt_id"`
	ProductUuid            string   `json:"product_uuid"`
	InvoiceMaintenanceID   *uint    `json:"invoice_maintenance_id"`
	ItemUnitID             *uint    `json:"item_unit_id"`
	VatID                  *uint    `json:"vat_id"`
	Pph23ID                *uint    `json:"pph23_id"`
	RefID                  *uint    `json:"ref_id"`
	RefDtID                *uint    `json:"ref_dt_id"`
	ProductID              *uint    `json:"product_id"`
	RefType                *string  `json:"ref_type"`
	ProductType            *string  `json:"product_type"`
	Remark                 *string  `json:"remark"`
	IsVat                  *uint    `json:"is_vat"`
	IsPph23                *uint    `json:"is_pph23"`
	Qty                    *float64 `json:"qty"`
	Price                  *float64 `json:"price"`
	Subtotal               *float64 `json:"subtotal"`
	Discount               *float64 `json:"discount"`
	TotalAmount            *float64 `json:"total_amount"`
	TotalDp                *float64 `json:"total_dp"`
	TotalBalance           *float64 `json:"total_balance"`
}

type UpdateInvoiceMaintenanceRequest struct {
	ID                       uint                                `json:"id"`
	InvoiceMaintenanceID     *uint                               `json:"invoice_maintenance_id"`
	CustomerID               *uint                               `json:"customer_id"`
	CurrencyID               *uint                               `json:"currency_id"`
	PaymentTermID            *uint                               `json:"payment_term_id"`
	VatID                    *uint                               `json:"vat_id"`
	Pph23ID                  *uint                               `json:"pph23_id"`
	BranchID                 *uint                               `json:"branch_id"`
	BankID                   *uint                               `json:"bank_id"`
	Title                    *string                             `json:"title"`
	InvoiceNo                *string                             `json:"invoice_no"`
	InvoiceDate              *string                             `json:"invoice_date"`
	DueDate                  *string                             `json:"due_date"`
	ExchangeRate             *float64                            `json:"exchange_rate"`
	Remark                   *string                             `json:"remark"`
	Status                   *string                             `json:"status"`
	ApprovedStatus           *string                             `json:"approved_status"`
	RevNo                    *int                                `json:"rev_no"`
	Pph23Percentage          *float64                            `json:"pph23_percentage"`
	VatPercentage            *float64                            `json:"vat_percentage"`
	DiscountAmount           *float64                            `json:"discount_amount"`
	DiscountPercentage       *float64                            `json:"discount_percentage"`
	DiscountPercentageAmount *float64                            `json:"discount_percentage_amount"`
	DiscountFinal            *float64                            `json:"discount_final"`
	DiscountType             *string                             `json:"discount_type"`
	TotalAmountProducts      *float64                            `json:"total_amount_products"`
	TotalDpProducts          *float64                            `json:"total_dp_products"`
	TotalBalanceProducts     *float64                            `json:"total_balance_products"`
	Subtotal                 *float64                            `json:"subtotal"`
	TotalQty                 *float64                            `json:"total_qty"`
	TotalDiscount            *float64                            `json:"total_discount"`
	TotalPph23               *float64                            `json:"total_pph23"`
	TotalVat                 *float64                            `json:"total_vat"`
	GrandTotal               *float64                            `json:"grand_total"`
	InvoiceMaintenanceDts    []UpdateInvoiceMaintenanceDtRequest `json:"invoice_maintenance_dts"`
}

type GetInvoiceMaintenanceByIDRequest struct {
	ID uint `json:"id"`
}

type GetInvoiceMaintenanceParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetInvoiceMaintenanceParams(id uint) *GetInvoiceMaintenanceParams {
	defaultIsDeleted := 0
	return &GetInvoiceMaintenanceParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type GetInvoiceMaintenanceDtParams struct {
	ID                   uint
	InvoiceMaintenanceID uint
	IsDeleted            *int
}

type DeleteInvoiceMaintenanceRequest struct {
	ID uint `json:"id"`
}

type InvoiceMaintenanceListDTO struct {
	ID                       uint     `json:"id" db:"id"`
	CustomerID               *uint    `json:"customer_id" db:"customer_id"`
	CurrencyID               *uint    `json:"currency_id" db:"currency_id"`
	PaymentTermID            *uint    `json:"payment_term_id" db:"payment_term_id"`
	VatID                    *uint    `json:"vat_id" db:"vat_id"`
	Pph23ID                  *uint    `json:"pph23_id" db:"pph23_id"`
	BranchID                 *uint    `json:"branch_id" db:"branch_id"`
	BankID                   *uint    `json:"bank_id" db:"bank_id"`
	BankName                 *string  `json:"bank_name" db:"bank_name"`
	AccountNumber            *string  `json:"account_number" db:"account_number"`
	AccountName              *string  `json:"account_name" db:"account_name"`
	Title                    *string  `json:"title" db:"title"`
	InvoiceNo                *string  `json:"invoice_no" db:"invoice_no"`
	InvoiceDate              *string  `json:"invoice_date" db:"invoice_date"`
	DueDate                  *string  `json:"due_date" db:"due_date"`
	Remark                   *string  `json:"remark" db:"remark"`
	Status                   *string  `json:"status" db:"status"`
	ApprovedStatus           *string  `json:"approved_status" db:"approved_status"`
	ApprovedByID             *string  `json:"approved_by_id" db:"approved_by_id"`
	RevNo                    *int     `json:"rev_no" db:"rev_no"`
	ExchangeRate             *float64 `json:"exchange_rate" db:"exchange_rate"`
	VatPercentage            *float64 `json:"vat_percentage" db:"vat_percentage"`
	Pph23Percentage          *float64 `json:"pph23_percentage" db:"pph23_percentage"`
	DiscountAmount           *float64 `json:"discount_amount" db:"discount_amount"`
	DiscountPercentage       *float64 `json:"discount_percentage" db:"discount_percentage"`
	DiscountPercentageAmount *float64 `json:"discount_percentage_amount" db:"discount_percentage_amount"`
	DiscountFinal            *float64 `json:"discount_final" db:"discount_final"`
	DiscountType             *string  `json:"discount_type" db:"discount_type"`
	TotalAmountProducts      *float64 `json:"total_amount_products" db:"total_amount_products"`
	TotalDpProducts          *float64 `json:"total_dp_products" db:"total_dp_products"`
	TotalBalanceProducts     *float64 `json:"total_balance_products" db:"total_balance_products"`
	TotalQty                 *float64 `json:"total_qty" db:"total_qty"`
	Subtotal                 *float64 `json:"subtotal" db:"subtotal"`
	TotalDiscount            *float64 `json:"total_discount" db:"total_discount"`
	TotalPph23               *float64 `json:"total_pph23" db:"total_pph23"`
	TotalVat                 *float64 `json:"total_vat" db:"total_vat"`
	GrandTotal               *float64 `json:"grand_total" db:"grand_total"`
	TotalAdjustment          *float64 `json:"total_adjustment" db:"total_adjustment"`
	CreatedByID              *uint    `json:"created_by_id" db:"created_by_id"`
	UpdatedByID              *uint    `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID              *uint    `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName            *string  `json:"created_by_name" db:"created_by_name"`
	UpdatedByName            *string  `json:"updated_by_name" db:"updated_by_name"`
	ApprovedByName           *string  `json:"approved_by_name" db:"approved_by_name"`
	CreatedAt                *string  `json:"created_at" db:"created_at"`
	UpdatedAt                *string  `json:"updated_at" db:"updated_at"`
	DeleteAt                 *string  `json:"deleted_at" db:"deleted_at"`

	CurrencyName               *string `json:"currency_name" db:"currency_name"`
	CustomerName               *string `json:"customer_name" db:"customer_name"`
	CustomerEmail              *string `json:"customer_email" db:"customer_email"`
	CustomerAddress            *string `json:"customer_address" db:"customer_address"`
	PaymentTermName            *string `json:"payment_term_name" db:"payment_term_name"`
	VatName                    *string `json:"vat_name" db:"vat_name"`
	Pph23Name                  *string `json:"pph23_name" db:"pph23_name"`
	BranchName                 *string `json:"branch_name" db:"branch_name"`
	InvoiceMaintenanceDtRemark *string `json:"invoice_maintenance_dt_remark" db:"invoice_maintenance_dt_remark"`
	DaysRemaining              *int    `json:"days_remaining" db:"days_remaining"`
	StatusExpired              *string `json:"status_expired" db:"status_expired"`

	// InvoiceMaintenanceDts []*InvoiceMaintenanceDtListDTO `json:"invoice_maintenance_dts,omitempty" db:"invoice_maintenance_dts" gorm:"-"`
	// InvoiceMaintenanceDts []*InvoiceMaintenanceDtListDTO `json:"invoice_maintenance_dts,omitempty"`
	// InvoiceMaintenanceDts *[]*InvoiceMaintenanceDtListDTO `json:"invoice_maintenance_dts,omitempty" db:"invoice_maintenance_dts" gorm:"foreignKey:InvoiceMaintenanceID"`
	// InvoiceMaintenanceDts interface{} `json:"invoice_maintenance_dts" db:"invoice_maintenance_dts"`
}

type InvoiceMaintenanceDetailDTO struct {
	ID                       uint     `json:"id" db:"id"`
	CustomerID               *uint    `json:"customer_id" db:"customer_id"`
	CurrencyID               *uint    `json:"currency_id" db:"currency_id"`
	PaymentTermID            *uint    `json:"payment_term_id" db:"payment_term_id"`
	VatID                    *uint    `json:"vat_id" db:"vat_id"`
	Pph23ID                  *uint    `json:"pph23_id" db:"pph23_id"`
	BranchID                 *uint    `json:"branch_id" db:"branch_id"`
	BankID                   *uint    `json:"bank_id" db:"bank_id"`
	Title                    *string  `json:"title" db:"title"`
	InvoiceNo                *string  `json:"invoice_no" db:"invoice_no"`
	InvoiceDate              *string  `json:"invoice_date" db:"invoice_date"`
	DueDate                  *string  `json:"due_date" db:"due_date"`
	Remark                   *string  `json:"remark" db:"remark"`
	Status                   *string  `json:"status" db:"status"`
	ApprovedStatus           *string  `json:"approved_status" db:"approved_status"`
	RevNo                    *int     `json:"rev_no" db:"rev_no"`
	ExchangeRate             float64  `json:"exchange_rate" db:"exchange_rate"`
	VatPercentage            *float64 `json:"vat_percentage" db:"vat_percentage"`
	Pph23Percentage          *float64 `json:"pph23_percentage" db:"pph23_percentage"`
	DiscountAmount           *float64 `json:"discount_amount" db:"discount_amount"`
	DiscountPercentage       *float64 `json:"discount_percentage" db:"discount_percentage"`
	DiscountPercentageAmount *float64 `json:"discount_percentage_amount" db:"discount_percentage_amount"`
	DiscountFinal            *float64 `json:"discount_final" db:"discount_final"`
	DiscountType             *string  `json:"discount_type" db:"discount_type"`
	TotalAmountProducts      *float64 `json:"total_amount_products" db:"total_amount_products"`
	TotalDpProducts          *float64 `json:"total_dp_products" db:"total_dp_products"`
	TotalBalanceProducts     *float64 `json:"total_balance_products" db:"total_balance_products"`
	TotalQty                 float64  `json:"total_qty" db:"total_qty"`
	Subtotal                 float64  `json:"subtotal" db:"subtotal"`
	TotalDiscount            float64  `json:"total_discount" db:"total_discount"`
	TotalPph23               float64  `json:"total_pph23" db:"total_pph23"`
	TotalVat                 float64  `json:"total_vat" db:"total_vat"`
	GrandTotal               float64  `json:"grand_total" db:"grand_total"`

	CreatedByID           *uint                         `json:"created_by_id" db:"created_by_id"`
	UpdatedByID           *uint                         `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID           *uint                         `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName         *string                       `json:"created_by_name" db:"created_by_name"`
	UpdatedByName         *string                       `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt             *string                       `json:"created_at" db:"created_at"`
	UpdatedAt             *string                       `json:"updated_at" db:"updated_at"`
	DeleteAt              *string                       `json:"deleted_at" db:"deleted_at"`
	InvoiceMaintenanceDts []InvoiceMaintenanceDtListDTO `json:"invoice_maintenance_dts"`

	CompanyProfileID *uint                   `json:"company_profile_id" db:"company_profile_id"`
	CustomerCode     *string                 `json:"customer_code" db:"customer_code"`
	CustomerName     *string                 `json:"customer_name" db:"customer_name"`
	Phone            *string                 `json:"phone" db:"phone"`
	Pic              *string                 `json:"pic" db:"pic"`
	Address          *string                 `json:"address" db:"address"`
	Company          CompanyProfileDetailDTO `json:"company"`
	OrderTypeName    *string                 `json:"order_type_name" db:"order_type_name"`
	CurrencyName     *string                 `json:"currency_name" db:"currency_name"`
	VatName          *string                 `json:"vat_name" db:"vat_name"`
	Pph23Name        *string                 `json:"pph23_name" db:"pph23_name"`
	BankName         *string                 `json:"bank_name" db:"bank_name"`
	AccountName      *string                 `json:"account_name" db:"account_name"`
	TotalAfterDisc   float64                 `json:"total_after_disc" db:"total_after_disc"`
	IsIDOnly         *int                    `json:"is_id_only" db:"is_id_only"`
}

type InvoiceMaintenanceDetailNoBomDTO struct {
	ID                       uint     `json:"id" db:"id"`
	CustomerID               *uint    `json:"customer_id" db:"customer_id"`
	CurrencyID               *uint    `json:"currency_id" db:"currency_id"`
	PaymentTermID            *uint    `json:"payment_term_id" db:"payment_term_id"`
	VatID                    *uint    `json:"vat_id" db:"vat_id"`
	Pph23ID                  *uint    `json:"pph23_id" db:"pph23_id"`
	BranchID                 *uint    `json:"branch_id" db:"branch_id"`
	BankID                   *uint    `json:"bank_id" db:"bank_id"`
	Title                    *string  `json:"title" db:"title"`
	InvoiceNo                *string  `json:"invoice_no" db:"invoice_no"`
	InvoiceDate              *string  `json:"invoice_date" db:"invoice_date"`
	DueDate                  *string  `json:"due_date" db:"due_date"`
	Remark                   *string  `json:"remark" db:"remark"`
	Status                   *string  `json:"status" db:"status"`
	ApprovedStatus           *string  `json:"approved_status" db:"approved_status"`
	RevNo                    *int     `json:"rev_no" db:"rev_no"`
	ExchangeRate             float64  `json:"exchange_rate" db:"exchange_rate"`
	VatPercentage            *float64 `json:"vat_percentage" db:"vat_percentage"`
	Pph23Percentage          *float64 `json:"pph23_percentage" db:"pph23_percentage"`
	DiscountAmount           *float64 `json:"discount_amount" db:"discount_amount"`
	DiscountPercentage       *float64 `json:"discount_percentage" db:"discount_percentage"`
	DiscountPercentageAmount *float64 `json:"discount_percentage_amount" db:"discount_percentage_amount"`
	DiscountFinal            *float64 `json:"discount_final" db:"discount_final"`
	DiscountType             *string  `json:"discount_type" db:"discount_type"`
	TotalAmountProducts      *float64 `json:"total_amount_products" db:"total_amount_products"`
	TotalDpProducts          *float64 `json:"total_dp_products" db:"total_dp_products"`
	TotalBalanceProducts     *float64 `json:"total_balance_products" db:"total_balance_products"`
	TotalQty                 float64  `json:"total_qty" db:"total_qty"`
	Subtotal                 float64  `json:"subtotal" db:"subtotal"`
	TotalDiscount            float64  `json:"total_discount" db:"total_discount"`
	TotalPph23               float64  `json:"total_pph23" db:"total_pph23"`
	TotalVat                 float64  `json:"total_vat" db:"total_vat"`
	GrandTotal               float64  `json:"grand_total" db:"grand_total"`

	CreatedByID           *uint                              `json:"created_by_id" db:"created_by_id"`
	UpdatedByID           *uint                              `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID           *uint                              `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName         *string                            `json:"created_by_name" db:"created_by_name"`
	UpdatedByName         *string                            `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt             *string                            `json:"created_at" db:"created_at"`
	UpdatedAt             *string                            `json:"updated_at" db:"updated_at"`
	DeleteAt              *string                            `json:"deleted_at" db:"deleted_at"`
	InvoiceMaintenanceDts []InvoiceMaintenanceDtListNoBomDTO `json:"invoice_maintenance_dts"`

	CompanyProfileID *uint                   `json:"company_profile_id" db:"company_profile_id"`
	CustomerCode     *string                 `json:"customer_code" db:"customer_code"`
	CustomerName     *string                 `json:"customer_name" db:"customer_name"`
	Phone            *string                 `json:"phone" db:"phone"`
	Pic              *string                 `json:"pic" db:"pic"`
	Address          *string                 `json:"address" db:"address"`
	Company          CompanyProfileDetailDTO `json:"company"`
	OrderTypeName    *string                 `json:"order_type_name" db:"order_type_name"`
	CurrencyName     *string                 `json:"currency_name" db:"currency_name"`
	VatName          *string                 `json:"vat_name" db:"vat_name"`
	Pph23Name        *string                 `json:"pph23_name" db:"pph23_name"`
	BankName         *string                 `json:"bank_name" db:"bank_name"`
	AccountName      *string                 `json:"account_name" db:"account_name"`
	TotalAfterDisc   float64                 `json:"total_after_disc" db:"total_after_disc"`
	IsIDOnly         *int                    `json:"is_id_only" db:"is_id_only"`
}

type InvoiceMaintenanceDtListDTO struct {
	ID                     *uint    `json:"id" db:"id"`
	InvoiceMaintenanceDtID *uint    `json:"invoice_maintenance_dt_id" db:"invoice_maintenance_dt_id"`
	ProductUuid            *string  `json:"product_uuid" db:"product_uuid"`
	CustomerID             *uint    `json:"customer_id" db:"customer_id"`
	InvoiceMaintenanceID   *uint    `json:"invoice_maintenance_id" db:"invoice_maintenance_id"`
	ItemUnitID             *uint    `json:"item_unit_id" db:"item_unit_id"`
	VatID                  *uint    `json:"vat_id" db:"vat_id"`
	Pph23ID                *uint    `json:"pph23_id" db:"pph23_id"`
	RefID                  *uint    `json:"ref_id" db:"ref_id"`
	RefDtID                *uint    `json:"ref_dt_id" db:"ref_dt_id"`
	ProductID              *uint    `json:"product_id" db:"product_id"`
	ItemName               *string  `json:"item_name" db:"item_name"`
	ItemCode               *string  `json:"item_code" db:"item_code"`
	UnitName               *string  `json:"unit_name" db:"unit_name"`
	RefJSON                *string  `json:"ref_json" db:"ref_json"`
	RefType                *string  `json:"ref_type" db:"ref_type"`
	ProductType            *string  `json:"product_type" db:"product_type"`
	Remark                 *string  `json:"remark" db:"remark"`
	IsVat                  *uint    `json:"is_vat" db:"is_vat"`
	IsPph23                *uint    `json:"is_pph23" db:"is_pph23"`
	VatName                *string  `json:"vat_name" db:"vat_name"`
	Pph23Name              *string  `json:"pph23_name" db:"pph23_name"`
	Qty                    *float64 `json:"qty" db:"qty"`
	Price                  *float64 `json:"price" db:"price"`
	Subtotal               *float64 `json:"subtotal" db:"subtotal"`
	Discount               *float64 `json:"discount" db:"discount"`
	TotalAmount            *float64 `json:"total_amount" db:"total_amount"`
	TotalDp                *float64 `json:"total_dp" db:"total_dp"`
	TotalBalance           *float64 `json:"total_balance" db:"total_balance"`

	CreatedByName *string   `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string   `json:"updated_by_name" db:"updated_by_name"`
	CreatedByID   *uint     `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint     `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint     `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *string   `json:"updated_at" db:"updated_at"`
	DeleteAt      *string   `json:"deleted_at" db:"deleted_at"`

	RefNum *string `json:"ref_num" db:"ref_num"`

	SoDtsBoms []SalesOrderSoDtBomListDTO `json:"so_dts_boms,omitempty"`
	ItemID    *uint                      `json:"item_id" db:"item_id"`
}

type InvoiceMaintenanceDtListNoBomDTO struct {
	ID                     *uint    `json:"id" db:"id"`
	InvoiceMaintenanceDtID *uint    `json:"invoice_maintenance_dt_id" db:"invoice_maintenance_dt_id"`
	ProductUuid            *string  `json:"product_uuid" db:"product_uuid"`
	CustomerID             *uint    `json:"customer_id" db:"customer_id"`
	InvoiceMaintenanceID   *uint    `json:"invoice_maintenance_id" db:"invoice_maintenance_id"`
	ItemUnitID             *uint    `json:"item_unit_id" db:"item_unit_id"`
	VatID                  *uint    `json:"vat_id" db:"vat_id"`
	Pph23ID                *uint    `json:"pph23_id" db:"pph23_id"`
	RefID                  *uint    `json:"ref_id" db:"ref_id"`
	RefDtID                *uint    `json:"ref_dt_id" db:"ref_dt_id"`
	ProductID              *uint    `json:"product_id" db:"product_id"`
	ItemName               *string  `json:"item_name" db:"item_name"`
	ItemCode               *string  `json:"item_code" db:"item_code"`
	UnitName               *string  `json:"unit_name" db:"unit_name"`
	RefJSON                *string  `json:"ref_json" db:"ref_json"`
	RefType                *string  `json:"ref_type" db:"ref_type"`
	ProductType            *string  `json:"product_type" db:"product_type"`
	Remark                 *string  `json:"remark" db:"remark"`
	IsVat                  *uint    `json:"is_vat" db:"is_vat"`
	IsPph23                *uint    `json:"is_pph23" db:"is_pph23"`
	VatName                *string  `json:"vat_name" db:"vat_name"`
	Pph23Name              *string  `json:"pph23_name" db:"pph23_name"`
	Qty                    *float64 `json:"qty" db:"qty"`
	Price                  *float64 `json:"price" db:"price"`
	Subtotal               *float64 `json:"subtotal" db:"subtotal"`
	Discount               *float64 `json:"discount" db:"discount"`
	TotalAmount            *float64 `json:"total_amount" db:"total_amount"`
	TotalDp                *float64 `json:"total_dp" db:"total_dp"`
	TotalBalance           *float64 `json:"total_balance" db:"total_balance"`

	CreatedByName *string   `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string   `json:"updated_by_name" db:"updated_by_name"`
	CreatedByID   *uint     `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint     `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint     `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *string   `json:"updated_at" db:"updated_at"`
	DeleteAt      *string   `json:"deleted_at" db:"deleted_at"`

	RefNum *string `json:"ref_num" db:"ref_num"`
}

type InvoiceMaintenanceDtListUpdateDTO struct {
	ID                     *uint    `json:"id" db:"id"`
	InvoiceMaintenanceDtID *uint    `json:"invoice_maintenance_dt_id" db:"invoice_maintenance_dt_id"`
	ProductUuid            *string  `json:"product_uuid" db:"product_uuid"`
	InvoiceMaintenanceID   *uint    `json:"invoice_maintenance_id" db:"invoice_maintenance_id"`
	ItemUnitID             *uint    `json:"item_unit_id" db:"item_unit_id"`
	VatID                  *uint    `json:"vat_id" db:"vat_id"`
	Pph23ID                *uint    `json:"pph23_id" db:"pph23_id"`
	RefID                  *uint    `json:"ref_id" db:"ref_id"`
	RefDtID                *uint    `json:"ref_dt_id"`
	ProductID              *uint    `json:"product_id" db:"product_id"`
	ItemName               *string  `json:"item_name" db:"item_name"`
	ItemCode               *string  `json:"item_code" db:"item_code"`
	UnitName               *string  `json:"unit_name" db:"unit_name"`
	RefJSON                *string  `json:"ref_json" db:"ref_json"`
	RefType                *string  `json:"ref_type" db:"ref_type"`
	ProductType            *string  `json:"product_type" db:"product_type"`
	Remark                 *string  `json:"remark" db:"remark"`
	IsVat                  *uint    `json:"is_vat" db:"is_vat"`
	IsPph23                *uint    `json:"is_pph23" db:"is_pph23"`
	Qty                    *float64 `json:"qty" db:"qty"`
	Price                  *float64 `json:"price" db:"price"`
	Subtotal               *float64 `json:"subtotal" db:"subtotal"`
	DiscountFinal          *float64 `json:"discount" db:"discount"`
	TotalAmount            *float64 `json:"total_amount" db:"total_amount"`
	TotalDp                *float64 `json:"total_dp" db:"total_dp"`
	TotalBalance           *float64 `json:"total_balance" db:"total_balance"`

	CreatedByName *string   `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string   `json:"updated_by_name" db:"updated_by_name"`
	CreatedByID   *uint     `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint     `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint     `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *string   `json:"updated_at" db:"updated_at"`
	DeleteAt      *string   `json:"deleted_at" db:"deleted_at"`
}

type GetInvoiceMaintenancesResult struct {
	InvoiceMaintenances []InvoiceMaintenanceListDTO
	Total               int
	Err                 error
}

type GetRefSalesOrderForInvoiceMaintenanceRequest struct {
	Global         *string `json:"global"`
	InvoiceID      *string `json:"invoice_id"`
	SalesOrderNo   *string `json:"sales_order_no"`
	PoBuyerNo      *string `json:"po_buyer_no"`
	Remark         *string `json:"remark"`
	CustomerID     *int    `json:"customer_id"`
	OrderTypeID    *int    `json:"order_type_id"`
	CurrencyID     *int    `json:"currency_id"`
	VatID          *int    `json:"vat_id"`
	PaymentID      *int    `json:"payment_id"`
	Pph23ID        *int    `json:"pph23_id"`
	BranchID       *int    `json:"branch_id"`
	Status         *string `json:"status"`
	DateType       *string `json:"date_type"`
	StartDate      *string `json:"start_date"`
	EndDate        *string `json:"end_date"`
	SpecificIDs    *string `json:"specific_ids"`
	PerPage        *string `json:"per_page" default:"10"`
	Page           *string `json:"page" default:"1"`
	OrderColumn    *string `json:"order_column" default:"id"`
	OrderDirection *string `json:"order_direction" default:"asc"`
}

type RefSalesOrderForInvoiceMaintenanceListDTO struct {
	ID              *uint     `json:"id" db:"id"`
	SoDtID          *uint     `json:"so_dt_id" db:"so_dt_id"`
	ProductUuid     *string   `json:"product_uuid" db:"product_uuid"`
	SalesOrderID    *uint     `json:"sales_order_id" db:"sales_order_id"`
	ItemUnitID      *uint     `json:"item_unit_id" db:"item_unit_id"`
	VatID           *uint     `json:"vat_id" db:"vat_id"`
	Pph23ID         *uint     `json:"pph23_id" db:"pph23_id"`
	RefID           *uint     `json:"ref_id" db:"ref_id"`
	ItemID          *uint     `json:"item_id" db:"item_id"`
	ItemName        *string   `json:"item_name" db:"item_name"`
	ItemCode        *string   `json:"item_code" db:"item_code"`
	UnitName        *string   `json:"unit_name" db:"unit_name"`
	RefJSON         *string   `json:"ref_json" db:"ref_json"`
	RefType         *string   `json:"ref_type" db:"ref_type"`
	ItemType        *string   `json:"item_type" db:"item_type"`
	GenCode         *string   `json:"gen_code" db:"gen_code"`
	Remark          *string   `json:"remark" db:"remark"`
	VatPerc         *float64  `json:"vat_perc" db:"vat_perc"`
	VatPercAm       *float64  `json:"vat_perc_am" db:"vat_perc_am"`
	VatName         *string   `json:"vat_name" db:"vat_name"`
	Pph23Name       *string   `json:"pph23_name" db:"pph23_name"`
	Pph23Perc       *float64  `json:"pph23_perc" db:"pph23_perc"`
	Pph23PercAm     *float64  `json:"pph23_perc_am" db:"pph23_perc_am"`
	DiscAm          *float64  `json:"disc_am" db:"disc_am"`
	DiscPerc        *float64  `json:"disc_perc" db:"disc_perc"`
	DiscPercNum     *float64  `json:"disc_perc_num" db:"disc_perc_num"`
	DiscPercAm      *float64  `json:"disc_perc_am" db:"disc_perc_am"`
	DiscType        *string   `json:"disc_type" db:"disc_type"`
	MarkupPerc      *float64  `json:"markup_perc" db:"markup_perc"`
	MarkupPercAm    *float64  `json:"markup_perc_am" db:"markup_perc_am"`
	IsLockMarkup    *int8     `json:"is_lock_markup" db:"is_lock_markup"`
	IsLockPriceSell *int8     `json:"is_lock_price_sell" db:"is_lock_price_sell"`
	IsVat           *int8     `json:"is_vat" db:"is_vat"`
	IsPph23         *int8     `json:"is_pph23" db:"is_pph23"`
	Qty             *float64  `json:"qty" db:"qty"`
	QtyOut          *float64  `json:"qty_out" db:"qty_out"`
	PriceBuy        *float64  `json:"price_buy" db:"price_buy"`
	PriceSell       *float64  `json:"price_sell" db:"price_sell"`
	SubtotalBuy     *float64  `json:"subtotal_buy" db:"subtotal_buy"`
	SubtotalSell    *float64  `json:"subtotal_sell" db:"subtotal_sell"`
	DiscFinal       *float64  `json:"disc_final" db:"disc_final"`
	TotalAm         *float64  `json:"total_am" db:"total_am"`
	TotalDp         *float64  `json:"total_dp" db:"total_dp"`
	TotalBalance    *float64  `json:"total_balance" db:"total_balance"`
	InvoiceStatus   *string   `json:"invoice_status" db:"invoice_status"`
	CreatedByName   *string   `json:"created_by_name" db:"created_by_name"`
	UpdatedByName   *string   `json:"updated_by_name" db:"updated_by_name"`
	CreatedByID     *uint     `json:"created_by_id" db:"created_by_id"`
	UpdatedByID     *uint     `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID     *uint     `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       *string   `json:"updated_at" db:"updated_at"`
	DeleteAt        *string   `json:"deleted_at" db:"deleted_at"`

	CustomerID       *uint    `json:"customer_id" db:"customer_id"`
	OrderTypeID      *uint    `json:"order_type_id" db:"order_type_id"`
	CurrencyID       *uint    `json:"currency_id" db:"currency_id"`
	HeadVatID        *uint    `json:"head_vat_id" db:"head_vat_id"`
	HeadPph23ID      *uint    `json:"head_pph23_id" db:"head_pph23_id"`
	HeadVatPerc      *float64 `json:"head_vat_perc" db:"head_vat_perc"`
	HeadPph23Perc    *float64 `json:"head_pph23_perc" db:"head_pph23_perc"`
	HeadDiscAm       *float64 `json:"head_disc_am" db:"head_disc_am"`
	HeadDiscPerc     *float64 `json:"head_disc_perc" db:"head_disc_perc"`
	HeadMarkupPerc   *float64 `json:"head_markup_perc" db:"head_markup_perc"`
	HeadMarkupPercAm *float64 `json:"head_markup_perc_am" db:"head_markup_perc_am"`
	HeadRemark       *string  `json:"head_remark" db:"head_remark"`
	ExchangeRate     *float64 `json:"exchange_rate" db:"exchange_rate"`
	SalesOrderNo     *string  `json:"sales_order_no" db:"sales_order_no"`
	PoBuyerNo        *string  `json:"po_buyer_no" db:"po_buyer_no"`
	CustomerName     *string  `json:"customer_name" db:"customer_name"`
	OrderTypeName    *string  `json:"order_type_name" db:"order_type_name"`
	OrderDate        *string  `json:"order_date" db:"order_date"`
	ShippingDate     *string  `json:"shipping_date" db:"shipping_date"`
	ItemSku          *string  `json:"item_sku" db:"item_sku"`
	AgreeAt          *string  `json:"agree_at" db:"agree_at"`
	DueAt            *string  `json:"due_at" db:"due_at"`
	PaymentID        *uint    `json:"payment_id" db:"payment_id"`

	SoDtsBoms []SalesOrderSoDtBomListDTO `json:"so_dts_boms"`
}

type UpdateSalesOrderStatusForInvoiceMaintenanceRequest struct {
	ID     uint   `json:"id"`
	Status string `json:"status"`
}

type ApproveInvoiceMaintenancesRequest struct {
	IDs []uint `json:"ids" validate:"required,min=1"`
}

type CancelApproveInvoiceMaintenancesRequest struct {
	IDs []uint `json:"ids" validate:"required,min=1"`
}

type BulkSendEmailApprovedInvoiceMaintenancesRequest struct {
	IDs        []uint                 `json:"ids" validate:"required,min=1"`
	SenderID   uint                   `json:"sender_id"`
	SentEmails []FormSentEmailRequest `json:"sent_emails"`
}

type InvoiceMaintenanceStatusWidget struct {
	Status     string  `json:"status" db:"status"`
	OrderCount int     `json:"order_count" db:"order_count"`
	TotalQty   float64 `json:"total_qty" db:"total_qty"`
	GrandTotal float64 `json:"grand_total" db:"grand_total"`
}

type RepeatInvoiceMaintenanceItem struct {
	ID          uint    `json:"id" validate:"required"`
	Title       *string `json:"title"`
	InvoiceDate *string `json:"invoice_date"`
	DueDate     *string `json:"due_date"`
	Remark      *string `json:"remark"`
}

type RepeatInvoiceMaintenanceRequest struct {
	Invoices []RepeatInvoiceMaintenanceItem `json:"invoices" validate:"required,min=1"`
}

type RepeatInvoiceMaintenanceResult struct {
	OriginalID  uint `json:"original_id"`
	DuplicateID uint `json:"duplicate_id"`
}

type RepeatInvoiceMaintenanceResponse struct {
	Results []RepeatInvoiceMaintenanceResult `json:"results"`
}

type BulkSendEmailApprovedInvoiceMaintenancesEmailData struct {
	Req          InvoiceMaintenanceListDTO          `json:"req"`
	Dts          []InvoiceMaintenanceDtListNoBomDTO `json:"dts"`
	CustomerName string                             `json:"customer_name"`
	Message      string                             `json:"message"`
	Subject      string                             `json:"subject"`
	ButtonURL    string                             `json:"button_url"`
	ButtonText   string                             `json:"button_string"`
	SentAt       string                             `json:"sent_at"`
	Company      *CompanyProfileDetailDTO           `json:"company"`

	Attachments []BulkSendEmailApprovedInvoiceMaintenancesEmailAttachment `json:"attachments"`
}

type BulkSendEmailApprovedInvoiceMaintenancesEmailAttachment struct {
	Label string `json:"label"`
	Path  string `json:"path"`
}

type InvoiceMaintenancePDFData struct {
	Num string `json:"num"`
	// Form FormInvoiceMaintenanceRequest
	Form   InvoiceMaintenanceDetailNoBomDTO
	QrCode string `json:"qr_code"`
}
