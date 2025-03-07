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

type RoleRepository struct {
	db     *gorm.DB
	sqlDB  *sqlx.DB
	tracer opentracing.Tracer
}

func NewRoleRepository(db *gorm.DB, sqlDB *sqlx.DB, tracer opentracing.Tracer) *RoleRepository {
	return &RoleRepository{
		db:     db,
		sqlDB:  sqlDB,
		tracer: tracer,
	}
}

func (r *RoleRepository) GetRoles(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RoleListDTO, int, error) {
	// Create a child span for the controller
	childSpan := opentracing.StartSpan("RoleRepository-GetRoles", opentracing.ChildOf(span.Context()))

	roles := []dtos.RoleListDTO{}
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
                WHERE g.name = 'roles'
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM mix_values m
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
                LEFT JOIN groups g ON m.group_id = g.id
                WHERE g.name = 'roles'
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

		err := r.sqlDB.SelectContext(ctx.Context(), &roles, query, args...)
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

	return roles, total, nil
}

func (r *RoleRepository) GetRoleByID(ctx *fiber.Ctx, params *dtos.GetRoleParams, span opentracing.Span) (*dtos.RoleDetailDTO, error) {
	childSpan := opentracing.StartSpan("RoleRepository-GetRoleByID", opentracing.ChildOf(span.Context()))
	var role dtos.RoleDetailDTO

	query := `SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
	cu.name as created_by_name,
	uu.name as updated_by_name

	FROM mix_values m
	LEFT JOIN users cu ON m.created_by_id = cu.id
	LEFT JOIN users uu ON m.updated_by_id = uu.id
	LEFT JOIN groups g ON m.group_id = g.id
	WHERE g.name = 'roles'`

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

	if err := r.sqlDB.Get(&role, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return &role, nil
}

// BeginTransaction starts a new transaction
func (r *RoleRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *RoleRepository) CreateRole(tx *gorm.DB, role *models.MixValue, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("RoleRepository-CreateRole", opentracing.ChildOf(span.Context()))
	if err := tx.Create(role).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *RoleRepository) UpdateRole(tx *gorm.DB, role *models.MixValue, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("RoleRepository-UpdateRole", opentracing.ChildOf(span.Context()))

	if err := tx.Select("*").Omit("created_at", "created_by_id").Updates(role).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *RoleRepository) DeleteRole(tx *gorm.DB, params *dtos.GetRoleParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("RoleRepository-DeleteRole", opentracing.ChildOf(span.Context()))

	if err := tx.Delete(&models.MixValue{}, params.ID).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (s *RoleRepository) RestoreRole(tx *gorm.DB, params *dtos.GetRoleParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("RoleRepository-RestoreRole", opentracing.ChildOf(span.Context()))

	var role models.MixValue
	if err := tx.Unscoped().Model(&role).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}
