package dtos

import "time"

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
	ProductUuid  *string                   `json:"product_uuid"`
	ItemUnitID   *uint                     `json:"item_unit_id"`
	VatID        *uint                     `json:"vat_id"`
	RefID        *uint                     `json:"ref_id"`
	ItemID       uint                      `json:"item_id"`
	RefType      *string                   `json:"ref_type"`
	ItemType     *string                   `json:"item_type"`
	GenCode      *string                   `json:"gen_code"`
	Remark       *string                   `json:"remark"`
	VatPerc      *float64                  `json:"vat_perc"`
	VatPercAm    *float64                  `json:"vat_perc_am"`
	QtySO        *float64                  `json:"qty_so"`
	Qty          *float64                  `json:"qty"`
	PriceSell    *float64                  `json:"price_sell"`
	PriceBuy     *float64                  `json:"price_buy"`
	SubtotalSell *float64                  `json:"subtotal_sell"`
	SubtotalBuy  *float64                  `json:"subtotal_buy"`
	DiscAm       *float64                  `json:"disc_am"`
	DiscPerc     *float64                  `json:"disc_perc"`
	DiscPercNum  *float64                  `json:"disc_perc_num"`
	DiscPercAm   *float64                  `json:"disc_perc_am"`
	DiscFinal    *float64                  `json:"disc_final"`
	DiscType     *string                   `json:"disc_type"`
	TotalAm      *float64                  `json:"total_am"`
	QuoDtsBoms   []CreateQuoDtsBomsRequest `json:"quo_dts_boms"`
}
type CreateQuoDtsBomsRequest struct {
	ProductUuid   *string `json:"product_uuid"`
	ProductID     uint    `json:"product_id"`
	ProductItemID uint    `json:"product_item_id"`
	ItemUnitID    *uint   `json:"item_unit_id"`
	RefType       *string `json:"ref_type"`
	GenCode       *string `json:"gen_code"`
	Remark        *string `json:"remark"`
	Qty           float64 `json:"qty"`
	PriceSell     float64 `json:"price_sell"`
	PriceBuy      float64 `json:"price_buy"`
	SubtotalSell  float64 `json:"subtotal_sell"`
	SubtotalBuy   float64 `json:"subtotal_buy"`
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
	ID           *uint                     `json:"id"`
	QuotationID  *uint                     `json:"quotation_id"`
	ItemUnitID   *uint                     `json:"item_unit_id"`
	VatID        *uint                     `json:"vat_id"`
	RefID        *uint                     `json:"ref_id"`
	ItemID       uint                      `json:"item_id"`
	RefType      *string                   `json:"ref_type"`
	ItemType     *string                   `json:"item_type"`
	GenCode      *string                   `json:"gen_code"`
	Remark       *string                   `json:"remark"`
	VatPerc      *float64                  `json:"vat_perc"`
	VatPercAm    *float64                  `json:"vat_perc_am"`
	QtySO        *float64                  `json:"qty_so"`
	Qty          *float64                  `json:"qty"`
	PriceSell    *float64                  `json:"price_sell"`
	PriceBuy     *float64                  `json:"price_buy"`
	SubtotalSell *float64                  `json:"subtotal_sell"`
	SubtotalBuy  *float64                  `json:"subtotal_buy"`
	DiscAm       *float64                  `json:"disc_am"`
	DiscPerc     *float64                  `json:"disc_perc"`
	DiscPercNum  *float64                  `json:"disc_perc_num"`
	DiscPercAm   *float64                  `json:"disc_perc_am"`
	DiscFinal    *float64                  `json:"disc_final"`
	DiscType     *string                   `json:"disc_type"`
	TotalAm      *float64                  `json:"total_am"`
	QuoDtsBoms   []UpdateQuoDtsBomsRequest `json:"quo_dts_boms"`
}

type UpdateQuoDtsBomsRequest struct {
	ID           *uint    `json:"id"`
	QuotationID  *uint    `json:"quotation_id"`
	QuoDtID      *uint    `json:"quo_dt_id"`
	ProductID    uint     `json:"product_id"`
	ItemID       uint     `json:"item_id"`
	ItemUnitID   *uint    `json:"item_unit_id"`
	RefType      *string  `json:"ref_type"`
	GenCode      *string  `json:"gen_code"`
	Remark       *string  `json:"remark"`
	Qty          *float64 `json:"qty"`
	PriceSell    *float64 `json:"price_sell"`
	PriceBuy     *float64 `json:"price_buy"`
	SubtotalSell *float64 `json:"subtotal_sell"`
	SubtotalBuy  *float64 `json:"subtotal_buy"`
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
	QuoDts        []UpdateQuoDtsRequest `json:"quo_dts"`
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
	CreatedByID   *uint    `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint    `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint    `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName *string  `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string  `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string  `json:"created_at" db:"created_at"`
	UpdatedAt     *string  `json:"updated_at" db:"updated_at"`
	DeleteAt      *string  `json:"deleted_at" db:"deleted_at"`

	// quo_dt_vat_id, currency_name vat_name pph23_name
	ProductID       *string `json:"product_id" db:"product_id"`
	ItemID          *string `json:"item_id" db:"item_id"`
	QuoDtVatID      *string `json:"quo_dt_vat_id" db:"quo_dt_vat_id"`
	CurrencyName    *string `json:"currency_name" db:"currency_name"`
	ProductName     *string `json:"product_name" db:"product_name"`
	ItemName        *string `json:"item_name" db:"item_name"`
	VatName         *string `json:"vat_name" db:"vat_name"`
	Pph23Name       *string `json:"pph23_name" db:"pph23_name"`
	QuoDtRemark     *string `json:"quo_dt_remark" db:"quo_dt_remark"`
	QuoDtGenCode    *string `json:"quo_dt_gen_code" db:"quo_dt_gen_code"`
	QuoDtBomGenCode *string `json:"quo_dt_bom_gen_code" db:"quo_dt_bom_gen_code"`
	QuoDtBomRemark  *string `json:"quo_dt_bom_remark" db:"quo_dt_bom_remark"`
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
	CreatedByID   *uint                   `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint                   `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint                   `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName *string                 `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string                 `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string                 `json:"created_at" db:"created_at"`
	UpdatedAt     *string                 `json:"updated_at" db:"updated_at"`
	DeleteAt      *string                 `json:"deleted_at" db:"deleted_at"`
	QuoDts        []QuotationQuoDtListDTO `json:"quo_dts"`
}

type QuotationQuoDtListDTO struct {
	ID               *uint     `json:"id" db:"id"`
	QuoDtID          *uint     `json:"quo_dt_id" db:"quo_dt_id"`
	ProductUuid      *string   `json:"product_uuid" db:"product_uuid"`
	QuotationID      *uint     `json:"quotation_id" db:"quotation_id"`
	ItemUnitID       *uint     `json:"item_unit_id" db:"item_unit_id"`
	VatID            *uint     `json:"vat_id" db:"vat_id"`
	RefID            *uint     `json:"ref_id" db:"ref_id"`
	ItemID           *uint     `json:"item_id" db:"item_id"`
	ItemSubGroupID   *uint     `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID      *uint     `json:"item_group_id" db:"item_group_id"`
	ItemSubGroupName *string   `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName    *string   `json:"item_group_name" db:"item_group_name"`
	ItemName         *string   `json:"item_name" db:"item_name"`
	UnitName         *string   `json:"unit_name" db:"unit_name"`
	RefJSON          *string   `json:"ref_json" db:"ref_json"`
	RefType          *string   `json:"ref_type" db:"ref_type"`
	ItemType         *string   `json:"item_type" db:"item_type"`
	GenCode          *string   `json:"gen_code" db:"gen_code"`
	Remark           *string   `json:"remark" db:"remark"`
	VatPerc          *float64  `json:"vat_perc" db:"vat_perc"`
	VatPercAm        *float64  `json:"vat_perc_am" db:"vat_perc_am"`
	QtySO            *float64  `json:"qty_so" db:"qty_so"`
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

	QuoDtsBoms []QuotationQuoDtBomListDTO `json:"quo_dts_boms"`
}

type QuotationQuoDtBomListDTO struct {
	ID               *uint     `json:"id" db:"id"`
	QuoDtBomID       *uint     `json:"quo_dt_bom_id" db:"quo_dt_bom_id"`
	QuotationID      *uint     `json:"quotation_id" db:"quotation_id"`
	QuoDtID          *uint     `json:"quo_dt_id" db:"quo_dt_id"`
	ProductID        uint      `json:"product_id" db:"product_id"`
	ProductUuid      string    `json:"product_uuid" db:"product_uuid"`
	ItemID           uint      `json:"item_id" db:"item_id"`
	ItemSubGroupID   *uint     `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID      *uint     `json:"item_group_id" db:"item_group_id"`
	ItemSubGroupName *string   `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName    *string   `json:"item_group_name" db:"item_group_name"`
	ItemName         *string   `json:"item_name" db:"item_name"`
	UnitName         *string   `json:"unit_name" db:"unit_name"`
	ItemUnitID       *uint     `json:"item_unit_id" db:"item_unit_id"`
	RefJSON          *string   `json:"ref_json" db:"ref_json"`
	GenCode          *string   `json:"gen_code" db:"gen_code"`
	Remark           *string   `json:"remark" db:"remark"`
	Qty              float64   `json:"qty" db:"qty"`
	PriceSell        float64   `json:"price_sell" db:"price_sell"`
	PriceBuy         float64   `json:"price_buy" db:"price_buy"`
	SubtotalSell     float64   `json:"subtotal_sell" db:"subtotal_sell"`
	SubtotalBuy      float64   `json:"subtotal_buy" db:"subtotal_buy"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	CreatedByID      *uint     `json:"created_by_id" db:"created_by_id"`
}

type GetQuotationsResult struct {
	Quotations []QuotationListDTO
	Total      int
	Err        error
}
