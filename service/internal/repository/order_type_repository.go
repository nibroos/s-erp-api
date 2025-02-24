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

type OrderTypeRepository struct {
	db     *gorm.DB
	sqlDB  *sqlx.DB
	tracer opentracing.Tracer
}

func NewOrderTypeRepository(db *gorm.DB, sqlDB *sqlx.DB, tracer opentracing.Tracer) *OrderTypeRepository {
	return &OrderTypeRepository{
		db:     db,
		sqlDB:  sqlDB,
		tracer: tracer,
	}
}

func (r *OrderTypeRepository) GetOrderTypes(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.OrderTypeListDTO, int, error) {
	// Create a child span for the controller
	childSpan := opentracing.StartSpan("OrderTypeRepository-GetOrderTypes", opentracing.ChildOf(span.Context()))

	orderTypes := []dtos.OrderTypeListDTO{}
	var total int

	// Simulate an error for testing Jaeger tracing
	if filters["simulate_error"] == "true" {
		utils.LogErrors(childSpan, fmt.Errorf("simulated error"))

		return nil, 0, fmt.Errorf("simulated error")
	}

	query := `SELECT *
    FROM ( 
        SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM mix_values m
                LEFT JOIN groups g ON m.group_id = g.id
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
                WHERE g.name = 'order_types'
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM mix_values m
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
                LEFT JOIN groups g ON m.group_id = g.id
                WHERE g.name = 'order_types'
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	var args []interface{}

	i := 1
	for key, value := range filters {
		switch key {
		case "name", "description", "remark":
			if value != "" {
				query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				countQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

	if value, ok := filters["global"]; ok && value != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d OR remark ILIKE $%d)", i, i+1, i+2)
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d OR remark ILIKE $%d)", i, i+1, i+2)
		args = append(args, "%"+value+"%", "%"+value+"%", "%"+value+"%")
		i += 3
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

		err := r.sqlDB.SelectContext(ctx.Context(), &orderTypes, query, args...)
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

	return orderTypes, total, nil
}

func (r *OrderTypeRepository) GetOrderTypeByID(ctx *fiber.Ctx, params *dtos.GetOrderTypeParams, span opentracing.Span) (*dtos.OrderTypeDetailDTO, error) {
	childSpan := opentracing.StartSpan("OrderTypeRepository-GetOrderTypeByID", opentracing.ChildOf(span.Context()))
	var orderType dtos.OrderTypeDetailDTO

	query := `SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
	cu.name as created_by_name,
	uu.name as updated_by_name

	FROM mix_values m
	LEFT JOIN users cu ON m.created_by_id = cu.id
	LEFT JOIN users uu ON m.updated_by_id = uu.id
	LEFT JOIN groups g ON m.group_id = g.id
	WHERE g.name = 'order_types'`

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

	if err := r.sqlDB.Get(&orderType, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return &orderType, nil
}

// BeginTransaction starts a new transaction
func (r *OrderTypeRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *OrderTypeRepository) CreateOrderType(tx *gorm.DB, orderType *models.MixValue, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("OrderTypeRepository-CreateOrderType", opentracing.ChildOf(span.Context()))
	if err := tx.Create(orderType).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *OrderTypeRepository) UpdateOrderType(tx *gorm.DB, orderType *models.MixValue, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("OrderTypeRepository-UpdateOrderType", opentracing.ChildOf(span.Context()))

	if err := tx.Select("*").Omit("created_at", "created_by_id").Updates(orderType).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *OrderTypeRepository) DeleteOrderType(tx *gorm.DB, params *dtos.GetOrderTypeParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("OrderTypeRepository-DeleteOrderType", opentracing.ChildOf(span.Context()))

	if err := tx.Delete(&models.MixValue{}, params.ID).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (s *OrderTypeRepository) RestoreOrderType(tx *gorm.DB, params *dtos.GetOrderTypeParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("OrderTypeRepository-RestoreOrderType", opentracing.ChildOf(span.Context()))

	var orderType models.MixValue
	if err := tx.Unscoped().Model(&orderType).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}
