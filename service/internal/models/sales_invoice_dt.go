package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

type SalesInvoiceDt struct {
	gorm.Model
	ID             uint             `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ProductUuid    string           `json:"product_uuid" gorm:"column:product_uuid"`
	SalesInvoiceID *uint            `json:"sales_invoice_id" gorm:"column:sales_invoice_id"`
	ItemUnitID     *uint            `json:"item_unit_id" gorm:"column:item_unit_id"`
	VatID          *uint            `json:"vat_id" gorm:"column:vat_id"`
	Pph23ID        *uint            `json:"pph23_id" gorm:"column:pph23_id"`
	RefID          *uint            `json:"ref_id" gorm:"column:ref_id"`
	RefDtID        *uint            `json:"ref_dt_id" gorm:"column:ref_dt_id"`
	ProductID      *uint            `json:"product_id" gorm:"column:product_id"`
	RefType        *string          `json:"ref_type" gorm:"column:ref_type"`
	RefJSON        *json.RawMessage `json:"ref_json" gorm:"column:ref_json"`
	ProductType    *string          `json:"product_type" gorm:"column:product_type"`
	ProductJSON    *json.RawMessage `json:"product_json" gorm:"column:product_json"`
	Remark         *string          `json:"remark" gorm:"column:remark"`
	IsVat          *uint            `json:"is_vat" gorm:"column:is_vat"`
	IsPph23        *uint            `json:"is_pph23" gorm:"column:is_pph23"`
	Qty            *float64         `json:"qty" gorm:"column:qty"`
	Price          *float64         `json:"price" gorm:"column:price"`
	Subtotal       *float64         `json:"subtotal" gorm:"column:subtotal"`
	Discount       *float64         `json:"discount" gorm:"column:discount"`
	TotalAmount    *float64         `json:"total_amount" gorm:"column:total_amount"`
	TotalDp        *float64         `json:"total_dp" gorm:"column:total_dp"`
	TotalBalance   *float64         `json:"total_balance" gorm:"column:total_balance"`
	CreatedByID    *uint            `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID    *uint            `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID    *uint            `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt      gorm.DeletedAt   `json:"deleted_at" gorm:"index"`
}

func (SalesInvoiceDt) TableName() string {
	return "sales_invoice_dts"
}
