package models

import (
	"gorm.io/gorm"
)

type Customer struct {
	gorm.Model
	ID             uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CustomerTypeID *uint          `json:"customer_type_id" gorm:"column:customer_type_id"`
	AgentID        *uint          `json:"agent_id" gorm:"column:agent_id"`
	Code           *string        `json:"code" gorm:"column:code"`
	Name           string         `json:"name" gorm:"column:name"`
	Address        *string        `json:"address" gorm:"column:address"`
	Phone          *string        `json:"phone" gorm:"column:phone"`
	Email          *string        `json:"email" gorm:"column:email"`
	Pic            *string        `json:"pic" gorm:"column:pic"`
	Status         int8           `json:"status" gorm:"column:status"`
	CreatedByID    *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID    *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID    *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
