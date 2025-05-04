package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

type PoDt struct {
	gorm.Model
	ID                       uint             `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ProductUuid              string           `json:"product_uuid" gorm:"column:product_uuid"`
	PoID                     *uint            `json:"po_id" gorm:"column:po_id"`
	ItemUnitID               *uint            `json:"item_unit_id" gorm:"column:item_unit_id"`
	VatID                    *uint            `json:"vat_id" gorm:"column:vat_id"`
	Pph23ID                  *uint            `json:"pph23_id" gorm:"column:pph23_id"`
	RefID                    *uint            `json:"ref_id" gorm:"column:ref_id"`
	RefSoDtID                *uint            `json:"ref_so_dt_id" gorm:"column:ref_so_dt_id"`
	RefSoDtBomID             *uint            `json:"ref_so_dt_bom_id" gorm:"column:ref_so_dt_bom_id"`
	RefProductID             *uint            `json:"ref_product_id" gorm:"column:ref_product_id"`
	ProductID                *uint            `json:"product_id" gorm:"column:product_id"`
	BomID                    *uint            `json:"bom_id" gorm:"column:bom_id"`
	ProductType              *string          `json:"product_type" gorm:"column:product_type"`
	ProductJSON              *json.RawMessage `json:"product_json" gorm:"column:product_json"`
	RefType                  *string          `json:"ref_type" gorm:"column:ref_type"`
	RefJSON                  *json.RawMessage `json:"ref_json" gorm:"column:ref_json"`
	GenCode                  *string          `json:"gen_code" gorm:"column:gen_code"`
	VatPerc                  *float64         `json:"vat_perc" gorm:"column:vat_perc"`
	VatPercAm                *float64         `json:"vat_perc_am" gorm:"column:vat_perc_am"`
	Pph23Perc                *float64         `json:"pph23_perc" gorm:"column:pph23_perc"`
	Pph23PercAm              *float64         `json:"pph23_perc_am" gorm:"column:pph23_perc_am"`
	IsVat                    *uint            `json:"is_vat" gorm:"column:is_vat"`
	IsPph23                  *uint            `json:"is_pph23" gorm:"column:is_pph23"`
	Remark                   *string          `json:"remark" gorm:"column:remark"`
	NeedQty                  *float64         `json:"need_qty" gorm:"column:need_qty"`
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
	CreatedByID              *uint            `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID              *uint            `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID              *uint            `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt                gorm.DeletedAt   `json:"deleted_at" gorm:"index"`
}

func (PoDt) TableName() string {
	return "purchase_order_dts"
}
