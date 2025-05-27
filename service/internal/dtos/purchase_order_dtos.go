package dtos

import "time"

type GetPurchaseOrdersRequest struct {
	Global         *string `json:"global"`
	PoNo           *string `json:"po_no"`
	Remark         *string `json:"remark"`
	CustomerID     *int    `json:"customer_id"`
	PurchaseTypeID *int    `json:"purchase_type_id"`
	CurrencyID     *int    `json:"currency_id"`
	VatID          *int    `json:"vat_id"`
	PaymentTermID  *int    `json:"payment_term_id"`
	ShippingTermID *int    `json:"shipping_term_id"`
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

type CreatePoDtsRequest struct {
	ProductUuid              string   `json:"product_uuid"`
	ItemUnitID               *uint    `json:"item_unit_id"`
	VatID                    *uint    `json:"vat_id"`
	Pph23ID                  *uint    `json:"pph23_id"`
	RefID                    *uint    `json:"ref_id"`
	RefSoDtID                *uint    `json:"ref_so_dt_id"`
	RefSoDtBomID             *uint    `json:"ref_so_dt_bom_id"`
	RefProductID             *uint    `json:"ref_product_id"`
	ProductID                uint     `json:"product_id"`
	BomID                    *uint    `json:"bom_id"`
	ProductType              *string  `json:"product_type"`
	ProductJSON              *string  `json:"product_json"`
	RefType                  *string  `json:"ref_type"`
	RefJSON                  *string  `json:"ref_json"`
	GenCode                  *string  `json:"gen_code"`
	Remark                   *string  `json:"remark"`
	NeedQty                  *float64 `json:"need_qty"`
	Qty                      *float64 `json:"qty"`
	Price                    *float64 `json:"price"`
	Subtotal                 *float64 `json:"subtotal"`
	DiscountAmount           *float64 `json:"discount_amount"`
	DiscountPercentage       *float64 `json:"discount_percentage"`
	DiscountPercentageNum    *float64 `json:"discount_percentage_num"`
	DiscountPercentageAmount *float64 `json:"discount_percentage_amount"`
	DiscountFinal            *float64 `json:"discount_final"`
	DiscountType             *string  `json:"discount_type"`
	VatPerc                  *float64 `json:"vat_perc"`
	VatPercAm                *float64 `json:"vat_perc_am"`
	Pph23Perc                *float64 `json:"pph23_perc"`
	Pph23PercAm              *float64 `json:"pph23_perc_am"`
	IsVat                    *int8    `json:"is_vat"`
	IsPph23                  *int8    `json:"is_pph23"`
	TotalAmount              *float64 `json:"total_amount"`
}

type FormPoDtsRequest struct {
	ID                       *uint    `json:"id"`
	PoDtID                   *uint    `json:"po_dt_id"`
	PoID                     *uint    `json:"po_id"`
	ProductUuid              string   `json:"product_uuid"`
	ItemUnitID               *uint    `json:"item_unit_id"`
	VatID                    *uint    `json:"vat_id"`
	Pph23ID                  *uint    `json:"pph23_id"`
	RefID                    *uint    `json:"ref_id"`
	RefSoDtID                *uint    `json:"ref_so_dt_id"`
	RefSoDtBomID             *uint    `json:"ref_so_dt_bom_id"`
	RefRoDtID                *uint    `json:"ref_ro_dt_id"`
	RefProductID             *uint    `json:"ref_product_id"`
	RefProductBomID          *uint    `json:"ref_product_bom_id"`
	ProductID                uint     `json:"product_id"`
	BomID                    *uint    `json:"bom_id"`
	ProductType              *string  `json:"product_type"`
	ProductJSON              *string  `json:"product_json"`
	RefType                  *string  `json:"ref_type"`
	RefJSON                  *string  `json:"ref_json"`
	GenCode                  *string  `json:"gen_code"`
	Remark                   *string  `json:"remark"`
	NeedQty                  *float64 `json:"need_qty"`
	Qty                      *float64 `json:"qty"`
	Price                    *float64 `json:"price"`
	Subtotal                 *float64 `json:"subtotal"`
	DiscountAmount           *float64 `json:"discount_amount"`
	DiscountPercentage       *float64 `json:"discount_percentage"`
	DiscountPercentageNum    *float64 `json:"discount_percentage_num"`
	DiscountPercentageAmount *float64 `json:"discount_percentage_amount"`
	DiscountFinal            *float64 `json:"discount_final"`
	DiscountType             *string  `json:"discount_type"`
	VatPerc                  *float64 `json:"vat_perc"`
	VatPercAm                *float64 `json:"vat_perc_am"`
	Pph23Perc                *float64 `json:"pph23_perc"`
	Pph23PercAm              *float64 `json:"pph23_perc_am"`
	IsVat                    *int8    `json:"is_vat"`
	IsPph23                  *int8    `json:"is_pph23"`
	TotalAmount              *float64 `json:"total_amount"`
}

type CreatePurchaseOrderRequest struct {
	CustomerID               *uint                `json:"customer_id"`
	PurchaseTypeID           *uint                `json:"purchase_type_id"`
	CurrencyID               *uint                `json:"currency_id"`
	VatID                    *uint                `json:"vat_id"`
	PaymentTermID            *uint                `json:"payment_term_id"`
	ShippingTermID           *uint                `json:"shipping_term_id"`
	Pph23ID                  *uint                `json:"pph23_id"`
	BranchID                 *uint                `json:"branch_id"`
	IsVat                    *int                 `json:"is_vat"`
	PoNo                     *string              `json:"po_no"`
	PoDate                   *string              `json:"po_date"`
	DeliveryDate             *string              `json:"delivery_date"`
	ShippingDestination      *string              `json:"shipping_destination"`
	Remark                   *string              `json:"remark"`
	Status                   string               `json:"status"`
	ExchangeRate             *float64             `json:"exchange_rate"`
	DiscountPercentage       *float64             `json:"discount_percentage"`
	DiscountAmount           *float64             `json:"discount_amount"`
	DiscountPercentageAmount *float64             `json:"discount_percentage_amount"`
	DiscountFinalHeader      *float64             `json:"discount_final_header"`
	DiscountAmountProduct    *float64             `json:"discount_amount_product"`
	DiscountType             *string              `json:"discount_type"`
	Pph23Percentage          *float64             `json:"pph23_percentage"`
	VatPercentage            *float64             `json:"vat_percentage"`
	TotalAmountProducts      *float64             `json:"total_amount_products"`
	Subtotal                 *float64             `json:"subtotal"`
	TotalQty                 *float64             `json:"total_qty"`
	TotalDiscount            *float64             `json:"total_discount"`
	TotalPph23               *float64             `json:"total_pph23"`
	TotalVat                 *float64             `json:"total_vat"`
	GrandTotal               *float64             `json:"grand_total"`
	PoDts                    []CreatePoDtsRequest `json:"po_dts"`

	CustomerCode string `json:"customer_code"`
}

type FormPurchaseOrderRequest struct {
	ID                       *uint              `json:"id"`
	CustomerID               *uint              `json:"customer_id"`
	PurchaseTypeID           *uint              `json:"purchase_type_id"`
	CurrencyID               *uint              `json:"currency_id"`
	VatID                    *uint              `json:"vat_id"`
	PaymentTermID            *uint              `json:"payment_term_id"`
	ShippingTermID           *uint              `json:"shipping_term_id"`
	Pph23ID                  *uint              `json:"pph23_id"`
	BranchID                 *uint              `json:"branch_id"`
	IsVat                    *int               `json:"is_vat"`
	RevNo                    *int               `json:"rev_no"`
	PoNo                     *string            `json:"po_no"`
	PoNoOri                  *string            `json:"po_no_ori"`
	PoDate                   *string            `json:"po_date"`
	DeliveryDate             *string            `json:"delivery_date"`
	ShippingDestination      *string            `json:"shipping_destination"`
	Remark                   *string            `json:"remark"`
	Status                   string             `json:"status"`
	ExchangeRate             *float64           `json:"exchange_rate"`
	DiscountPercentage       *float64           `json:"discount_percentage"`
	DiscountAmount           *float64           `json:"discount_amount"`
	DiscountPercentageAmount *float64           `json:"discount_percentage_amount"`
	DiscountFinalHeader      *float64           `json:"discount_final_header"`
	DiscountAmountProduct    *float64           `json:"discount_amount_product"`
	DiscountType             *string            `json:"discount_type"`
	Pph23Percentage          *float64           `json:"pph23_percentage"`
	VatPercentage            *float64           `json:"vat_percentage"`
	TotalAmountProducts      *float64           `json:"total_amount_products"`
	Subtotal                 *float64           `json:"subtotal"`
	TotalQty                 *float64           `json:"total_qty"`
	TotalDiscount            *float64           `json:"total_discount"`
	TotalPph23               *float64           `json:"total_pph23"`
	TotalVat                 *float64           `json:"total_vat"`
	GrandTotal               *float64           `json:"grand_total"`
	PoDts                    []FormPoDtsRequest `json:"po_dts"`

	CustomerCode string `json:"customer_code"`
}

type UpdatePoDtsRequest struct {
	ID                       *uint    `json:"id"`
	PoDtID                   *uint    `json:"po_dt_id"`
	ProductUuid              string   `json:"product_uuid"`
	PoID                     *uint    `json:"po_id"`
	ItemUnitID               *uint    `json:"item_unit_id"`
	VatID                    *uint    `json:"vat_id"`
	Pph23ID                  *uint    `json:"pph23_id"`
	RefID                    *uint    `json:"ref_id"`
	RefSoDtID                *uint    `json:"ref_so_dt_id"`
	RefSoDtBomID             *uint    `json:"ref_so_dt_bom_id"`
	RefProductID             *uint    `json:"ref_product_id"`
	ProductID                uint     `json:"product_id"`
	BomID                    *uint    `json:"bom_id"`
	ProductType              *string  `json:"product_type"`
	ProductJSON              *string  `json:"product_json"`
	RefType                  *string  `json:"ref_type"`
	RefJSON                  *string  `json:"ref_json"`
	GenCode                  *string  `json:"gen_code"`
	Remark                   *string  `json:"remark"`
	NeedQty                  *float64 `json:"need_qty"`
	Qty                      *float64 `json:"qty"`
	Price                    *float64 `json:"price"`
	Subtotal                 *float64 `json:"subtotal"`
	DiscountAmount           *float64 `json:"discount_amount"`
	DiscountPercentage       *float64 `json:"discount_percentage"`
	DiscountPercentageNum    *float64 `json:"discount_percentage_num"`
	DiscountPercentageAmount *float64 `json:"discount_percentage_amount"`
	DiscountFinal            *float64 `json:"discount_final"`
	DiscountType             *string  `json:"discount_type"`
	VatPerc                  *float64 `json:"vat_perc"`
	VatPercAm                *float64 `json:"vat_perc_am"`
	Pph23Perc                *float64 `json:"pph23_perc"`
	Pph23PercAm              *float64 `json:"pph23_perc_am"`
	IsVat                    *int8    `json:"is_vat"`
	IsPph23                  *int8    `json:"is_pph23"`
	TotalAmount              *float64 `json:"total_amount"`
}

type UpdatePurchaseOrderRequest struct {
	ID                       uint                 `json:"id"`
	PoID                     *uint                `json:"po_id"`
	CustomerID               *uint                `json:"customer_id"`
	PurchaseTypeID           *uint                `json:"purchase_type_id"`
	CurrencyID               *uint                `json:"currency_id"`
	VatID                    *uint                `json:"vat_id"`
	PaymentTermID            *uint                `json:"payment_term_id"`
	ShippingTermID           *uint                `json:"shipping_term_id"`
	Pph23ID                  *uint                `json:"pph23_id"`
	BranchID                 *uint                `json:"branch_id"`
	IsVat                    *int                 `json:"is_vat"`
	RevNo                    *int                 `json:"rev_no"`
	PoNo                     string               `json:"po_no"`
	PoNoOri                  *string              `json:"po_no_ori"`
	PoDate                   *string              `json:"po_date"`
	DeliveryDate             *string              `json:"delivery_date"`
	ShippingDestination      *string              `json:"shipping_destination"`
	Remark                   *string              `json:"remark"`
	Status                   string               `json:"status"`
	ExchangeRate             *float64             `json:"exchange_rate"`
	DiscountPercentage       *float64             `json:"discount_percentage"`
	DiscountAmount           *float64             `json:"discount_amount"`
	DiscountPercentageAmount *float64             `json:"discount_percentage_amount"`
	DiscountFinalHeader      *float64             `json:"discount_final_header"`
	DiscountAmountProduct    *float64             `json:"discount_amount_product"`
	DiscountType             *string              `json:"discount_type"`
	Pph23Percentage          *float64             `json:"pph23_percentage"`
	VatPercentage            *float64             `json:"vat_percentage"`
	TotalAmountProducts      *float64             `json:"total_amount_products"`
	Subtotal                 *float64             `json:"subtotal"`
	TotalQty                 *float64             `json:"total_qty"`
	TotalDiscount            *float64             `json:"total_discount"`
	TotalPph23               *float64             `json:"total_pph23"`
	TotalVat                 *float64             `json:"total_vat"`
	GrandTotal               *float64             `json:"grand_total"`
	PoDts                    []UpdatePoDtsRequest `json:"po_dts"`

	CustomerCode string `json:"customer_code"`
}

type GetPurchaseOrderByIDRequest struct {
	ID uint `json:"id"`
}

type GetPurchaseOrderParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetPurchaseOrderParams(id uint) *GetPurchaseOrderParams {
	defaultIsDeleted := 0
	return &GetPurchaseOrderParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type GetPurchaseOrderPoDtParams struct {
	ID              uint
	PurchaseOrderID uint
	IsDeleted       *int
}

type DeletePurchaseOrderRequest struct {
	ID uint `json:"id"`
}

type PurchaseOrderListDTO struct {
	ID                       int      `json:"id" db:"id"`
	PoID                     *uint    `json:"po_id" db:"po_id"`
	CustomerID               *uint    `json:"customer_id" db:"customer_id"`
	PurchaseTypeID           *uint    `json:"purchase_type_id" db:"purchase_type_id"`
	CurrencyID               *uint    `json:"currency_id" db:"currency_id"`
	VatID                    *uint    `json:"vat_id" db:"vat_id"`
	PaymentTermID            *uint    `json:"payment_term_id" db:"payment_term_id"`
	ShippingTermID           *uint    `json:"shipping_term_id" db:"shipping_term_id"`
	Pph23ID                  *uint    `json:"pph23_id" db:"pph23_id"`
	BranchID                 *uint    `json:"branch_id" db:"branch_id"`
	PoNo                     *string  `json:"po_no" db:"po_no"`
	PoDate                   *string  `json:"po_date" db:"po_date"`
	DeliveryDate             *string  `json:"delivery_date" db:"delivery_date"`
	ShippingDestination      *string  `json:"shipping_destination" db:"shipping_destination"`
	Remark                   *string  `json:"remark" db:"remark"`
	Status                   string   `json:"status" db:"status"`
	ExchangeRate             *float64 `json:"exchange_rate" db:"exchange_rate"`
	DiscountPercentage       *float64 `json:"discount_percentage" db:"discount_percentage"`
	DiscountAmount           *float64 `json:"discount_amount" db:"discount_amount"`
	DiscountPercentageAmount *float64 `json:"discount_percentage_amount" db:"discount_percentage_amount"`
	DiscountFinalHeader      *float64 `json:"discount_final_header" db:"discount_final_header"`
	DiscountAmountProduct    *float64 `json:"discount_amount_product" db:"discount_amount_product"`
	DiscountType             *string  `json:"discount_type" db:"discount_type"`
	Pph23Percentage          *float64 `json:"pph23_percentage" db:"pph23_percentage"`
	VatPercentage            *float64 `json:"vat_percentage" db:"vat_percentage"`
	TotalAmountProducts      *float64 `json:"total_amount_products" db:"total_amount_products"`
	Subtotal                 *float64 `json:"subtotal" db:"subtotal"`
	TotalQty                 *float64 `json:"total_qty" db:"total_qty"`
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

	ProductID        *string `json:"product_id" db:"product_id"`
	PoDtVatID        *string `json:"po_dt_vat_id" db:"po_dt_vat_id"`
	CurrencyName     *string `json:"currency_name" db:"currency_name"`
	PurchaseTypeName *string `json:"purchase_type_name" db:"purchase_type_name"`
	CustomerName     *string `json:"customer_name" db:"customer_name"`
	ProductName      *string `json:"product_name" db:"product_name"`
	VatName          *string `json:"vat_name" db:"vat_name"`
	Pph23Name        *string `json:"pph23_name" db:"pph23_name"`
	PaymentTermName  *string `json:"payment_term_name" db:"payment_term_name"`
	ShippingTermName *string `json:"shipping_term_name" db:"shipping_term_name"`
	PoDtRemark       *string `json:"po_dt_remark" db:"po_dt_remark"`
	PoDtGenCode      *string `json:"po_dt_gen_code" db:"po_dt_gen_code"`
}

type PurchaseOrderDetailDTO struct {
	ID                       uint     `json:"id" db:"id"`
	PoID                     *uint    `json:"po_id" db:"po_id"`
	CustomerID               *uint    `json:"customer_id" db:"customer_id"`
	PurchaseTypeID           *uint    `json:"purchase_type_id" db:"purchase_type_id"`
	CurrencyID               *uint    `json:"currency_id" db:"currency_id"`
	VatID                    *uint    `json:"vat_id" db:"vat_id"`
	PaymentTermID            *uint    `json:"payment_term_id" db:"payment_term_id"`
	ShippingTermID           *uint    `json:"shipping_term_id" db:"shipping_term_id"`
	Pph23ID                  *uint    `json:"pph23_id" db:"pph23_id"`
	BranchID                 *uint    `json:"branch_id" db:"branch_id"`
	IsVat                    *int     `json:"is_vat" db:"is_vat"`
	RevNo                    *int     `json:"rev_no" db:"rev_no"`
	PoNo                     *string  `json:"po_no" db:"po_no"`
	PoNoOri                  *string  `json:"po_no_ori" db:"po_no_ori"`
	PoDate                   *string  `json:"po_date" db:"po_date"`
	DeliveryDate             *string  `json:"delivery_date" db:"delivery_date"`
	ShippingDestination      *string  `json:"shipping_destination" db:"shipping_destination"`
	Remark                   *string  `json:"remark" db:"remark"`
	Status                   string   `json:"status" db:"status"`
	ExchangeRate             float64  `json:"exchange_rate" db:"exchange_rate"`
	DiscountPercentage       *float64 `json:"discount_percentage" db:"discount_percentage"`
	DiscountAmount           *float64 `json:"discount_amount" db:"discount_amount"`
	DiscountPercentageAmount *float64 `json:"discount_percentage_amount" db:"discount_percentage_amount"`
	DiscountFinalHeader      *float64 `json:"discount_final_header" db:"discount_final_header"`
	DiscountAmountProduct    *float64 `json:"discount_amount_product" db:"discount_amount_product"`
	DiscountType             *string  `json:"discount_type" db:"discount_type"`
	Pph23Percentage          *float64 `json:"pph23_percentage" db:"pph23_percentage"`
	VatPercentage            *float64 `json:"vat_percentage" db:"vat_percentage"`
	TotalAmountProducts      *float64 `json:"total_amount_products" db:"total_amount_products"`
	Subtotal                 float64  `json:"subtotal" db:"subtotal"`
	TotalQty                 float64  `json:"total_qty" db:"total_qty"`
	TotalDiscount            float64  `json:"total_discount" db:"total_discount"`
	TotalPph23               float64  `json:"total_pph23" db:"total_pph23"`
	TotalVat                 float64  `json:"total_vat" db:"total_vat"`
	GrandTotal               float64  `json:"grand_total" db:"grand_total"`

	CreatedByID   *uint                      `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint                      `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint                      `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName *string                    `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string                    `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string                    `json:"created_at" db:"created_at"`
	UpdatedAt     *string                    `json:"updated_at" db:"updated_at"`
	DeleteAt      *string                    `json:"deleted_at" db:"deleted_at"`
	PoDts         []PurchaseOrderPoDtListDTO `json:"po_dts"`
}

type PurchaseOrderPoDtListDTO struct {
	ID                       *uint    `json:"id" db:"id"`
	PoDtID                   *uint    `json:"po_dt_id" db:"po_dt_id"`
	ProductUuid              *string  `json:"product_uuid" db:"product_uuid"`
	CustomerID               *uint    `json:"customer_id" db:"customer_id"`
	PoID                     *uint    `json:"po_id" db:"po_id"`
	ItemUnitID               *uint    `json:"item_unit_id" db:"item_unit_id"`
	VatID                    *uint    `json:"vat_id" db:"vat_id"`
	Pph23ID                  *uint    `json:"pph23_id" db:"pph23_id"`
	RefID                    *uint    `json:"ref_id" db:"ref_id"`
	ProductID                *uint    `json:"product_id" db:"product_id"`
	BomID                    *uint    `json:"bom_id" db:"bom_id"`
	RefSoDtID                *uint    `json:"ref_so_dt_id" db:"ref_so_dt_id"`
	RefSoDtBomID             *uint    `json:"ref_so_dt_bom_id" db:"ref_so_dt_bom_id"`
	RefRoDtID                *uint    `json:"ref_ro_dt_id" db:"ref_ro_dt_id"`
	RefRoDtBomID             *uint    `json:"ref_ro_dt_bom_id" db:"ref_ro_dt_bom_id"`
	RefProductID             *uint    `json:"ref_product_id" db:"ref_product_id"`
	RefProductBomID          *uint    `json:"ref_product_bom_id" db:"ref_product_bom_id"`
	ProductType              *string  `json:"product_type" db:"product_type"`
	ProductJSON              *string  `json:"product_json" db:"product_json"`
	RefType                  *string  `json:"ref_type" db:"ref_type"`
	RefJSON                  *string  `json:"ref_json" db:"ref_json"`
	GenCode                  *string  `json:"gen_code" db:"gen_code"`
	Remark                   *string  `json:"remark" db:"remark"`
	NeedQty                  *float64 `json:"need_qty" db:"need_qty"`
	Qty                      *float64 `json:"qty" db:"qty"`
	QtyPo                    *float64 `json:"qty_po" db:"qty_po"`
	Price                    *float64 `json:"price" db:"price"`
	Subtotal                 *float64 `json:"subtotal" db:"subtotal"`
	DiscountAmount           *float64 `json:"discount_amount" db:"discount_amount"`
	DiscountPercentage       *float64 `json:"discount_percentage" db:"discount_percentage"`
	DiscountPercentageNum    *float64 `json:"discount_percentage_num" db:"discount_percentage_num"`
	DiscountPercentageAmount *float64 `json:"discount_percentage_amount" db:"discount_percentage_amount"`
	DiscountFinal            *float64 `json:"discount_final" db:"discount_final"`
	DiscountType             *string  `json:"discount_type" db:"discount_type"`
	IsVat                    *int8    `json:"is_vat" db:"is_vat"`
	IsPph23                  *int8    `json:"is_pph23" db:"is_pph23"`
	TotalAmount              *float64 `json:"total_amount" db:"total_amount"`

	ProductCode *string `json:"product_code" db:"product_code"`
	UnitName    *string `json:"unit_name" db:"unit_name"`
	ProductName *string `json:"product_name" db:"product_name"`

	CreatedByName *string   `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string   `json:"updated_by_name" db:"updated_by_name"`
	CreatedByID   *uint     `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint     `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint     `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *string   `json:"updated_at" db:"updated_at"`
	DeleteAt      *string   `json:"deleted_at" db:"deleted_at"`

	RefQty *float64 `json:"ref_qty" db:"ref_qty"`
	RefNum *string  `json:"ref_num" db:"ref_num"`
}

type PurchaseOrderPoDtListUpdateDTO struct {
	ID                       *uint    `json:"id" db:"id"`
	PoDtID                   *uint    `json:"po_dt_id" db:"po_dt_id"`
	ProductUuid              *string  `json:"product_uuid" db:"product_uuid"`
	PoID                     *uint    `json:"po_id" db:"po_id"`
	ItemUnitID               *uint    `json:"item_unit_id" db:"item_unit_id"`
	VatID                    *uint    `json:"vat_id" db:"vat_id"`
	Pph23ID                  *uint    `json:"pph23_id" db:"pph23_id"`
	RefID                    *uint    `json:"ref_id" db:"ref_id"`
	ProductID                *uint    `json:"product_id" db:"product_id"`
	BomID                    *uint    `json:"bom_id" db:"bom_id"`
	ProductType              *string  `json:"product_type" db:"product_type"`
	ProductJSON              *string  `json:"product_json" db:"product_json"`
	RefType                  *string  `json:"ref_type" db:"ref_type"`
	RefJSON                  *string  `json:"ref_json" db:"ref_json"`
	GenCode                  *string  `json:"gen_code" db:"gen_code"`
	Remark                   *string  `json:"remark" db:"remark"`
	NeedQty                  *float64 `json:"need_qty" db:"need_qty"`
	Qty                      *float64 `json:"qty" db:"qty"`
	Price                    *float64 `json:"price" db:"price"`
	Subtotal                 *float64 `json:"subtotal" db:"subtotal"`
	DiscountAmount           *float64 `json:"discount_amount" db:"discount_amount"`
	DiscountPercentage       *float64 `json:"discount_percentage" db:"discount_percentage"`
	DiscountPercentageNum    *float64 `json:"discount_percentage_num" db:"discount_percentage_num"`
	DiscountPercentageAmount *float64 `json:"discount_percentage_amount" db:"discount_percentage_amount"`
	DiscountFinal            *float64 `json:"discount_final" db:"discount_final"`
	DiscountType             *string  `json:"discount_type" db:"discount_type"`
	IsVat                    *int8    `json:"is_vat" db:"is_vat"`
	IsPph23                  *int8    `json:"is_pph23" db:"is_pph23"`
	TotalAmount              *float64 `json:"total_amount" db:"total_amount"`

	ItemName    *string `json:"item_name" db:"item_name"`
	ItemCode    *string `json:"item_code" db:"item_code"`
	UnitName    *string `json:"unit_name" db:"unit_name"`
	ProductName *string `json:"product_name" db:"product_name"`

	CreatedByName *string   `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string   `json:"updated_by_name" db:"updated_by_name"`
	CreatedByID   *uint     `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint     `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint     `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *string   `json:"updated_at" db:"updated_at"`
	DeleteAt      *string   `json:"deleted_at" db:"deleted_at"`
}

type GetPurchaseOrdersResult struct {
	PurchaseOrders []PurchaseOrderListDTO
	Total          int
	Err            error
}

type UpdatePurchaseOrderStatusRequest struct {
	ID     uint   `json:"id"`
	Status string `json:"status"`
}

type PurchaseOrderStatusWidget struct {
	Status     string  `json:"status" db:"status"`
	OrderCount int     `json:"order_count" db:"order_count"`
	TotalQty   float64 `json:"total_qty" db:"total_qty"`
	GrandTotal float64 `json:"grand_total" db:"grand_total"`
}

type RefPoIndexSoDtListDTO struct {
	// ID               *uint    `json:"id" db:"id"`
	ProductUuid      *string  `json:"product_uuid" db:"product_uuid"`
	SalesOrderID     *uint    `json:"sales_order_id" db:"sales_order_id"`
	CustomerID       *uint    `json:"customer_id" db:"customer_id"`
	ItemUnitID       *uint    `json:"item_unit_id" db:"item_unit_id"`
	ItemID           *uint    `json:"item_id" db:"item_id"`
	ItemSubGroupName *string  `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName    *string  `json:"item_group_name" db:"item_group_name"`
	ItemName         *string  `json:"item_name" db:"item_name"`
	ItemCode         *string  `json:"item_code" db:"item_code"`
	UnitName         *string  `json:"unit_name" db:"unit_name"`
	GenCode          *string  `json:"gen_code" db:"gen_code"`
	Remark           *string  `json:"remark" db:"remark"`
	QtyPo            *float64 `json:"qty_po" db:"qty_po"`
	PriceSell        *float64 `json:"price_sell" db:"price_sell"`
	PriceBuy         *float64 `json:"price_buy" db:"price_buy"`
	SubtotalSell     *float64 `json:"subtotal_sell" db:"subtotal_sell"`
	SubtotalBuy      *float64 `json:"subtotal_buy" db:"subtotal_buy"`
	VatPerc          *float64 `json:"vat_perc" db:"vat_perc"`
	VatPercAm        *float64 `json:"vat_perc_am" db:"vat_perc_am"`
	Pph23Perc        *float64 `json:"pph23_perc" db:"pph23_perc"`
	Pph23PercAm      *float64 `json:"pph23_perc_am" db:"pph23_perc_am"`
	IsVat            *int8    `json:"is_vat" db:"is_vat"`
	IsPph23          *int8    `json:"is_pph23" db:"is_pph23"`

	CustomerName *string  `json:"customer_name" db:"customer_name"`
	RefType      *string  `json:"ref_type" db:"ref_type"`
	RefSoDtID    *uint    `json:"ref_so_dt_id" db:"ref_so_dt_id"`
	RefSoDtBomID *uint    `json:"ref_so_dt_bom_id" db:"ref_so_dt_bom_id"`
	RefPoDtID    *uint    `json:"ref_po_dt_id" db:"ref_po_dt_id"`
	RefPoDtBomID *uint    `json:"ref_po_dt_bom_id" db:"ref_po_dt_bom_id"`
	RefInvDtID   *uint    `json:"ref_inv_dt_id" db:"ref_inv_dt_id"`
	RefProductID *uint    `json:"ref_product_id" db:"ref_product_id"`
	RefNum       *string  `json:"ref_num" db:"ref_num"`
	OrderAt      *string  `json:"order_at" db:"order_at"`
	ShippingAt   *string  `json:"shipping_at" db:"shipping_at"`
	AgreeAt      *string  `json:"agree_at" db:"agree_at"`
	DueAt        *string  `json:"due_at" db:"due_at"`
	ItemSku      *string  `json:"item_sku" db:"item_sku"`
	RefQty       *float64 `json:"ref_qty" db:"ref_qty"`
	ItemType     *string  `json:"item_type" db:"item_type"`
	Balance      *float64 `json:"balance" db:"balance"`
}

type RefPoIndexRoDtListDTO struct {
	// ID               *uint    `json:"id" db:"id"`
	RefRoDtID        *uint    `json:"ref_ro_dt_id" db:"ref_ro_dt_id"`
	RequestOrderID   *uint    `json:"request_order_id" db:"request_order_id"`
	CustomerID       *uint    `json:"customer_id" db:"customer_id"`
	ItemUnitID       *uint    `json:"item_unit_id" db:"item_unit_id"`
	ItemID           *uint    `json:"item_id" db:"item_id"`
	ItemSubGroupName *string  `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName    *string  `json:"item_group_name" db:"item_group_name"`
	ItemName         *string  `json:"item_name" db:"item_name"`
	ItemCode         *string  `json:"item_code" db:"item_code"`
	UnitName         *string  `json:"unit_name" db:"unit_name"`
	Remark           *string  `json:"remark" db:"remark"`
	QtyPo            *float64 `json:"qty_po" db:"qty_po"`
	PriceSell        *float64 `json:"price_sell" db:"price_sell"`
	PriceBuy         *float64 `json:"price_buy" db:"price_buy"`
	SubtotalSell     *float64 `json:"subtotal_sell" db:"subtotal_sell"`
	SubtotalBuy      *float64 `json:"subtotal_buy" db:"subtotal_buy"`

	RefID        *uint    `json:"ref_id" db:"ref_id"`
	CustomerName *string  `json:"customer_name" db:"customer_name"`
	RefType      *string  `json:"ref_type" db:"ref_type"`
	RefProductID *uint    `json:"ref_product_id" db:"ref_product_id"`
	RefNum       *string  `json:"ref_num" db:"ref_num"`
	RequestDate  *string  `json:"request_date" db:"request_date"`
	ItemSku      *string  `json:"item_sku" db:"item_sku"`
	RefQty       *float64 `json:"ref_qty" db:"ref_qty"`
	ItemType     *string  `json:"item_type" db:"item_type"`
	Balance      *float64 `json:"balance" db:"balance"`
}
