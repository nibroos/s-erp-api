package models

import (
	"gorm.io/gorm"
)

type Letter struct {
	gorm.Model
	ID          uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	RefID       *uint          `json:"ref_id" gorm:"column:ref_id"`
	RefType     string         `json:"ref_type" gorm:"column:ref_type"`
	FileType    string         `json:"file_type" gorm:"column:file_type"`
	FileUrl     string         `json:"file_url" gorm:"column:file_url"`
	FileName    string         `json:"file_name" gorm:"column:file_name"`
	Remark      *string        `json:"remark" gorm:"column:remark"`
	FileProp    string         `json:"file_prop" gorm:"column:file_prop"`
	CreatedByID *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
