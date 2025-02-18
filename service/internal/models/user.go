package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID          uint    `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name        string  `json:"name" gorm:"column:name"`
	Username    *string `json:"username" gorm:"column:username;unique"`
	Email       string  `json:"email" gorm:"column:email;unique"`
	Password    string  `json:"-" gorm:"column:password"`
	Address     *string `json:"address" gorm:"column:address"`
	Roles       []Role  `json:"roles,omitempty" gorm:"many2many:user_roles"`
	CreatedByID uint    `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID uint    `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID uint    `json:"deleted_by_id" gorm:"column:deleted_by_id"`
}
