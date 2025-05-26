package models

import (
	"gorm.io/gorm"
)

type Customer struct {
	gorm.Model
	ID             uint    `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CustomerTypeID *uint   `json:"customer_type_id" gorm:"column:customer_type_id"`
	AgentID        *uint   `json:"agent_id" gorm:"column:agent_id"`
	CurrencyID     *uint   `json:"currency_id" gorm:"column:currency_id"`
	Shortname      *string `json:"shortname" gorm:"column:shortname"`
	Code           *string `json:"code" gorm:"column:code"`
	Name           string  `json:"name" gorm:"column:name"`
	Address        *string `json:"address" gorm:"column:address"`
	Phone          *string `json:"phone" gorm:"column:phone"`
	Email          *string `json:"email" gorm:"column:email"`
	Pic            *string `json:"pic" gorm:"column:pic"`
	Status         int8    `json:"status" gorm:"column:status"`
	IsCrm          int8    `json:"is_crm" gorm:"column:is_crm"`

	Remark         *string `json:"remark" gorm:"column:remark"`
	OwnerName      *string `json:"owner_name" gorm:"column:owner_name"`
	OwnerPhone     *string `json:"owner_phone" gorm:"column:owner_phone"`
	OwnerEmail     *string `json:"owner_email" gorm:"column:owner_email"`
	CategoryTypeID *uint   `json:"category_type_id" gorm:"column:category_type_id"`
	ContractDate   *string `json:"contract_date" gorm:"column:contract_date"`
	IsContract     *int8   `json:"is_contract" gorm:"column:is_contract"`
	PicName        *string `json:"pic_name" gorm:"column:pic_name"`
	PicPhone       *string `json:"pic_phone" gorm:"column:pic_phone"`

	CreatedByID *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
