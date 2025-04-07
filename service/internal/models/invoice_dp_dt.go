package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

type InvoiceDpDt struct {
	gorm.Model
	ID                       uint             `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ProductUuid              string           `json:"product_uuid" gorm:"column:product_uuid"`
	InvoiceDpID              *uint            `json:"invoice_dp_id" gorm:"column:invoice_dp_id"`
	ItemUnitID               *uint            `json:"item_unit_id" gorm:"column:item_unit_id"`
	VatID                    *uint            `json:"vat_id" gorm:"column:vat_id"`
	Pph23ID                  *uint            `json:"pph23_id" gorm:"column:pph23_id"`
	RefID                    *uint            `json:"ref_id" gorm:"column:ref_id"`
	RefDtID                  *uint            `json:"ref_dt_id" gorm:"column:ref_dt_id"`
	ProductID                *uint            `json:"product_id" gorm:"column:product_id"`
	RefType                  *string          `json:"ref_type" gorm:"column:ref_type"`
	RefJSON                  *json.RawMessage `json:"ref_json" gorm:"column:ref_json"`
	ProductType              *string          `json:"product_type" gorm:"column:product_type"`
	ProductJSON              *json.RawMessage `json:"product_json" gorm:"column:product_json"`
	Remark                   *string          `json:"remark" gorm:"column:remark"`
	DpPercentage             *float64         `json:"dp_percentage" gorm:"column:dp_percentage"`
	IsVat                    *uint            `json:"is_vat" gorm:"column:is_vat"`
	IsPph23                  *uint            `json:"is_pph23" gorm:"column:is_pph23"`
	Qty                      *float64         `json:"qty" gorm:"column:qty"`
	Price                    *float64         `json:"price" gorm:"column:price"`
	Subtotal                 *float64         `json:"subtotal" gorm:"column:subtotal"`
	DiscountAmount           *float64         `json:"discount_amount" gorm:"column:discount_amount"`
	DiscountPercentage       *float64         `json:"discount_percentage" gorm:"column:discount_percentage"`
	DiscountPercentageNum    *float64         `json:"discount_percentage_num" gorm:"column:discount_percentage_num"`
	DiscountPercentageAmount *float64         `json:"discount_percentage_amount" gorm:"column:discount_percentage_amount"`
	DiscountFinal            *float64         `json:"discount_final" gorm:"column:discount_final"`
	DiscountType             *string          `json:"discount_type" gorm:"column:discount_type"`
	TotalAmount              *float64         `json:"total_amount" gorm:"column:total_amount"`
	TotalDp                  *float64         `json:"total_dp" gorm:"column:total_dp"`
	CreatedByID              *uint            `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID              *uint            `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID              *uint            `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt                gorm.DeletedAt   `json:"deleted_at" gorm:"index"`
}

func (InvoiceDpDt) TableName() string {
	return "invoice_dp_dts"
}
