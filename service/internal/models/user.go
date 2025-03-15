package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID              uint    `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	BranchID        *uint   `json:"branch_id" gorm:"column:branch_id"`
	Name            string  `json:"name" gorm:"column:name"`
	Username        *string `json:"username" gorm:"column:username;unique"`
	Email           string  `json:"email" gorm:"column:email;unique"`
	Password        string  `json:"-" gorm:"column:password"`
	Status          *int    `json:"status" gorm:"column:status"`
	Address         *string `json:"address" gorm:"column:address"`
	PhoneNumber     *string `json:"phone_number" gorm:"column:phone_number"`
	ProfileImageURL *string `json:"profile_image_url" gorm:"column:profile_image_url"`
	Roles           []Role  `json:"roles,omitempty" gorm:"many2many:user_roles"`
	CreatedByID     uint    `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID     uint    `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID     uint    `json:"deleted_by_id" gorm:"column:deleted_by_id"`
}
