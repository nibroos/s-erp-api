package models

import (
	"gorm.io/gorm"
)

type MixValue struct {
	gorm.Model
	ID          uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	GroupID     uint           `json:"group_id" gorm:"column:group_id"`
	ParentID    *uint          `json:"parent_id" gorm:"column:parent_id"`
	Name        string         `json:"name" gorm:"column:name"`
	Description *string        `json:"description" gorm:"column:description"`
	Remark      *string        `json:"remark" gorm:"column:remark"`
	Num         float64        `json:"num" gorm:"column:num"`
	Status      int8           `json:"status" gorm:"column:status"`
	OptionsJSON string         `json:"options_json" gorm:"column:options_json"`
	CreatedByID *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
