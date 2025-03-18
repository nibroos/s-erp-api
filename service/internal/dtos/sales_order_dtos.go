package dtos

import "time"

type GetSalesOrdersRequest struct {
	Global         *string `json:"global"`
	Title          *string `json:"title"`
	PoBuyerNo      *string `json:"po_buyer_no"`
	SalesOrderNo   *string `json:"sales_order_no"`
	Remark         *string `json:"remark"`
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

type CreateSoDtsRequest struct {
	ProductUuid  string                   `json:"product_uuid"`
	ItemUnitID   *uint                    `json:"item_unit_id"`
	VatID        *uint                    `json:"vat_id"`
	RefID        *uint                    `json:"ref_id"`
	ItemID       uint                     `json:"item_id"`
	RefType      *string                  `json:"ref_type"`
	ItemType     *string                  `json:"item_type"`
	GenCode      *string                  `json:"gen_code"`
	Remark       *string                  `json:"remark"`
	VatPerc      *float64                 `json:"vat_perc"`
	VatPercAm    *float64                 `json:"vat_perc_am"`
	Qty          *float64                 `json:"qty"`
	PriceSell    *float64                 `json:"price_sell"`
	PriceBuy     *float64                 `json:"price_buy"`
	SubtotalSell *float64                 `json:"subtotal_sell"`
	SubtotalBuy  *float64                 `json:"subtotal_buy"`
	DiscAm       *float64                 `json:"disc_am"`
	DiscPerc     *float64                 `json:"disc_perc"`
	DiscPercNum  *float64                 `json:"disc_perc_num"`
	DiscPercAm   *float64                 `json:"disc_perc_am"`
	DiscFinal    *float64                 `json:"disc_final"`
	DiscType     *string                  `json:"disc_type"`
	TotalAm      *float64                 `json:"total_am"`
	SoDtsBoms    []CreateSoDtsBomsRequest `json:"so_dts_boms"`
}
type CreateSoDtsBomsRequest struct {
	ProductUuid  string  `json:"product_uuid"`
	ProductID    uint    `json:"product_id"`
	ItemID       uint    `json:"item_id"`
	ItemUnitID   *uint   `json:"item_unit_id"`
	GenCode      *string `json:"gen_code"`
	Remark       *string `json:"remark"`
	Qty          float64 `json:"qty"`
	PriceSell    float64 `json:"price_sell"`
	PriceBuy     float64 `json:"price_buy"`
	SubtotalSell float64 `json:"subtotal_sell"`
	SubtotalBuy  float64 `json:"subtotal_buy"`
}

type CreateSalesOrderRequest struct {
	CustomerID    *uint                `json:"customer_id"`
	OrderTypeID   *uint                `json:"order_type_id"`
	CurrencyID    *uint                `json:"currency_id"`
	WarehouseID   *uint                `json:"warehouse_id"`
	VatID         *uint                `json:"vat_id"`
	PaymentID     *uint                `json:"payment_id"`
	Pph23ID       *uint                `json:"pph23_id"`
	BranchID      *uint                `json:"branch_id"`
	PoBuyerNo     string               `json:"po_buyer_no"`
	SalesOrderNo  *string              `json:"sales_order_no"`
	Remark        *string              `json:"remark"`
	ShipDest      *string              `json:"ship_dest"`
	Status        string               `json:"status"`
	ExchangeRate  *float64             `json:"exchange_rate"`
	VatPerc       *float64             `json:"vat_perc"`
	DiscAm        *float64             `json:"disc_am"`
	DiscPerc      *float64             `json:"disc_perc"`
	DiscPercAm    *float64             `json:"disc_perc_am"`
	DiscPercNum   *float64             `json:"disc_perc_num"`
	DiscFinal     *float64             `json:"disc_final"`
	DiscType      *string              `json:"disc_type"`
	Pph23Perc     *float64             `json:"pph23_perc"`
	TotalQty      *float64             `json:"total_qty"`
	Subtotal      *float64             `json:"subtotal"`
	TotalDiscount *float64             `json:"total_discount"`
	TotalPph23    *float64             `json:"total_pph23"`
	TotalVat      *float64             `json:"total_vat"`
	GrandTotal    *float64             `json:"grand_total"`
	OrderAt       *string              `json:"order_at"`
	ShippingAt    *string              `json:"shipping_at"`
	AgreeAt       *string              `json:"agree_at"`
	DueAt         *string              `json:"due_at"`
	ExpiredAt     *string              `json:"expired_at"`
	SoDts         []CreateSoDtsRequest `json:"so_dts"`
}

type UpdateSoDtsRequest struct {
	ID           *uint                     `json:"id"`
	SoDtID       *uint                     `json:"so_dt_id"`
	ProductUuid  string                    `json:"product_uuid"`
	SalesOrderID *uint                     `json:"sales_order_id"`
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
	SoDtsBoms    []*UpdateSoDtsBomsRequest `json:"so_dts_boms" gorm:"-"`
}

type UpdateSoDtsBomsRequest struct {
	ID           *uint   `json:"id"`
	SalesOrderID *uint   `json:"sales_order_id"`
	SoDtID       *uint   `json:"so_dt_id"`
	SoDtBomID    *uint   `json:"so_dt_bom_id"`
	ProductUuid  *string `json:"product_uuid"`
	ProductID    uint    `json:"product_id"`
	ItemID       *uint   `json:"item_id"`
	ItemUnitID   *uint   `json:"item_unit_id"`
	GenCode      *string `json:"gen_code"`
	Remark       *string `json:"remark"`
	Qty          float64 `json:"qty"`
	PriceSell    float64 `json:"price_sell"`
	PriceBuy     float64 `json:"price_buy"`
	SubtotalSell float64 `json:"subtotal_sell"`
	SubtotalBuy  float64 `json:"subtotal_buy"`
}

type UpdateSalesOrderRequest struct {
	ID            uint                 `json:"id"`
	SalesOrderID  *uint                `json:"sales_order_id"`
	CustomerID    *uint                `json:"customer_id"`
	OrderTypeID   *uint                `json:"order_type_id"`
	CurrencyID    *uint                `json:"currency_id"`
	WarehouseID   *uint                `json:"warehouse_id"`
	VatID         *uint                `json:"vat_id"`
	PaymentID     *uint                `json:"payment_id"`
	Pph23ID       *uint                `json:"pph23_id"`
	BranchID      *uint                `json:"branch_id"`
	PoBuyerNo     string               `json:"po_buyer_no"`
	SalesOrderNo  *string              `json:"sales_order_no"`
	Remark        *string              `json:"remark"`
	ShipDest      *string              `json:"ship_dest"`
	Status        string               `json:"status"`
	ExchangeRate  *float64             `json:"exchange_rate"`
	VatPerc       *float64             `json:"vat_perc"`
	DiscAm        *float64             `json:"disc_am"`
	DiscPerc      *float64             `json:"disc_perc"`
	DiscPercAm    *float64             `json:"disc_perc_am"`
	DiscPercNum   *float64             `json:"disc_perc_num"`
	DiscFinal     *float64             `json:"disc_final"`
	DiscType      *string              `json:"disc_type"`
	Pph23Perc     *float64             `json:"pph23_perc"`
	TotalQty      *float64             `json:"total_qty"`
	Subtotal      *float64             `json:"subtotal"`
	TotalDiscount *float64             `json:"total_discount"`
	TotalPph23    *float64             `json:"total_pph23"`
	TotalVat      *float64             `json:"total_vat"`
	GrandTotal    *float64             `json:"grand_total"`
	OrderAt       *string              `json:"order_at"`
	ShippingAt    *string              `json:"shipping_at"`
	AgreeAt       *string              `json:"agree_at"`
	DueAt         *string              `json:"due_at"`
	ExpiredAt     *string              `json:"expired_at"`
	SoDts         []UpdateSoDtsRequest `json:"so_dts"`
}

type GetSalesOrderByIDRequest struct {
	ID uint `json:"id"`
}

type GetSalesOrderParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetSalesOrderParams(id uint) *GetSalesOrderParams {
	defaultIsDeleted := 0
	return &GetSalesOrderParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type GetSalesOrderSoDtParams struct {
	ID           uint
	SalesOrderID uint
	IsDeleted    *int
}

type DeleteSalesOrderRequest struct {
	ID uint `json:"id"`
}

type SalesOrderListDTO struct {
	ID            int      `json:"id" db:"id"`
	SalesOrderID  *uint    `json:"sales_order_id" db:"sales_order_id"`
	CustomerID    *uint    `json:"customer_id" db:"customer_id"`
	OrderTypeID   *uint    `json:"order_type_id" db:"order_type_id"`
	CurrencyID    *uint    `json:"currency_id" db:"currency_id"`
	WarehouseID   *uint    `json:"warehouse_id" db:"warehouse_id"`
	VatID         *uint    `json:"vat_id" db:"vat_id"`
	PaymentID     *uint    `json:"payment_id" db:"payment_id"`
	Pph23ID       *uint    `json:"pph23_id" db:"pph23_id"`
	BranchID      *uint    `json:"branch_id" db:"branch_id"`
	SalesOrderNo  *string  `json:"sales_order_no" db:"sales_order_no"`
	PoBuyerNo     string   `json:"po_buyer_no" db:"po_buyer_no"`
	ShipDest      *string  `json:"ship_dest" db:"ship_dest"`
	Remark        *string  `json:"remark" db:"remark"`
	Status        string   `json:"status" db:"status"`
	ExchangeRate  *float64 `json:"exchange_rate" db:"exchange_rate"`
	VatPerc       *float64 `json:"vat_perc" db:"vat_perc"`
	Pph23Perc     *float64 `json:"pph23_perc" db:"pph23_perc"`
	DiscAm        *float64 `json:"disc_am" db:"disc_am"`
	DiscPerc      *float64 `json:"disc_perc" db:"disc_perc"`
	DiscPercAm    *float64 `json:"disc_perc_am" db:"disc_perc_am"`
	DiscFinal     *float64 `json:"disc_final" db:"disc_final"`
	DiscType      *string  `json:"disc_type" db:"disc_type"`
	TotalQty      *float64 `json:"total_qty" db:"total_qty"`
	Subtotal      *float64 `json:"subtotal" db:"subtotal"`
	QtyOut        *float64 `json:"qty_out" db:"qty_out"`
	TotalDiscount *float64 `json:"total_discount" db:"total_discount"`
	TotalPph23    *float64 `json:"total_pph23" db:"total_pph23"`
	TotalVat      *float64 `json:"total_vat" db:"total_vat"`
	GrandTotal    *float64 `json:"grand_total" db:"grand_total"`
	SiTotalAm     *float64 `json:"si_total_am" db:"si_total_am"`
	SaTotalAm     *float64 `json:"sa_total_am" db:"sa_total_am"`
	OrderAt       *string  `json:"order_at" db:"order_at"`
	ShippingAt    *string  `json:"shipping_at" db:"shipping_at"`
	AgreeAt       *string  `json:"agree_at" db:"agree_at"`
	DueAt         *string  `json:"due_at" db:"due_at"`
	ExpiredAt     *string  `json:"expired_at" db:"expired_at"`
	CreatedByID   *uint    `json:"crweated_by_id" db:"created_by_id"`
	UpdatedByID   *uint    `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint    `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName *string  `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string  `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string  `json:"created_at" db:"created_at"`
	UpdatedAt     *string  `json:"updated_at" db:"updated_at"`
	DeleteAt      *string  `json:"deleted_at" db:"deleted_at"`

	// so_dt_vat_id, currency_name vat_name pph23_name
	ProductID      *string `json:"product_id" db:"product_id"`
	ItemID         *string `json:"item_id" db:"item_id"`
	SoDtVatID      *string `json:"so_dt_vat_id" db:"so_dt_vat_id"`
	CurrencyName   *string `json:"currency_name" db:"currency_name"`
	WarehouseName  *string `json:"warehouse_name" db:"warehouse_name"`
	OrderTypeName  *string `json:"order_type_name" db:"order_type_name"`
	CustomerName   *string `json:"customer_name" db:"customer_name"`
	ProductName    *string `json:"product_name" db:"product_name"`
	ItemName       *string `json:"item_name" db:"item_name"`
	VatName        *string `json:"vat_name" db:"vat_name"`
	Pph23Name      *string `json:"pph23_name" db:"pph23_name"`
	SoDtRemark     *string `json:"so_dt_remark" db:"so_dt_remark"`
	SoDtGenCode    *string `json:"so_dt_gen_code" db:"so_dt_gen_code"`
	SoDtBomGenCode *string `json:"so_dt_bom_gen_code" db:"so_dt_bom_gen_code"`
	SoDtBomRemark  *string `json:"so_dt_bom_remark" db:"so_dt_bom_remark"`
}

type SalesOrderDetailDTO struct {
	ID            uint     `json:"id" db:"id"`
	SalesOrderID  *uint    `json:"sales_order_id" db:"sales_order_id"`
	CustomerID    *uint    `json:"customer_id" db:"customer_id"`
	OrderTypeID   *uint    `json:"order_type_id" db:"order_type_id"`
	CurrencyID    *uint    `json:"currency_id" db:"currency_id"`
	WarehouseID   *uint    `json:"warehouse_id" db:"warehouse_id"`
	VatID         *uint    `json:"vat_id" db:"vat_id"`
	PaymentID     *uint    `json:"payment_id" db:"payment_id"`
	Pph23ID       *uint    `json:"pph23_id" db:"pph23_id"`
	BranchID      *uint    `json:"branch_id" db:"branch_id"`
	SalesOrderNo  *string  `json:"sales_order_no" db:"sales_order_no"`
	PoBuyerNo     string   `json:"po_buyer_no" db:"po_buyer_no"`
	ShipDest      *string  `json:"ship_dest" db:"ship_dest"`
	Remark        *string  `json:"remark" db:"remark"`
	Status        string   `json:"status" db:"status"`
	IsApproved    int8     `json:"is_approved" db:"is_approved"`
	ExchangeRate  float64  `json:"exchange_rate" db:"exchange_rate"`
	VatPerc       *float64 `json:"vat_perc" db:"vat_perc"`
	Pph23Perc     *float64 `json:"pph23_perc" db:"pph23_perc"`
	DiscAm        *float64 `json:"disc_am" db:"disc_am"`
	DiscPerc      *float64 `json:"disc_perc" db:"disc_perc"`
	DiscPercAm    *float64 `json:"disc_perc_am" db:"disc_perc_am"`
	DiscFinal     *float64 `json:"disc_final" db:"disc_final"`
	DiscType      *string  `json:"disc_type" db:"disc_type"`
	TotalQty      float64  `json:"total_qty" db:"total_qty"`
	QtyOut        *float64 `json:"qty_out" db:"qty_out"`
	Subtotal      float64  `json:"subtotal" db:"subtotal"`
	TotalDiscount float64  `json:"total_discount" db:"total_discount"`
	TotalPph23    float64  `json:"total_pph23" db:"total_pph23"`
	TotalVat      float64  `json:"total_vat" db:"total_vat"`
	GrandTotal    float64  `json:"grand_total" db:"grand_total"`
	OrderAt       *string  `json:"order_at" db:"order_at"`
	ShippingAt    *string  `json:"shipping_at" db:"shipping_at"`
	AgreeAt       *string  `json:"agree_at" db:"agree_at"`
	DueAt         *string  `json:"due_at" db:"due_at"`
	ExpiredAt     *string  `json:"expired_at" db:"expired_at"`

	SiTotalAm *float64 `json:"si_total_am" db:"si_total_am"`
	SaTotalAm *float64 `json:"sa_total_am" db:"sa_total_am"`

	CreatedByID   *uint                   `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint                   `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint                   `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName *string                 `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string                 `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string                 `json:"created_at" db:"created_at"`
	UpdatedAt     *string                 `json:"updated_at" db:"updated_at"`
	DeleteAt      *string                 `json:"deleted_at" db:"deleted_at"`
	SoDts         []SalesOrderSoDtListDTO `json:"so_dts"`
}

type SalesOrderSoDtListDTO struct {
	ID               *uint    `json:"id" db:"id"`
	SoDtID           *uint    `json:"so_dt_id" db:"so_dt_id"`
	ProductUuid      *string  `json:"product_uuid" db:"product_uuid"`
	SalesOrderID     *uint    `json:"sales_order_id" db:"sales_order_id"`
	ItemUnitID       *uint    `json:"item_unit_id" db:"item_unit_id"`
	VatID            *uint    `json:"vat_id" db:"vat_id"`
	RefID            *uint    `json:"ref_id" db:"ref_id"`
	ItemID           *uint    `json:"item_id" db:"item_id"`
	ItemSubGroupID   *uint    `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID      *uint    `json:"item_group_id" db:"item_group_id"`
	ItemSubGroupName *string  `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName    *string  `json:"item_group_name" db:"item_group_name"`
	ItemName         *string  `json:"item_name" db:"item_name"`
	ItemCode         *string  `json:"item_code" db:"item_code"`
	UnitName         *string  `json:"unit_name" db:"unit_name"`
	RefJSON          *string  `json:"ref_json" db:"ref_json"`
	RefType          *string  `json:"ref_type" db:"ref_type"`
	ItemType         *string  `json:"item_type" db:"item_type"`
	GenCode          *string  `json:"gen_code" db:"gen_code"`
	Remark           *string  `json:"remark" db:"remark"`
	VatPerc          *float64 `json:"vat_perc" db:"vat_perc"`
	VatPercAm        *float64 `json:"vat_perc_am" db:"vat_perc_am"`
	QtyOut           *float64 `json:"qty_out" db:"qty_out"`
	Qty              *float64 `json:"qty" db:"qty"`
	PriceSell        *float64 `json:"price_sell" db:"price_sell"`
	PriceBuy         *float64 `json:"price_buy" db:"price_buy"`
	SubtotalSell     *float64 `json:"subtotal_sell" db:"subtotal_sell"`
	SubtotalBuy      *float64 `json:"subtotal_buy" db:"subtotal_buy"`
	DiscAm           *float64 `json:"disc_am" db:"disc_am"`
	DiscPerc         *float64 `json:"disc_perc" db:"disc_perc"`
	DiscPercNum      *float64 `json:"disc_perc_num" db:"disc_perc_num"`
	DiscPercAm       *float64 `json:"disc_perc_am" db:"disc_perc_am"`
	DiscFinal        *float64 `json:"disc_final" db:"disc_final"`
	DiscType         *string  `json:"disc_type" db:"disc_type"`
	TotalAm          *float64 `json:"total_am" db:"total_am"`

	SiTotalAm *float64 `json:"si_total_am" db:"si_total_am"`
	SaTotalAm *float64 `json:"sa_total_am" db:"sa_total_am"`

	CreatedByName *string   `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string   `json:"updated_by_name" db:"updated_by_name"`
	CreatedByID   *uint     `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint     `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint     `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *string   `json:"updated_at" db:"updated_at"`
	DeleteAt      *string   `json:"deleted_at" db:"deleted_at"`

	SoDtsBoms []SalesOrderSoDtBomListDTO `json:"so_dts_boms"`
}

type SalesOrderSoDtListUpdateDTO struct {
	ID               *uint    `json:"id" db:"id"`
	SoDtID           *uint    `json:"so_dt_id" db:"so_dt_id"`
	ProductUuid      *string  `json:"product_uuid" db:"product_uuid"`
	SalesOrderID     *uint    `json:"sales_order_id" db:"sales_order_id"`
	ItemUnitID       *uint    `json:"item_unit_id" db:"item_unit_id"`
	VatID            *uint    `json:"vat_id" db:"vat_id"`
	RefID            *uint    `json:"ref_id" db:"ref_id"`
	ItemID           *uint    `json:"item_id" db:"item_id"`
	ItemSubGroupID   *uint    `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID      *uint    `json:"item_group_id" db:"item_group_id"`
	ItemSubGroupName *string  `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName    *string  `json:"item_group_name" db:"item_group_name"`
	ItemName         *string  `json:"item_name" db:"item_name"`
	ItemCode         *string  `json:"item_code" db:"item_code"`
	UnitName         *string  `json:"unit_name" db:"unit_name"`
	RefJSON          *string  `json:"ref_json" db:"ref_json"`
	RefType          *string  `json:"ref_type" db:"ref_type"`
	ItemType         *string  `json:"item_type" db:"item_type"`
	GenCode          *string  `json:"gen_code" db:"gen_code"`
	Remark           *string  `json:"remark" db:"remark"`
	VatPerc          *float64 `json:"vat_perc" db:"vat_perc"`
	VatPercAm        *float64 `json:"vat_perc_am" db:"vat_perc_am"`
	QtyOut           *float64 `json:"qty_out" db:"qty_out"`
	Qty              *float64 `json:"qty" db:"qty"`
	PriceSell        *float64 `json:"price_sell" db:"price_sell"`
	PriceBuy         *float64 `json:"price_buy" db:"price_buy"`
	SubtotalSell     *float64 `json:"subtotal_sell" db:"subtotal_sell"`
	SubtotalBuy      *float64 `json:"subtotal_buy" db:"subtotal_buy"`
	DiscAm           *float64 `json:"disc_am" db:"disc_am"`
	DiscPerc         *float64 `json:"disc_perc" db:"disc_perc"`
	DiscPercNum      *float64 `json:"disc_perc_num" db:"disc_perc_num"`
	DiscPercAm       *float64 `json:"disc_perc_am" db:"disc_perc_am"`
	DiscFinal        *float64 `json:"disc_final" db:"disc_final"`
	DiscType         *string  `json:"disc_type" db:"disc_type"`
	TotalAm          *float64 `json:"total_am" db:"total_am"`

	SiTotalAm *float64 `json:"si_total_am" db:"si_total_am"`
	SaTotalAm *float64 `json:"sa_total_am" db:"sa_total_am"`

	CreatedByName *string   `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string   `json:"updated_by_name" db:"updated_by_name"`
	CreatedByID   *uint     `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint     `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint     `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *string   `json:"updated_at" db:"updated_at"`
	DeleteAt      *string   `json:"deleted_at" db:"deleted_at"`
}

type SalesOrderSoDtBomListDTO struct {
	ID                *uint    `json:"id" db:"id"`
	SoDtBomID         *uint    `json:"so_dt_bom_id" db:"so_dt_bom_id"`
	SalesOrderID      *uint    `json:"sales_order_id" db:"sales_order_id"`
	SoDtID            *uint    `json:"so_dt_id" db:"so_dt_id"`
	ProductID         uint     `json:"product_id" db:"product_id"`
	ProductUuid       string   `json:"product_uuid" db:"product_uuid"`
	ItemID            uint     `json:"item_id" db:"item_id"`
	ItemSubGroupID    *uint    `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID       *uint    `json:"item_group_id" db:"item_group_id"`
	ItemSubGroupName  *string  `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName     *string  `json:"item_group_name" db:"item_group_name"`
	ItemName          *string  `json:"item_name" db:"item_name"`
	ItemCode          *string  `json:"item_code" db:"item_code"`
	ItemBarcode       *string  `json:"item_barcode" db:"item_barcode"`
	ItemSku           *string  `json:"item_sku" db:"item_sku"`
	ItemFactoryCode   *string  `json:"item_factory_code" db:"item_factory_code"`
	ItemSpecification *string  `json:"item_specification" db:"item_specification"`
	ItemQtyStock      *string  `json:"item_qty_stock" db:"item_qty_stock"`
	UnitName          *string  `json:"unit_name" db:"unit_name"`
	ItemUnitID        *uint    `json:"item_unit_id" db:"item_unit_id"`
	RefJSON           *string  `json:"ref_json" db:"ref_json"`
	GenCode           *string  `json:"gen_code" db:"gen_code"`
	Remark            *string  `json:"remark" db:"remark"`
	QtyOut            *float64 `json:"qty_out" db:"qty_out"`
	Qty               float64  `json:"qty" db:"qty"`
	PriceSell         float64  `json:"price_sell" db:"price_sell"`
	PriceBuy          float64  `json:"price_buy" db:"price_buy"`
	SubtotalSell      float64  `json:"subtotal_sell" db:"subtotal_sell"`
	SubtotalBuy       float64  `json:"subtotal_buy" db:"subtotal_buy"`

	CreatedByID   *uint     `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint     `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint     `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName *string   `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string   `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *string   `json:"updated_at" db:"updated_at"`
	DeletedAt     *string   `json:"deleted_at" db:"deleted_at"`
}

type GetSalesOrdersResult struct {
	SalesOrders []SalesOrderListDTO
	Total       int
	Err         error
}
