package dtos

import "time"

type GetInvoiceDpsRequest struct {
	Global         *string `json:"global"`
	InvoiceNo      *string `json:"invoice_no"`
	Remark         *string `json:"remark"`
	Status         *string `json:"status"`
	CustomerID     *int    `json:"customer_id"`
	CurrencyID     *int    `json:"currency_id"`
	PaymentTermID  *int    `json:"payment_term_id"`
	VatID          *int    `json:"vat_id"`
	Pph23ID        *int    `json:"pph23_id"`
	BranchID       *int    `json:"branch_id"`
	StartDate      *string `json:"start_date"`
	EndDate        *string `json:"end_date"`
	PerPage        *string `json:"per_page" default:"10"`
	Page           *string `json:"page" default:"1"`
	OrderColumn    *string `json:"order_column" default:"id"`
	OrderDirection *string `json:"order_direction" default:"asc"`
}

type CreateInvoiceDpDtRequest struct {
	ProductUuid              string   `json:"product_uuid"`
	ItemUnitID               *uint    `json:"item_unit_id"`
	VatID                    *uint    `json:"vat_id"`
	Pph23ID                  *uint    `json:"pph23_id"`
	RefID                    *uint    `json:"ref_id"`
	RefDtID                  *uint    `json:"ref_dt_id"`
	ProductID                *uint    `json:"product_id"`
	RefType                  *string  `json:"ref_type"`
	ProductType              *string  `json:"product_type"`
	Remark                   *string  `json:"remark"`
	DpPercentage             *float64 `json:"dp_percentage"`
	IsVat                    *uint    `json:"is_vat"`
	IsPph23                  *uint    `json:"is_pph23"`
	Qty                      *float64 `json:"qty"`
	Price                    *float64 `json:"price"`
	Subtotal                 *float64 `json:"subtotal"`
	DiscountAmount           *float64 `json:"discount_amount"`
	DiscountPercentage       *float64 `json:"discount_percentage"`
	DiscountPercentageNum    *float64 `json:"discount_percentage_num"`
	DiscountPercentageAmount *float64 `json:"discount_percentage_amount"`
	DiscountFinal            *float64 `json:"discount_final"`
	DiscountType             *string  `json:"discount_type"`
	TotalAmount              *float64 `json:"total_amount"`
	TotalDp                  *float64 `json:"total_dp"`
}

type CreateInvoiceDpRequest struct {
	CustomerID               *uint                      `json:"customer_id"`
	CurrencyID               *uint                      `json:"currency_id"`
	PaymentTermID            *uint                      `json:"payment_term_id"`
	VatID                    *uint                      `json:"vat_id"`
	Pph23ID                  *uint                      `json:"pph23_id"`
	BranchID                 *uint                      `json:"branch_id"`
	InvoiceNo                *string                    `json:"invoice_no"`
	InvoiceDate              *string                    `json:"invoice_date"`
	ExchangeRate             *float64                   `json:"exchange_rate"`
	Remark                   *string                    `json:"remark"`
	Status                   *string                    `json:"status"`
	Pph23Percentage          *float64                   `json:"pph23_percentage"`
	VatPercentage            *float64                   `json:"vat_percentage"`
	DiscountAmount           *float64                   `json:"discount_amount"`
	DiscountPercentage       *float64                   `json:"discount_percentage"`
	DiscountPercentageAmount *float64                   `json:"discount_percentage_amount"`
	DiscountFinal            *float64                   `json:"discount_final"`
	DiscountType             *string                    `json:"discount_type"`
	DpPercentage             *float64                   `json:"dp_percentage"`
	TotalAmountProducts      *float64                   `json:"total_amount_products"`
	TotalDpProducts          *float64                   `json:"total_dp_products"`
	Subtotal                 *float64                   `json:"subtotal"`
	TotalQty                 *float64                   `json:"total_qty"`
	TotalDiscount            *float64                   `json:"total_discount"`
	TotalPph23               *float64                   `json:"total_pph23"`
	TotalVat                 *float64                   `json:"total_vat"`
	GrandTotal               *float64                   `json:"grand_total"`
	InvoiceDpDts             []CreateInvoiceDpDtRequest `json:"invoice_dp_dts"`
}

type UpdateInvoiceDpDtRequest struct {
	ID                       *uint    `json:"id"`
	InvoiceDpDtID            *uint    `json:"invoice_dp_dt_id"`
	ProductUuid              string   `json:"product_uuid"`
	InvoiceDpID              *uint    `json:"invoice_dp_id"`
	ItemUnitID               *uint    `json:"item_unit_id"`
	VatID                    *uint    `json:"vat_id"`
	Pph23ID                  *uint    `json:"pph23_id"`
	RefID                    *uint    `json:"ref_id"`
	RefDtID                  *uint    `json:"ref_dt_id"`
	ProductID                *uint    `json:"product_id"`
	RefType                  *string  `json:"ref_type"`
	ProductType              *string  `json:"product_type"`
	Remark                   *string  `json:"remark"`
	DpPercentage             *float64 `json:"dp_percentage"`
	IsVat                    *uint    `json:"is_vat"`
	IsPph23                  *uint    `json:"is_pph23"`
	Qty                      *float64 `json:"qty"`
	Price                    *float64 `json:"price"`
	Subtotal                 *float64 `json:"subtotal"`
	DiscountAmount           *float64 `json:"discount_amount"`
	DiscountPercentage       *float64 `json:"discount_percentage"`
	DiscountPercentageNum    *float64 `json:"discount_percentage_num"`
	DiscountPercentageAmount *float64 `json:"discount_percentage_amount"`
	DiscountFinal            *float64 `json:"discount_final"`
	DiscountType             *string  `json:"discount_type"`
	TotalAmount              *float64 `json:"total_amount"`
	TotalDp                  *float64 `json:"total_dp"`
}

type UpdateInvoiceDpRequest struct {
	ID                       uint                       `json:"id"`
	InvoiceDpID              *uint                      `json:"invoice_dp_id"`
	CustomerID               *uint                      `json:"customer_id"`
	CurrencyID               *uint                      `json:"currency_id"`
	PaymentTermID            *uint                      `json:"payment_term_id"`
	VatID                    *uint                      `json:"vat_id"`
	Pph23ID                  *uint                      `json:"pph23_id"`
	BranchID                 *uint                      `json:"branch_id"`
	InvoiceNo                *string                    `json:"invoice_no"`
	InvoiceDate              *string                    `json:"invoice_date"`
	ExchangeRate             *float64                   `json:"exchange_rate"`
	Remark                   *string                    `json:"remark"`
	Status                   *string                    `json:"status"`
	Pph23Percentage          *float64                   `json:"pph23_percentage"`
	VatPercentage            *float64                   `json:"vat_percentage"`
	DiscountAmount           *float64                   `json:"discount_amount"`
	DiscountPercentage       *float64                   `json:"discount_percentage"`
	DiscountPercentageAmount *float64                   `json:"discount_percentage_amount"`
	DiscountFinal            *float64                   `json:"discount_final"`
	DiscountType             *string                    `json:"discount_type"`
	DpPercentage             *float64                   `json:"dp_percentage"`
	TotalAmountProducts      *float64                   `json:"total_amount_products"`
	TotalDpProducts          *float64                   `json:"total_dp_products"`
	Subtotal                 *float64                   `json:"subtotal"`
	TotalQty                 *float64                   `json:"total_qty"`
	TotalDiscount            *float64                   `json:"total_discount"`
	TotalPph23               *float64                   `json:"total_pph23"`
	TotalVat                 *float64                   `json:"total_vat"`
	GrandTotal               *float64                   `json:"grand_total"`
	InvoiceDpDts             []UpdateInvoiceDpDtRequest `json:"invoice_dp_dts"`
}

type GetInvoiceDpByIDRequest struct {
	ID uint `json:"id"`
}

type GetInvoiceDpParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetInvoiceDpParams(id uint) *GetInvoiceDpParams {
	defaultIsDeleted := 0
	return &GetInvoiceDpParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type GetInvoiceDpDtParams struct {
	ID          uint
	InvoiceDpID uint
	IsDeleted   *int
}

type DeleteInvoiceDpRequest struct {
	ID uint `json:"id"`
}

type InvoiceDpListDTO struct {
	ID                       int      `json:"id" db:"id"`
	InvoiceDpID              *uint    `json:"invoice_dp_id" db:"invoice_dp_id"`
	CustomerID               *uint    `json:"customer_id" db:"customer_id"`
	CurrencyID               *uint    `json:"currency_id" db:"currency_id"`
	PaymentTermID            *uint    `json:"payment_term_id" db:"payment_term_id"`
	VatID                    *uint    `json:"vat_id" db:"vat_id"`
	Pph23ID                  *uint    `json:"pph23_id" db:"pph23_id"`
	BranchID                 *uint    `json:"branch_id" db:"branch_id"`
	InvoiceNo                *string  `json:"invoice_no" db:"invoice_no"`
	InvoiceDate              *string  `json:"invoice_date" db:"invoice_date"`
	Remark                   *string  `json:"remark" db:"remark"`
	Status                   *string  `json:"status" db:"status"`
	ExchangeRate             *float64 `json:"exchange_rate" db:"exchange_rate"`
	VatPercentage            *float64 `json:"vat_percentage" db:"vat_percentage"`
	Pph23Percentage          *float64 `json:"pph23_percentage" db:"pph23_percentage"`
	DiscountAmount           *float64 `json:"discount_amount" db:"discount_amount"`
	DiscountPercentage       *float64 `json:"discount_percentage" db:"discount_percentage"`
	DiscountPercentageAmount *float64 `json:"discount_percentage_amount" db:"discount_percentage_amount"`
	DiscountFinal            *float64 `json:"discount_final" db:"discount_final"`
	DiscountType             *string  `json:"discount_type" db:"discount_type"`
	DpPercentage             *float64 `json:"dp_percentage" db:"dp_percentage"`
	TotalAmountProducts      *float64 `json:"total_amount_products" db:"total_amount_products"`
	TotalDpProducts          *float64 `json:"total_dp_products" db:"total_dp_products"`
	TotalQty                 *float64 `json:"total_qty" db:"total_qty"`
	Subtotal                 *float64 `json:"subtotal" db:"subtotal"`
	TotalDiscount            *float64 `json:"total_discount" db:"total_discount"`
	TotalPph23               *float64 `json:"total_pph23" db:"total_pph23"`
	TotalVat                 *float64 `json:"total_vat" db:"total_vat"`
	GrandTotal               *float64 `json:"grand_total" db:"grand_total"`
	CreatedByID              *uint    `json:"created_by_id" db:"created_by_id"`
	UpdatedByID              *uint    `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID              *uint    `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName            *string  `json:"created_by_name" db:"created_by_name"`
	UpdatedByName            *string  `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt                *string  `json:"created_at" db:"created_at"`
	UpdatedAt                *string  `json:"updated_at" db:"updated_at"`
	DeleteAt                 *string  `json:"deleted_at" db:"deleted_at"`

	CurrencyName      *string `json:"currency_name" db:"currency_name"`
	CustomerName      *string `json:"customer_name" db:"customer_name"`
	PaymentTermName   *string `json:"payment_term_name" db:"payment_term_name"`
	VatName           *string `json:"vat_name" db:"vat_name"`
	Pph23Name         *string `json:"pph23_name" db:"pph23_name"`
	BranchName        *string `json:"branch_name" db:"branch_name"`
	InvoiceDpDtRemark *string `json:"invoice_dp_dt_remark" db:"invoice_dp_dt_remark"`
}

type InvoiceDpDetailDTO struct {
	ID                       uint     `json:"id" db:"id"`
	InvoiceDpID              *uint    `json:"invoice_dp_id" db:"invoice_dp_id"`
	CustomerID               *uint    `json:"customer_id" db:"customer_id"`
	CurrencyID               *uint    `json:"currency_id" db:"currency_id"`
	PaymentTermID            *uint    `json:"payment_term_id" db:"payment_term_id"`
	VatID                    *uint    `json:"vat_id" db:"vat_id"`
	Pph23ID                  *uint    `json:"pph23_id" db:"pph23_id"`
	BranchID                 *uint    `json:"branch_id" db:"branch_id"`
	InvoiceNo                *string  `json:"invoice_no" db:"invoice_no"`
	InvoiceDate              *string  `json:"invoice_date" db:"invoice_date"`
	Remark                   *string  `json:"remark" db:"remark"`
	Status                   *string  `json:"status" db:"status"`
	ExchangeRate             float64  `json:"exchange_rate" db:"exchange_rate"`
	VatPercentage            *float64 `json:"vat_percentage" db:"vat_percentage"`
	Pph23Percentage          *float64 `json:"pph23_percentage" db:"pph23_percentage"`
	DiscountAmount           *float64 `json:"discount_amount" db:"discount_amount"`
	DiscountPercentage       *float64 `json:"discount_percentage" db:"discount_percentage"`
	DiscountPercentageAmount *float64 `json:"discount_percentage_amount" db:"discount_percentage_amount"`
	DiscountFinal            *float64 `json:"discount_final" db:"discount_final"`
	DiscountType             *string  `json:"discount_type" db:"discount_type"`
	DpPercentage             *float64 `json:"dp_percentage" db:"dp_percentage"`
	TotalAmountProducts      *float64 `json:"total_amount_products" db:"total_amount_products"`
	TotalDpProducts          *float64 `json:"total_dp_products" db:"total_dp_products"`
	TotalQty                 float64  `json:"total_qty" db:"total_qty"`
	Subtotal                 float64  `json:"subtotal" db:"subtotal"`
	TotalDiscount            float64  `json:"total_discount" db:"total_discount"`
	TotalPph23               float64  `json:"total_pph23" db:"total_pph23"`
	TotalVat                 float64  `json:"total_vat" db:"total_vat"`
	GrandTotal               float64  `json:"grand_total" db:"grand_total"`

	CreatedByID   *uint                `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint                `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint                `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName *string              `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string              `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string              `json:"created_at" db:"created_at"`
	UpdatedAt     *string              `json:"updated_at" db:"updated_at"`
	DeleteAt      *string              `json:"deleted_at" db:"deleted_at"`
	InvoiceDpDts  []InvoiceDpDtListDTO `json:"invoice_dp_dts"`
}

type InvoiceDpDtListDTO struct {
	ID                       *uint    `json:"id" db:"id"`
	InvoiceDpDtID            *uint    `json:"invoice_dp_dt_id" db:"invoice_dp_dt_id"`
	ProductUuid              *string  `json:"product_uuid" db:"product_uuid"`
	CustomerID               *uint    `json:"customer_id" db:"customer_id"`
	InvoiceDpID              *uint    `json:"invoice_dp_id" db:"invoice_dp_id"`
	ItemUnitID               *uint    `json:"item_unit_id" db:"item_unit_id"`
	VatID                    *uint    `json:"vat_id" db:"vat_id"`
	Pph23ID                  *uint    `json:"pph23_id" db:"pph23_id"`
	RefID                    *uint    `json:"ref_id" db:"ref_id"`
	RefDtID                  *uint    `json:"ref_dt_id" db:"ref_dt_id"`
	ProductID                *uint    `json:"product_id" db:"product_id"`
	ItemSubGroupID           *uint    `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID              *uint    `json:"item_group_id" db:"item_group_id"`
	ItemSubGroupName         *string  `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName            *string  `json:"item_group_name" db:"item_group_name"`
	ItemName                 *string  `json:"item_name" db:"item_name"`
	ItemCode                 *string  `json:"item_code" db:"item_code"`
	UnitName                 *string  `json:"unit_name" db:"unit_name"`
	RefJSON                  *string  `json:"ref_json" db:"ref_json"`
	RefType                  *string  `json:"ref_type" db:"ref_type"`
	ProductType              *string  `json:"product_type" db:"product_type"`
	Remark                   *string  `json:"remark" db:"remark"`
	DpPercentage             *float64 `json:"dp_percentage" db:"dp_percentage"`
	IsVat                    *uint    `json:"is_vat" db:"is_vat"`
	IsPph23                  *uint    `json:"is_pph23" db:"is_pph23"`
	VatName                  *string  `json:"vat_name" db:"vat_name"`
	Pph23Name                *string  `json:"pph23_name" db:"pph23_name"`
	Qty                      *float64 `json:"qty" db:"qty"`
	Price                    *float64 `json:"price" db:"price"`
	Subtotal                 *float64 `json:"subtotal" db:"subtotal"`
	DiscountAmount           *float64 `json:"discount_amount" db:"discount_amount"`
	DiscountPercentage       *float64 `json:"discount_percentage" db:"discount_percentage"`
	DiscountPercentageNum    *float64 `json:"discount_percentage_num" db:"discount_percentage_num"`
	DiscountPercentageAmount *float64 `json:"discount_percentage_amount" db:"discount_percentage_amount"`
	DiscountFinal            *float64 `json:"discount_final" db:"discount_final"`
	DiscountType             *string  `json:"discount_type" db:"discount_type"`
	TotalAmount              *float64 `json:"total_amount" db:"total_amount"`
	TotalDp                  *float64 `json:"total_dp" db:"total_dp"`

	CreatedByName *string   `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string   `json:"updated_by_name" db:"updated_by_name"`
	CreatedByID   *uint     `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint     `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint     `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *string   `json:"updated_at" db:"updated_at"`
	DeleteAt      *string   `json:"deleted_at" db:"deleted_at"`

	RefNum *string `json:"ref_num" db:"ref_num"`

	SoDtsBoms []SalesOrderSoDtBomListDTO `json:"invoice_dp_dt_boms"`
}

type InvoiceDpDtListUpdateDTO struct {
	ID                       *uint    `json:"id" db:"id"`
	InvoiceDpDtID            *uint    `json:"invoice_dp_dt_id" db:"invoice_dp_dt_id"`
	ProductUuid              *string  `json:"product_uuid" db:"product_uuid"`
	InvoiceDpID              *uint    `json:"invoice_dp_id" db:"invoice_dp_id"`
	ItemUnitID               *uint    `json:"item_unit_id" db:"item_unit_id"`
	VatID                    *uint    `json:"vat_id" db:"vat_id"`
	Pph23ID                  *uint    `json:"pph23_id" db:"pph23_id"`
	RefID                    *uint    `json:"ref_id" db:"ref_id"`
	RefDtID                  *uint    `json:"ref_dt_id"`
	ProductID                *uint    `json:"product_id" db:"product_id"`
	ItemSubGroupID           *uint    `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID              *uint    `json:"item_group_id" db:"item_group_id"`
	ItemSubGroupName         *string  `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName            *string  `json:"item_group_name" db:"item_group_name"`
	ItemName                 *string  `json:"item_name" db:"item_name"`
	ItemCode                 *string  `json:"item_code" db:"item_code"`
	UnitName                 *string  `json:"unit_name" db:"unit_name"`
	RefJSON                  *string  `json:"ref_json" db:"ref_json"`
	RefType                  *string  `json:"ref_type" db:"ref_type"`
	ProductType              *string  `json:"product_type" db:"product_type"`
	Remark                   *string  `json:"remark" db:"remark"`
	DpPercentage             *float64 `json:"dp_percentage" db:"dp_percentage"`
	IsVat                    *uint    `json:"is_vat" db:"is_vat"`
	IsPph23                  *uint    `json:"is_pph23" db:"is_pph23"`
	Qty                      *float64 `json:"qty" db:"qty"`
	Price                    *float64 `json:"price" db:"price"`
	Subtotal                 *float64 `json:"subtotal" db:"subtotal"`
	DiscountAmount           *float64 `json:"discount_amount" db:"discount_amount"`
	DiscountPercentage       *float64 `json:"discount_percentage" db:"discount_percentage"`
	DiscountPercentageNum    *float64 `json:"discount_percentage_num" db:"discount_percentage_num"`
	DiscountPercentageAmount *float64 `json:"discount_percentage_amount" db:"discount_percentage_amount"`
	DiscountFinal            *float64 `json:"discount_final" db:"discount_final"`
	DiscountType             *string  `json:"discount_type" db:"discount_type"`
	TotalAmount              *float64 `json:"total_amount" db:"total_amount"`
	TotalDp                  *float64 `json:"total_dp" db:"total_dp"`

	CreatedByName *string   `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string   `json:"updated_by_name" db:"updated_by_name"`
	CreatedByID   *uint     `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint     `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint     `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *string   `json:"updated_at" db:"updated_at"`
	DeleteAt      *string   `json:"deleted_at" db:"deleted_at"`
}

type GetInvoiceDpsResult struct {
	InvoiceDps []InvoiceDpListDTO
	Total      int
	Err        error
}

type GetRefSalesOrderDtsRequest struct {
	Global         *string `json:"global"`
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
	PerPage        *string `json:"per_page" default:"10"`
	Page           *string `json:"page" default:"1"`
	OrderColumn    *string `json:"order_column" default:"id"`
	OrderDirection *string `json:"order_direction" default:"asc"`
}

type RefSalesOrderDtListDTO struct {
	ID               *uint     `json:"id" db:"id"`
	SoDtID           *uint     `json:"so_dt_id" db:"so_dt_id"`
	ProductUuid      *string   `json:"product_uuid" db:"product_uuid"`
	SalesOrderID     *uint     `json:"sales_order_id" db:"sales_order_id"`
	ItemUnitID       *uint     `json:"item_unit_id" db:"item_unit_id"`
	VatID            *uint     `json:"vat_id" db:"vat_id"`
	Pph23ID          *uint     `json:"pph23_id" db:"pph23_id"`
	RefID            *uint     `json:"ref_id" db:"ref_id"`
	ItemID           *uint     `json:"item_id" db:"item_id"`
	ItemSubGroupID   *uint     `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID      *uint     `json:"item_group_id" db:"item_group_id"`
	ItemSubGroupName *string   `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName    *string   `json:"item_group_name" db:"item_group_name"`
	ItemName         *string   `json:"item_name" db:"item_name"`
	ItemCode         *string   `json:"item_code" db:"item_code"`
	UnitName         *string   `json:"unit_name" db:"unit_name"`
	RefJSON          *string   `json:"ref_json" db:"ref_json"`
	RefType          *string   `json:"ref_type" db:"ref_type"`
	ItemType         *string   `json:"item_type" db:"item_type"`
	GenCode          *string   `json:"gen_code" db:"gen_code"`
	Remark           *string   `json:"remark" db:"remark"`
	VatPerc          *float64  `json:"vat_perc" db:"vat_perc"`
	VatPercAm        *float64  `json:"vat_perc_am" db:"vat_perc_am"`
	VatName          *string   `json:"vat_name" db:"vat_name"`
	Pph23Name        *string   `json:"pph23_name" db:"pph23_name"`
	Pph23Perc        *float64  `json:"pph23_perc" db:"pph23_perc"`
	Pph23PercAm      *float64  `json:"pph23_perc_am" db:"pph23_perc_am"`
	MarkupPerc       *float64  `json:"markup_perc" db:"markup_perc"`
	MarkupPercAm     *float64  `json:"markup_perc_am" db:"markup_perc_am"`
	IsVat            *int8     `json:"is_vat" db:"is_vat"`
	IsPph23          *int8     `json:"is_pph23" db:"is_pph23"`
	IsLockMarkup     *int8     `json:"is_lock_markup" db:"is_lock_markup"`
	IsLockPriceSell  *int8     `json:"is_lock_price_sell" db:"is_lock_price_sell"`
	QtyOut           *float64  `json:"qty_out" db:"qty_out"`
	Qty              *float64  `json:"qty" db:"qty"`
	PriceSell        *float64  `json:"price_sell" db:"price_sell"`
	PriceBuy         *float64  `json:"price_buy" db:"price_buy"`
	SubtotalSell     *float64  `json:"subtotal_sell" db:"subtotal_sell"`
	SubtotalBuy      *float64  `json:"subtotal_buy" db:"subtotal_buy"`
	DiscAm           *float64  `json:"disc_am" db:"disc_am"`
	DiscPerc         *float64  `json:"disc_perc" db:"disc_perc"`
	DiscPercNum      *float64  `json:"disc_perc_num" db:"disc_perc_num"`
	DiscPercAm       *float64  `json:"disc_perc_am" db:"disc_perc_am"`
	DiscFinal        *float64  `json:"disc_final" db:"disc_final"`
	DiscType         *string   `json:"disc_type" db:"disc_type"`
	TotalAm          *float64  `json:"total_am" db:"total_am"`
	CreatedByName    *string   `json:"created_by_name" db:"created_by_name"`
	UpdatedByName    *string   `json:"updated_by_name" db:"updated_by_name"`
	CreatedByID      *uint     `json:"created_by_id" db:"created_by_id"`
	UpdatedByID      *uint     `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID      *uint     `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        *string   `json:"updated_at" db:"updated_at"`
	DeleteAt         *string   `json:"deleted_at" db:"deleted_at"`

	CustomerID     *uint    `json:"customer_id" db:"customer_id"`
	OrderTypeID    *uint    `json:"order_type_id" db:"order_type_id"`
	CurrencyID     *uint    `json:"currency_id" db:"currency_id"`
	HeadVatID      *uint    `json:"head_vat_id" db:"head_vat_id"`
	HeadPph23ID    *uint    `json:"head_pph23_id" db:"head_pph23_id"`
	HeadVatPerc    *float64 `json:"head_vat_perc" db:"head_vat_perc"`
	HeadPph23Perc  *float64 `json:"head_pph23_perc" db:"head_pph23_perc"`
	HeadDiscAm     *float64 `json:"head_disc_am" db:"head_disc_am"`
	HeadDiscPerc   *float64 `json:"head_disc_perc" db:"head_disc_perc"`
	HeadMarkupPerc *float64 `json:"head_markup_perc" db:"head_markup_perc"`
	HeadRemark     *string  `json:"head_remark" db:"head_remark"`
	ExchangeRate   *float64 `json:"exchange_rate" db:"exchange_rate"`
	SalesOrderNo   *string  `json:"sales_order_no" db:"sales_order_no"`
	CustomerName   *string  `json:"customer_name" db:"customer_name"`
	ItemSku        *string  `json:"item_sku" db:"item_sku"`
	DueAt          *string  `json:"due_at" db:"due_at"`

	SoDtsBoms []SalesOrderSoDtBomListDTO `json:"so_dts_boms"`
}

type GetSoDtQtyUpdateForInvoiceDTO struct {
	ID           *uint    `json:"id" db:"id"`
	SoDtID       *uint    `json:"so_dt_id" db:"so_dt_id"`
	QtyInvoiced  *float64 `json:"qty_invoiced" db:"qty_invoiced"`
	SalesOrderID *uint    `json:"sales_order_id" db:"sales_order_id"`
}

type UpdateSalesOrderStatusForInvoiceRequest struct {
	ID     uint   `json:"id"`
	Status string `json:"status"`
}
