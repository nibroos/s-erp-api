package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type InvoiceAdjustmentDt struct {
	gorm.Model
	ID                  uint             `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	InvoiceUUID         string           `json:"invoice_uuid" gorm:"column:invoice_uuid"`
	InvoiceAdjustmentID *uint            `json:"invoice_adjustment_id" gorm:"column:invoice_adjustment_id"`
	RefID               *uint            `json:"ref_id" gorm:"column:ref_id"`
	RefType             *string          `json:"ref_type" gorm:"column:ref_type"`
	RefJSON             *json.RawMessage `json:"ref_json" gorm:"column:ref_json"`
	InvoiceNo           *string          `json:"invoice_no" gorm:"column:invoice_no"`
	InvoiceDate         *time.Time       `json:"invoice_date" gorm:"column:invoice_date"`
	InvoiceAmount       *float64         `json:"invoice_amount" gorm:"column:invoice_amount"`
	TotalAdjustment     *float64         `json:"total_adjustment" gorm:"column:total_adjustment"`
	BalanceAmount       *float64         `json:"balance_amount" gorm:"column:balance_amount"`
	AdjustmentAmount    *float64         `json:"adjustment_amount" gorm:"column:adjustment_amount"`
	AdminBank           *float64         `json:"admin_bank" gorm:"column:admin_bank"`
	TotalAmount         *float64         `json:"total_amount" gorm:"column:total_amount"`
	CreatedByID         *uint            `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID         *uint            `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID         *uint            `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt           gorm.DeletedAt   `json:"deleted_at" gorm:"index"`
}

func (InvoiceAdjustmentDt) TableName() string {
	return "invoice_adjustment_dts"
}
