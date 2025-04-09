package models

import (
	"gorm.io/gorm"
)

type ScheduleTask struct {
	ID          uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ScheduleID  uint           `json:"schedule_id" gorm:"column:schedule_id"`
	AssigneeID  *uint          `json:"assignee_id" gorm:"column:assignee_id"`
	ParentID    *uint          `json:"parent_id" gorm:"column:parent_id"`
	EntityID    *uint          `json:"entity_id" gorm:"column:entity_id"`
	EntityType  string         `json:"entity_type" gorm:"column:entity_type"`
	UUID        *string        `json:"uuid" gorm:"column:uuid"`
	ParentUUID  *string        `json:"parent_uuid" gorm:"column:parent_uuid"`
	Title       string         `json:"title" gorm:"column:title"`
	Remark      *string        `json:"remark" gorm:"column:remark"`
	OrderItem   *int           `json:"order_item" gorm:"column:order_item"`
	Color       *string        `json:"color" gorm:"column:color"`
	IsChecked   *int           `json:"is_checked" gorm:"column:is_checked"`
	Locations   *string        `json:"locations" gorm:"column:locations"`
	StartAt     *string        `json:"start_at" gorm:"column:start_at"`
	EndAt       *string        `json:"end_at" gorm:"column:end_at"`
	OptionsJSON *string        `json:"options_json" gorm:"column:options_json"`
	CreatedByID *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
