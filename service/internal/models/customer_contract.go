package models

import (
	"gorm.io/gorm"
)

type CustomerContract struct {
	gorm.Model
	ID             uint     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID     *uint    `json:"customer_id" gorm:"column:customer_id"`
	ProductID      *uint    `json:"product_id" gorm:"column:product_id"`
	PaymentTypeID  *uint    `json:"payment_type_id" gorm:"column:payment_type_id"`
	AgreeAt        *string  `json:"agree_at" gorm:"column:agree_at"`
	DueAt          *string  `json:"due_at" gorm:"column:due_at"`
	Price          *float64 `json:"price" gorm:"column:price"`
	Qty            *float64 `json:"qty" gorm:"column:qty"`
	InstallationAt *string  `json:"installation_at" gorm:"column:installation_at"`
	WarrantyAt     *string  `json:"warranty_at" gorm:"column:warranty_at"`
	Remark         *string  `json:"remark" gorm:"column:remark"`

	CreatedByID *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (CustomerContract) TableName() string {
	return "customer_contracts"
}
