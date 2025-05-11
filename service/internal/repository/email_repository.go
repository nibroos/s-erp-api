package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type EmailRepository struct {
	db     *gorm.DB
	sqlDB  *sqlx.DB
	tracer opentracing.Tracer
}

func NewEmailRepository(db *gorm.DB, sqlDB *sqlx.DB, tracer opentracing.Tracer) *EmailRepository {
	return &EmailRepository{
		db:     db,
		sqlDB:  sqlDB,
		tracer: tracer,
	}
}

func (r *EmailRepository) CreateEmail(tx *gorm.DB, unit *models.MixValue, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("EmailRepository-CreateEmail", opentracing.ChildOf(span.Context()))
	if err := tx.Create(unit).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

// func (r *EmailRepository) GetEmails(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.EmailListDTO, int, error) {
// 	// Create a child span for the controller
// 	childSpan := opentracing.StartSpan("EmailRepository-GetEmails", opentracing.ChildOf(span.Context()))

// 	units := []dtos.EmailListDTO{}
// 	var total int

// 	condition := ""
// 	var args []interface{}
// 	i := 1

// 	filterEqual := map[string]string{
// 		"status":    "m.status",
// 		"is_active": "m.status",
// 	}
// 	for key, colDB := range filterEqual {
// 		if value, ok := filters[key]; ok && value != "" {
// 			condition += fmt.Sprintf(" AND %s = $%d", colDB, i)
// 			args = append(args, value)
// 			i++
// 		}
// 	}

// 	query := `SELECT *
//     FROM (
//         SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,

// 				m.id as unit_id,

//         cu.name as created_by_name,
//         uu.name as updated_by_name

//         FROM mix_values m
//                 LEFT JOIN groups g ON m.group_id = g.id
//         LEFT JOIN users cu ON m.created_by_id = cu.id
//         LEFT JOIN users uu ON m.updated_by_id = uu.id
//                 WHERE g.name = 'units' ` + condition + `
//     ) AS alias WHERE 1=1 AND deleted_at IS NULL`

// 	countQuery := `SELECT COUNT(*) FROM (
//         SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
//         cu.name as created_by_name,
//         uu.name as updated_by_name

//         FROM mix_values m
//         LEFT JOIN users cu ON m.created_by_id = cu.id
//         LEFT JOIN users uu ON m.updated_by_id = uu.id
//                 LEFT JOIN groups g ON m.group_id = g.id
//                 WHERE g.name = 'units' ` + condition + `
//     ) AS alias WHERE 1=1 AND deleted_at IS NULL`

// 	for key, value := range filters {
// 		switch key {
// 		case "name", "description", "remark":
// 			if value != "" {
// 				query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
// 				countQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
// 				args = append(args, "%"+value+"%")
// 				i++
// 			}
// 		}
// 	}

// 	if filters["ids"] != "" {
// 		query += fmt.Sprintf(" AND id IN (%s)", filters["ids"])
// 	}
// 	if value, ok := filters["global"]; ok && value != "" {
// 		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d OR remark ILIKE $%d)", i, i+1, i+2)
// 		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d OR remark ILIKE $%d)", i, i+1, i+2)
// 		args = append(args, "%"+value+"%", "%"+value+"%", "%"+value+"%")
// 		i += 3
// 	}

// 	countArgs := append([]interface{}{}, args...)

// 	var wg sync.WaitGroup
// 	var countErr, selectErr error

// 	// Goroutine for count query
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		if filters["is_csv"] != "1" {
// 			// Create a span for the count query
// 			countSpan := opentracing.StartSpan("CountQuery", opentracing.ChildOf(childSpan.Context()))

// 			err := r.sqlDB.GetContext(ctx.Context(), &total, countQuery, countArgs...)
// 			if err != nil {
// 				utils.LogErrors(countSpan, err)
// 				countSpan.LogKV("query", countQuery)
// 				countErr = err
// 			}
// 		}
// 	}()

// 	if countErr != nil {
// 		return nil, 0, countErr
// 	}

// 	orderColumn := utils.GetStringOrDefault(filters["order_column"], "name")
// 	orderDirection := utils.GetStringOrDefault(filters["order_direction"], "asc")
// 	query += fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)

// 	perPage := utils.GetIntOrDefault(filters["per_page"], 10)
// 	currentPage := utils.GetIntOrDefault(filters["page"], 1)

// 	// if is_csv
// 	if filters["is_csv"] != "1" {
// 		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
// 		args = append(args, perPage, (currentPage-1)*perPage)
// 	}

// 	// Goroutine for select query
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		// Create a span for the select query
// 		selectSpan := opentracing.StartSpan("SelectQuery", opentracing.ChildOf(childSpan.Context()))

// 		err := r.sqlDB.SelectContext(ctx.Context(), &units, query, args...)
// 		if err != nil {
// 			selectSpan.LogKV("query", query)
// 			utils.LogErrors(selectSpan, err)
// 			selectErr = err
// 		}
// 	}()

// 	// Wait for both goroutines to finish
// 	wg.Wait()

// 	if countErr != nil || selectErr != nil {
// 		defer childSpan.Finish()
// 	}

// 	if countErr != nil {
// 		return nil, 0, countErr
// 	}

// 	if selectErr != nil {
// 		return nil, 0, selectErr
// 	}

// 	return units, total, nil
// }

// func (r *EmailRepository) GetEmailByID(ctx *fiber.Ctx, params *dtos.GetEmailParams, span opentracing.Span) (*dtos.EmailDetailDTO, error) {
// 	childSpan := opentracing.StartSpan("EmailRepository-GetEmailByID", opentracing.ChildOf(span.Context()))
// 	var unit dtos.EmailDetailDTO

// 	query := `SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
// 	cu.name as created_by_name,
// 	uu.name as updated_by_name

// 	FROM mix_values m
// 	LEFT JOIN users cu ON m.created_by_id = cu.id
// 	LEFT JOIN users uu ON m.updated_by_id = uu.id
// 	LEFT JOIN groups g ON m.group_id = g.id
// 	WHERE g.name = 'units'`

// 	var args []interface{}

// 	i := 1
// 	query += " AND m.id = $1"
// 	args = append(args, params.ID)
// 	i++

// 	isDeletedQuery := ` AND m.deleted_at IS NULL`
// 	if params.IsDeleted != nil && *params.IsDeleted == 1 {
// 		isDeletedQuery = " AND m.deleted_at IS NOT NULL"
// 	}

// 	query += isDeletedQuery

// 	if err := r.sqlDB.Get(&unit, query, args...); err != nil {
// 		utils.LogErrors(childSpan, err)
// 		return nil, err
// 	}

// 	return &unit, nil
// }

// BeginTransaction starts a new transaction
func (r *EmailRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *EmailRepository) UpdateEmail(tx *gorm.DB, unit *models.MixValue, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("EmailRepository-UpdateEmail", opentracing.ChildOf(span.Context()))

	if err := tx.Select("*").Omit("created_at", "created_by_id").Updates(unit).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

// func (r *EmailRepository) DeleteEmail(tx *gorm.DB, params *dtos.GetEmailParams, span opentracing.Span) error {
// 	childSpan := opentracing.StartSpan("EmailRepository-DeleteEmail", opentracing.ChildOf(span.Context()))

// 	if err := tx.Delete(&models.MixValue{}, params.ID).Error; err != nil {
// 		defer childSpan.Finish()
// 		utils.LogErrors(childSpan, err)
// 		return err
// 	}
// 	return nil
// }

// func (s *EmailRepository) RestoreEmail(tx *gorm.DB, params *dtos.GetEmailParams, span opentracing.Span) error {
// 	childSpan := opentracing.StartSpan("EmailRepository-RestoreEmail", opentracing.ChildOf(span.Context()))

// 	var unit models.MixValue
// 	if err := tx.Unscoped().Model(&unit).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
// 		defer childSpan.Finish()
// 		utils.LogErrors(childSpan, err)
// 		return err
// 	}
// 	return nil
// }
