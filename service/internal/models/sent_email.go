package models

import (
	"gorm.io/gorm"
)

type SentEmail struct {
	gorm.Model
	ID           uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	RefID        *uint          `json:"ref_id" gorm:"column:ref_id"`
	SenderID     *uint          `json:"sender_id" gorm:"column:sender_id"`
	RefType      string         `json:"ref_type" gorm:"column:ref_type"`
	FromEmail    string         `json:"from_email" gorm:"column:from_email"`
	ToEmail      string         `json:"to_email" gorm:"column:to_email"`
	Subject      string         `json:"subject" gorm:"column:subject"`
	Remark       *string        `json:"remark" gorm:"column:remark"`
	ErrorMessage *string        `json:"error_message" gorm:"column:error_message"`
	LogJson      *string        `json:"log_json" gorm:"column:log_json"`
	Status       string         `json:"status" gorm:"column:status;default:'PROCESS'"`
	CreatedByID  *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID  *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID  *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (SentEmail) TableName() string {
	return "sent_emails"
}
