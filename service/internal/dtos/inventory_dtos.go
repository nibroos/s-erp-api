package dtos

import "time"

type GetInventoriesRequest struct {
	Global         *string `json:"global"`
	Title          *string `json:"title"`
	PoBuyerNo      *string `json:"po_buyer_no"`
	InventoryNo    *string `json:"sales_order_no"`
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

type CreateInvDtsRequest struct {
	ProductUuid  string                    `json:"product_uuid"`
	ItemUnitID   uint                      `json:"item_unit_id"`
	VatID        *uint                     `json:"vat_id"`
	Pph23ID      *uint                     `json:"pph23_id"`
	RefID        uint                      `json:"ref_id"`
	ItemID       uint                      `json:"item_id"`
	RefType      string                    `json:"ref_type"`
	ItemType     string                    `json:"item_type"`
	GenCode      *string                   `json:"gen_code"`
	Remark       *string                   `json:"remark"`
	VatPerc      *float64                  `json:"vat_perc"`
	VatPercAm    *float64                  `json:"vat_perc_am"`
	Pph23Perc    *float64                  `json:"pph23_perc"`
	Pph23PercAm  *float64                  `json:"pph23_perc_am"`
	MarkupPerc   *float64                  `json:"markup_perc"`
	MarkupPercAm *float64                  `json:"markup_perc_am"`
	IsVat        *int                      `json:"is_vat"`
	IsPph23      *int                      `json:"is_pph23"`
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
	InvDtsBoms   []CreateInvDtsBomsRequest `json:"so_dts_boms"`
}
type CreateInvDtsBomsRequest struct {
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

type CreateInventoryRequest struct {
	ID             *uint    `json:"id"`
	CustomerID     *uint    `json:"customer_id"`
	IoTypeID       *uint    `json:"io_type_id"`
	CurrencyID     *uint    `json:"currency_id"`
	PaymentTermID  *uint    `json:"payment_term_id"`
	WarehouseID    *uint    `json:"warehouse_id"`
	VatID          *uint    `json:"vat_id"`
	Pph23ID        *uint    `json:"pph23_id"`
	BranchID       *uint    `json:"branch_id"`
	InventoryNo    *string  `json:"sales_order_no"`
	InventoryNoOri *string  `json:"sales_order_no_ori"`
	DoNo           *string  `json:"do_no"`
	SuratJalanNo   *string  `json:"surat_jalan_no"`
	InvoiceNo      *string  `json:"invoice_no"`
	ShipDest       *string  `json:"ship_dest"`
	Remark         *string  `json:"remark"`
	Status         string   `json:"status"`
	ExchangeRate   *float64 `json:"exchange_rate"`
	IsVat          *int     `json:"is_vat"`
	VatPerc        *float64 `json:"vat_perc"`
	Pph23Perc      *float64 `json:"pph23_perc"`
	TotalQty       *float64 `json:"total_qty"`
	Subtotal       *float64 `json:"subtotal"`
	TotalPph23     *float64 `json:"total_pph23"`
	TotalVat       *float64 `json:"total_vat"`
	GrandTotal     *float64 `json:"grand_total"`
	DoAt           *string  `json:"do_at"`
	IngoingAt      *string  `json:"ingoing_at"`
	InvoiceAt      *string  `json:"invoice_at"`

	InvDts []CreateInvDtsRequest `json:"so_dts"`

	CustomerCode string `json:"customer_code"`
}

type UpdateInvDtsRequest struct {
	ID              *uint                      `json:"id"`
	InvDtID         *uint                      `json:"so_dt_id"`
	ProductUuid     string                     `json:"product_uuid"`
	InventoryID     uint                       `json:"sales_order_id"`
	ItemUnitID      uint                       `json:"item_unit_id"`
	VatID           *uint                      `json:"vat_id"`
	Pph23ID         *uint                      `json:"pph23_id"`
	RefID           uint                       `json:"ref_id"`
	ItemID          uint                       `json:"item_id"`
	RefType         string                     `json:"ref_type"`
	ItemType        string                     `json:"item_type"`
	GenCode         *string                    `json:"gen_code"`
	Remark          *string                    `json:"remark"`
	VatPerc         *float64                   `json:"vat_perc"`
	VatPercAm       *float64                   `json:"vat_perc_am"`
	Pph23Perc       *float64                   `json:"pph23_perc"`
	Pph23PercAm     *float64                   `json:"pph23_perc_am"`
	MarkupPerc      *float64                   `json:"markup_perc"`
	MarkupPercAm    *float64                   `json:"markup_perc_am"`
	IsVat           *int                       `json:"is_vat"`
	IsPph23         *int                       `json:"is_pph23"`
	IsLockMarkup    *int                       `json:"is_lock_markup"`
	IsLockPriceSell *int                       `json:"is_lock_price_sell"`
	Qty             *float64                   `json:"qty"`
	PriceSell       *float64                   `json:"price_sell"`
	PriceBuy        *float64                   `json:"price_buy"`
	SubtotalSell    *float64                   `json:"subtotal_sell"`
	SubtotalBuy     *float64                   `json:"subtotal_buy"`
	DiscAm          *float64                   `json:"disc_am"`
	DiscPerc        *float64                   `json:"disc_perc"`
	DiscPercNum     *float64                   `json:"disc_perc_num"`
	DiscPercAm      *float64                   `json:"disc_perc_am"`
	DiscFinal       *float64                   `json:"disc_final"`
	DiscType        *string                    `json:"disc_type"`
	TotalAm         *float64                   `json:"total_am"`
	InvDtsBoms      []*UpdateInvDtsBomsRequest `json:"so_dts_boms" gorm:"-"`
}

type UpdateInvDtsBomsRequest struct {
	ID           *uint   `json:"id"`
	InventoryID  *uint   `json:"sales_order_id"`
	InvDtID      *uint   `json:"so_dt_id"`
	InvDtBomID   *uint   `json:"so_dt_bom_id"`
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

type UpdateInventoryRequest struct {
	ID             *uint                 `json:"id"`
	CustomerID     *uint                 `json:"customer_id"`
	IoTypeID       *uint                 `json:"io_type_id"`
	CurrencyID     *uint                 `json:"currency_id"`
	PaymentTermID  *uint                 `json:"payment_term_id"`
	WarehouseID    *uint                 `json:"warehouse_id"`
	VatID          *uint                 `json:"vat_id"`
	Pph23ID        *uint                 `json:"pph23_id"`
	BranchID       *uint                 `json:"branch_id"`
	InventoryNo    *string               `json:"sales_order_no"`
	InventoryNoOri *string               `json:"sales_order_no_ori"`
	DoNo           *string               `json:"do_no"`
	SuratJalanNo   *string               `json:"surat_jalan_no"`
	InvoiceNo      *string               `json:"invoice_no"`
	ShipDest       *string               `json:"ship_dest"`
	Remark         *string               `json:"remark"`
	RevNo          *int                  `json:"rev_no" db:"rev_no"`
	Status         string                `json:"status"`
	ExchangeRate   *float64              `json:"exchange_rate"`
	IsVat          *int                  `json:"is_vat"`
	VatPerc        *float64              `json:"vat_perc"`
	Pph23Perc      *float64              `json:"pph23_perc"`
	TotalQty       *float64              `json:"total_qty"`
	Subtotal       *float64              `json:"subtotal"`
	TotalPph23     *float64              `json:"total_pph23"`
	TotalVat       *float64              `json:"total_vat"`
	GrandTotal     *float64              `json:"grand_total"`
	DoAt           *string               `json:"do_at"`
	IngoingAt      *string               `json:"ingoing_at"`
	InvoiceAt      *string               `json:"invoice_at"`
	InvDts         []UpdateInvDtsRequest `json:"so_dts"`

	CustomerCode string `json:"customer_code"`
}

type UpdateInventoryAttachmentsDTO struct {
	ID       *uint   `json:"id" db:"id"`
	RefID    *uint   `json:"ref_id" db:"ref_id"`
	RefType  *string `json:"ref_type" db:"ref_type"`
	FileType *string `json:"file_type" db:"file_type"`
	FileUrl  *string `json:"file_url" db:"file_url"`
	FileName *string `json:"file_name" db:"file_name"`
	Remark   *string `json:"remark" db:"remark"`
}

type GetInventoryByIDRequest struct {
	ID interface{} `json:"id"`
}

type GetInventoryParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetInventoryParams(id uint) *GetInventoryParams {
	defaultIsDeleted := 0
	return &GetInventoryParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type GetInventoryInvDtParams struct {
	ID          uint
	InventoryID uint
	IsDeleted   *int
}

type DeleteInventoryRequest struct {
	ID uint `json:"id"`
}

type InventoryListDTO struct {
	ID            int      `json:"id" db:"id"`
	InventoryID   *uint    `json:"sales_order_id" db:"sales_order_id"`
	CustomerID    *uint    `json:"customer_id" db:"customer_id"`
	OrderTypeID   *uint    `json:"order_type_id" db:"order_type_id"`
	CurrencyID    *uint    `json:"currency_id" db:"currency_id"`
	WarehouseID   *uint    `json:"warehouse_id" db:"warehouse_id"`
	VatID         *uint    `json:"vat_id" db:"vat_id"`
	PaymentID     *uint    `json:"payment_id" db:"payment_id"`
	Pph23ID       *uint    `json:"pph23_id" db:"pph23_id"`
	BranchID      *uint    `json:"branch_id" db:"branch_id"`
	InventoryNo   *string  `json:"sales_order_no" db:"sales_order_no"`
	PoBuyerNo     string   `json:"po_buyer_no" db:"po_buyer_no"`
	PoBuyerNoOri  *string  `json:"po_buyer_no_ori" db:"po_buyer_no_ori"`
	ShipDest      *string  `json:"ship_dest" db:"ship_dest"`
	Remark        *string  `json:"remark" db:"remark"`
	Status        string   `json:"status" db:"status"`
	ExchangeRate  *float64 `json:"exchange_rate" db:"exchange_rate"`
	VatPerc       *float64 `json:"vat_perc" db:"vat_perc"`
	Pph23Perc     *float64 `json:"pph23_perc" db:"pph23_perc"`
	MarkupPerc    *float64 `json:"markup_perc" db:"markup_perc"`
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
	CreatedByID   *uint    `json:"crweated_by_id" db:"created_by_id"`
	UpdatedByID   *uint    `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint    `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName *string  `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string  `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string  `json:"created_at" db:"created_at"`
	UpdatedAt     *string  `json:"updated_at" db:"updated_at"`
	DeleteAt      *string  `json:"deleted_at" db:"deleted_at"`

	// so_dt_vat_id, currency_name vat_name pph23_name
	ProductID       *string `json:"product_id" db:"product_id"`
	ItemID          *string `json:"item_id" db:"item_id"`
	InvDtVatID      *string `json:"so_dt_vat_id" db:"so_dt_vat_id"`
	CurrencyName    *string `json:"currency_name" db:"currency_name"`
	WarehouseName   *string `json:"warehouse_name" db:"warehouse_name"`
	OrderTypeName   *string `json:"order_type_name" db:"order_type_name"`
	CustomerName    *string `json:"customer_name" db:"customer_name"`
	ProductName     *string `json:"product_name" db:"product_name"`
	ItemName        *string `json:"item_name" db:"item_name"`
	VatName         *string `json:"vat_name" db:"vat_name"`
	Pph23Name       *string `json:"pph23_name" db:"pph23_name"`
	InvDtRemark     *string `json:"so_dt_remark" db:"so_dt_remark"`
	InvDtGenCode    *string `json:"so_dt_gen_code" db:"so_dt_gen_code"`
	InvDtBomGenCode *string `json:"so_dt_bom_gen_code" db:"so_dt_bom_gen_code"`
	InvDtBomRemark  *string `json:"so_dt_bom_remark" db:"so_dt_bom_remark"`
}

type InventoryDetailDTO struct {
	ID            uint     `json:"id" db:"id"`
	InventoryID   *uint    `json:"sales_order_id" db:"sales_order_id"`
	CustomerID    *uint    `json:"customer_id" db:"customer_id"`
	OrderTypeID   *uint    `json:"order_type_id" db:"order_type_id"`
	CurrencyID    *uint    `json:"currency_id" db:"currency_id"`
	WarehouseID   *uint    `json:"warehouse_id" db:"warehouse_id"`
	VatID         *uint    `json:"vat_id" db:"vat_id"`
	PaymentID     *uint    `json:"payment_id" db:"payment_id"`
	Pph23ID       *uint    `json:"pph23_id" db:"pph23_id"`
	BranchID      *uint    `json:"branch_id" db:"branch_id"`
	InventoryNo   *string  `json:"sales_order_no" db:"sales_order_no"`
	PoBuyerNo     string   `json:"po_buyer_no" db:"po_buyer_no"`
	PoBuyerNoOri  *string  `json:"po_buyer_no_ori" db:"po_buyer_no_ori"`
	ShipDest      *string  `json:"ship_dest" db:"ship_dest"`
	Remark        *string  `json:"remark" db:"remark"`
	RevNo         *int     `json:"rev_no" db:"rev_no"`
	Status        string   `json:"status" db:"status"`
	ExchangeRate  float64  `json:"exchange_rate" db:"exchange_rate"`
	VatPerc       *float64 `json:"vat_perc" db:"vat_perc"`
	Pph23Perc     *float64 `json:"pph23_perc" db:"pph23_perc"`
	MarkupPerc    *float64 `json:"markup_perc" db:"markup_perc"`
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

	SiTotalAm *float64 `json:"si_total_am" db:"si_total_am"`
	SaTotalAm *float64 `json:"sa_total_am" db:"sa_total_am"`

	CreatedByID   *uint                     `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint                     `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint                     `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName *string                   `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string                   `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string                   `json:"created_at" db:"created_at"`
	UpdatedAt     *string                   `json:"updated_at" db:"updated_at"`
	DeleteAt      *string                   `json:"deleted_at" db:"deleted_at"`
	InvDts        []InventoryInvDtListDTO   `json:"so_dts"`
	Schedule      *ScheduleDetailDTO        `json:"schedule"`
	Attachments   []InventoryAttachmentsDTO `json:"attachments"`
}

type InventoryAttachmentsDTO struct {
	ID         *uint   `json:"id" db:"id"`
	RefID      *uint   `json:"ref_id" db:"ref_id"`
	RefType    *string `json:"ref_type" db:"ref_type"`
	FileType   *string `json:"file_type" db:"file_type"`
	FileUrl    *string `json:"file_url" db:"file_url"`
	FileName   *string `json:"file_name" db:"file_name"`
	Remark     *string `json:"remark" db:"remark"`
	FileSize   *int64  `json:"file_size" db:"file_size"`
	DeviceType *string `json:"device_type" db:"device_type"`
	CreatedAt  *string `json:"created_at" db:"created_at"`
	DeletedAt  *string `json:"deleted_at" db:"deleted_at"`

	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
}

type InventoryInvDtListDTO struct {
	ID               *uint    `json:"id" db:"id"`
	InvDtID          *uint    `json:"so_dt_id" db:"so_dt_id"`
	ProductUuid      *string  `json:"product_uuid" db:"product_uuid"`
	CustomerID       *uint    `json:"customer_id" db:"customer_id"`
	InventoryID      *uint    `json:"sales_order_id" db:"sales_order_id"`
	ItemUnitID       *uint    `json:"item_unit_id" db:"item_unit_id"`
	VatID            *uint    `json:"vat_id" db:"vat_id"`
	Pph23ID          *uint    `json:"pph23_id" db:"pph23_id"`
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
	Pph23Perc        *float64 `json:"pph23_perc" db:"pph23_perc"`
	Pph23PercAm      *float64 `json:"pph23_perc_am" db:"pph23_perc_am"`
	MarkupPerc       *float64 `json:"markup_perc" db:"markup_perc"`
	MarkupPercAm     *float64 `json:"markup_perc_am" db:"markup_perc_am"`
	IsVat            *int8    `json:"is_vat" db:"is_vat"`
	IsPph23          *int8    `json:"is_pph23" db:"is_pph23"`
	IsLockMarkup     *int8    `json:"is_lock_markup" db:"is_lock_markup"`
	IsLockPriceSell  *int8    `json:"is_lock_price_sell" db:"is_lock_price_sell"`
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

	InvDtsBoms []InventoryInvDtBomListDTO `json:"so_dts_boms"`

	RefNum *string `json:"ref_num" db:"ref_num"`
}

type InventoryInvDtListUpdateDTO struct {
	ID               *uint    `json:"id" db:"id"`
	InvDtID          *uint    `json:"so_dt_id" db:"so_dt_id"`
	ProductUuid      *string  `json:"product_uuid" db:"product_uuid"`
	InventoryID      *uint    `json:"sales_order_id" db:"sales_order_id"`
	ItemUnitID       *uint    `json:"item_unit_id" db:"item_unit_id"`
	VatID            *uint    `json:"vat_id" db:"vat_id"`
	Pph23ID          *uint    `json:"pph23_id" db:"pph23_id"`
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
	Pph23Perc        *float64 `json:"pph23_perc" db:"pph23_perc"`
	Pph23PercAm      *float64 `json:"pph23_perc_am" db:"pph23_perc_am"`
	MarkupPerc       *float64 `json:"markup_perc" db:"markup_perc"`
	MarkupPercAm     *float64 `json:"markup_perc_am" db:"markup_perc_am"`
	IsVat            *int8    `json:"is_vat" db:"is_vat"`
	IsPph23          *int8    `json:"is_pph23" db:"is_pph23"`
	IsLockMarkup     *int8    `json:"is_lock_markup" db:"is_lock_markup"`
	IsLockPriceSell  *int8    `json:"is_lock_price_sell" db:"is_lock_price_sell"`
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

type InventoryInvDtBomListDTO struct {
	ID                *uint    `json:"id" db:"id"`
	InvDtBomID        *uint    `json:"so_dt_bom_id" db:"so_dt_bom_id"`
	InventoryID       *uint    `json:"sales_order_id" db:"sales_order_id"`
	InvDtID           *uint    `json:"so_dt_id" db:"so_dt_id"`
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

type GetInventoriesResult struct {
	Inventories []InventoryListDTO
	Total       int
	Err         error
}

type UpdateInventoryScheduleRequest struct {
	ID          uint    `json:"id"`
	AssigneeID  *uint   `json:"assignee_id"`
	InventoryID uint    `json:"sales_order_id"`
	StepsID     *uint   `json:"steps_id"`
	UUID        *string `json:"uuid"`
	Title       string  `json:"title"`
	ModuleType  string  `json:"module_type"`
	Remark      *string `json:"remark"`
	Status      string  `json:"status"`
	StartAt     *string `json:"start_at"`
	EndAt       *string `json:"end_at"`
	Color       *string `json:"color"`
	CreatedByID *uint   `json:"created_by_id"`
	UpdatedByID *uint   `json:"updated_by_id"`
	DeletedByID *uint   `json:"deleted_by_id"`
	IsDelete    *int    `json:"is_delete"`

	Steps []UpdateScheduleStepRequest `json:"steps"`
}

type UpdateInventoryScheduleAppRequest struct {
	// sales_orders | feedbacks
	RefType      string                                  `json:"ref_type"`
	InventoryID  uint                                    `json:"sales_order_id"`
	ScheduleID   uint                                    `json:"schedule_id"`
	DeviceType   string                                  `json:"device_type"`
	DeletedFiles []uint                                  `json:"deleted_files"`
	Tasks        []UpdateInventoryScheduleAppTaskRequest `json:"tasks"`
	Attachments  []UpdateInventoryAttachmentsDTO         `json:"attachments"`
}

type UpdateInventoryScheduleAppTaskRequest struct {
	ID         *uint   `json:"id"`
	AssigneeID *uint   `json:"assignee_id"`
	Title      *string `json:"title"`
	Remark     *string `json:"remark"`
	IsChecked  *int    `json:"is_checked"`
}

type UpdateInvSalesOrderStatusRequest struct {
	ID     uint   `json:"id"`
	Status string `json:"status"`
}

type RefInvIndexSoDtListDTO struct {
	ID               *uint    `json:"id" db:"id"`
	ProductUuid      *string  `json:"product_uuid" db:"product_uuid"`
	SalesOrderID     *uint    `json:"sales_order_id" db:"sales_order_id"`
	ItemUnitID       *uint    `json:"item_unit_id" db:"item_unit_id"`
	ItemID           *uint    `json:"item_id" db:"item_id"`
	ItemSubGroupName *string  `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName    *string  `json:"item_group_name" db:"item_group_name"`
	ItemName         *string  `json:"item_name" db:"item_name"`
	ItemCode         *string  `json:"item_code" db:"item_code"`
	UnitName         *string  `json:"unit_name" db:"unit_name"`
	GenCode          *string  `json:"gen_code" db:"gen_code"`
	Remark           *string  `json:"remark" db:"remark"`
	QtyOut           *float64 `json:"qty_out" db:"qty_out"`
	PriceSell        *float64 `json:"price_sell" db:"price_sell"`
	PriceBuy         *float64 `json:"price_buy" db:"price_buy"`
	SubtotalSell     *float64 `json:"subtotal_sell" db:"subtotal_sell"`
	SubtotalBuy      *float64 `json:"subtotal_buy" db:"subtotal_buy"`

	CustomerName *string  `json:"customer_name" db:"customer_name"`
	RefNum       *string  `json:"ref_num" db:"ref_num"`
	OrderAt      *string  `json:"order_at" db:"order_at"`
	ItemSku      *string  `json:"item_sku" db:"item_sku"`
	RefQty       *float64 `json:"ref_qty" db:"ref_qty"`
	Balance      *float64 `json:"balance" db:"balance"`
}

type InvSalesOrderSoDtBomListDTO struct {
	ID                *uint   `json:"id" db:"id"`
	SoDtBomID         *uint   `json:"so_dt_bom_id" db:"so_dt_bom_id"`
	SalesOrderID      *uint   `json:"sales_order_id" db:"sales_order_id"`
	SoDtID            *uint   `json:"so_dt_id" db:"so_dt_id"`
	ProductID         uint    `json:"product_id" db:"product_id"`
	ProductUuid       string  `json:"product_uuid" db:"product_uuid"`
	ItemID            uint    `json:"item_id" db:"item_id"`
	ItemSubGroupID    *uint   `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID       *uint   `json:"item_group_id" db:"item_group_id"`
	ItemSubGroupName  *string `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName     *string `json:"item_group_name" db:"item_group_name"`
	ItemName          *string `json:"item_name" db:"item_name"`
	ItemCode          *string `json:"item_code" db:"item_code"`
	ItemBarcode       *string `json:"item_barcode" db:"item_barcode"`
	ItemSku           *string `json:"item_sku" db:"item_sku"`
	ItemFactoryCode   *string `json:"item_factory_code" db:"item_factory_code"`
	ItemSpecification *string `json:"item_specification" db:"item_specification"`
	ItemQtyStock      *string `json:"item_qty_stock" db:"item_qty_stock"`
	UnitName          *string `json:"unit_name" db:"unit_name"`
	ItemUnitID        *uint   `json:"item_unit_id" db:"item_unit_id"`
	RefJSON           *string `json:"ref_json" db:"ref_json"`
	GenCode           *string `json:"gen_code" db:"gen_code"`
	Remark            *string `json:"remark" db:"remark"`
	Qty               float64 `json:"qty" db:"qty"`
	PriceSell         float64 `json:"price_sell" db:"price_sell"`
	PriceBuy          float64 `json:"price_buy" db:"price_buy"`
	SubtotalSell      float64 `json:"subtotal_sell" db:"subtotal_sell"`
	SubtotalBuy       float64 `json:"subtotal_buy" db:"subtotal_buy"`

	CreatedByID   *uint     `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint     `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint     `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName *string   `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string   `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *string   `json:"updated_at" db:"updated_at"`
	DeletedAt     *string   `json:"deleted_at" db:"deleted_at"`
}
