package models

import (
	"gorm.io/gorm"
)

type CompanyProfile struct {
	gorm.Model
	ID                 uint    `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ParentID           *uint   `json:"parent_id" gorm:"column:parent_id"`
	IsPrimary          int     `json:"is_primary" gorm:"column:is_primary"`
	CompanyOwnerName   *string `json:"company_owner_name" gorm:"column:company_owner_name"`
	CompanySignName    *string `json:"company_sign_name" gorm:"column:company_sign_name"`
	CompanyName        string  `json:"company_name" gorm:"column:company_name"`
	CompanyAddress     *string `json:"company_address" gorm:"column:company_address"`
	CompanyPhone       *string `json:"company_phone" gorm:"column:company_phone"`
	CompanyEmail       *string `json:"company_email" gorm:"column:company_email"`
	CompanyWebsite     *string `json:"company_website" gorm:"column:company_website"`
	CompanyLogo        *string `json:"company_logo" gorm:"column:company_logo"`
	CompanySign        *string `json:"company_sign" gorm:"column:company_sign"`
	CompanyDescription *string `json:"company_description" gorm:"column:company_description"`
	CompanyRemark      *string `json:"company_remark" gorm:"column:company_remark"`
	CompanyStatus      int     `json:"company_status" gorm:"column:company_status"`
	CompanyOptionsJSON string  `json:"company_options_json" gorm:"column:company_options_json"`
	CreatedByID        uint    `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID        uint    `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID        uint    `json:"deleted_by_id" gorm:"column:deleted_by_id"`
}
