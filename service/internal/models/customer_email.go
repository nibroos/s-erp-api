package models

import (
	"gorm.io/gorm"
)

type PicEmail struct {
	gorm.Model
	ID         uint    `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID *uint   `json:"customer_id" gorm:"column:customer_id"`
	Name       *string `json:"name" gorm:"column:name"`
	IsMain     *int8   `json:"is_main" gorm:"column:is_main"`

	CreatedByID *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (PicEmail) TableName() string {
	return "pic_emails"
}
