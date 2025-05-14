package dtos

import "time"

type GetRequestOrdersRequest struct {
	Global         *string `json:"global"`
	RequestNo      *string `json:"request_no"`
	Remark         *string `json:"remark"`
	Requested      *string `json:"requested"`
	RevNo          *int    `json:"rev_no"`
	BranchID       *int    `json:"branch_id"`
	WarehouseID    *int    `json:"warehouse_id"`
	Status         *string `json:"status"`
	StartDate      *string `json:"start_date"`
	EndDate        *string `json:"end_date"`
	PerPage        *string `json:"per_page" default:"10"`
	Page           *string `json:"page" default:"1"`
	OrderColumn    *string `json:"order_column" default:"id"`
	OrderDirection *string `json:"order_direction" default:"asc"`
}

type CreateRequestOrderDtRequest struct {
	ProductUuid     string   `json:"product_uuid"`
	ItemUnitID      *uint    `json:"item_unit_id"`
	RefID           *uint    `json:"ref_id"`
	ProductID       *uint    `json:"product_id"`
	ItemID          *uint    `json:"item_id"`
	RefType         *string  `json:"ref_type"`
	RefJSON         *string  `json:"ref_json"`
	ProductType     *string  `json:"product_type"`
	ProductName     *string  `json:"product_name"`
	ItemName        *string  `json:"item_name"`
	UnitName        *string  `json:"unit_name"`
	PriceSell       *float64 `json:"price_sell"`
	Remark          *string  `json:"remark"`
	ProductJSON     *string  `json:"product_json"`
	OrderProductQty *float64 `json:"order_product_qty"`
	OrderItemQty    *float64 `json:"order_item_qty"`
	WhQty           *float64 `json:"wh_qty"`
	ReqQty          *float64 `json:"req_qty"`
}

type CreateRequestOrderRequest struct {
	BranchID                  *uint                         `json:"branch_id"`
	WarehouseID               *uint                         `json:"warehouse_id"`
	RequestNo                 *string                       `json:"request_no"`
	RequestDate               *string                       `json:"request_date"`
	Remark                    *string                       `json:"remark"`
	Requested                 *string                       `json:"requested"`
	RevNo                     *int                          `json:"rev_no"`
	Status                    *string                       `json:"status"`
	GrandTotalOrderProductQty *float64                      `json:"grand_total_order_product_qty"`
	GrandTotalOrderItemQty    *float64                      `json:"grand_total_order_item_qty"`
	GrandTotalWhQty           *float64                      `json:"grand_total_wh_qty"`
	GrandTotalReqQty          *float64                      `json:"grand_total_req_qty"`
	RequestOrderDts           []CreateRequestOrderDtRequest `json:"request_order_dts"`
}

type UpdateRequestOrderDtRequest struct {
	ID               *uint    `json:"id"`
	RequestOrderDtID *uint    `json:"request_order_dt_id"`
	ProductUuid      string   `json:"product_uuid"`
	RequestOrderID   *uint    `json:"request_order_id"`
	ItemUnitID       *uint    `json:"item_unit_id"`
	RefID            *uint    `json:"ref_id"`
	ProductID        *uint    `json:"product_id"`
	ItemID           *uint    `json:"item_id"`
	RefType          *string  `json:"ref_type"`
	RefJSON          *string  `json:"ref_json"`
	ProductType      *string  `json:"product_type"`
	ProductName      *string  `json:"product_name"`
	ItemName         *string  `json:"item_name"`
	UnitName         *string  `json:"unit_name"`
	PriceSell        *float64 `json:"price_sell"`
	Remark           *string  `json:"remark"`
	ProductJSON      *string  `json:"product_json"`
	OrderProductQty  *float64 `json:"order_product_qty"`
	OrderItemQty     *float64 `json:"order_item_qty"`
	WhQty            *float64 `json:"wh_qty"`
	ReqQty           *float64 `json:"req_qty"`
}

type UpdateRequestOrderRequest struct {
	ID                        uint                          `json:"id"`
	RequestOrderID            *uint                         `json:"request_order_id"`
	BranchID                  *uint                         `json:"branch_id"`
	WarehouseID               *uint                         `json:"warehouse_id"`
	RequestNo                 *string                       `json:"request_no"`
	RequestDate               *string                       `json:"request_date"`
	Remark                    *string                       `json:"remark"`
	Requested                 *string                       `json:"requested"`
	RevNo                     *int                          `json:"rev_no"`
	Status                    *string                       `json:"status"`
	GrandTotalOrderProductQty *float64                      `json:"grand_total_order_product_qty"`
	GrandTotalOrderItemQty    *float64                      `json:"grand_total_order_item_qty"`
	GrandTotalWhQty           *float64                      `json:"grand_total_wh_qty"`
	GrandTotalReqQty          *float64                      `json:"grand_total_req_qty"`
	RequestOrderDts           []UpdateRequestOrderDtRequest `json:"request_order_dts"`
}

type GetRequestOrderByIDRequest struct {
	ID uint `json:"id"`
}

type GetRequestOrderParams struct {
	ID        uint
	IsDeleted *int
}

func NewGetRequestOrderParams(id uint) *GetRequestOrderParams {
	defaultIsDeleted := 0
	return &GetRequestOrderParams{
		ID:        id,
		IsDeleted: &defaultIsDeleted,
	}
}

type GetRequestOrderDtParams struct {
	ID             uint
	RequestOrderID uint
	IsDeleted      *int
}

type DeleteRequestOrderRequest struct {
	ID uint `json:"id"`
}

type RequestOrderListDTO struct {
	ID                        int      `json:"id" db:"id"`
	RequestOrderID            *uint    `json:"request_order_id" db:"request_order_id"`
	BranchID                  *uint    `json:"branch_id" db:"branch_id"`
	WarehouseID               *uint    `json:"warehouse_id" db:"warehouse_id"`
	RequestNo                 *string  `json:"request_no" db:"request_no"`
	RequestDate               *string  `json:"request_date" db:"request_date"`
	Remark                    *string  `json:"remark" db:"remark"`
	Requested                 *string  `json:"requested" db:"requested"`
	RevNo                     *int     `json:"rev_no" db:"rev_no"`
	Status                    *string  `json:"status" db:"status"`
	GrandTotalOrderProductQty *float64 `json:"grand_total_order_product_qty" db:"grand_total_order_product_qty"`
	GrandTotalOrderItemQty    *float64 `json:"grand_total_order_item_qty" db:"grand_total_order_item_qty"`
	GrandTotalWhQty           *float64 `json:"grand_total_wh_qty" db:"grand_total_wh_qty"`
	GrandTotalReqQty          *float64 `json:"grand_total_req_qty" db:"grand_total_req_qty"`
	CreatedByID               *uint    `json:"created_by_id" db:"created_by_id"`
	UpdatedByID               *uint    `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID               *uint    `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName             *string  `json:"created_by_name" db:"created_by_name"`
	UpdatedByName             *string  `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt                 *string  `json:"created_at" db:"created_at"`
	UpdatedAt                 *string  `json:"updated_at" db:"updated_at"`
	DeleteAt                  *string  `json:"deleted_at" db:"deleted_at"`

	BranchName    *string `json:"branch_name" db:"branch_name"`
	WarehouseName *string `json:"warehouse_name" db:"warehouse_name"`
}

type RequestOrderDetailDTO struct {
	ID                        uint     `json:"id" db:"id"`
	RequestOrderID            *uint    `json:"request_order_id" db:"request_order_id"`
	BranchID                  *uint    `json:"branch_id" db:"branch_id"`
	WarehouseID               *uint    `json:"warehouse_id" db:"warehouse_id"`
	RequestNo                 *string  `json:"request_no" db:"request_no"`
	RequestDate               *string  `json:"request_date" db:"request_date"`
	Remark                    *string  `json:"remark" db:"remark"`
	Requested                 *string  `json:"requested" db:"requested"`
	RevNo                     *int     `json:"rev_no" db:"rev_no"`
	Status                    *string  `json:"status" db:"status"`
	GrandTotalOrderProductQty *float64 `json:"grand_total_order_product_qty" db:"grand_total_order_product_qty"`
	GrandTotalOrderItemQty    *float64 `json:"grand_total_order_item_qty" db:"grand_total_order_item_qty"`
	GrandTotalWhQty           *float64 `json:"grand_total_wh_qty" db:"grand_total_wh_qty"`
	GrandTotalReqQty          *float64 `json:"grand_total_req_qty" db:"grand_total_req_qty"`

	CreatedByID     *uint                   `json:"created_by_id" db:"created_by_id"`
	UpdatedByID     *uint                   `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID     *uint                   `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedByName   *string                 `json:"created_by_name" db:"created_by_name"`
	UpdatedByName   *string                 `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt       *string                 `json:"created_at" db:"created_at"`
	UpdatedAt       *string                 `json:"updated_at" db:"updated_at"`
	DeleteAt        *string                 `json:"deleted_at" db:"deleted_at"`
	RequestOrderDts []RequestOrderDtListDTO `json:"request_order_dts"`
}

type RequestOrderDtListDTO struct {
	ID               *uint    `json:"id" db:"id"`
	RequestOrderDtID *uint    `json:"request_order_dt_id" db:"request_order_dt_id"`
	ProductUuid      *string  `json:"product_uuid" db:"product_uuid"`
	RequestOrderID   *uint    `json:"request_order_id" db:"request_order_id"`
	ItemUnitID       *uint    `json:"item_unit_id" db:"item_unit_id"`
	RefID            *uint    `json:"ref_id" db:"ref_id"`
	ProductID        *uint    `json:"product_id" db:"product_id"`
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
	ProductType      *string  `json:"product_type" db:"product_type"`
	ProductName      *string  `json:"product_name" db:"product_name"`
	ProductCode      *string  `json:"product_code" db:"product_code"`
	Remark           *string  `json:"remark" db:"remark"`
	PriceSell        *float64 `json:"price_sell" db:"price_sell"`
	OrderProductQty  *float64 `json:"order_product_qty" db:"order_product_qty"`
	OrderItemQty     *float64 `json:"order_item_qty" db:"order_item_qty"`
	WhQty            *float64 `json:"wh_qty" db:"wh_qty"`
	ReqQty           *float64 `json:"req_qty" db:"req_qty"`

	CreatedByName *string   `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string   `json:"updated_by_name" db:"updated_by_name"`
	CreatedByID   *uint     `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint     `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint     `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *string   `json:"updated_at" db:"updated_at"`
	DeleteAt      *string   `json:"deleted_at" db:"deleted_at"`

	RefNum       *string `json:"ref_num" db:"ref_num"`
	SalesOrderID *uint   `json:"sales_order_id" db:"sales_order_id"`
}

type RequestOrderDtListUpdateDTO struct {
	ID               *uint    `json:"id" db:"id"`
	RequestOrderDtID *uint    `json:"request_order_dt_id" db:"request_order_dt_id"`
	ProductUuid      *string  `json:"product_uuid" db:"product_uuid"`
	RequestOrderID   *uint    `json:"request_order_id" db:"request_order_id"`
	ItemUnitID       *uint    `json:"item_unit_id" db:"item_unit_id"`
	RefID            *uint    `json:"ref_id" db:"ref_id"`
	ProductID        *uint    `json:"product_id" db:"product_id"`
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
	ProductType      *string  `json:"product_type" db:"product_type"`
	ProductName      *string  `json:"product_name" db:"product_name"`
	Remark           *string  `json:"remark" db:"remark"`
	PriceSell        *float64 `json:"price_sell" db:"price_sell"`
	OrderProductQty  *float64 `json:"order_product_qty" db:"order_product_qty"`
	OrderItemQty     *float64 `json:"order_item_qty" db:"order_item_qty"`
	WhQty            *float64 `json:"wh_qty" db:"wh_qty"`
	ReqQty           *float64 `json:"req_qty" db:"req_qty"`

	CreatedByName *string   `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string   `json:"updated_by_name" db:"updated_by_name"`
	CreatedByID   *uint     `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint     `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint     `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *string   `json:"updated_at" db:"updated_at"`
	DeleteAt      *string   `json:"deleted_at" db:"deleted_at"`
}

type GetRequestOrdersResult struct {
	RequestOrders []RequestOrderListDTO
	Total         int
	Err           error
}

type GetRefSalesOrderForRequestOrderRequest struct {
	Global         *string `json:"global"`
	RequestOrderID *string `json:"request_order_id"`
	SalesOrderNo   *string `json:"sales_order_no"`
	PoBuyerNo      *string `json:"po_buyer_no"`
	Remark         *string `json:"remark"`
	CustomerID     *int    `json:"customer_id"`
	WarehouseID    *int    `json:"warehouse_id"`
	OrderTypeID    *int    `json:"order_type_id"`
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

type RefSalesOrderForRequestOrderListDTO struct {
	ID       *uint `json:"id" db:"id"`
	BranchID *uint `json:"branch_id" db:"branch_id"`
	SoDtID   *uint `json:"so_dt_id" db:"so_dt_id"`
	// ProductUuid      *string   `json:"product_uuid" db:"product_uuid"`
	SalesOrderID     *uint     `json:"sales_order_id" db:"sales_order_id"`
	ItemUnitID       *uint     `json:"item_unit_id" db:"item_unit_id"`
	RefID            *uint     `json:"ref_id" db:"ref_id"`
	ItemID           *uint     `json:"item_id" db:"item_id"`
	ProductID        *uint     `json:"product_id" db:"product_id"`
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
	ProductType      *string   `json:"product_type" db:"product_type"`
	ProductName      *string   `json:"product_name" db:"product_name"`
	ProductCode      *string   `json:"product_code" db:"product_code"`
	GenCode          *string   `json:"gen_code" db:"gen_code"`
	Remark           *string   `json:"remark" db:"remark"`
	PriceSell        *float64  `json:"price_sell" db:"price_sell"`
	Qty              *float64  `json:"qty" db:"qty"`
	QtyRequested     *float64  `json:"qty_requested" db:"qty_requested"`
	OrderProductQty  *float64  `json:"order_product_qty" db:"order_product_qty"`
	OrderItemQty     *float64  `json:"order_item_qty" db:"order_item_qty"`
	WhQty            *float64  `json:"wh_qty" db:"wh_qty"`
	ReqQty           *float64  `json:"req_qty" db:"req_qty"`
	RequestStatus    *string   `json:"request_status" db:"request_status"`
	CreatedByName    *string   `json:"created_by_name" db:"created_by_name"`
	UpdatedByName    *string   `json:"updated_by_name" db:"updated_by_name"`
	CreatedByID      *uint     `json:"created_by_id" db:"created_by_id"`
	UpdatedByID      *uint     `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID      *uint     `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        *string   `json:"updated_at" db:"updated_at"`
	DeleteAt         *string   `json:"deleted_at" db:"deleted_at"`

	CustomerID    *uint   `json:"customer_id" db:"customer_id"`
	OrderTypeID   *uint   `json:"order_type_id" db:"order_type_id"`
	HeadRemark    *string `json:"head_remark" db:"head_remark"`
	SalesOrderNo  *string `json:"sales_order_no" db:"sales_order_no"`
	PoBuyerNo     *string `json:"po_buyer_no" db:"po_buyer_no"`
	CustomerName  *string `json:"customer_name" db:"customer_name"`
	OrderTypeName *string `json:"order_type_name" db:"order_type_name"`
	OrderDate     *string `json:"order_date" db:"order_date"`
	ShippingDate  *string `json:"shipping_date" db:"shipping_date"`
	ItemSku       *string `json:"item_sku" db:"item_sku"`
	DueAt         *string `json:"due_at" db:"due_at"`
}

type GetRefProductForRequestOrderRequest struct {
	Global         *string `json:"global"`
	RequestOrderID *string `json:"request_order_id"`
	ProductCode    *string `json:"product_code"`
	ProductName    *string `json:"product_name"`
	ItemCode       *string `json:"item_code"`
	ItemName       *string `json:"item_name"`
	ItemGroupID    *int    `json:"item_group_id"`
	ItemSubGroupID *int    `json:"item_sub_group_id"`
	BranchID       *int    `json:"branch_id"`
	Status         *string `json:"status"`
	SpecificIDs    *string `json:"specific_ids"`
	PerPage        *string `json:"per_page" default:"10"`
	Page           *string `json:"page" default:"1"`
	OrderColumn    *string `json:"order_column" default:"id"`
	OrderDirection *string `json:"order_direction" default:"asc"`
}

type RefProductForRequestOrderListDTO struct {
	ID          *uint   `json:"id" db:"id"`
	ProductID   *uint   `json:"product_id" db:"product_id"`
	ProductCode *string `json:"product_code" db:"product_code"`
	ProductName *string `json:"product_name" db:"product_name"`
	// ProductType      *string  `json:"product_type" db:"product_type"`
	ItemUnitID       *uint    `json:"item_unit_id" db:"item_unit_id"`
	ItemID           *uint    `json:"item_id" db:"item_id"`
	ItemCode         *string  `json:"item_code" db:"item_code"`
	ItemName         *string  `json:"item_name" db:"item_name"`
	ItemType         *string  `json:"item_type" db:"item_type"`
	ItemSku          *string  `json:"item_sku" db:"item_sku"`
	ItemSubGroupID   *uint    `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID      *uint    `json:"item_group_id" db:"item_group_id"`
	ItemSubGroupName *string  `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName    *string  `json:"item_group_name" db:"item_group_name"`
	UnitName         *string  `json:"unit_name" db:"unit_name"`
	GenCode          *string  `json:"gen_code" db:"gen_code"`
	Remark           *string  `json:"remark" db:"remark"`
	PriceSell        *float64 `json:"price_sell" db:"price_sell"`
	OrderProductQty  *float64 `json:"order_product_qty" db:"order_product_qty"`
	OrderItemQty     *float64 `json:"order_item_qty" db:"order_item_qty"`
	WhQty            *float64 `json:"wh_qty" db:"wh_qty"`
	ReqQty           *float64 `json:"req_qty" db:"req_qty"`

	RefID   *uint   `json:"ref_id" db:"ref_id"`
	RefType *string `json:"ref_type" db:"ref_type"`
	RefJSON *string `json:"ref_json" db:"ref_json"`
	RefNum  *string `json:"ref_num" db:"ref_num"`

	BranchID *uint `json:"branch_id" db:"branch_id"`

	CreatedByName *string   `json:"created_by_name" db:"created_by_name"`
	UpdatedByName *string   `json:"updated_by_name" db:"updated_by_name"`
	CreatedByID   *uint     `json:"created_by_id" db:"created_by_id"`
	UpdatedByID   *uint     `json:"updated_by_id" db:"updated_by_id"`
	DeletedByID   *uint     `json:"deleted_by_id" db:"deleted_by_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *string   `json:"updated_at" db:"updated_at"`
	DeleteAt      *string   `json:"deleted_at" db:"deleted_at"`
}

type RequestOrderStatusWidget struct {
	Status     string  `json:"status" db:"status"`
	OrderCount int     `json:"order_count" db:"order_count"`
	TotalQty   float64 `json:"total_qty" db:"total_qty"`
}
