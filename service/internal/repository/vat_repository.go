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

type VatRepository struct {
	db     *gorm.DB
	sqlDB  *sqlx.DB
	tracer opentracing.Tracer
}

func NewVatRepository(db *gorm.DB, sqlDB *sqlx.DB, tracer opentracing.Tracer) *VatRepository {
	return &VatRepository{
		db:     db,
		sqlDB:  sqlDB,
		tracer: tracer,
	}
}

func (r *VatRepository) GetVats(ctx context.Context, filters map[string]string, span opentracing.Span) ([]dtos.VatListDTO, int, error) {
	// Create a child span for the controller
	childSpan := opentracing.StartSpan("VatRepository-GetVats", opentracing.ChildOf(span.Context()))

	vats := []dtos.VatListDTO{}
	var total int

	query := `SELECT *
    FROM ( 
        SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM mix_values m
                LEFT JOIN groups g ON m.group_id = g.id
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
                WHERE g.name = 'vats'
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM mix_values m
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
                LEFT JOIN groups g ON m.group_id = g.id
                WHERE g.name = 'vats'
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

			err := r.sqlDB.GetContext(ctx, &total, countQuery, countArgs...)
			if err != nil {
				utils.LogErrors(countSpan, err)
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

		err := r.sqlDB.SelectContext(ctx, &vats, query, args...)
		if err != nil {
			utils.LogErrors(selectSpan, err)
			selectErr = err
		}
	}()

	// Wait for both goroutines to finish
	wg.Wait()

	if countErr != nil {
		return nil, 0, countErr
	}

	if selectErr != nil {
		return nil, 0, selectErr
	}

	return vats, total, nil
}

func (r *VatRepository) GetVatByID(ctx context.Context, params *dtos.GetVatParams, span opentracing.Span) (*dtos.VatDetailDTO, error) {
	childSpan := opentracing.StartSpan("VatRepository-GetVatByID", opentracing.ChildOf(span.Context()))
	var vat dtos.VatDetailDTO

	query := `SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
	cu.name as created_by_name,
	uu.name as updated_by_name

	FROM mix_values m
	LEFT JOIN users cu ON m.created_by_id = cu.id
	LEFT JOIN users uu ON m.updated_by_id = uu.id
	LEFT JOIN groups g ON m.group_id = g.id
	WHERE g.name = 'vats'`

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

	if err := r.sqlDB.Get(&vat, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return &vat, nil
}

// BeginTransaction starts a new transaction
func (r *VatRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *VatRepository) CreateVat(tx *gorm.DB, vat *models.MixValue, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("VatRepository-CreateVat", opentracing.ChildOf(span.Context()))
	if err := tx.Create(vat).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *VatRepository) UpdateVat(tx *gorm.DB, vat *models.MixValue, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("VatRepository-UpdateVat", opentracing.ChildOf(span.Context()))
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("*").Omit("created_at", "created_by_id").Updates(vat).Error; err != nil {
			defer childSpan.Finish()
			utils.LogErrors(childSpan, err)
			return err
		}
		return nil
	})

}

func (r *VatRepository) DeleteVat(tx *gorm.DB, params *dtos.GetVatParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("VatRepository-DeleteVat", opentracing.ChildOf(span.Context()))
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.MixValue{}, params.ID).Error; err != nil {
			defer childSpan.Finish()
			utils.LogErrors(childSpan, err)
			return err
		}
		return nil
	})
}

func (s *VatRepository) RestoreVat(tx *gorm.DB, params *dtos.GetVatParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("VatRepository-RestoreVat", opentracing.ChildOf(span.Context()))
	return s.db.Transaction(func(tx *gorm.DB) error {
		var vat models.MixValue
		if err := tx.Unscoped().Model(&vat).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
			defer childSpan.Finish()
			utils.LogErrors(childSpan, err)
			return err
		}
		return nil
	})
}
