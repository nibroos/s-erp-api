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

func (r *VatRepository) GetVats(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.VatListDTO, int, error) {
	// Create a child span for the controller
	childSpan := opentracing.StartSpan("VatRepository-GetVats", opentracing.ChildOf(span.Context()))

	vats := []dtos.VatListDTO{}
	var total int

	condition := ""

	var args []interface{}

	i := 1

	filterDBColumnKey := []string{
		"m.name",
		"m.description",
		"m.remark",
	}

	queryGlobal := ""

	if value, ok := filters["global"]; ok && value != "" {
		queryGlobal = " AND ("
		for idx, column := range filterDBColumnKey {
			if idx > 0 {
				queryGlobal += " OR"
			}
			queryGlobal += fmt.Sprintf(" %s ILIKE $%d", column, i)
			args = append(args, "%"+value+"%")
			i++
		}
		queryGlobal += ")"
	}

	filterDateAt := map[string]string{
		"date_at": "m.date_at",
	}

	// for key, colDB := range filterDateAt {
	// 	if value, ok := filters[key]; ok && value != "" {
	// 		// date before or after
	// 		// condition += fmt.Sprintf(" AND %s = $%d", value, i)
	// 		condition += fmt.Sprintf(" AND %s <= $%d", colDB, i)
	// 		args = append(args, value)
	// 		i++
	// 	}
	// }
	for key, colDB := range filterDateAt {
		if value, ok := filters[key]; ok && value != "" {
			if value < "2022-04-01" {
				value = "2022-04-01"
			}

			condition += fmt.Sprintf(" AND %s <= $%d", colDB, i)
			args = append(args, value)
			i++
		}
	}

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

	filtersParams := map[string]string{
		"name":        "m.name",
		"description": "m.description",
		"remark":      "m.remark",
	}

	for key, colDB := range filtersParams {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s ILIKE $%d", colDB, i)
			// countQuery += fmt.Sprintf(" AND %s ILIKE $%d", value, i)
			args = append(args, "%"+value+"%")
			i++
		}
	}

	filterIDsKey := map[string]string{
		"ids": "m.id",
	}

	for key, valueID := range filterIDsKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s IN (%d)", valueID, i)
			args = append(args, value)
			i++
		}
	}

	query := `SELECT *
    FROM ( 
        SELECT DISTINCT ON (m.id)
				m.id, m.name, m.num, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
				TO_CHAR(m.date_at, 'YYYY-MM-DD') as date_at,

				vh.multiplier, vh.divider,
        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM mix_values m
				LEFT JOIN (
						SELECT * FROM vat_histories vh
						ORDER BY vh.changed_at DESC
				) vh ON m.id = vh.vat_id
				LEFT JOIN groups g ON m.group_id = g.id
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
				WHERE g.name = 'vats' ` + queryGlobal + condition + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT DISTINCT ON (m.id)
				m.id, m.name, m.num, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
				TO_CHAR(m.date_at, 'YYYY-MM-DD') as date_at,

				vh.multiplier, vh.divider,
        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM mix_values m
				LEFT JOIN (
						SELECT * FROM vat_histories vh
						ORDER BY vh.changed_at DESC
				) vh ON m.id = vh.vat_id
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
				LEFT JOIN groups g ON m.group_id = g.id
				WHERE g.name = 'vats' ` + queryGlobal + condition + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

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

		err := r.sqlDB.SelectContext(ctx.Context(), &vats, query, args...)
		if err != nil {
			selectSpan.LogKV("query", query)
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

func (r *VatRepository) GetVatByID(ctx *fiber.Ctx, params *dtos.GetVatParams, span opentracing.Span) (*dtos.VatDetailDTO, error) {
	childSpan := opentracing.StartSpan("VatRepository-GetVatByID", opentracing.ChildOf(span.Context()))
	var vat dtos.VatDetailDTO

	query := `SELECT m.id, m.name, m.num, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
	TO_CHAR(m.date_at, 'YYYY-MM-DD') as date_at,

	vh.multiplier, vh.divider,
	cu.name as created_by_name,
	uu.name as updated_by_name

	FROM mix_values m
	LEFT JOIN vat_histories vh ON m.id = vh.vat_id
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

	// ORDER BY
	query += " ORDER BY vh.changed_at DESC"

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

	if err := tx.Select("*").Omit("created_at", "created_by_id").Updates(vat).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *VatRepository) DeleteVat(tx *gorm.DB, params *dtos.GetVatParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("VatRepository-DeleteVat", opentracing.ChildOf(span.Context()))

	if err := tx.Delete(&models.MixValue{}, params.ID).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (s *VatRepository) RestoreVat(tx *gorm.DB, params *dtos.GetVatParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("VatRepository-RestoreVat", opentracing.ChildOf(span.Context()))

	var vat models.MixValue
	if err := tx.Unscoped().Model(&vat).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *VatRepository) GetVatHistories(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.VatHistoryListDTO, int, error) {
	// Create a child span for the controller
	childSpan := opentracing.StartSpan("VatRepository-GetVatHistories", opentracing.ChildOf(span.Context()))

	vatHistories := []dtos.VatHistoryListDTO{}
	var total int

	query := `SELECT *
    FROM ( 
        SELECT 
					vh.id, vh.vat_id, vh.num, vh.divider, vh.multiplier, vh.changed_at, vh.status, vh.remark, vh.created_at, vh.updated_at, vh.deleted_at,
					m.name,
					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM vat_histories vh
				LEFT JOIN mix_values m ON vh.vat_id = m.id
				LEFT JOIN groups g ON m.group_id = g.id
        LEFT JOIN users cu ON vh.created_by_id = cu.id
        LEFT JOIN users uu ON vh.updated_by_id = uu.id
				WHERE g.name = 'vats'
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM (
				SELECT
					vh.id, vh.vat_id, vh.num, vh.divider, vh.multiplier, vh.changed_at, vh.status, vh.remark, vh.created_at, vh.updated_at, vh.deleted_at,
					m.name,
					cu.name as created_by_name,
					uu.name as updated_by_name

				FROM vat_histories vh
				LEFT JOIN mix_values m ON vh.vat_id = m.id
				LEFT JOIN users cu ON vh.created_by_id = cu.id
				LEFT JOIN users uu ON vh.updated_by_id = uu.id
				LEFT JOIN groups g ON m.group_id = g.id
				WHERE g.name = 'vats'
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	var args []interface{}

	i := 1
	for key, value := range filters {
		switch key {
		case "name", "remark":
			if value != "" {
				query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				countQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

	if value, ok := filters["vat_id"]; ok && value != "" {
		query += fmt.Sprintf(" AND vat_id = $%d", i)
		countQuery += fmt.Sprintf(" AND vat_id = $%d", i)
		args = append(args, value)
		i++
	}

	if filters["ids"] != "" {
		query += fmt.Sprintf(" AND id IN (%s)", filters["ids"])
	}
	if value, ok := filters["global"]; ok && value != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR remark ILIKE $%d)", i, i+1, i+2)
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR remark ILIKE $%d)", i, i+1, i+2)
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "changed_at")
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

		err := r.sqlDB.SelectContext(ctx.Context(), &vatHistories, query, args...)
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

	return vatHistories, total, nil
}

func (r *VatRepository) CreateVatHistory(tx *gorm.DB, vatHistory *models.VatHistory, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("VatRepository-CreateVatHistory", opentracing.ChildOf(span.Context()))
	if err := tx.Create(vatHistory).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *VatRepository) GetVatHistoryByID(ctx *fiber.Ctx, params *dtos.GetVatHistoryParams, span opentracing.Span) (*dtos.VatHistoryDetailDTO, error) {
	childSpan := opentracing.StartSpan("VatRepository-GetVatHistoryByVatID", opentracing.ChildOf(span.Context()))
	var vatHistory dtos.VatHistoryDetailDTO

	query := `SELECT 
	vh.id, vh.vat_id, vh.num, vh.divider, vh.multiplier, vh.changed_at, vh.status, vh.remark, vh.created_at, vh.updated_at, vh.deleted_at,
	m.name,
	cu.name as created_by_name,
	uu.name as updated_by_name

	FROM vat_histories vh
	LEFT JOIN mix_values m ON vh.vat_id = m.id
	LEFT JOIN users cu ON vh.created_by_id = cu.id
	LEFT JOIN users uu ON vh.updated_by_id = uu.id
	LEFT JOIN groups g ON m.group_id = g.id
	WHERE g.name = 'vats'`

	var args []interface{}

	i := 1
	if params.ID != nil {
		query += " AND vh.id = $1"
		args = append(args, params.ID)
		i++
	}

	// if params.VatID != nil {
	// 	query += " AND vh.vat_id = $2"
	// 	args = append(args, params.VatID)
	// 	i++
	// }

	isDeletedQuery := ` AND m.deleted_at IS NULL`
	query += isDeletedQuery

	// order by changed_at desc
	if params.IsLatest != nil && *params.IsLatest == 1 {
		query += " ORDER BY vh.changed_at DESC, vh.updated_at DESC, vh.created_at DESC"
	}

	if err := r.sqlDB.Get(&vatHistory, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return &vatHistory, nil
}

func (r *VatRepository) UpdateVatHistory(tx *gorm.DB, vat *models.VatHistory, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("VatRepository-UpdateVatHistory", opentracing.ChildOf(span.Context()))

	if err := tx.Select("*").Omit("vat_id", "created_at", "created_by_id").Updates(vat).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *VatRepository) DeleteVatHistory(tx *gorm.DB, params *dtos.GetVatHistoryParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("VatRepository-DeleteVatHistory", opentracing.ChildOf(span.Context()))

	if err := tx.Delete(&models.VatHistory{}, params.ID).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (s *VatRepository) RestoreVatHistory(tx *gorm.DB, params *dtos.GetVatHistoryParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("VatRepository-RestoreVatHistory", opentracing.ChildOf(span.Context()))

	var vat models.VatHistory
	if err := tx.Unscoped().Model(&vat).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}
