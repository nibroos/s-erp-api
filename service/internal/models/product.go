package models

import (
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	ID             uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ItemSubGroupID uint           `json:"item_sub_group_id" gorm:"column:item_sub_group_id"`
	ItemUnitID     uint           `json:"item_unit_id" gorm:"column:item_unit_id"`
	Code           *string        `json:"code" gorm:"column:code"`
	FactoryCode    *string        `json:"factory_code" gorm:"column:factory_code"`
	Name           string         `json:"name" gorm:"column:name"`
	Sku            *string        `json:"sku" gorm:"column:sku"`
	Barcode        *string        `json:"barcode" gorm:"column:barcode"`
	Specification  *string        `json:"specification" gorm:"column:specification"`
	Description    *string        `json:"description" gorm:"column:description"`
	TpbCode        *string        `json:"tpb_code" gorm:"column:tpb_code"`
	MinimumStock   *float64       `json:"minimum_stock" gorm:"column:minimum_stock"`
	IsAllBranch    *int           `json:"is_all_branch" gorm:"column:is_all_branch"`
	Remark         *string        `json:"remark" gorm:"column:remark"`
	Status         int8           `json:"status" gorm:"column:status"`
	ExpiredAt      *string        `json:"expired_at" gorm:"column:expired_at"`
	CreatedByID    *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID    *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID    *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
