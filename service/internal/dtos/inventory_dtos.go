package dtos

import "time"

type GetInventoriesRequest struct {
	Global         *string `json:"global"`
	Title          *string `json:"title"`
	PoBuyerNo      *string `json:"po_buyer_no"`
	InventoryNo    *string `json:"sales_order_no"`
	Remark         *string `json:"remark"`
	CustomerID     *int    `json:"customer_id"`
	IoTypeID       *int    `json:"io_type_id"`
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

type FormInvDtsRequest struct {
	ID              *uint    `json:"id"`
	InvDtID         *uint    `json:"inv_dt_id"`
	ProductUuid     string   `json:"product_uuid"`
	ItemUnitID      uint     `json:"item_unit_id"`
	VatID           *uint    `json:"vat_id"`
	Pph23ID         *uint    `json:"pph23_id"`
	RefSoDtID       *uint    `json:"ref_so_dt_id"`
	RefSoDtBomID    *uint    `json:"ref_so_dt_bom_id"`
	RefRoDtID       *uint    `json:"ref_ro_dt_id"`
	RefPoDtID       *uint    `json:"ref_po_dt_id"`
	RefPoDtBomID    *uint    `json:"ref_po_dt_bom_id"`
	RefInvDtID      *uint    `json:"ref_inv_dt_id"`
	RefProductID    *uint    `json:"ref_product_id"`
	RefProductBomID *uint    `json:"ref_product_bom_id"`
	ItemID          uint     `json:"item_id"`
	RefType         string   `json:"ref_type"`
	ItemType        string   `json:"item_type"`
	GenCode         *string  `json:"gen_code"`
	Remark          *string  `json:"remark"`
	ExpiredAt       *string  `json:"expired_at"`
	VatPerc         *float64 `json:"vat_perc"`
	VatPercAm       *float64 `json:"vat_perc_am"`
	Pph23Perc       *float64 `json:"pph23_perc"`
	Pph23PercAm     *float64 `json:"pph23_perc_am"`
	IsVat           *int     `json:"is_vat"`
	IsPph23         *int     `json:"is_pph23"`
	QtyInvoice      *float64 `json:"qty_invoice"`
	QtyOut          *float64 `json:"qty_out"`
	Qty             *float64 `json:"qty"`
	PriceSell       *float64 `json:"price_sell"`
	PriceBuy        *float64 `json:"price_buy"`
	SubtotalSell    *float64 `json:"subtotal_sell"`
	SubtotalBuy     *float64 `json:"subtotal_buy"`
	TotalAm         *float64 `json:"total_am"`

	ItemName *string `json:"item_name"`
	ItemCode *string `json:"item_code"`
}
type FormInventoryRequest struct {
	ID             *uint    `json:"id"`
	CustomerID     *uint    `json:"customer_id"`
	IoTypeID       *uint    `json:"io_type_id"`
	CurrencyID     *uint    `json:"currency_id"`
	PaymentTermID  *uint    `json:"payment_term_id"`
	WarehouseID    uint     `json:"warehouse_id"`
	VatID          *uint    `json:"vat_id"`
	Pph23ID        *uint    `json:"pph23_id"`
	BranchID       *uint    `json:"branch_id"`
	IoType         string   `json:"io_type"`
	InventoryNo    *string  `json:"inventory_no"`
	InventoryNoOri *string  `json:"inventory_no_ori"`
	RevNo          *int     `json:"rev_no"`
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

	InvDts []FormInvDtsRequest `json:"inv_dts"`

	CustomerCode string `json:"customer_code"`
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
	ID             int      `json:"id" db:"id"`
	InventoryID    *uint    `json:"sales_order_id" db:"sales_order_id"`
	CustomerID     *uint    `json:"customer_id" db:"customer_id"`
	IoTypeID       *uint    `json:"io_type_id" db:"io_type_id"`
	CurrencyID     *uint    `json:"currency_id" db:"currency_id"`
	WarehouseID    *uint    `json:"warehouse_id" db:"warehouse_id"`
	VatID          *uint    `json:"vat_id" db:"vat_id"`
	PaymentTermID  *uint    `json:"payment_term_id" db:"payment_term_id"`
	Pph23ID        *uint    `json:"pph23_id" db:"pph23_id"`
	BranchID       *uint    `json:"branch_id" db:"branch_id"`
	InventoryNo    string   `json:"inventory_no" db:"inventory_no"`
	InventoryNoOri *string  `json:"inventory_no_ori" db:"inventory_no_ori"`
	DoNo           *string  `json:"do_no" db:"do_no"`
	SuratJalanNo   *string  `json:"surat_jalan_no" db:"surat_jalan_no"`
	InvoiceNo      *string  `json:"invoice_no" db:"invoice_no"`
	ShipDest       *string  `json:"ship_dest" db:"ship_dest"`
	Remark         *string  `json:"remark" db:"remark"`
	Status         string   `json:"status" db:"status"`
	ExchangeRate   *float64 `json:"exchange_rate" db:"exchange_rate"`
	VatPerc        *float64 `json:"vat_perc" db:"vat_perc"`
	Pph23Perc      *float64 `json:"pph23_perc" db:"pph23_perc"`
	TotalQty       *float64 `json:"total_qty" db:"total_qty"`
	Subtotal       *float64 `json:"subtotal" db:"subtotal"`
	TotalPph23     *float64 `json:"total_pph23" db:"total_pph23"`
	TotalVat       *float64 `json:"total_vat" db:"total_vat"`
	GrandTotal     *float64 `json:"grand_total" db:"grand_total"`
	DoAt           *string  `json:"do_at" db:"do_at"`
	IngoingAt      *string  `json:"ingoing_at" db:"ingoing_at"`
	InvoiceAt      *string  `json:"invoice_at" db:"invoice_at"`
	CreatedByID    *uint    `json:"created_by_id" db:"created_by_id"`
	UpdatedByID    *uint    `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID    *uint    `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName  *string  `json:"created_by_name" db:"created_by_name"`
	UpdatedByName  *string  `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt      *string  `json:"created_at" db:"created_at"`
	UpdatedAt      *string  `json:"updated_at" db:"updated_at"`
	DeleteAt       *string  `json:"deleted_at" db:"deleted_at"`

	CurrencyName  *string `json:"currency_name" db:"currency_name"`
	WarehouseName *string `json:"warehouse_name" db:"warehouse_name"`
	IoTypeName    *string `json:"io_type_name" db:"io_type_name"`
	CustomerName  *string `json:"customer_name" db:"customer_name"`
	ItemName      *string `json:"item_name" db:"item_name"`
	VatName       *string `json:"vat_name" db:"vat_name"`
	Pph23Name     *string `json:"pph23_name" db:"pph23_name"`
}

type InventoryDetailDTO struct {
	ID             uint     `json:"id" db:"id"`
	InventoryID    *uint    `json:"inventory_id" db:"inventory_id"`
	CustomerID     *uint    `json:"customer_id" db:"customer_id"`
	IoTypeID       *uint    `json:"io_type_id" db:"io_type_id"`
	CurrencyID     *uint    `json:"currency_id" db:"currency_id"`
	WarehouseID    *uint    `json:"warehouse_id" db:"warehouse_id"`
	VatID          *uint    `json:"vat_id" db:"vat_id"`
	PaymentTermID  *uint    `json:"payment_term_id" db:"payment_term_id"`
	Pph23ID        *uint    `json:"pph23_id" db:"pph23_id"`
	BranchID       *uint    `json:"branch_id" db:"branch_id"`
	InventoryNo    string   `json:"inventory_no" db:"inventory_no"`
	InventoryNoOri *string  `json:"inventory_no_ori" db:"inventory_no_ori"`
	IoType         *string  `json:"io_type" db:"io_type"`
	DoNo           *string  `json:"do_no" db:"do_no"`
	SuratJalanNo   *string  `json:"surat_jalan_no" db:"surat_jalan_no"`
	InvoiceNo      *string  `json:"invoice_no" db:"invoice_no"`
	ShipDest       *string  `json:"ship_dest" db:"ship_dest"`
	Remark         *string  `json:"remark" db:"remark"`
	RevNo          *int     `json:"rev_no" db:"rev_no"`
	Status         string   `json:"status" db:"status"`
	ExchangeRate   float64  `json:"exchange_rate" db:"exchange_rate"`
	VatPerc        *float64 `json:"vat_perc" db:"vat_perc"`
	Pph23Perc      *float64 `json:"pph23_perc" db:"pph23_perc"`
	TotalQty       float64  `json:"total_qty" db:"total_qty"`
	QtyOut         *float64 `json:"qty_out" db:"qty_out"`
	Subtotal       float64  `json:"subtotal" db:"subtotal"`
	TotalPph23     float64  `json:"total_pph23" db:"total_pph23"`
	TotalVat       float64  `json:"total_vat" db:"total_vat"`
	GrandTotal     float64  `json:"grand_total" db:"grand_total"`
	DoAt           *string  `json:"do_at" db:"do_at"`
	IngoingAt      *string  `json:"ingoing_at" db:"ingoing_at"`
	InvoiceAt      *string  `json:"invoice_at" db:"invoice_at"`

	CreatedByID   *uint                   `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint                   `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint                   `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName *string                 `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string                 `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string                 `json:"created_at" db:"created_at"`
	UpdatedAt     *string                 `json:"updated_at" db:"updated_at"`
	DeleteAt      *string                 `json:"deleted_at" db:"deleted_at"`
	InvDts        []InventoryInvDtListDTO `json:"inv_dts,omitempty" gorm:"-"`

	CompanyProfileID *uint                   `json:"company_profile_id" db:"company_profile_id"`
	CustomerCode     *string                 `json:"customer_code" db:"customer_code"`
	CustomerName     *string                 `json:"customer_name" db:"customer_name"`
	Phone            *string                 `json:"phone" db:"phone"`
	Address          *string                 `json:"address" db:"address"`
	Company          CompanyProfileDetailDTO `json:"company,omitempty" gorm:"-"`
	OrderTypeName    *string                 `json:"order_type_name" db:"order_type_name"`
	CurrencyName     *string                 `json:"currency_name" db:"currency_name"`
	VatName          *string                 `json:"vat_name" db:"vat_name"`
	Pph23Name        *string                 `json:"pph23_name" db:"pph23_name"`
	BankName         *string                 `json:"bank_name" db:"bank_name"`
	AccountName      *string                 `json:"account_name" db:"account_name"`
	TotalAfterDisc   float64                 `json:"total_after_disc" db:"total_after_disc"`
	IsIDOnly         *int                    `json:"is_id_only" db:"is_id_only"`

	QtyOutOnIn *float64 `json:"qty_out_on_in" db:"qty_out_on_in"`
}

type InventoryInvDtListDTO struct {
	ID               *uint    `json:"id" db:"id"`
	InvDtID          *uint    `json:"inv_dt_id" db:"inv_dt_id"`
	ProductUuid      *string  `json:"product_uuid" db:"product_uuid"`
	CustomerID       *uint    `json:"customer_id" db:"customer_id"`
	InventoryID      *uint    `json:"inventory_id" db:"inventory_id"`
	ItemUnitID       *uint    `json:"item_unit_id" db:"item_unit_id"`
	VatID            *uint    `json:"vat_id" db:"vat_id"`
	Pph23ID          *uint    `json:"pph23_id" db:"pph23_id"`
	RefID            *uint    `json:"ref_id" db:"ref_id"`
	ItemID           uint     `json:"item_id" db:"item_id"`
	RefSoDtID        *uint    `json:"ref_so_dt_id" db:"ref_so_dt_id"`
	RefSoDtBomID     *uint    `json:"ref_so_dt_bom_id" db:"ref_so_dt_bom_id"`
	RefRoDtID        *uint    `json:"ref_ro_dt_id" db:"ref_ro_dt_id"`
	RefPoDtID        *uint    `json:"ref_po_dt_id" db:"ref_po_dt_id"`
	RefPoDtBomID     *uint    `json:"ref_po_dt_bom_id" db:"ref_po_dt_bom_id"`
	RefInvDtID       *uint    `json:"ref_inv_dt_id" db:"ref_inv_dt_id"`
	RefProductID     *uint    `json:"ref_product_id" db:"ref_product_id"`
	RefProductBomID  *uint    `json:"ref_product_bom_id" db:"ref_product_bom_id"`
	ItemSubGroupID   *uint    `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID      *uint    `json:"item_group_id" db:"item_group_id"`
	ItemSubGroupName *string  `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName    *string  `json:"item_group_name" db:"item_group_name"`
	ItemName         *string  `json:"item_name" db:"item_name"`
	ItemCode         *string  `json:"item_code" db:"item_code"`
	UnitName         *string  `json:"unit_name" db:"unit_name"`
	RefType          *string  `json:"ref_type" db:"ref_type"`
	ItemType         *string  `json:"item_type" db:"item_type"`
	GenCode          *string  `json:"gen_code" db:"gen_code"`
	Remark           *string  `json:"remark" db:"remark"`
	VatPerc          *float64 `json:"vat_perc" db:"vat_perc"`
	VatPercAm        *float64 `json:"vat_perc_am" db:"vat_perc_am"`
	Pph23Perc        *float64 `json:"pph23_perc" db:"pph23_perc"`
	Pph23PercAm      *float64 `json:"pph23_perc_am" db:"pph23_perc_am"`
	IsVat            *int8    `json:"is_vat" db:"is_vat"`
	IsPph23          *int8    `json:"is_pph23" db:"is_pph23"`
	QtyInvoice       *float64 `json:"qty_invoice" db:"qty_invoice"`
	QtyOut           *float64 `json:"qty_out" db:"qty_out"`
	QtyOutOnIn       *float64 `json:"qty_out_on_in" db:"qty_out_on_in"`
	QtyIn            *float64 `json:"qty_in" db:"qty_in"`
	Qty              *float64 `json:"qty" db:"qty"`
	QtyInit          *float64 `json:"qty_init" db:"qty_init"`
	PriceSell        *float64 `json:"price_sell" db:"price_sell"`
	PriceBuy         *float64 `json:"price_buy" db:"price_buy"`
	SubtotalSell     *float64 `json:"subtotal_sell" db:"subtotal_sell"`
	SubtotalBuy      *float64 `json:"subtotal_buy" db:"subtotal_buy"`
	TotalAm          *float64 `json:"total_am" db:"total_am"`
	ExpiredAt        *string  `json:"expired_at" db:"expired_at"`

	RefQty        *float64  `json:"ref_qty" db:"ref_qty"`
	ItemSku       *string   `json:"item_sku" db:"item_sku"`
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

type RefInvIndexInvDtListDTO struct {
	ID               *uint    `json:"id" db:"id"`
	BranchID         *uint    `json:"branch_id" db:"branch_id"`
	InvDtID          *uint    `json:"inv_dt_id" db:"inv_dt_id"`
	ProductUuid      *string  `json:"product_uuid" db:"product_uuid"`
	CustomerID       *uint    `json:"customer_id" db:"customer_id"`
	InventoryID      *uint    `json:"inventory_id" db:"inventory_id"`
	ItemUnitID       *uint    `json:"item_unit_id" db:"item_unit_id"`
	VatID            *uint    `json:"vat_id" db:"vat_id"`
	Pph23ID          *uint    `json:"pph23_id" db:"pph23_id"`
	RefID            *uint    `json:"ref_id" db:"ref_id"`
	ItemID           uint     `json:"item_id" db:"item_id"`
	RefSoDtID        *uint    `json:"ref_so_dt_id" db:"ref_so_dt_id"`
	RefSoDtBomID     *uint    `json:"ref_so_dt_bom_id" db:"ref_so_dt_bom_id"`
	RefPoDtID        *uint    `json:"ref_po_dt_id" db:"ref_po_dt_id"`
	RefPoDtBomID     *uint    `json:"ref_po_dt_bom_id" db:"ref_po_dt_bom_id"`
	RefInvDtID       *uint    `json:"ref_inv_dt_id" db:"ref_inv_dt_id"`
	RefProductID     *uint    `json:"ref_product_id" db:"ref_product_id"`
	ItemSubGroupID   *uint    `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID      *uint    `json:"item_group_id" db:"item_group_id"`
	ItemSubGroupName *string  `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName    *string  `json:"item_group_name" db:"item_group_name"`
	ItemName         *string  `json:"item_name" db:"item_name"`
	ItemCode         *string  `json:"item_code" db:"item_code"`
	UnitName         *string  `json:"unit_name" db:"unit_name"`
	RefType          *string  `json:"ref_type" db:"ref_type"`
	ItemType         *string  `json:"item_type" db:"item_type"`
	GenCode          *string  `json:"gen_code" db:"gen_code"`
	Remark           *string  `json:"remark" db:"remark"`
	VatPerc          *float64 `json:"vat_perc" db:"vat_perc"`
	VatPercAm        *float64 `json:"vat_perc_am" db:"vat_perc_am"`
	Pph23Perc        *float64 `json:"pph23_perc" db:"pph23_perc"`
	Pph23PercAm      *float64 `json:"pph23_perc_am" db:"pph23_perc_am"`
	IsVat            *int8    `json:"is_vat" db:"is_vat"`
	IsPph23          *int8    `json:"is_pph23" db:"is_pph23"`
	QtyInvoice       *float64 `json:"qty_invoice" db:"qty_invoice"`
	QtyOut           *float64 `json:"qty_out" db:"qty_out"`
	Qty              *float64 `json:"qty" db:"qty"`
	PriceSell        *float64 `json:"price_sell" db:"price_sell"`
	PriceBuy         *float64 `json:"price_buy" db:"price_buy"`
	SubtotalSell     *float64 `json:"subtotal_sell" db:"subtotal_sell"`
	SubtotalBuy      *float64 `json:"subtotal_buy" db:"subtotal_buy"`
	TotalAm          *float64 `json:"total_am" db:"total_am"`
	ExpiredAt        *string  `json:"expired_at" db:"expired_at"`

	RefQty    *float64  `json:"ref_qty" db:"ref_qty"`
	Balance   *float64  `json:"balance" db:"balance"`
	ItemSku   *string   `json:"item_sku" db:"item_sku"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *string   `json:"updated_at" db:"updated_at"`
	DeleteAt  *string   `json:"deleted_at" db:"deleted_at"`

	CustomerName *string `json:"customer_name" db:"customer_name"`
	IoTypeName   *string `json:"io_type_name" db:"io_type_name"`
	InventoryNo  *string `json:"inventory_no" db:"inventory_no"`
	SuratJalanNo *string `json:"surat_jalan_no" db:"surat_jalan_no"`
	DoNo         *string `json:"do_no" db:"do_no"`
	InvoiceNo    *string `json:"invoice_no" db:"invoice_no"`
	IngoingAt    *string `json:"ingoing_at" db:"ingoing_at"`
	DoAt         *string `json:"do_at" db:"do_at"`
	InvoiceAt    *string `json:"invoice_at" db:"invoice_at"`

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

type GetInventoriesResult struct {
	Inventories []InventoryListDTO
	Total       int
	Err         error
}

type UpdateInvSalesOrderStatusRequest struct {
	ID     uint   `json:"id"`
	Status string `json:"status"`
}

type RefInvIndexSoDtListDTO struct {
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
	QtyOut           *float64 `json:"qty_out" db:"qty_out"`
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
	ShipDest     *string  `json:"ship_dest" db:"ship_dest"`
	ItemSku      *string  `json:"item_sku" db:"item_sku"`
	RefQty       *float64 `json:"ref_qty" db:"ref_qty"`
	ItemType     *string  `json:"item_type" db:"item_type"`
	Balance      *float64 `json:"balance" db:"balance"`
}

type RefInvIndexRoDtListDTO struct {
	// ID               *uint    `json:"id" db:"id"`
	ProductUuid      *string  `json:"product_uuid" db:"product_uuid"`
	RequestOrderID   *uint    `json:"request_order_id" db:"request_order_id"`
	WarehouseID      *uint    `json:"warehouse_id" db:"warehouse_id"`
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
	VatPerc          *float64 `json:"vat_perc" db:"vat_perc"`
	VatPercAm        *float64 `json:"vat_perc_am" db:"vat_perc_am"`
	Pph23Perc        *float64 `json:"pph23_perc" db:"pph23_perc"`
	Pph23PercAm      *float64 `json:"pph23_perc_am" db:"pph23_perc_am"`
	IsVat            *int8    `json:"is_vat" db:"is_vat"`
	IsPph23          *int8    `json:"is_pph23" db:"is_pph23"`

	WarehouseName *string  `json:"warehouse_name" db:"warehouse_name"`
	RefType       *string  `json:"ref_type" db:"ref_type"`
	RefSoDtID     *uint    `json:"ref_so_dt_id" db:"ref_so_dt_id"`
	RefSoDtBomID  *uint    `json:"ref_so_dt_bom_id" db:"ref_so_dt_bom_id"`
	RefRoDtID     *uint    `json:"ref_ro_dt_id" db:"ref_ro_dt_id"`
	RefRoDtBomID  *uint    `json:"ref_ro_dt_bom_id" db:"ref_ro_dt_bom_id"`
	RefPoDtID     *uint    `json:"ref_po_dt_id" db:"ref_po_dt_id"`
	RefPoDtBomID  *uint    `json:"ref_po_dt_bom_id" db:"ref_po_dt_bom_id"`
	RefInvDtID    *uint    `json:"ref_inv_dt_id" db:"ref_inv_dt_id"`
	RefProductID  *uint    `json:"ref_product_id" db:"ref_product_id"`
	RefNum        *string  `json:"ref_num" db:"ref_num"`
	RequestDate   *string  `json:"request_date" db:"request_date"`
	ItemSku       *string  `json:"item_sku" db:"item_sku"`
	RefQty        *float64 `json:"ref_qty" db:"ref_qty"`
	ItemType      *string  `json:"item_type" db:"item_type"`
	Balance       *float64 `json:"balance" db:"balance"`
}

type RefInvIndexPoDtListDTO struct {
	// ID               *uint    `json:"id" db:"id"`
	ProductUuid      *string  `json:"product_uuid" db:"product_uuid"`
	PoID             *uint    `json:"po_id" db:"po_id"`
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
	QtyIn            *float64 `json:"qty_in" db:"qty_in"`
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

	CustomerName  *string  `json:"customer_name" db:"customer_name"`
	RefType       *string  `json:"ref_type" db:"ref_type"`
	RefSoDtID     *uint    `json:"ref_so_dt_id" db:"ref_so_dt_id"`
	RefSoDtBomID  *uint    `json:"ref_so_dt_bom_id" db:"ref_so_dt_bom_id"`
	RefPoDtID     *uint    `json:"ref_po_dt_id" db:"ref_po_dt_id"`
	RefPoDtBomID  *uint    `json:"ref_po_dt_bom_id" db:"ref_po_dt_bom_id"`
	RefInvDtID    *uint    `json:"ref_inv_dt_id" db:"ref_inv_dt_id"`
	RefProductID  *uint    `json:"ref_product_id" db:"ref_product_id"`
	CurrencyID    *uint    `json:"currency_id" db:"currency_id"`
	VatID         *uint    `json:"vat_id" db:"vat_id"`
	Pph23ID       *uint    `json:"pph23_id" db:"pph23_id"`
	PaymentTermID *uint    `json:"payment_term_id" db:"payment_term_id"`
	ExchangeRate  *float64 `json:"exchange_rate" db:"exchange_rate"`
	RefNum        *string  `json:"ref_num" db:"ref_num"`
	PoDate        *string  `json:"po_date" db:"po_date"`
	DeliveryDate  *string  `json:"delivery_date" db:"delivery_date"`
	ShipDest      *string  `json:"ship_dest" db:"ship_dest"`
	ItemSku       *string  `json:"item_sku" db:"item_sku"`
	RefQty        *float64 `json:"ref_qty" db:"ref_qty"`
	ItemType      *string  `json:"item_type" db:"item_type"`
	Balance       *float64 `json:"balance" db:"balance"`
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

type GetInvSoDtQtyUpdateDTO struct {
	ID           *uint    `json:"id" db:"id"`
	SoDtID       *uint    `json:"so_dt_id" db:"so_dt_id"`
	SoDtBomID    *uint    `json:"so_dt_bom_id" db:"so_dt_bom_id"`
	QtyOut       *float64 `json:"qty_out" db:"qty_out"`
	SalesOrderID *uint    `json:"sales_order_id" db:"sales_order_id"`
}

type StockListDTO struct {
	ID            int      `json:"id" db:"id"`
	ItemID        *uint    `json:"item_id" db:"item_id"`
	WarehouseID   *uint    `json:"warehouse_id" db:"warehouse_id"`
	BranchID      *uint    `json:"branch_id" db:"branch_id"`
	Qty           *float64 `json:"qty" db:"qty"`
	CreatedByID   *uint    `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint    `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint    `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName *string  `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string  `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string  `json:"created_at" db:"created_at"`
	UpdatedAt     *string  `json:"updated_at" db:"updated_at"`
	DeleteAt      *string  `json:"deleted_at" db:"deleted_at"`

	WarehouseName    *string `json:"warehouse_name" db:"warehouse_name"`
	ItemName         *string `json:"item_name" db:"item_name"`
	UnitName         *string `json:"unit_name" db:"unit_name"`
	BranchName       *string `json:"branch_name" db:"branch_name"`
	ItemGroupName    *string `json:"item_group_name" db:"item_group_name"`
	ItemSubGroupName *string `json:"item_sub_group_name" db:"item_sub_group_name"`
}

type StockClosingListDTO struct {
	ID             int      `json:"id" db:"id"`
	ItemID         *uint    `json:"item_id" db:"item_id"`
	WarehouseID    *uint    `json:"warehouse_id" db:"warehouse_id"`
	ClosingAt      *string  `json:"closing_at" db:"closing_at"`
	LastClosingAt  *string  `json:"last_closing_at" db:"last_closing_at"`
	BeginQty       *float64 `json:"begin_qty" db:"begin_qty"`
	InQty          *float64 `json:"in_qty" db:"in_qty"`
	OutQty         *float64 `json:"out_qty" db:"out_qty"`
	AdjustmentQty  *float64 `json:"adjustment_qty" db:"adjustment_qty"`
	EndQty         *float64 `json:"end_qty" db:"end_qty"`
	PriceSell      *float64 `json:"price_sell" db:"price_sell"`
	PriceBuy       *float64 `json:"price_buy" db:"price_buy"`
	TotalValueSell *float64 `json:"total_value_sell" db:"total_value_sell"`
	TotalValueBuy  *float64 `json:"total_value_buy" db:"total_value_buy"`
	CreatedByName  *string  `json:"created_by_name" db:"created_by_name"`
	UpdatedByName  *string  `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt      *string  `json:"created_at" db:"created_at"`
	UpdatedAt      *string  `json:"updated_at" db:"updated_at"`
	DeleteAt       *string  `json:"deleted_at" db:"deleted_at"`

	WarehouseName    *string `json:"warehouse_name" db:"warehouse_name"`
	ItemName         *string `json:"item_name" db:"item_name"`
	UnitName         *string `json:"unit_name" db:"unit_name"`
	ItemGroupName    *string `json:"item_group_name" db:"item_group_name"`
	ItemSubGroupName *string `json:"item_sub_group_name" db:"item_sub_group_name"`
}

type FormClosingStockStoreRequest struct {
	StartClosingAt *string `json:"start_closing_at"`
	EndClosingAt   string  `json:"end_closing_at"`
	Password       string  `json:"password"`
	IsFinalized    int8    `json:"is_finalized"`
}

type InventoryStatusDTO struct {
	ItemID    uint   `json:"item_id" db:"item_id"`
	BranchID  uint   `json:"branch_id" db:"branch_id"`
	IngoingAt string `json:"ingoing_at" db:"ingoing_at"`
}

type GetInventoriesStatusDtParams struct {
	ItemIDs []uint `json:"item_ids"`
}

type InventoryStatusDtDTO struct {
	ID        uint    `json:"id" db:"id"`
	ItemID    uint    `json:"item_id" db:"item_id"`
	BranchID  *uint   `json:"branch_id" db:"branch_id"`
	IngoingAt *string `json:"ingoing_at" db:"ingoing_at"`

	IoType           string   `json:"io_type" db:"io_type"`
	ItemGroupName    *string  `json:"item_group_name" db:"item_group_name"`
	ItemSubGroupName *string  `json:"item_sub_group_name" db:"item_sub_group_name"`
	WarehouseName    *string  `json:"warehouse_name" db:"warehouse_name"`
	IoTypeName       *string  `json:"io_type_name" db:"io_type_name"`
	CustomerName     *string  `json:"customer_name" db:"customer_name"`
	ItemCode         *string  `json:"item_code" db:"item_code"`
	ItemName         *string  `json:"item_name" db:"item_name"`
	UnitName         *string  `json:"unit_name" db:"unit_name"`
	CurrencyName     *string  `json:"currency_name" db:"currency_name"`
	PriceSell        *float64 `json:"price_sell" db:"price_sell"`
	PriceBuy         *float64 `json:"price_buy" db:"price_buy"`
	Price            *float64 `json:"price" db:"price"`
	Qty              float64  `json:"qty" db:"qty"`
	QtyIn            float64  `json:"qty_in" db:"qty_in"`
	QtyOut           float64  `json:"qty_out" db:"qty_out"`
	Balance          float64  `json:"balance" db:"balance"`
	IsTotal          bool     `json:"is_total" db:"is_total"`
	InTotal          float64  `json:"in_total" db:"in_total"`
	OutTotal         float64  `json:"out_total" db:"out_total"`
	BalanceTotal     float64  `json:"balance_total" db:"balance_total"`
}

type SyncStockInventoryRequest struct {
	StartClosingAt *string `json:"start_closing_at"`
	EndClosingAt   string  `json:"end_closing_at"`
	Password       string  `json:"password"`
	UserID         uint    `json:"user_id"`
	BranchID       uint    `json:"branch_id"`
	IsFinalized    int8    `json:"is_finalized"`
}

type InventoryPDFData struct {
	Num string `json:"num"`
	// Form FormInventoryRequest
	Form InventoryDetailDTO
}
