package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jLog "github.com/opentracing/opentracing-go/log"
	"gorm.io/gorm"
)

type CustomerTypeRepository struct {
	db     *gorm.DB
	sqlDB  *sqlx.DB
	tracer opentracing.Span
}

func NewCustomerTypeRepository(db *gorm.DB, sqlDB *sqlx.DB, tracer opentracing.Span) *CustomerTypeRepository {
	return &CustomerTypeRepository{
		db:     db,
		sqlDB:  sqlDB,
		tracer: tracer,
	}
}

func (r *CustomerTypeRepository) GetCustomerTypes(ctx context.Context, filters map[string]string, span opentracing.Span) ([]dtos.CustomerTypeListDTO, int, error) {
	// Create a child span for the controller
	childSpan := opentracing.StartSpan("CustomerTypeRepository-GetCustomerTypes", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	customerTypes := []dtos.CustomerTypeListDTO{}
	var total int

	// Simulate an error for testing Jaeger tracing
	if filters["simulate_error"] == "true" {
		childSpan.LogFields(
			jLog.String("event", "error"),
			jLog.String("message", "simulated error"),
		)
		ext.Error.Set(childSpan, true)

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
				WHERE g.name = 'customer_types'
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM mix_values m
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
				LEFT JOIN groups g ON m.group_id = g.id
				WHERE g.name = 'customer_types'
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

	// Channels for concurrent execution
	countChan := make(chan error)
	selectChan := make(chan error)

	// Goroutine for count query
	go func() {
		if filters["is_csv"] != "1" {
			err := r.sqlDB.GetContext(ctx, &total, countQuery, countArgs...)
			countChan <- err
		} else {
			countChan <- nil
		}
	}()

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
	go func() {
		err := r.sqlDB.SelectContext(ctx, &customerTypes, query, args...)
		selectChan <- err
	}()

	// Wait for both goroutines to finish
	countErr := <-countChan
	selectErr := <-selectChan

	if countErr != nil {
		return nil, 0, countErr
	}

	if selectErr != nil {
		return nil, 0, selectErr
	}

	return customerTypes, total, nil
}

func (r *CustomerTypeRepository) GetCustomerTypeByID(ctx context.Context, params *dtos.GetCustomerTypeParams) (*dtos.CustomerTypeDetailDTO, error) {
	var customerType dtos.CustomerTypeDetailDTO

	query := `SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
	cu.name as created_by_name,
	uu.name as updated_by_name

	FROM mix_values m
	LEFT JOIN users cu ON m.created_by_id = cu.id
	LEFT JOIN users uu ON m.updated_by_id = uu.id
	LEFT JOIN groups g ON m.group_id = g.id
	WHERE g.name = 'customer_types'`

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

	if err := r.sqlDB.Get(&customerType, query, args...); err != nil {
		return nil, err
	}

	return &customerType, nil
}

// BeginTransaction starts a new transaction
func (r *CustomerTypeRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *CustomerTypeRepository) CreateCustomerType(tx *gorm.DB, customerType *models.MixValue) error {
	if err := tx.Create(customerType).Error; err != nil {
		return err
	}
	return nil
}

func (r *CustomerTypeRepository) UpdateCustomerType(tx *gorm.DB, customerType *models.MixValue) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Updates(customerType).Error; err != nil {
			return err
		}
		return nil
	})

}

func (r *CustomerTypeRepository) DeleteCustomerType(tx *gorm.DB, params *dtos.GetCustomerTypeParams) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.MixValue{}, params.ID).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *CustomerTypeRepository) RestoreCustomerType(tx *gorm.DB, params *dtos.GetCustomerTypeParams) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var customerType models.MixValue
		if err := tx.Unscoped().Model(&customerType).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		return nil
	})
}
