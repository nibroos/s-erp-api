package models

import (
	"gorm.io/gorm"
)

type Ticket struct {
	gorm.Model
	ID            *uint          `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID    *uint          `json:"customer_id" gorm:"column:customer_id"`
	ProductID     *uint          `json:"product_id" gorm:"column:product_id"`
	BranchID      *uint          `json:"branch_id" gorm:"column:branch_id"`
	PriorityType  *string        `json:"priority_type" gorm:"column:priority_type"`
	RevNo         *int           `json:"rev_no" gorm:"column:rev_no"`
	TicketNo      *string        `json:"ticket_no" gorm:"column:ticket_no"`
	TicketNoOri   *string        `json:"ticket_no_ori" gorm:"column:ticket_no_ori"`
	Title         string         `json:"title" gorm:"column:title"`
	IssueDesc     *string        `json:"issue_desc" gorm:"column:issue_desc"`
	IssueSolution *string        `json:"issue_solution" gorm:"column:issue_solution"`
	Remark        *string        `json:"remark" gorm:"column:remark"`
	Status        string         `json:"status" gorm:"column:status"`
	ReportedAt    *string        `json:"reported_at" gorm:"column:reported_at"`
	CreatedByID   *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID   *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID   *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
