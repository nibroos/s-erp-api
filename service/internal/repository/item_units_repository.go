package repository

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type ItemUnitRepository struct {
	db     *gorm.DB
	sqlDB  *sqlx.DB
	tracer opentracing.Tracer
}

func NewItemUnitRepository(db *gorm.DB, sqlDB *sqlx.DB, tracer opentracing.Tracer) *ItemUnitRepository {
	return &ItemUnitRepository{
		db:     db,
		sqlDB:  sqlDB,
		tracer: tracer,
	}
}

func (r *ItemUnitRepository) GetItemUnits(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.ItemUnitListDTO, int, error) {
	// Create a child span for the controller
	childSpan := opentracing.StartSpan("ItemUnitRepository-GetItemUnits", opentracing.ChildOf(span.Context()))

	itemUnits := []dtos.ItemUnitListDTO{}
	var total int

	query := `SELECT *
    FROM ( 
        SELECT DISTINCT ON (m.id)
					m.id, m.ms_item_id, m.unit_id, m.conversion, m.price_sell, m.price_buy, m.status, m.created_at, m.updated_at, m.deleted_at,
					mi.name as ms_item_name,
					u.name as unit_name,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM item_units m
				LEFT JOIN ms_items mi ON m.ms_item_id = mi.id
				LEFT JOIN units u ON m.unit_id = u.id
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT DISTINCT ON (m.id) 
					
				mi.name as ms_item_name,
				u.name as unit_name,

        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM item_units m
				LEFT JOIN ms_items mi ON m.ms_item_id = mi.id
				LEFT JOIN units u ON m.unit_id = u.id
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	var args []interface{}

	i := 1

	filterKey := map[string]string{
		"ms_item_id": "ms_item_id",
		"unit_id":    "unit_id",
		"status":     "status",
	}

	for key, _ := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			query += fmt.Sprintf(" AND %s = $%d", value, i)
			countQuery += fmt.Sprintf(" AND %s = $%d", value, i)
			args = append(args, value)
			i++
		}
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

			err := r.sqlDB.GetContext(ctx.Context(), &total, countQuery, countArgs...)
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

		err := r.sqlDB.SelectContext(ctx.Context(), &itemUnits, query, args...)
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

	return itemUnits, total, nil
}

func (r *ItemUnitRepository) GetItemUnitByID(ctx *fiber.Ctx, params *dtos.GetItemUnitParams, span opentracing.Span) (*dtos.ItemUnitDetailDTO, error) {
	childSpan := opentracing.StartSpan("ItemUnitRepository-GetItemUnitByID", opentracing.ChildOf(span.Context()))
	var itemUnit dtos.ItemUnitDetailDTO

	query := `
	SELECT DISTINCT ON (m.id) 
		m.id, m.ms_item_id, m.unit_id, m.conversion, m.price_sell, m.price_buy, m.status, m.created_at, m.updated_at, m.deleted_at,
		mi.name as ms_item_name,
		u.name as unit_name,

		cu.name as created_by_name,
		uu.name as updated_by_name

	FROM item_units m
	LEFT JOIN ms_items mi ON m.ms_item_id = mi.id
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

	if err := r.sqlDB.Get(&itemUnit, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return &itemUnit, nil
}

// BeginTransaction starts a new transaction
func (r *ItemUnitRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *ItemUnitRepository) CreateItemUnit(tx *gorm.DB, itemUnit *models.ItemUnit, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ItemUnitRepository-CreateItemUnit", opentracing.ChildOf(span.Context()))
	if err := tx.Create(itemUnit).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *ItemUnitRepository) UpdateItemUnit(tx *gorm.DB, itemUnit *models.ItemUnit, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ItemUnitRepository-UpdateItemUnit", opentracing.ChildOf(span.Context()))

	if err := tx.Select("*").Omit("created_at", "created_by_id").Updates(itemUnit).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *ItemUnitRepository) DeleteItemUnit(tx *gorm.DB, params *dtos.GetItemUnitParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ItemUnitRepository-DeleteItemUnit", opentracing.ChildOf(span.Context()))

	if err := tx.Delete(&models.ItemUnit{}, params.ID).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (s *ItemUnitRepository) RestoreItemUnit(tx *gorm.DB, params *dtos.GetItemUnitParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ItemUnitRepository-RestoreItemUnit", opentracing.ChildOf(span.Context()))

	var itemUnit models.ItemUnit
	if err := tx.Unscoped().Model(&itemUnit).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}
