package models

import (
	"gorm.io/gorm"
)

type Branch struct {
	gorm.Model
	ID               uint    `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ParentID         *uint   `json:"parent_id" gorm:"column:parent_id"`
	CompanyProfileID *uint   `json:"company_profile_id" gorm:"column:company_profile_id"`
	OwnerName        *string `json:"owner_name" gorm:"column:owner_name"`
	SignName         *string `json:"sign_name" gorm:"column:sign_name"`
	Name             string  `json:"name" gorm:"column:name"`
	Address          *string `json:"address" gorm:"column:address"`
	Phone            *string `json:"phone" gorm:"column:phone"`
	Email            *string `json:"email" gorm:"column:email"`
	Website          *string `json:"website" gorm:"column:website"`
	Logo             *string `json:"logo" gorm:"column:logo"`
	Sign             *string `json:"sign" gorm:"column:sign"`
	Description      *string `json:"description" gorm:"column:description"`
	Remark           *string `json:"remark" gorm:"column:remark"`
	Status           *int    `json:"status" gorm:"column:status"`
	OptionsJSON      string  `json:"options_json" gorm:"column:options_json"`
	CreatedByID      uint    `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID      uint    `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID      uint    `json:"deleted_by_id" gorm:"column:deleted_by_id"`
}
