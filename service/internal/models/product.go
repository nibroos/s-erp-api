package models

import (
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	ID            uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	UnitID        *uint          `json:"unit_id" gorm:"column:unit_id"`
	BranchID      *uint          `json:"branch_id" gorm:"column:branch_id"`
	Code          *string        `json:"code" gorm:"column:code"`
	FactoryCode   *string        `json:"factory_code" gorm:"column:factory_code"`
	Name          string         `json:"name" gorm:"column:name"`
	Sku           *string        `json:"sku" gorm:"column:sku"`
	Barcode       *string        `json:"barcode" gorm:"column:barcode"`
	Specification *string        `json:"specification" gorm:"column:specification"`
	Description   *string        `json:"description" gorm:"column:description"`
	Remark        *string        `json:"remark" gorm:"column:remark"`
	PriceSell     *float64       `json:"price_sell" gorm:"column:price_sell"`
	PriceBuy      *float64       `json:"price_buy" gorm:"column:price_buy"`
	Margin        *float64       `json:"margin" gorm:"column:margin"`
	Status        int8           `json:"status" gorm:"column:status"`
	ExpiredAt     *string        `json:"expired_at" gorm:"column:expired_at"`
	CreatedByID   *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID   *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID   *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
