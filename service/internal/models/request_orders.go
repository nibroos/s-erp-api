package models

import (
	"time"

	"gorm.io/gorm"
)

type RequestOrder struct {
	gorm.Model
	ID                        uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	BranchID                  *uint          `json:"branch_id" gorm:"column:branch_id"`
	WarehouseID               *uint          `json:"warehouse_id" gorm:"column:warehouse_id"`
	RequestNo                 *string        `json:"request_no" gorm:"column:request_no"`
	RequestDate               *time.Time     `json:"request_date" gorm:"column:request_date"`
	Remark                    *string        `json:"remark" gorm:"column:remark"`
	Requested                 *string        `json:"requested" gorm:"column:requested"`
	RevNo                     *int           `json:"rev_no" gorm:"column:rev_no"`
	Status                    *string        `json:"status" gorm:"column:status"`
	GrandTotalOrderProductQty *float64       `json:"grand_total_order_product_qty" gorm:"column:grand_total_order_product_qty"`
	GrandTotalOrderItemQty    *float64       `json:"grand_total_order_item_qty" gorm:"column:grand_total_order_item_qty"`
	GrandTotalWhQty           *float64       `json:"grand_total_wh_qty" gorm:"column:grand_total_wh_qty"`
	GrandTotalReqQty          *float64       `json:"grand_total_req_qty" gorm:"column:grand_total_req_qty"`
	CreatedByID               *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID               *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID               *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt                 gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (RequestOrder) TableName() string {
	return "request_orders"
}
