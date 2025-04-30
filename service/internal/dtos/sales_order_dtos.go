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
	ProductUuid     string                   `json:"product_uuid"`
	ItemUnitID      *uint                    `json:"item_unit_id"`
	VatID           *uint                    `json:"vat_id"`
	Pph23ID         *uint                    `json:"pph23_id"`
	RefID           *uint                    `json:"ref_id"`
	ItemID          uint                     `json:"item_id"`
	RefType         *string                  `json:"ref_type"`
	ItemType        *string                  `json:"item_type"`
	GenCode         *string                  `json:"gen_code"`
	Remark          *string                  `json:"remark"`
	VatPerc         *float64                 `json:"vat_perc"`
	VatPercAm       *float64                 `json:"vat_perc_am"`
	Pph23Perc       *float64                 `json:"pph23_perc"`
	Pph23PercAm     *float64                 `json:"pph23_perc_am"`
	MarkupPerc      *float64                 `json:"markup_perc"`
	MarkupPercAm    *float64                 `json:"markup_perc_am"`
	IsVat           *int8                    `json:"is_vat"`
	IsPph23         *int8                    `json:"is_pph23"`
	IsLockMarkup    *int8                    `json:"is_lock_markup"`
	IsLockPriceSell *int8                    `json:"is_lock_price_sell"`
	Qty             *float64                 `json:"qty"`
	PriceSell       *float64                 `json:"price_sell"`
	PriceBuy        *float64                 `json:"price_buy"`
	SubtotalSell    *float64                 `json:"subtotal_sell"`
	SubtotalBuy     *float64                 `json:"subtotal_buy"`
	DiscAm          *float64                 `json:"disc_am"`
	DiscPerc        *float64                 `json:"disc_perc"`
	DiscPercNum     *float64                 `json:"disc_perc_num"`
	DiscPercAm      *float64                 `json:"disc_perc_am"`
	DiscFinal       *float64                 `json:"disc_final"`
	DiscType        *string                  `json:"disc_type"`
	TotalAm         *float64                 `json:"total_am"`
	SoDtsBoms       []CreateSoDtsBomsRequest `json:"so_dts_boms"`
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
	CustomerID    *uint                  `json:"customer_id"`
	OrderTypeID   *uint                  `json:"order_type_id"`
	CurrencyID    *uint                  `json:"currency_id"`
	WarehouseID   *uint                  `json:"warehouse_id"`
	VatID         *uint                  `json:"vat_id"`
	PaymentID     *uint                  `json:"payment_id"`
	Pph23ID       *uint                  `json:"pph23_id"`
	BranchID      *uint                  `json:"branch_id"`
	RevNo         *int                   `json:"rev_no"`
	PoBuyerNo     *string                `json:"po_buyer_no"`
	SalesOrderNo  *string                `json:"sales_order_no"`
	Remark        *string                `json:"remark"`
	ShipDest      *string                `json:"ship_dest"`
	Status        string                 `json:"status"`
	ExchangeRate  *float64               `json:"exchange_rate"`
	VatPerc       *float64               `json:"vat_perc"`
	Pph23Perc     *float64               `json:"pph23_perc"`
	MarkupPerc    *float64               `json:"markup_perc"`
	IsVat         *int                   `json:"is_vat"`
	IsPph23       *int                   `json:"is_pph23"`
	DiscAm        *float64               `json:"disc_am"`
	DiscPerc      *float64               `json:"disc_perc"`
	DiscPercAm    *float64               `json:"disc_perc_am"`
	DiscFinal     *float64               `json:"disc_final"`
	DiscType      *string                `json:"disc_type"`
	TotalQty      *float64               `json:"total_qty"`
	Subtotal      *float64               `json:"subtotal"`
	TotalDiscount *float64               `json:"total_discount"`
	TotalPph23    *float64               `json:"total_pph23"`
	TotalVat      *float64               `json:"total_vat"`
	GrandTotal    *float64               `json:"grand_total"`
	OrderAt       *string                `json:"order_at"`
	ShippingAt    *string                `json:"shipping_at"`
	AgreeAt       *string                `json:"agree_at"`
	DueAt         *string                `json:"due_at"`
	SoDts         []CreateSoDtsRequest   `json:"so_dts"`
	Schedule      *CreateScheduleRequest `json:"schedule"`

	CustomerCode string `json:"customer_code"`
}

type CreateScheduleRequest struct {
	AssigneeID *uint `json:"assignee_id"`
	// SalesOrderID  uint    `json:"sales_order_id"`
	UUID       *string `json:"uuid"`
	StepsID    *uint   `json:"steps_id"`
	Title      string  `json:"title"`
	ModuleType string  `json:"module_type"`
	Remark     *string `json:"remark"`
	// Status        string  `json:"status"`
	StartAt *string `json:"start_at"`
	EndAt   *string `json:"end_at"`
	Color   *string `json:"color"`

	Steps []UpdateScheduleStepRequest `json:"steps"`
}

type CreateScheduleNoRefRequest struct {
	AssigneeID *uint `json:"assignee_id"`
	// SalesOrderID  uint    `json:"sales_order_id"`
	CustomerID *uint   `json:"customer_id"`
	UUID       *string `json:"uuid"`
	StepsID    *uint   `json:"steps_id"`
	Title      string  `json:"title"`
	ModuleType string  `json:"module_type"`
	Remark     *string `json:"remark"`
	// Status        string  `json:"status"`
	StartAt *string `json:"start_at"`
	EndAt   *string `json:"end_at"`
	Color   *string `json:"color"`

	Steps []UpdateScheduleStepRequest `json:"steps"`
}

type CreateScheduleStepRequest struct {
	AssigneeID *uint   `json:"assignee_id"`
	ParentID   *uint   `json:"parent_id"`
	EntityID   *uint   `json:"entity_id"`
	EntityType string  `json:"entity_type"`
	UUID       *string `json:"uuid"`
	ParentUUID *string `json:"parent_uuid"`
	Title      string  `json:"title"`
	Remark     *string `json:"remark"`
	OrderItem  *int    `json:"order_item"`
	Color      *string `json:"color"`
	IsChecked  *int    `json:"is_checked"`
	// Locations   *string `json:"locations"`
	StartAt *string `json:"start_at"`
	EndAt   *string `json:"end_at"`

	Tasks []CreateScheduleTaskRequest `json:"tasks"`
}

type CreateScheduleTaskRequest struct {
	AssigneeID *uint   `json:"assignee_id"`
	ParentID   *uint   `json:"parent_id"`
	EntityID   *uint   `json:"entity_id"`
	EntityType string  `json:"entity_type"`
	UUID       *string `json:"uuid"`
	ParentUUID *string `json:"parent_uuid"`
	Title      string  `json:"title"`
	Remark     *string `json:"remark"`
	OrderItem  *int    `json:"order_item"`
	Color      *string `json:"color"`
	IsChecked  *int    `json:"is_checked"`
	// Locations   *string `json:"locations"`
	StartAt *string `json:"start_at"`
	EndAt   *string `json:"end_at"`
}

type UpdateSoDtsRequest struct {
	ID              *uint                     `json:"id"`
	SoDtID          *uint                     `json:"so_dt_id"`
	ProductUuid     string                    `json:"product_uuid"`
	SalesOrderID    *uint                     `json:"sales_order_id"`
	ItemUnitID      *uint                     `json:"item_unit_id"`
	VatID           *uint                     `json:"vat_id"`
	Pph23ID         *uint                     `json:"pph23_id"`
	RefID           *uint                     `json:"ref_id"`
	ItemID          uint                      `json:"item_id"`
	RefType         *string                   `json:"ref_type"`
	ItemType        *string                   `json:"item_type"`
	GenCode         *string                   `json:"gen_code"`
	Remark          *string                   `json:"remark"`
	VatPerc         *float64                  `json:"vat_perc"`
	VatPercAm       *float64                  `json:"vat_perc_am"`
	Pph23Perc       *float64                  `json:"pph23_perc"`
	Pph23PercAm     *float64                  `json:"pph23_perc_am"`
	MarkupPerc      *float64                  `json:"markup_perc"`
	MarkupPercAm    *float64                  `json:"markup_perc_am"`
	IsVat           *int8                     `json:"is_vat"`
	IsPph23         *int8                     `json:"is_pph23"`
	IsLockMarkup    *int8                     `json:"is_lock_markup"`
	IsLockPriceSell *int8                     `json:"is_lock_price_sell"`
	Qty             *float64                  `json:"qty"`
	PriceSell       *float64                  `json:"price_sell"`
	PriceBuy        *float64                  `json:"price_buy"`
	SubtotalSell    *float64                  `json:"subtotal_sell"`
	SubtotalBuy     *float64                  `json:"subtotal_buy"`
	DiscAm          *float64                  `json:"disc_am"`
	DiscPerc        *float64                  `json:"disc_perc"`
	DiscPercNum     *float64                  `json:"disc_perc_num"`
	DiscPercAm      *float64                  `json:"disc_perc_am"`
	DiscFinal       *float64                  `json:"disc_final"`
	DiscType        *string                   `json:"disc_type"`
	TotalAm         *float64                  `json:"total_am"`
	SoDtsBoms       []*UpdateSoDtsBomsRequest `json:"so_dts_boms" gorm:"-"`
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
	ID            uint                             `json:"id"`
	SalesOrderID  *uint                            `json:"sales_order_id"`
	CustomerID    *uint                            `json:"customer_id"`
	OrderTypeID   *uint                            `json:"order_type_id"`
	CurrencyID    *uint                            `json:"currency_id"`
	WarehouseID   *uint                            `json:"warehouse_id"`
	VatID         *uint                            `json:"vat_id"`
	PaymentID     *uint                            `json:"payment_id"`
	Pph23ID       *uint                            `json:"pph23_id"`
	BranchID      *uint                            `json:"branch_id"`
	RevNo         *int                             `json:"rev_no"`
	PoBuyerNo     *string                          `json:"po_buyer_no"`
	PoBuyerNoOri  *string                          `json:"po_buyer_no_ori"`
	SalesOrderNo  *string                          `json:"sales_order_no"`
	Remark        *string                          `json:"remark"`
	ShipDest      *string                          `json:"ship_dest"`
	Status        string                           `json:"status"`
	ExchangeRate  *float64                         `json:"exchange_rate"`
	VatPerc       *float64                         `json:"vat_perc"`
	Pph23Perc     *float64                         `json:"pph23_perc"`
	MarkupPerc    *float64                         `json:"markup_perc"`
	IsVat         *int                             `json:"is_vat"`
	IsPph23       *int                             `json:"is_pph23"`
	DiscAm        *float64                         `json:"disc_am"`
	DiscPerc      *float64                         `json:"disc_perc"`
	DiscPercAm    *float64                         `json:"disc_perc_am"`
	DiscFinal     *float64                         `json:"disc_final"`
	DiscType      *string                          `json:"disc_type"`
	TotalQty      *float64                         `json:"total_qty"`
	Subtotal      *float64                         `json:"subtotal"`
	TotalDiscount *float64                         `json:"total_discount"`
	TotalPph23    *float64                         `json:"total_pph23"`
	TotalVat      *float64                         `json:"total_vat"`
	GrandTotal    *float64                         `json:"grand_total"`
	OrderAt       *string                          `json:"order_at"`
	ShippingAt    *string                          `json:"shipping_at"`
	AgreeAt       *string                          `json:"agree_at"`
	DueAt         *string                          `json:"due_at"`
	SoDts         []UpdateSoDtsRequest             `json:"so_dts"`
	Attachments   []UpdateSalesOrderAttachmentsDTO `json:"attachments"`
	DeletedFiles  []uint                           `json:"deleted_files"`

	CustomerCode string `json:"customer_code"`
}

type UpdateSalesOrderAttachmentsDTO struct {
	ID       *uint   `json:"id" db:"id"`
	RefID    *uint   `json:"ref_id" db:"ref_id"`
	RefType  *string `json:"ref_type" db:"ref_type"`
	FileType *string `json:"file_type" db:"file_type"`
	FileUrl  *string `json:"file_url" db:"file_url"`
	FileName *string `json:"file_name" db:"file_name"`
	Remark   *string `json:"remark" db:"remark"`
}

type GetSalesOrderByIDRequest struct {
	ID interface{} `json:"id"`
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

	CreatedByID   *uint                      `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint                      `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint                      `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName *string                    `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string                    `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string                    `json:"created_at" db:"created_at"`
	UpdatedAt     *string                    `json:"updated_at" db:"updated_at"`
	DeleteAt      *string                    `json:"deleted_at" db:"deleted_at"`
	SoDts         []SalesOrderSoDtListDTO    `json:"so_dts"`
	Schedule      *ScheduleDetailDTO         `json:"schedule"`
	Attachments   []SalesOrderAttachmentsDTO `json:"attachments"`
}

type AppScheduleDetailDTO struct {
	ID          uint                     `json:"id" db:"id"`
	Schedule    *ScheduleSingleDetailDTO `json:"schedule"`
	Attachments []ScheduleAttachmentsDTO `json:"attachments"`
}

type SalesOrderAttachmentsDTO struct {
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

type ScheduleDetailDTO struct {
	ID                 uint    `json:"id" db:"id"`
	AssigneeID         *uint   `json:"assignee_id" db:"assignee_id"`
	CustomerID         *uint   `json:"customer_id" db:"customer_id"`
	SalesOrderID       uint    `json:"sales_order_id" db:"sales_order_id"`
	StepsID            *uint   `json:"steps_id" db:"steps_id"`
	UUID               *string `json:"uuid" db:"uuid"`
	Title              string  `json:"title" db:"title"`
	ModuleType         *string `json:"module_type" db:"module_type"`
	Remark             *string `json:"remark" db:"remark"`
	Status             string  `json:"status" db:"status"`
	StartAt            *string `json:"start_at" db:"start_at"`
	EndAt              *string `json:"end_at" db:"end_at"`
	Color              *string `json:"color" db:"color"`
	TotalTaskStep4Done *int    `json:"total_task_step_4_done" db:"total_task_step_4_done"`
	TotalAllTasksDone  *int    `json:"total_all_tasks_done" db:"total_all_tasks_done"`
	TotalTasks         *int    `json:"total_tasks" db:"total_tasks"`
	CreatedByID        *uint   `json:"created_by_id" db:"created_by_id"`
	UpdatedByID        *uint   `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID        *uint   `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName      *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName      *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt          *string `json:"created_at" db:"created_at"`
	UpdatedAt          *string `json:"updated_at" db:"updated_at"`
	DeleteAt           *string `json:"deleted_at" db:"deleted_at"`

	AssigneeName *string `json:"assignee_name" db:"assignee_name"`

	Steps []ScheduleStepListDTO `json:"steps"`
}

type ScheduleStepListDTO struct {
	ID         *uint   `json:"id" db:"id"`
	Uuid       *string `json:"uuid" db:"uuid"`
	StepIndex  *int    `json:"step_index" db:"step_index"`
	ScheduleID *uint   `json:"schedule_id" db:"schedule_id"`
	AssigneeID *uint   `json:"assignee_id" db:"assignee_id"`
	ParentID   *uint   `json:"parent_id" db:"parent_id"`
	EntityID   *uint   `json:"entity_id" db:"entity_id"`
	EntityType *string `json:"entity_type" db:"entity_type"`
	ParentUUID *string `json:"parent_uuid" db:"parent_uuid"`
	Title      *string `json:"title" db:"title"`
	Remark     *string `json:"remark" db:"remark"`
	OrderItem  *int    `json:"order_item" db:"order_item"`
	Color      *string `json:"color" db:"color"`
	StartAt    *string `json:"start_at" db:"start_at"`
	EndAt      *string `json:"end_at" db:"end_at"`

	CreatedByID   *uint   `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint   `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint   `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeleteAt      *string `json:"deleted_at" db:"deleted_at"`

	Tasks []ScheduleTaskListDTO `json:"tasks"`
}

type UpdatedScheduleStepListDTO struct {
	ID         *uint   `json:"id" db:"id"`
	Uuid       *string `json:"uuid" db:"uuid"`
	StepIndex  *int    `json:"step_index" db:"step_index"`
	ScheduleID *uint   `json:"schedule_id" db:"schedule_id"`
	AssigneeID *uint   `json:"assignee_id" db:"assignee_id"`
	ParentID   *uint   `json:"parent_id" db:"parent_id"`
	EntityID   *uint   `json:"entity_id" db:"entity_id"`
	EntityType *string `json:"entity_type" db:"entity_type"`
	ParentUUID *string `json:"parent_uuid" db:"parent_uuid"`
	Title      *string `json:"title" db:"title"`
	Remark     *string `json:"remark" db:"remark"`
	OrderItem  *int    `json:"order_item" db:"order_item"`
	Color      *string `json:"color" db:"color"`
	StartAt    *string `json:"start_at" db:"start_at"`
	EndAt      *string `json:"end_at" db:"end_at"`

	CreatedByID   *uint   `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint   `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint   `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeleteAt      *string `json:"deleted_at" db:"deleted_at"`
}

type ScheduleTaskListDTO struct {
	ID         *uint   `json:"id" db:"id"`
	Uuid       *string `json:"uuid" db:"uuid"`
	ParentID   *uint   `json:"parent_id" db:"parent_id"`
	ParentUUID *string `json:"parent_uuid" db:"parent_uuid"`
	ScheduleID *uint   `json:"schedule_id" db:"schedule_id"`
	AssigneeID *uint   `json:"assignee_id" db:"assignee_id"`
	EntityID   *uint   `json:"entity_id" db:"entity_id"`
	EntityType *string `json:"entity_type" db:"entity_type"`
	Title      *string `json:"title" db:"title"`
	Remark     *string `json:"remark" db:"remark"`
	OrderItem  *int    `json:"order_item" db:"order_item"`
	Color      *string `json:"color" db:"color"`
	StartAt    *string `json:"start_at" db:"start_at"`
	EndAt      *string `json:"end_at" db:"end_at"`
	IsChecked  *int    `json:"is_checked" db:"is_checked"`

	CreatedByID   *uint   `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint   `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint   `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt     *string `json:"created_at" db:"created_at"`
	UpdatedAt     *string `json:"updated_at" db:"updated_at"`
	DeleteAt      *string `json:"deleted_at" db:"deleted_at"`
}

type SalesOrderSoDtListDTO struct {
	ID               *uint    `json:"id" db:"id"`
	SoDtID           *uint    `json:"so_dt_id" db:"so_dt_id"`
	ProductUuid      *string  `json:"product_uuid" db:"product_uuid"`
	CustomerID       *uint    `json:"customer_id" db:"customer_id"`
	SalesOrderID     *uint    `json:"sales_order_id" db:"sales_order_id"`
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

	SoDtsBoms []SalesOrderSoDtBomListDTO `json:"so_dts_boms"`

	RefNum *string `json:"ref_num" db:"ref_num"`
}

type SalesOrderSoDtListUpdateDTO struct {
	ID               *uint    `json:"id" db:"id"`
	SoDtID           *uint    `json:"so_dt_id" db:"so_dt_id"`
	ProductUuid      *string  `json:"product_uuid" db:"product_uuid"`
	SalesOrderID     *uint    `json:"sales_order_id" db:"sales_order_id"`
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

type GetRefIndexQuoDtsRequest struct {
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

type RefIndexQuoDtListDTO struct {
	ID               *uint     `json:"id" db:"id"`
	QuoDtID          *uint     `json:"quo_dt_id" db:"quo_dt_id"`
	ProductUuid      *string   `json:"product_uuid" db:"product_uuid"`
	QuotationID      *uint     `json:"quotation_id" db:"quotation_id"`
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
	Pph23Perc        *float64  `json:"pph23_perc" db:"pph23_perc"`
	Pph23PercAm      *float64  `json:"pph23_perc_am" db:"pph23_perc_am"`
	MarkupPerc       *float64  `json:"markup_perc" db:"markup_perc"`
	MarkupPercAm     *float64  `json:"markup_perc_am" db:"markup_perc_am"`
	IsVat            *int8     `json:"is_vat" db:"is_vat"`
	IsPph23          *int8     `json:"is_pph23" db:"is_pph23"`
	IsLockMarkup     *int8     `json:"is_lock_markup" db:"is_lock_markup"`
	IsLockPriceSell  *int8     `json:"is_lock_price_sell" db:"is_lock_price_sell"`
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
	RefNum         *string  `json:"ref_num" db:"ref_num"`
	HeadIsVat      *int     `json:"head_is_vat" db:"head_is_vat"`
	ExchangeRate   *float64 `json:"exchange_rate" db:"exchange_rate"`
	QuoNo          *string  `json:"quo_no" db:"quo_no"`
	CustomerName   *string  `json:"customer_name" db:"customer_name"`
	ItemSku        *string  `json:"item_sku" db:"item_sku"`
	DueAt          *string  `json:"due_at" db:"due_at"`

	QuoDtsBoms []QuotationQuoDtBomListDTO `json:"quo_dts_boms"`
}

type GetQuoDtQtyUpdateDTO struct {
	ID          *uint    `json:"id" db:"id"`
	QuoDtID     *uint    `json:"quo_dt_id" db:"quo_dt_id"`
	QtySO       *float64 `json:"qty_so" db:"qty_so"`
	QuotationID *uint    `json:"quotation_id" db:"quotation_id"`
}

type UpdateQuotationStatusRequest struct {
	ID     uint   `json:"id"`
	Status string `json:"status"`
}

type UpdateSalesOrderScheduleRequest struct {
	ID           uint    `json:"id"`
	AssigneeID   *uint   `json:"assignee_id"`
	CustomerID   *uint   `json:"customer_id"`
	SalesOrderID uint    `json:"sales_order_id"`
	StepsID      *uint   `json:"steps_id"`
	UUID         *string `json:"uuid"`
	Title        string  `json:"title"`
	ModuleType   string  `json:"module_type"`
	Remark       *string `json:"remark"`
	Status       string  `json:"status"`
	StartAt      *string `json:"start_at"`
	EndAt        *string `json:"end_at"`
	Color        *string `json:"color"`
	CreatedByID  *uint   `json:"created_by_id"`
	UpdatedByID  *uint   `json:"updated_by_id"`
	DeletedByID  *uint   `json:"deleted_by_id"`
	IsDelete     *int    `json:"is_delete"`

	Steps        []UpdateScheduleStepRequest      `json:"steps"`
	Attachments  []UpdateSalesOrderAttachmentsDTO `json:"attachments"`
	DeletedFiles []uint                           `json:"deleted_files"`
}

type UpdateScheduleRequest struct {
	ID           uint    `json:"id"`
	AssigneeID   *uint   `json:"assignee_id"`
	CustomerID   *uint   `json:"customer_id"`
	SalesOrderID *uint   `json:"sales_order_id"`
	StepsID      *uint   `json:"steps_id"`
	UUID         *string `json:"uuid"`
	Title        string  `json:"title"`
	ModuleType   string  `json:"module_type"`
	Remark       *string `json:"remark"`
	Status       string  `json:"status"`
	StartAt      *string `json:"start_at"`
	EndAt        *string `json:"end_at"`
	Color        *string `json:"color"`
	CreatedByID  *uint   `json:"created_by_id"`
	UpdatedByID  *uint   `json:"updated_by_id"`
	DeletedByID  *uint   `json:"deleted_by_id"`
	IsDelete     *int    `json:"is_delete"`

	Steps []UpdateScheduleStepRequest `json:"steps"`
}

type ListScheduleTaskByScheduleID struct {
	ID            *uint `json:"id" db:"id"`
	ScheduleID    *uint `json:"schedule_id" db:"schedule_id"`
	StepOrderItem *int  `json:"step_order_item" db:"step_order_item"`
	IsChecked     *int  `json:"is_checked" db:"is_checked"`
}

type UpdateSalesOrderScheduleAppRequest struct {
	// sales_orders | feedbacks
	RefType      string                                   `json:"ref_type"`
	SalesOrderID *uint                                    `json:"sales_order_id"`
	ScheduleID   uint                                     `json:"schedule_id"`
	DeviceType   string                                   `json:"device_type"`
	DeletedFiles []uint                                   `json:"deleted_files"`
	Tasks        []UpdateSalesOrderScheduleAppTaskRequest `json:"tasks"`
	Attachments  []UpdateSalesOrderAttachmentsDTO         `json:"attachments"`
}

type UpdateSalesOrderScheduleAppTaskRequest struct {
	ID         *uint   `json:"id"`
	AssigneeID *uint   `json:"assignee_id"`
	Title      *string `json:"title"`
	Remark     *string `json:"remark"`
	IsChecked  *int    `json:"is_checked"`
}

type UpdateScheduleStepRequest struct {
	ID         *uint   `json:"id"`
	ScheduleID uint    `json:"schedule_id"`
	AssigneeID *uint   `json:"assignee_id"`
	ParentID   *uint   `json:"parent_id"`
	EntityID   *uint   `json:"entity_id"`
	EntityType string  `json:"entity_type"`
	UUID       *string `json:"uuid"`
	ParentUUID *string `json:"parent_uuid"`
	Title      string  `json:"title"`
	Remark     *string `json:"remark"`
	OrderItem  *int    `json:"order_item"`
	Color      *string `json:"color"`
	IsChecked  *int    `json:"is_checked"`
	StartAt    *string `json:"start_at"`
	EndAt      *string `json:"end_at"`

	Tasks []UpdateScheduleTaskRequest `json:"tasks"`
}

type UpdateScheduleTaskRequest struct {
	ID         *uint   `json:"id"`
	ScheduleID uint    `json:"schedule_id"`
	AssigneeID *uint   `json:"assignee_id"`
	ParentID   *uint   `json:"parent_id"`
	EntityID   *uint   `json:"entity_id"`
	EntityType string  `json:"entity_type"`
	UUID       *string `json:"uuid"`
	ParentUUID *string `json:"parent_uuid"`
	Title      string  `json:"title"`
	Remark     *string `json:"remark"`
	OrderItem  *int    `json:"order_item"`
	Color      *string `json:"color"`
	IsChecked  *int    `json:"is_checked"`
	StartAt    *string `json:"start_at"`
	EndAt      *string `json:"end_at"`
}
type UpdateScheduleTasksCheckAppRequest struct {
	Tasks []UpdateScheduleTaskCheckAppRequest      `json:"tasks"`
	Files []UpdateScheduleTasksCheckAppFileRequest `json:"files"`
}

type UpdateScheduleTasksCheckAppFileRequest struct {
	LastModified int64   `json:"last_modified"`
	Name         string  `json:"name"`
	Size         int64   `json:"size"`
	Type         string  `json:"type"`
	RefType      *string `json:"ref_type"`
	RefID        *uint   `json:"ref_id"`
	FileUrl      *string `json:"file_url"`
	FileName     *string `json:"file_name"`
	Remark       *string `json:"remark"`
	FileProp     *string `json:"file_prop"`
}

type UpdateScheduleTaskCheckAppRequest struct {
	ID         *uint   `json:"id"`
	AssigneeID *uint   `json:"assignee_id"`
	Remark     *string `json:"remark"`
	IsChecked  *int    `json:"is_checked"`
}

// {
// 	"tasks": [
// 		{
// 			"id": 1,
// 			"assignee_id": 1,
// 			"remark": "test",
// 			"is_checked": 1
// 		}
// 	],
// 	"deleted_files": [
// 		{
// 			"id": 1,
// 			"file_url": "https://example.com/file.jpg"
// 		}
// 	],
// 	"files": [
// 		{
// 			"last_modified": 1723451248056,
// 			"name": "PT Dalim.jpg",
// 			"size": 211233,
// 			"type": "image/jpeg"
// 		} // file & include
// 	]
// }

type ProjectAppListDTO struct {
	ID              int     `json:"id" db:"id"`
	Name            *string `json:"name" db:"name"`
	Client          *string `json:"client" db:"client"`
	Status          *string `json:"status" db:"status"`
	OrderAt         *string `json:"order_at" db:"order_at"`
	StartDate       *string `json:"start_date" db:"start_date"`
	EndDate         *string `json:"end_date" db:"end_date"`
	TotalTasks      *int    `json:"total_tasks" db:"total_tasks"`
	CompletedTasks  *int    `json:"completed_tasks" db:"completed_tasks"`
	ProgressPercent *int    `json:"progress_percent" db:"progress_percent"`
	OrderTypeName   *string `json:"order_type_name" db:"order_type_name"`
}

// "id": "p-001",
//
//	"name": "Project 2",
//	"description": "PT Project 2",
//	"status": "in_progress",
//	"start_date": "2023-01-20",
//	"end_date": "2023-02-20",
//	"task_completed": 10,
//	"task_total": 20,
//	"tasks": [
//	  {
//	    "id": "t-001",
//	    "title": "Task 1",
//	    "notes": "",
//	    "completed_by": "Ahmad",
//	    "completed_at": "2023-10-10T10:00:00",
//	    "is_done": true
//	  },
//	  {
//	    "id": "t-002",
//	    "title": "Task 2",
//	    "notes": "",
//	    "completed_by": "Ahmad",
//	    "completed_at": "2023-10-10T10:00:00",
//	    "is_done": true
//	  }
//	],
//	"attachments": [
//	  {
//	    "id": "a-001",
//	    "image_url": "https://example.com/image1.jpg",
//	    "uploaded_by": "Ahmad",
//	    "description": "Lorem ipsum dolor sit amet",
//	    "uploaded_at": "2022-10-10"
//	  },
//	  {
//	    "id": "a-002",
//	    "image_url": "https://example.com/image2.jpg",
//	    "uploaded_by": "Ahmad",
//	    "description": "Lorem ipsum dolor sit amet",
//	    "uploaded_at": "2022-10-10"
//	  }
//	]
type ProjectAppDetailDTO struct {
	ID              int     `json:"id" db:"id"`
	SalesOrderID    *uint   `json:"sales_order_id" db:"sales_order_id"`
	Name            *string `json:"name" db:"name"`
	Client          *string `json:"client" db:"client"`
	Status          *string `json:"status" db:"status"`
	OrderAt         *string `json:"order_at" db:"order_at"`
	StartDate       *string `json:"start_date" db:"start_date"`
	EndDate         *string `json:"end_date" db:"end_date"`
	TotalTasks      *int    `json:"total_tasks" db:"total_tasks"`
	CompletedTasks  *int    `json:"completed_tasks" db:"completed_tasks"`
	ProgressPercent *int    `json:"progress_percent" db:"progress_percent"`
	OrderTypeName   *string `json:"order_type_name" db:"order_type_name"`

	Tasks       []ScheduleTaskListDTO      `json:"tasks"`
	Attachments []SalesOrderAttachmentsDTO `json:"attachments"`
}

type ScheduleSingleDetailDTO struct {
	ID                 uint    `json:"id" db:"id"`
	AssigneeID         *uint   `json:"assignee_id" db:"assignee_id"`
	CustomerID         *uint   `json:"customer_id" db:"customer_id"`
	SalesOrderID       *uint   `json:"sales_order_id" db:"sales_order_id"`
	Title              string  `json:"title" db:"title"`
	ModuleType         *string `json:"module_type" db:"module_type"`
	CustomerName       *string `json:"customer_name" db:"customer_name"`
	StartAt            *string `json:"start_at" db:"start_at"`
	EndAt              *string `json:"end_at" db:"end_at"`
	Color              *string `json:"color" db:"color"`
	TotalTaskStep4Done *int    `json:"total_task_step_4_done" db:"total_task_step_4_done"`
	TotalAllTasksDone  *int    `json:"total_all_tasks_done" db:"total_all_tasks_done"`
	TotalTasks         *int    `json:"total_tasks" db:"total_tasks"`
	TotalTasks4        *int    `json:"total_tasks_4" db:"total_tasks_4"`
	CreatedByName      *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName      *string `json:"updated_by_name" db:"updated_by_name"`

	AssigneeName *string `json:"assignee_name" db:"assignee_name"`

	Steps       []ScheduleStepListDTO    `json:"steps"`
	Attachments []ScheduleAttachmentsDTO `json:"attachments"`
}

type CalendarListDTO struct {
	ID                 *uint   `json:"id" db:"id"`
	SalesOrderID       *uint   `json:"sales_order_id" db:"sales_order_id"`
	CustomerName       *string `json:"customer_name" db:"customer_name"`
	OrderAt            *string `json:"order_at" db:"order_at"`
	Start              *string `json:"start" db:"start"`
	End                *string `json:"end" db:"end"`
	Title              *string `json:"title" db:"title"`
	Color              *string `json:"color" db:"color"`
	TaskTitle          *string `json:"task_title" db:"task_title"`
	TotalTaskStep4Done *int    `json:"total_task_step_4_done" db:"total_task_step_4_done"`
	TotalAllTasksDone  *int    `json:"total_all_tasks_done" db:"total_all_tasks_done"`
	TotalTasks         *int    `json:"total_tasks" db:"total_tasks"`
	TotalTasks4        *int    `json:"total_tasks_4" db:"total_tasks_4"`
	OrderTypeName      *string `json:"order_type_name" db:"order_type_name"`
}

type ScheduleAttachmentsDTO struct {
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
