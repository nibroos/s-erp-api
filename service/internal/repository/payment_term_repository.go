package repository

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type PaymentTermRepository struct {
	db     *gorm.DB
	sqlDB  *sqlx.DB
	tracer opentracing.Tracer
}

func NewPaymentTermRepository(db *gorm.DB, sqlDB *sqlx.DB, tracer opentracing.Tracer) *PaymentTermRepository {
	return &PaymentTermRepository{
		db:     db,
		sqlDB:  sqlDB,
		tracer: tracer,
	}
}

func (r *PaymentTermRepository) GetPaymentTerms(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.PaymentTermListDTO, int, error) {
	childSpan := opentracing.StartSpan("PaymentTermRepository-GetPaymentTerms", opentracing.ChildOf(span.Context()))

	paymentTerms := []dtos.PaymentTermListDTO{}
	var total int

	condition := ""
	var args []interface{}
	i := 1

	filterEqual := map[string]string{
		"status":    "m.status",
		"is_active": "m.status",
	}
	for key, colDB := range filterEqual {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", colDB, i)
			args = append(args, value)
			i++
		}
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
        WHERE g.name = 'payment_terms' ` + condition + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM mix_values m
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
        LEFT JOIN groups g ON m.group_id = g.id
        WHERE g.name = 'payment_terms' ` + condition + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

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

	if filters["ids"] != "" {
		query += fmt.Sprintf(" AND id IN (%s)", filters["ids"])
	}
	if value, ok := filters["global"]; ok && value != "" {
		searchFields := []string{"name", "description", "remark"}
		searchConditions := make([]string, len(searchFields))

		for idx, field := range searchFields {
			searchConditions[idx] = fmt.Sprintf("%s ILIKE $%d", field, i+idx)
			args = append(args, "%"+value+"%")
		}

		query += fmt.Sprintf(" AND (%s)", strings.Join(searchConditions, " OR "))
		countQuery += fmt.Sprintf(" AND (%s)", strings.Join(searchConditions, " OR "))
		i += len(searchFields)
	}

	countArgs := append([]interface{}{}, args...)

	var wg sync.WaitGroup
	var countErr, selectErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		if filters["is_csv"] != "1" {
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

	if filters["is_csv"] != "1" {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
		args = append(args, perPage, (currentPage-1)*perPage)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		selectSpan := opentracing.StartSpan("SelectQuery", opentracing.ChildOf(childSpan.Context()))

		err := r.sqlDB.SelectContext(ctx.Context(), &paymentTerms, query, args...)
		if err != nil {
			selectSpan.LogKV("query", query)
			utils.LogErrors(selectSpan, err)
			selectErr = err
		}
	}()

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

	return paymentTerms, total, nil
}

func (r *PaymentTermRepository) GetPaymentTermByID(ctx *fiber.Ctx, params *dtos.GetPaymentTermParams, span opentracing.Span) (*dtos.PaymentTermDetailDTO, error) {
	childSpan := opentracing.StartSpan("PaymentTermRepository-GetPaymentTermByID", opentracing.ChildOf(span.Context()))
	var term dtos.PaymentTermDetailDTO

	query := `SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
    cu.name as created_by_name,
    uu.name as updated_by_name

    FROM mix_values m
    LEFT JOIN users cu ON m.created_by_id = cu.id
    LEFT JOIN users uu ON m.updated_by_id = uu.id
    LEFT JOIN groups g ON m.group_id = g.id
    WHERE g.name = 'payment_terms'`

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

	if err := r.sqlDB.Get(&term, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return &term, nil
}

func (r *PaymentTermRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *PaymentTermRepository) CreatePaymentTerm(tx *gorm.DB, term *models.MixValue, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PaymentTermRepository-CreatePaymentTerm", opentracing.ChildOf(span.Context()))
	if err := tx.Create(term).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *PaymentTermRepository) UpdatePaymentTerm(tx *gorm.DB, term *models.MixValue, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PaymentTermRepository-UpdatePaymentTerm", opentracing.ChildOf(span.Context()))

	if err := tx.Select("*").Omit("created_at", "created_by_id").Updates(term).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *PaymentTermRepository) DeletePaymentTerm(tx *gorm.DB, params *dtos.GetPaymentTermParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PaymentTermRepository-DeletePaymentTerm", opentracing.ChildOf(span.Context()))

	if err := tx.Model(&models.MixValue{}).Where("id = ?", params.ID).Updates(map[string]interface{}{
		"deleted_at":    time.Now(),
		"deleted_by_id": params.DeletedByID,
	}).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *PaymentTermRepository) RestorePaymentTerm(tx *gorm.DB, params *dtos.GetPaymentTermParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PaymentTermRepository-RestorePaymentTerm", opentracing.ChildOf(span.Context()))

	var term models.MixValue
	if err := tx.Unscoped().Model(&term).Where("id = ?", params.ID).Updates(map[string]interface{}{
		"deleted_at":    nil,
		"deleted_by_id": nil,
	}).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}
