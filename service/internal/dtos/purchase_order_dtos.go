package dtos

import "encoding/json"

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
	ItemUnitID            *uint                   `json:"item_unit_id"`
	VatID                 *uint                   `json:"vat_id"`
	RefID                 *uint                   `json:"ref_id"`
	ProductID             uint                    `json:"product_id"`
	ProductType           *string                 `json:"product_type"`
	ProductJSON           *json.RawMessage        `json:"product_json"`
	RefType               *string                 `json:"ref_type"`
	RefJSON               *json.RawMessage        `json:"ref_json"`
	Remark                *string                 `json:"remark"`
	NeedQty               *float64                `json:"need_qty"`
	Qty                   *float64                `json:"qty"`
	Price                 *float64                `json:"price"`
	Subtotal              *float64                `json:"subtotal"`
	DiscountAmount        *float64                `json:"discount_amount"`
	DiscountPercentage    *float64                `json:"discount_percentage"`
	DiscountPercentageNum *float64                `json:"discount_percentage_num"`
	DiscountPercentageAm  *float64                `json:"discount_percentage_amount"`
	DiscountFinal         *float64                `json:"discount_final"`
	VatPercentage         *float64                `json:"vat_percentage"`
	VatPercentageAmount   *float64                `json:"vat_percentage_amount"`
	TotalAmount           *float64                `json:"total_amount"`
	PoDtBoms              []CreatePoDtBomsRequest `json:"po_dt_boms"`
}

type CreatePoDtBomsRequest struct {
	ProductID   uint             `json:"product_id"`
	BomID       uint             `json:"bom_id"`
	ItemUnitID  *uint            `json:"item_unit_id"`
	ProductJSON *json.RawMessage `json:"product_json"`
	Remark      *string          `json:"remark"`
	Qty         *float64         `json:"qty"`
	Price       *float64         `json:"price"`
	Subtotal    *float64         `json:"subtotal"`
}

type CreatePurchaseOrderRequest struct {
	CustomerID               *uint                `json:"customer_id"`
	PurchaseTypeID           *uint                `json:"purchase_type_id"`
	CurrencyID               *uint                `json:"currency_id"`
	VatID                    *uint                `json:"vat_id"`
	VatPercentage            *float64             `json:"vat_percentage"`
	VatPercentageAmount      *float64             `json:"vat_percentage_amount"`
	PaymentTermID            *uint                `json:"payment_term_id"`
	ShippingTermID           *uint                `json:"shipping_term_id"`
	Pph23ID                  *uint                `json:"pph23_id"`
	Pph23Percentage          *float64             `json:"pph23_percentage"`
	Pph23PercentageAmount    *float64             `json:"pph23_percentage_amount"`
	BranchID                 *uint                `json:"branch_id"`
	PoNo                     *string              `json:"po_no"`
	PoDate                   *string              `json:"po_date"`
	DeliveryDate             *string              `json:"delivery_date"`
	ShippingDestination      *string              `json:"shipping_destination"`
	Remark                   *string              `json:"remark"`
	ExchangeRate             *float64             `json:"exchange_rate"`
	DiscountAmount           *float64             `json:"discount_amount"`
	DiscountPercentage       *float64             `json:"discount_percentage"`
	DiscountPercentageAmount *float64             `json:"discount_percentage_amount"`
	DiscountFinalHeader      *float64             `json:"discount_final_header"`
	DiscountAmountProduct    *float64             `json:"discount_amount_product"`
	Subtotal                 *float64             `json:"subtotal"`
	TotalAmountProduct       *float64             `json:"total_amount_product"`
	TotalQty                 *float64             `json:"total_qty"`
	TotalDiscount            *float64             `json:"total_discount"`
	TotalPph23               *float64             `json:"total_pph23"`
	TotalVat                 *float64             `json:"total_vat"`
	GrandTotal               *float64             `json:"grand_total"`
	Status                   string               `json:"status"`
	PoDts                    []CreatePoDtsRequest `json:"po_dts"`
}
