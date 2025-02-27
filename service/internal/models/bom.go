package models

import (
	"gorm.io/gorm"
)

type Bom struct {
	gorm.Model
	ID          *uint          `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ProductID   *uint          `json:"product_id" gorm:"column:product_id"`
	MsItemID    *uint          `json:"ms_item_id" gorm:"column:ms_item_id"`
	ItemUnitID  *uint          `json:"item_unit_id" gorm:"column:item_unit_id"`
	Qty         *float64       `json:"qty" gorm:"column:qty"`
	Remark      *string        `json:"remark" gorm:"column:remark"`
	CreatedByID *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
