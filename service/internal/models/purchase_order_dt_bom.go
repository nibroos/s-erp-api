package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

type PurchaseOrderDtBom struct {
	gorm.Model
	ID          uint             `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	PoID        *uint            `json:"po_id" gorm:"column:po_id"`
	PoDtID      *uint            `json:"po_dt_id" gorm:"column:po_dt_id"`
	BomID       *uint            `json:"bom_id" gorm:"column:bom_id"`
	ProductID   *uint            `json:"product_id" gorm:"column:product_id"`
	ItemUnitID  *uint            `json:"item_unit_id" gorm:"column:item_unit_id"`
	ProductJSON *json.RawMessage `json:"product_json" gorm:"column:product_json"`
	GenCode     *string          `json:"gen_code" gorm:"column:gen_code"`
	Remark      *string          `json:"remark" gorm:"column:remark"`
	Qty         *float64         `json:"qty" gorm:"column:qty"`
	Price       *float64         `json:"price" gorm:"column:price"`
	Subtotal    *float64         `json:"subtotal" gorm:"column:subtotal"`
	CreatedByID *uint            `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID *uint            `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID *uint            `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt   gorm.DeletedAt   `json:"deleted_at" gorm:"index"`
}
