package models

import (
	"gorm.io/gorm"
)

type VatHistory struct {
	gorm.Model
	ID          uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	VatID       *uint          `json:"vat_id" gorm:"column:vat_id"`
	Num         *float64       `json:"num" gorm:"column:num"`
	Divider     *float64       `json:"divider" gorm:"column:divider"`
	Multiplier  *float64       `json:"multiplier" gorm:"column:multiplier"`
	ChangedAt   string         `json:"changed_at" gorm:"column:changed_at"`
	Status      *int8          `json:"status" gorm:"column:status"`
	Remark      *string        `json:"remark" gorm:"column:remark"`
	CreatedByID *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
