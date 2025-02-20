package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type MsItemRepository struct {
	db     *gorm.DB
	sqlDB  *sqlx.DB
	tracer opentracing.Tracer
}

func NewMsItemRepository(db *gorm.DB, sqlDB *sqlx.DB, tracer opentracing.Tracer) *MsItemRepository {
	return &MsItemRepository{
		db:     db,
		sqlDB:  sqlDB,
		tracer: tracer,
	}
}

func (r *MsItemRepository) GetMsItems(ctx context.Context, filters map[string]string, span opentracing.Span) ([]dtos.MsItemListDTO, int, error) {
	// Create a child span for the controller
	childSpan := opentracing.StartSpan("MsItemRepository-GetMsItems", opentracing.ChildOf(span.Context()))

	msItems := []dtos.MsItemListDTO{}
	var total int

	query := `SELECT *
    FROM ( 
        SELECT DISTINCT ON (m.id)
					m.id, m.item_sub_group_id, ig.item_group_id, m.unit_id, m.name, m.specification, m.description, m.tpb_code, iu.price_sell, iu.price_buy, m.minimum_stock, m.is_all_branch, m.status, m.created_at, m.updated_at, m.deleted_at,
				isg.name as item_sub_group_name,
				ig.name as item_group_name,
				u.name as unit_name,
				iu.unit_id as item_unit_unit_id,
				bi.name as branch_item_name, bi.branch_id as branch_id, bi.specification as branch_item_specification, bi.description as branch_item_description, bi.tpb_code as branch_item_tpb_code, bi.price_sell as branch_item_price_sell, bi.price_buy as branch_item_price_buy, bi.minimum_stock as branch_item_minimum_stock, bi.status as branch_item_status, bi.created_at as branch_item_created_at, bi.updated_at as branch_item_updated_at, bi.deleted_at as branch_item_deleted_at,

        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM ms_items m
				LEFT JOIN item_sub_groups isg ON m.item_sub_group_id = isg.id
				LEFT JOIN item_groups ig ON isg.item_group_id = ig.id
				LEFT JOIN item_units iu ON iu.id = m.item_unit_id
				-- LEFT JOIN units u ON iu.unit_id = u.id
				LEFT JOIN mix_values u ON iu.unit_id = u.id
				LEFT JOIN branch_items bi ON bi.ms_item_id = m.id
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT DISTINCT ON (m.id) 
					m.id, m.item_sub_group_id, ig.item_group_id, m.unit_id, m.name, m.specification, m.description, m.tpb_code, iu.price_sell, iu.price_buy, m.minimum_stock, m.is_all_branch, m.status, m.created_at, m.updated_at, m.deleted_at,
				isg.name as item_sub_group_name,
				ig.name as item_group_name,
				u.name as unit_name,
				iu.unit_id as item_unit_unit_id,
				bi.name as branch_item_name, bi.branch_id as branch_id, bi.specification as branch_item_specification, bi.description as branch_item_description, bi.tpb_code as branch_item_tpb_code, bi.price_sell as branch_item_price_sell, bi.price_buy as branch_item_price_buy, bi.minimum_stock as branch_item_minimum_stock, bi.status as branch_item_status, bi.created_at as branch_item_created_at, bi.updated_at as branch_item_updated_at, bi.deleted_at as branch_item_deleted_at,

        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM ms_items m
				LEFT JOIN item_sub_groups isg ON m.item_sub_group_id = isg.id
				LEFT JOIN item_groups ig ON isg.item_group_id = ig.id
				LEFT JOIN item_units iu ON iu.id = m.item_unit_id
				LEFT JOIN units u ON iu.unit_id = u.id
				LEFT JOIN branch_items bi ON bi.ms_item_id = m.id
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	var args []interface{}

	i := 1
	for key, value := range filters {
		switch key {
		case "name", "specification", "tpb_code", "description":
			if value != "" {
				query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				countQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

	filterKey := map[string]string{
		"item_sub_group_id": "item_sub_group_id",
		"item_group_id":     "item_group_id",
		"branch_id":         "branch_id",
		"item_unit_unit_id": "item_unit_unit_id",
		"status":            "status",
	}

	for key, _ := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			query += fmt.Sprintf(" AND %s = $%d", value, i)
			countQuery += fmt.Sprintf(" AND %s = $%d", value, i)
			args = append(args, value)
			i++
		}
	}

	if value, ok := filters["global"]; ok && value != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR specification ILIKE $%d OR tpb_code ILIKE $%d OR description ILIKE $%d OR branch_item_name ILIKE $%d OR branch_item_specification ILIKE $%d OR branch_item_tpb_code ILIKE $%d OR branch_item_description ILIKE $%d)", i, i+1, i+2, i+3, i+4, i+5, i+6, i+7)
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR specification ILIKE $%d OR tpb_code ILIKE $%d OR description ILIKE $%d OR branch_item_name ILIKE $%d OR branch_item_specification ILIKE $%d OR branch_item_tpb_code ILIKE $%d OR branch_item_description ILIKE $%d)", i, i+1, i+2, i+3, i+4, i+5, i+6, i+7)
		args = append(args, "%"+value+"%", "%"+value+"%", "%"+value+"%", "%"+value+"%", "%"+value+"%", "%"+value+"%", "%"+value+"%", "%"+value+"%")
		i += 8
	}

	countArgs := append([]interface{}{}, args...)

	var wg sync.WaitGroup
	var countErr, selectErr error

	// Goroutine for count query
	wg.Add(1)
	go func() {
		defer wg.Done()
		if filters["is_csv"] != "1" {
			// Create a span for the count query
			countSpan := opentracing.StartSpan("CountQuery", opentracing.ChildOf(childSpan.Context()))

			err := r.sqlDB.GetContext(ctx, &total, countQuery, countArgs...)
			if err != nil {
				utils.LogErrors(countSpan, err)
				countSpan.LogKV("query", countQuery)
				countErr = err
			}
		}
	}()

	if countErr != nil {
		return nil, 0, countErr
	}

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "name")
	orderDirection := utils.GetStringOrDefault(filters["order_direction"], "asc")
	query += fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)

	perPage := utils.GetIntOrDefault(filters["per_page"], 10)
	currentPage := utils.GetIntOrDefault(filters["page"], 1)

	// if is_csv
	if filters["is_csv"] != "1" {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
		args = append(args, perPage, (currentPage-1)*perPage)
	}

	// Goroutine for select query
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Create a span for the select query
		selectSpan := opentracing.StartSpan("SelectQuery", opentracing.ChildOf(childSpan.Context()))

		err := r.sqlDB.SelectContext(ctx, &msItems, query, args...)
		if err != nil {
			selectSpan.LogKV("query", query)
			utils.LogErrors(selectSpan, err)
			selectErr = err
		}
	}()

	// Wait for both goroutines to finish
	wg.Wait()

	if countErr != nil || selectErr != nil {
		defer childSpan.Finish()
	}

	if countErr != nil {
		return nil, 0, countErr
	}

	if selectErr != nil {
		return nil, 0, selectErr
	}

	return msItems, total, nil
}

func (r *MsItemRepository) GetMsItemByID(ctx context.Context, params *dtos.GetMsItemParams, span opentracing.Span) (*dtos.MsItemDetailDTO, error) {
	childSpan := opentracing.StartSpan("MsItemRepository-GetMsItemByID", opentracing.ChildOf(span.Context()))
	var msItem dtos.MsItemDetailDTO

	query := `
	SELECT DISTINCT ON (m.id) 
		m.id, m.item_sub_group_id, ig.item_group_id, m.unit_id, m.name, m.specification, m.tpb_code, m.price_sell, m.price_buy, m.minimum_stock, m.is_all_branch, m.status, m.created_at, m.updated_at, m.deleted_at,
		isg.name as item_sub_group_name,
		ig.name as item_group_name,
		u.name as unit_name,

		cu.name as created_by_name,
		uu.name as updated_by_name

	FROM ms_items m
	LEFT JOIN item_sub_groups isg ON m.item_sub_group_id = isg.id
	LEFT JOIN item_groups ig ON isg.item_group_id = ig.id
	LEFT JOIN units u ON m.unit_id = u.id
	LEFT JOIN users cu ON m.created_by_id = cu.id
	LEFT JOIN users uu ON m.updated_by_id = uu.id
	WHERE 1=1`

	var args []interface{}

	i := 1
	query += " AND m.id = $1"
	args = append(args, params.ID)
	i++

	isDeletedQuery := ` AND m.deleted_at IS NULL`
	if params.IsDeleted != nil && *params.IsDeleted == 1 {
		isDeletedQuery = " AND m.deleted_at IS NOT NULL"
	}

	query += isDeletedQuery

	if err := r.sqlDB.Get(&msItem, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return &msItem, nil
}

// BeginTransaction starts a new transaction
func (r *MsItemRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *MsItemRepository) CreateMsItem(tx *gorm.DB, msItem *models.MsItem, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("MsItemRepository-CreateMsItem", opentracing.ChildOf(span.Context()))
	if err := tx.Create(msItem).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *MsItemRepository) UpdateMsItem(tx *gorm.DB, msItem *models.MsItem, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("MsItemRepository-UpdateMsItem", opentracing.ChildOf(span.Context()))
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("*").Omit("created_at", "created_by_id").Updates(msItem).Error; err != nil {
			defer childSpan.Finish()
			utils.LogErrors(childSpan, err)
			return err
		}
		return nil
	})

}

func (r *MsItemRepository) DeleteMsItem(tx *gorm.DB, params *dtos.GetMsItemParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("MsItemRepository-DeleteMsItem", opentracing.ChildOf(span.Context()))
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.MsItem{}, params.ID).Error; err != nil {
			defer childSpan.Finish()
			utils.LogErrors(childSpan, err)
			return err
		}
		return nil
	})
}

func (s *MsItemRepository) RestoreMsItem(tx *gorm.DB, params *dtos.GetMsItemParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("MsItemRepository-RestoreMsItem", opentracing.ChildOf(span.Context()))
	return s.db.Transaction(func(tx *gorm.DB) error {
		var msItem models.MsItem
		if err := tx.Unscoped().Model(&msItem).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
			defer childSpan.Finish()
			utils.LogErrors(childSpan, err)
			return err
		}
		return nil
	})
}
