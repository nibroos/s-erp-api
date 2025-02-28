package dtos

type GetQuotationsRequest struct {
	Global         string `json:"global"`
	Name           string `json:"name"`
	Sku            string `json:"sku"`
	Barcode        string `json:"barcode"`
	ExpiredAt      string `json:"expired_at"`
	Remark         string `json:"remark"`
	FactoryCode    string `json:"factory_code"`
	ItemSubGroupID string `json:"item_sub_group_id"`
	UnitID         string `json:"unit_id"`
	Status         string `json:"status"`
	PerPage        string `json:"per_page" default:"10"`         // Default per_page to 10
	Page           string `json:"page" default:"1"`              // Default page to 1
	OrderColumn    string `json:"order_column" default:"id"`     // Default order column to "id"
	OrderDirection string `json:"order_direction" default:"asc"` // Default order direction to "asc"
}

type CreateQuoDtsRequest struct {
	QuotationItemID uint    `json:"product_item_id"`
	ItemUnitID      uint    `json:"item_unit_id"`
	Qty             float64 `json:"qty"`
	Remark          *string `json:"remark"`
}

type CreateQuotationRequest struct {
	ItemSubGroupID uint                       `json:"item_sub_group_id"`
	ItemUnitID     uint                       `json:"item_unit_id"`
	Code           *string                    `json:"code"`
	FactoryCode    *string                    `json:"factory_code"`
	Name           string                     `json:"name"`
	Sku            *string                    `json:"sku"`
	Barcode        *string                    `json:"barcode"`
	Specification  *string                    `json:"specification"`
	Description    *string                    `json:"description"`
	Remark         *string                    `json:"remark"`
	PriceSell      *float64                   `json:"price_sell"`
	PriceBuy       *float64                   `json:"price_buy"`
	Margin         *float64                   `json:"margin"`
	TpbCode        *string                    `json:"tpb_code"`
	MinimumStock   *float64                   `json:"minimum_stock"`
	IsAllBranch    *int                       `json:"is_all_branch"`
	Status         int8                       `json:"status"`
	ExpiredAt      *string                    `json:"expired_at"`
	Units          []CreateMsItemUnitsRequest `json:"units"`
	QuoDts         []CreateQuoDtsRequest      `json:"boms"`
}

type UpdateQuoDtsRequest struct {
	ID              *uint   `json:"id"`
	QuotationItemID uint    `json:"product_item_id"`
	ItemUnitID      uint    `json:"item_unit_id"`
	Qty             float64 `json:"qty"`
	Remark          *string `json:"remark"`
}

type UpdateQuotationRequest struct {
	ID             uint                  `json:"id"`
	ItemSubGroupID uint                  `json:"item_sub_group_id"`
	ItemUnitID     uint                  `json:"item_unit_id"`
	Code           *string               `json:"code"`
	FactoryCode    *string               `json:"factory_code"`
	Name           string                `json:"name"`
	Sku            *string               `json:"sku"`
	Barcode        *string               `json:"barcode"`
	Specification  *string               `json:"specification"`
	Description    *string               `json:"description"`
	Remark         *string               `json:"remark"`
	PriceSell      *float64              `json:"price_sell"`
	PriceBuy       *float64              `json:"price_buy"`
	Margin         *float64              `json:"margin"`
	TpbCode        *string               `json:"tpb_code"`
	MinimumStock   *float64              `json:"minimum_stock"`
	IsAllBranch    *int                  `json:"is_all_branch"`
	Status         int8                  `json:"status"`
	ExpiredAt      *string               `json:"expired_at"`
	QuoDts         []UpdateQuoDtsRequest `json:"boms"`
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
	ID               int     `json:"id" db:"id"`
	QuotationID      *uint   `json:"product_id" db:"product_id"`
	ItemSubGroupID   uint    `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID      *uint   `json:"item_group_id" db:"item_group_id"`
	ItemUnitID       *uint   `json:"item_unit_id" db:"item_unit_id"`
	ItemUnitUnitID   *uint   `json:"item_unit_unit_id" db:"item_unit_unit_id"`
	BranchID         *uint   `json:"branch_id" db:"branch_id"`
	BranchItemID     *uint   `json:"branch_item_id" db:"branch_item_id"`
	ItemSubGroupName *string `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName    *string `json:"item_group_name" db:"item_group_name"`
	UnitName         *string `json:"unit_name" db:"unit_name"`
	BranchName       *string `json:"branch_name" db:"branch_name"`
	Code             *string `json:"code" db:"code"`
	FactoryCode      *string `json:"factory_code" db:"factory_code"`
	Name             string  `json:"name" db:"name"`
	Sku              *string `json:"sku" db:"sku"`
	Barcode          *string `json:"barcode" db:"barcode"`
	Specification    *string `json:"specification" db:"specification"`
	Description      *string `json:"description" db:"description"`
	TpbCode          *string `json:"tpb_code" db:"tpb_code"`
	MinimumStock     *string `json:"minimum_stock" db:"minimum_stock"`
	IsAllBranch      *int    `json:"is_all_branch" db:"is_all_branch"`
	Remark           *string `json:"remark" db:"remark"`
	PriceSell        *string `json:"price_sell" db:"price_sell"`
	PriceBuy         *string `json:"price_buy" db:"price_buy"`
	Margin           *string `json:"margin" db:"margin"`
	ExpiredAt        *string `json:"expired_at" db:"expired_at"`
	Status           int8    `json:"status" db:"status"`
	CreatedByName    *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName    *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt        *string `json:"created_at" db:"created_at"`
	UpdatedAt        *string `json:"updated_at" db:"updated_at"`
	DeleteAt         *string `json:"deleted_at" db:"deleted_at"`
}

type QuotationQuoDtListDTO struct {
	ID                *uint   `json:"id" db:"id"`
	QuoDtID           *uint   `json:"bom_id" db:"bom_id"`
	QuotationID       uint    `json:"product_id" db:"product_id"`
	QuotationItemID   uint    `json:"product_item_id" db:"product_item_id"`
	ItemSubGroupID    *uint   `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID       *uint   `json:"item_group_id" db:"item_group_id"`
	ItemUnitID        *uint   `json:"item_unit_id" db:"item_unit_id"`
	ItemSubGroupName  *string `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName     *string `json:"item_group_name" db:"item_group_name"`
	Qty               float64 `json:"qty" db:"qty"`
	Remark            *string `json:"remark" db:"remark"`
	QuotationItemName *string `json:"product_item_name" db:"product_item_name"`
	ItemUnitName      *string `json:"item_unit_name" db:"item_unit_name"`
	CreatedByName     *string `json:"created_by_name" db:"created_by_name"`
	UpdatedByName     *string `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt         *string `json:"created_at" db:"created_at"`
	UpdatedAt         *string `json:"updated_at" db:"updated_at"`
	DeleteAt          *string `json:"deleted_at" db:"deleted_at"`
}

type QuotationDetailDTO struct {
	ID               uint                    `json:"id" db:"id"`
	QuotationID      *uint                   `json:"product_id" db:"product_id"`
	ItemSubGroupID   uint                    `json:"item_sub_group_id" db:"item_sub_group_id"`
	ItemGroupID      *uint                   `json:"item_group_id" db:"item_group_id"`
	ItemUnitID       uint                    `json:"item_unit_id" db:"item_unit_id"`
	ItemUnitUnitID   *uint                   `json:"item_unit_unit_id" db:"item_unit_unit_id"`
	BranchID         *uint                   `json:"branch_id" db:"branch_id"`
	BranchItemID     *uint                   `json:"branch_item_id" db:"branch_item_id"`
	ItemSubGroupName *string                 `json:"item_sub_group_name" db:"item_sub_group_name"`
	ItemGroupName    *string                 `json:"item_group_name" db:"item_group_name"`
	UnitName         *string                 `json:"unit_name" db:"unit_name"`
	BranchName       *string                 `json:"branch_name" db:"branch_name"`
	Code             *string                 `json:"code" db:"code"`
	FactoryCode      *string                 `json:"factory_code" db:"factory_code"`
	Name             string                  `json:"name" db:"name"`
	Sku              *string                 `json:"sku" db:"sku"`
	Barcode          *string                 `json:"barcode" db:"barcode"`
	Specification    *string                 `json:"specification" db:"specification"`
	Description      *string                 `json:"description" db:"description"`
	TpbCode          *string                 `json:"tpb_code" db:"tpb_code"`
	MinimumStock     *string                 `json:"minimum_stock" db:"minimum_stock"`
	IsAllBranch      *int                    `json:"is_all_branch" db:"is_all_branch"`
	Remark           *string                 `json:"remark" db:"remark"`
	PriceSell        *string                 `json:"price_sell" db:"price_sell"`
	PriceBuy         *string                 `json:"price_buy" db:"price_buy"`
	Margin           *string                 `json:"margin" db:"margin"`
	ExpiredAt        *string                 `json:"expired_at" db:"expired_at"`
	Status           int8                    `json:"status" db:"status"`
	CreatedByName    *string                 `json:"created_by_name" db:"created_by_name"`
	UpdatedByName    *string                 `json:"updated_by_name" db:"updated_by_name"`
	CreatedAt        *string                 `json:"created_at" db:"created_at"`
	UpdatedAt        *string                 `json:"updated_at" db:"updated_at"`
	DeleteAt         *string                 `json:"deleted_at" db:"deleted_at"`
	QuoDts           []QuotationQuoDtListDTO `json:"boms"`
}

type GetQuotationsResult struct {
	Quotations []QuotationListDTO
	Total      int
	Err        error
}
