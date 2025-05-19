package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

type RequestOrderDt struct {
	gorm.Model
	ID              uint             `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ProductUuid     string           `json:"product_uuid" gorm:"column:product_uuid"`
	RequestOrderID  *uint            `json:"request_order_id" gorm:"column:request_order_id"`
	ItemUnitID      *uint            `json:"item_unit_id" gorm:"column:item_unit_id"`
	RefID           *uint            `json:"ref_id" gorm:"column:ref_id"`
	ProductID       *uint            `json:"product_id" gorm:"column:product_id"`
	ItemID          *uint            `json:"item_id" gorm:"column:item_id"`
	RefType         *string          `json:"ref_type" gorm:"column:ref_type"`
	RefJSON         *json.RawMessage `json:"ref_json" gorm:"column:ref_json"`
	ProductType     *string          `json:"product_type" gorm:"column:product_type"`
	ProductName     *string          `json:"product_name" gorm:"column:product_name"`
	ItemName        *string          `json:"item_name" gorm:"column:item_name"`
	UnitName        *string          `json:"unit_name" gorm:"column:unit_name"`
	PriceSell       *float64         `json:"price_sell" gorm:"column:price_sell"`
	Remark          *string          `json:"remark" gorm:"column:remark"`
	ProductJSON     *json.RawMessage `json:"product_json" gorm:"column:product_json"`
	OrderProductQty *float64         `json:"order_product_qty" gorm:"column:order_product_qty"`
	OrderItemQty    *float64         `json:"order_item_qty" gorm:"column:order_item_qty"`
	WhQty           *float64         `json:"wh_qty" gorm:"column:wh_qty"`
	ReqQty          *float64         `json:"req_qty" gorm:"column:req_qty"`
	QtyOut          *float64         `json:"qty_out" gorm:"column:qty_out"`
	CreatedByID     *uint            `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID     *uint            `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID     *uint            `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt       gorm.DeletedAt   `json:"deleted_at" gorm:"index"`
}

func (RequestOrderDt) TableName() string {
	return "request_order_dts"
}
