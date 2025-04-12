package models

import (
	"time"
)

type BankInformation struct {
	ID                uint       `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CommpanyProfileID *uint      `json:"commpany_profile_id" gorm:"column:commpany_profile_id"`
	Name              *string    `json:"name" gorm:"column:name"`
	AccountNumber     *string    `json:"account_number" gorm:"column:account_number"`
	AccountName       *string    `json:"account_name" gorm:"column:account_name"`
	Description       *string    `json:"description" gorm:"column:description"`
	CreatedByID       *uint      `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID       *uint      `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID       *uint      `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	CreatedAt         *time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         *time.Time `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt         *time.Time `json:"deleted_at" gorm:"column:deleted_at"`
}

func (BankInformation) TableName() string {
	return "bank_informations"
}
