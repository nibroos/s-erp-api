package dtos

type GetQuotationsRequest struct {
	Global         *string `json:"global"`
	Title          *string `json:"title"`
	QuoNo          *string `json:"quo_no"`
	Remark         *string `json:"remark"`
	IsApproved     *int    `json:"is_approved"` // 0 = All, 1 = Approved, 2 = Not Approved
	CustomerID     *int    `json:"customer_id"`
	OrderTypeID    *int    `json:"order_type_id"`
	CurrencyID     *int    `json:"currency_id"`
	VatID          *int    `json:"vat_id"`
	PaymentID      *int    `json:"payment_id"`
	Pph23ID        *int    `json:"pph23_id"`
	BranchID       *int    `json:"branch_id"`
	Status         *string `json:"status"`
	DateType       *string `json:"date_type"` // 1 = due_at, 2 = expired_at
	StartDate      *string `json:"start_date"`
	EndDate        *string `json:"end_date"`
	PerPage        *string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           *string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    *string `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection *string `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateQuoDtsRequest struct {
	RefID             *uint                     `json:"ref_id"`
	ItemUnitID        *uint                     `json:"item_unit_id"`
	VatID             *uint                     `json:"vat_id"`
	ProductID         *uint                     `json:"product_id"`
	ProductItemUnitID *uint                     `json:"product_item_unit_id"`
	RefType           *string                   `json:"ref_type"`
	Remark            *string                   `json:"remark"`
	VatPerc           *float64                  `json:"vat_perc"`
	QtySO             *float64                  `json:"qty_so"`
	Qty               *float64                  `json:"qty"`
	PriceSell         *float64                  `json:"price_sell"`
	Subtotal          *float64                  `json:"subtotal"`
	DiscAm            *float64                  `json:"disc_am"`
	DiscPerc          *float64                  `json:"disc_perc"`
	TotalAm           *float64                  `json:"total_am"`
	PQtySO            *float64                  `json:"p_qty_so"`
	PQty              *float64                  `json:"p_qty"`
	PPriceSell        *float64                  `json:"p_price_sell"`
	PSubtotal         *float64                  `json:"p_subtotal"`
	PDiscAm           *float64                  `json:"p_disc_am"`
	PDiscPerc         *float64                  `json:"p_disc_perc"`
	PTotalAm          *float64                  `json:"p_total_am"`
	Boms              []CreateQuoDtsBomsRequest `json:"boms"`
}
type CreateQuoDtsBomsRequest struct {
	RefID             *uint    `json:"ref_id"`
	ItemUnitID        *uint    `json:"item_unit_id"`
	VatID             *uint    `json:"vat_id"`
	ProductID         *uint    `json:"product_id"`
	ProductItemUnitID *uint    `json:"product_item_unit_id"`
	RefType           *string  `json:"ref_type"`
	Remark            *string  `json:"remark"`
	VatPerc           *float64 `json:"vat_perc"`
	QtySO             *float64 `json:"qty_so"`
	Qty               *float64 `json:"qty"`
	PriceSell         *float64 `json:"price_sell"`
	Subtotal          *float64 `json:"subtotal"`
	DiscAm            *float64 `json:"disc_am"`
	DiscPerc          *float64 `json:"disc_perc"`
	TotalAm           *float64 `json:"total_am"`
	PQtySO            *float64 `json:"p_qty_so"`
	PQty              *float64 `json:"p_qty"`
	PPriceSell        *float64 `json:"p_price_sell"`
	PSubtotal         *float64 `json:"p_subtotal"`
	PDiscAm           *float64 `json:"p_disc_am"`
	PDiscPerc         *float64 `json:"p_disc_perc"`
	PTotalAm          *float64 `json:"p_total_am"`
}

type CreateQuotationRequest struct {
	CustomerID    *uint                 `json:"customer_id"`
	OrderTypeID   *uint                 `json:"order_type_id"`
	CurrencyID    *uint                 `json:"currency_id"`
	VatID         *uint                 `json:"vat_id"`
	PaymentID     *uint                 `json:"payment_id"`
	Pph23ID       *uint                 `json:"pph23_id"`
	BranchID      *uint                 `json:"branch_id"`
	QuoNo         *string               `json:"quo_no"`
	Title         string                `json:"title"`
	Remark        *string               `json:"remark"`
	Status        string                `json:"status"`
	IsApproved    int8                  `json:"is_approved"`
	ExchangeRate  *float64              `json:"exchange_rate"`
	VatPerc       *float64              `json:"vat_perc"`
	Pph23Perc     *float64              `json:"pph23_perc"`
	TotalQty      *float64              `json:"total_qty"`
	Subtotal      *float64              `json:"subtotal"`
	TotalDiscount *float64              `json:"total_discount"`
	TotalPph23    *float64              `json:"total_pph23"`
	TotalVat      *float64              `json:"total_vat"`
	GrandTotal    *float64              `json:"grand_total"`
	DueAt         *string               `json:"due_at"`
	ExpiredAt     *string               `json:"expired_at"`
	QuoDts        []CreateQuoDtsRequest `json:"quo_dts"`
}

type UpdateQuoDtsRequest struct {
	ID                *uint    `json:"id"`
	QuotationID       *uint    `json:"quotation_id"`
	RefID             *uint    `json:"ref_id"`
	ItemUnitID        *uint    `json:"item_unit_id"`
	VatID             *uint    `json:"vat_id"`
	ProductID         *uint    `json:"product_id"`
	ProductItemUnitID *uint    `json:"product_item_unit_id"`
	RefType           *string  `json:"ref_type"`
	Remark            *string  `json:"remark"`
	VatPerc           *float64 `json:"vat_perc"`
	QtySO             *float64 `json:"qty_so"`
	Qty               *float64 `json:"qty"`
	PriceSell         *float64 `json:"price_sell"`
	Subtotal          *float64 `json:"subtotal"`
	DiscAm            *float64 `json:"disc_am"`
	DiscPerc          *float64 `json:"disc_perc"`
	TotalAm           *float64 `json:"total_am"`
	PQtySO            *float64 `json:"p_qty_so"`
	PQty              *float64 `json:"p_qty"`
	PPriceSell        *float64 `json:"p_price_sell"`
	PSubtotal         *float64 `json:"p_subtotal"`
	PDiscAm           *float64 `json:"p_disc_am"`
	PDiscPerc         *float64 `json:"p_disc_perc"`
	PTotalAm          *float64 `json:"p_total_am"`
}

type UpdateQuotationRequest struct {
	ID            uint                  `json:"id"`
	QuotationID   *uint                 `json:"quotation_id"`
	CustomerID    *uint                 `json:"customer_id"`
	OrderTypeID   *uint                 `json:"order_type_id"`
	CurrencyID    *uint                 `json:"currency_id"`
	VatID         *uint                 `json:"vat_id"`
	PaymentID     *uint                 `json:"payment_id"`
	Pph23ID       *uint                 `json:"pph23_id"`
	BranchID      *uint                 `json:"branch_id"`
	QuoNo         *string               `json:"quo_no"`
	Title         string                `json:"title"`
	Remark        *string               `json:"remark"`
	Status        string                `json:"status"`
	IsApproved    int8                  `json:"is_approved"`
	ExchangeRate  *float64              `json:"exchange_rate"`
	VatPerc       *float64              `json:"vat_perc"`
	Pph23Perc     *float64              `json:"pph23_perc"`
	TotalQty      *float64              `json:"total_qty"`
	Subtotal      *float64              `json:"subtotal"`
	TotalDiscount *float64              `json:"total_discount"`
	TotalPph23    *float64              `json:"total_pph23"`
	TotalVat      *float64              `json:"total_vat"`
	GrandTotal    *float64              `json:"grand_total"`
	DueAt         *string               `json:"due_at"`
	ExpiredAt     *string               `json:"expired_at"`
	QuoDts        []UpdateQuoDtsRequest `json:"boms"`
}

type GetQuotationByIDRequest struct {
	ID uint `json:"id"`
}

type GetQuotationParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetQuotationParams(id uint) *GetQuotationParams {
	defaultIsDeleted := 0
	return &GetQuotationParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type GetQuotationQuoDtParams struct {
	ID          uint
	QuotationID uint
	IsDeleted   *int
}

type DeleteQuotationRequest struct {
	ID uint `json:"id"`
}

type QuotationListDTO struct {
	ID            int      `json:"id" db:"id"`
	QuotationID   *uint    `json:"quotation_id" db:"quotation_id"`
	CustomerID    *uint    `json:"customer_id" db:"customer_id"`
	OrderTypeID   *uint    `json:"order_type_id" db:"order_type_id"`
	CurrencyID    *uint    `json:"currency_id" db:"currency_id"`
	VatID         *uint    `json:"vat_id" db:"vat_id"`
	PaymentID     *uint    `json:"payment_id" db:"payment_id"`
	Pph23ID       *uint    `json:"pph23_id" db:"pph23_id"`
	BranchID      *uint    `json:"branch_id" db:"branch_id"`
	QuoNo         *string  `json:"quo_no" db:"quo_no"`
	Title         string   `json:"title" db:"title"`
	Remark        *string  `json:"remark" db:"remark"`
	Status        string   `json:"status" db:"status"`
	IsApproved    int8     `json:"is_approved" db:"is_approved"`
	ExchangeRate  *float64 `json:"exchange_rate" db:"exchange_rate"`
	VatPerc       *float64 `json:"vat_perc" db:"vat_perc"`
	Pph23Perc     *float64 `json:"pph23_perc" db:"pph23_perc"`
	TotalQty      *float64 `json:"total_qty" db:"total_qty"`
	Subtotal      *float64 `json:"subtotal" db:"subtotal"`
	TotalDiscount *float64 `json:"total_discount" db:"total_discount"`
	TotalPph23    *float64 `json:"total_pph23" db:"total_pph23"`
	TotalVat      *float64 `json:"total_vat" db:"total_vat"`
	GrandTotal    *float64 `json:"grand_total" db:"grand_total"`
	DueAt         *string  `json:"due_at" db:"due_at"`
	ExpiredAt     *string  `json:"expired_at" db:"expired_at"`
	CreatedByName *string  `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string  `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string  `json:"created_at" db:"created_at"`
	UpdatedAt     *string  `json:"updated_at" db:"updated_at"`
	DeleteAt      *string  `json:"deleted_at" db:"deleted_at"`
}

type QuotationQuoDtListDTO struct {
	ID                *uint   `json:"id" db:"id"`
	QuoDtID           *uint   `json:"quo_dt_id" db:"quo_dt_id"`
	QuotationID       *uint   `json:"quotation_id" db:"quotation_id"`
	RefID             *uint   `json:"ref_id" db:"ref_id"`
	ItemUnitID        *uint   `json:"item_unit_id" db:"item_unit_id"`
	VatID             *uint   `json:"vat_id" db:"vat_id"`
	ProductID         *uint   `json:"product_id" db:"product_id"`
	ProductItemUnitID *uint   `json:"product_item_unit_id" db:"product_item_unit_id"`
	QuoDtRefID        *uint   `json:"quo_dt_ref_id" db:"quo_dt_ref_id"`
	RefJSON           *string `json:"ref_json" db:"ref_json"`
	RefType           *string `json:"ref_type" db:"ref_type"`
	Remark            *string `json:"remark" db:"remark"`
	VatPerc           *string `json:"vat_perc" db:"vat_perc"`
	QtySO             *string `json:"qty_so" db:"qty_so"`
	Qty               *string `json:"qty" db:"qty"`
	PriceSell         *string `json:"price_sell" db:"price_sell"`
	Subtotal          *string `json:"subtotal" db:"subtotal"`
	DiscAm            *string `json:"disc_am" db:"disc_am"`
	DiscPerc          *string `json:"disc_perc" db:"disc_perc"`
	TotalAm           *string `json:"total_am" db:"total_am"`
	PQtySO            *string `json:"p_qty_so" db:"p_qty_so"`
	PQty              *string `json:"p_qty" db:"p_qty"`
	PPriceSell        *string `json:"p_price_sell" db:"p_price_sell"`
	PSubtotal         *string `json:"p_subtotal" db:"p_subtotal"`
	PDiscAm           *string `json:"p_disc_am" db:"p_disc_am"`
	PDiscPerc         *string `json:"p_disc_perc" db:"p_disc_perc"`
	PTotalAm          *string `json:"p_total_am" db:"p_total_am"`
	CreatedByName     *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName     *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt         *string `json:"created_at" db:"created_at"`
	UpdatedAt         *string `json:"updated_at" db:"updated_at"`
	DeleteAt          *string `json:"deleted_at" db:"deleted_at"`
}

type QuotationDetailDTO struct {
	ID            uint                    `json:"id" db:"id"`
	QuotationID   *uint                   `json:"quotation_id" db:"quotation_id"`
	CustomerID    *uint                   `json:"customer_id" db:"customer_id"`
	OrderTypeID   *uint                   `json:"order_type_id" db:"order_type_id"`
	CurrencyID    *uint                   `json:"currency_id" db:"currency_id"`
	VatID         *uint                   `json:"vat_id" db:"vat_id"`
	PaymentID     *uint                   `json:"payment_id" db:"payment_id"`
	Pph23ID       *uint                   `json:"pph23_id" db:"pph23_id"`
	BranchID      *uint                   `json:"branch_id" db:"branch_id"`
	QuoNo         *string                 `json:"quo_no" db:"quo_no"`
	Title         string                  `json:"title" db:"title"`
	Remark        *string                 `json:"remark" db:"remark"`
	Status        string                  `json:"status" db:"status"`
	IsApproved    int8                    `json:"is_approved" db:"is_approved"`
	ExchangeRate  *float64                `json:"exchange_rate" db:"exchange_rate"`
	VatPerc       *float64                `json:"vat_perc" db:"vat_perc"`
	Pph23Perc     *float64                `json:"pph23_perc" db:"pph23_perc"`
	TotalQty      *float64                `json:"total_qty" db:"total_qty"`
	Subtotal      *float64                `json:"subtotal" db:"subtotal"`
	TotalDiscount *float64                `json:"total_discount" db:"total_discount"`
	TotalPph23    *float64                `json:"total_pph23" db:"total_pph23"`
	TotalVat      *float64                `json:"total_vat" db:"total_vat"`
	GrandTotal    *float64                `json:"grand_total" db:"grand_total"`
	DueAt         *string                 `json:"due_at" db:"due_at"`
	ExpiredAt     *string                 `json:"expired_at" db:"expired_at"`
	CreatedByName *string                 `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string                 `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string                 `json:"created_at" db:"created_at"`
	UpdatedAt     *string                 `json:"updated_at" db:"updated_at"`
	DeleteAt      *string                 `json:"deleted_at" db:"deleted_at"`
	QuoDts        []QuotationQuoDtListDTO `json:"quo_dts"`
}

type GetQuotationsResult struct {
	Quotations []QuotationListDTO
	Total      int
	Err        error
}
