package models

import (
	"gorm.io/gorm"
)

type Schedule struct {
	gorm.Model
	ID           uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	AssigneeID   *uint          `json:"assignee_id" gorm:"column:assignee_id"`
	SalesOrderID uint           `json:"sales_order_id" gorm:"column:sales_order_id"`
	UUID         *string        `json:"uuid" gorm:"column:uuid"`
	StepsID      *uint          `json:"steps_id" gorm:"column:steps_id"`
	Title        string         `json:"title" gorm:"column:title"`
	Remark       *string        `json:"remark" gorm:"column:remark"`
	Status       string         `json:"status" gorm:"column:status"`
	StartAt      *string        `json:"start_at" gorm:"column:start_at"`
	EndAt        *string        `json:"end_at" gorm:"column:end_at"`
	Color        *string        `json:"color" gorm:"column:color"`
	CreatedByID  *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID  *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID  *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
