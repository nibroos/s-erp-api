package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/auth"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type QuotationRepository struct {
	db       *gorm.DB
	sqlDB    *sqlx.DB
	utilRepo *UtilRepository
	tracer   opentracing.Tracer
}

func NewQuotationRepository(db *gorm.DB, sqlDB *sqlx.DB, utilRepo *UtilRepository, tracer opentracing.Tracer) *QuotationRepository {
	return &QuotationRepository{
		db:       db,
		sqlDB:    sqlDB,
		tracer:   tracer,
		utilRepo: utilRepo,
	}
}

func (r *QuotationRepository) GetQuotations(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.QuotationListDTO, int, error) {
	// Create a child span for the controller
	childSpan := opentracing.StartSpan("QuotationRepository-GetQuotations", opentracing.ChildOf(span.Context()))

	// Simulate an error for testing Jaeger tracing
	if filters["simulate_error"] == "true" {
		utils.LogErrors(childSpan, fmt.Errorf("simulated error"))

		return nil, 0, fmt.Errorf("simulated error")
	}

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	quotations := []dtos.QuotationListDTO{}

	var total int

	// select column
	cdSelect := `m.name, m.specification, m.description, m.tpb_code, iu.price_sell, iu.price_buy, m.minimum_stock,`
	if branchID != nil && !isAdmin {

		cdSelect = `
		COALESCE(bi.name, m.name) as name,
		COALESCE(bi.factory_code, m.factory_code) as factory_code,
		COALESCE(bi.sku, m.sku) as sku,
		COALESCE(bi.barcode, m.barcode) as barcode,
		COALESCE(bi.specification, m.specification) as specification,
		COALESCE(bi.description, m.description) as description,
		COALESCE(bi.remark, m.remark) as remark,
		COALESCE(bi.tpb_code, m.tpb_code) as tpb_code,
		COALESCE(bi.minimum_stock, m.minimum_stock) as minimum_stock,
		COALESCE(bi.price_sell, iu.price_sell) as price_sell,
		COALESCE(bi.price_buy, iu.price_buy) as price_buy,
		COALESCE(bi.margin, iu.margin) as margin,
		COALESCE(bi.status, m.status) as status,
		COALESCE(bi.expired_at, m.expired_at) as expired_at,
		COALESCE(bi.created_at, m.created_at) as created_at,
		COALESCE(bi.updated_at, m.updated_at) as updated_at,
		COALESCE(bi.deleted_at, m.deleted_at) as deleted_at,

		bi.id as branch_item_id,
		`
	} else {
		cdSelect = `
			m.name, m.factory_code, m.sku, m.barcode, m.specification, m.description, m.remark, m.tpb_code, m.minimum_stock, iu.price_sell, iu.price_buy, iu.margin, m.status, m.expired_at, m.created_at, m.updated_at, m.deleted_at,
		`
	}

	query := `SELECT *
    FROM ( 
        SELECT DISTINCT ON (m.id)
					m.id, m.item_sub_group_id, isg.parent_id as item_group_id, m.item_unit_id, m.code, m.is_all_branch,
					` + cdSelect + `

					m.id as quotation_id,
					u.name as unit_name,
					b.name as branch_name,
					isg.name as item_sub_group_name,
					ig.name as item_group_name,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM quotations m
				LEFT JOIN mix_values isg ON m.item_sub_group_id = isg.id
				LEFT JOIN mix_values ig ON isg.parent_id = ig.id
				LEFT JOIN item_units iu ON iu.id = m.item_unit_id
				LEFT JOIN mix_values u ON iu.unit_id = u.id
				LEFT JOIN branch_items bi ON bi.item_unit_id = iu.id
				LEFT JOIN branches b ON bi.branch_id = b.id
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT DISTINCT ON (m.id) 
					m.id, m.item_sub_group_id, isg.parent_id as item_group_id, m.item_unit_id, m.code, m.is_all_branch,
					` + cdSelect + `

					m.id as quotation_id,
					u.name as unit_name,
					bi.name as branch_name,
					isg.name as item_sub_group_name,
					ig.name as item_group_name,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM quotations m
				LEFT JOIN mix_values isg ON m.item_sub_group_id = isg.id
				LEFT JOIN mix_values ig ON isg.parent_id = ig.id
				LEFT JOIN item_units iu ON iu.id = m.item_unit_id
				LEFT JOIN mix_values u ON iu.unit_id = u.id
				LEFT JOIN branch_items bi ON bi.item_unit_id = iu.id
				LEFT JOIN branches b ON bi.branch_id = b.id
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	var args []interface{}

	i := 1
	for key, value := range filters {
		switch key {
		case "name", "code", "factory_code", "sku", "barcode", "specification", "description", "remark":
			if value != "" {
				query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				countQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

	if !isAdmin && branchID != nil {
		query += fmt.Sprintf(" AND (branch_id = $%d)", i)
		countQuery += fmt.Sprintf(" AND (branch_id = $%d)", i)
		args = append(args, branchID)
		i++
	}

	if isAdmin && filters["branch_id"] != "" {
		query += fmt.Sprintf(" AND (branch_id = $%d)", i)
		countQuery += fmt.Sprintf(" AND (branch_id = $%d)", i)
		args = append(args, filters["branch_id"])
		i++
	}

	filterKey := map[string]string{
		"unit_id": "unit_id",
		"status":  "status",
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
		query += fmt.Sprintf(" AND (name ILIKE $%d OR code ILIKE $%d OR factory_code ILIKE $%d OR sku ILIKE $%d OR barcode ILIKE $%d OR specification ILIKE $%d OR description ILIKE $%d OR remark ILIKE $%d)", i, i+1, i+2, i+3, i+4, i+5, i+6, i+7)
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR code ILIKE $%d OR factory_code ILIKE $%d OR sku ILIKE $%d OR barcode ILIKE $%d OR specification ILIKE $%d OR description ILIKE $%d OR remark ILIKE $%d)", i, i+1, i+2, i+3, i+4, i+5, i+6, i+7)
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

		err := r.sqlDB.SelectContext(ctx.Context(), &quotations, query, args...)
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

	return quotations, total, nil
}

func (r *QuotationRepository) GetQuotationByID(ctx *fiber.Ctx, params *dtos.GetQuotationParams, span opentracing.Span) (*dtos.QuotationDetailDTO, error) {
	childSpan := opentracing.StartSpan("QuotationRepository-GetQuotationByID", opentracing.ChildOf(span.Context()))
	var quotation dtos.QuotationDetailDTO

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	// select column
	cdSelect := `m.name, m.specification, m.description, m.tpb_code, iu.price_sell, iu.price_buy, m.minimum_stock,`
	if branchID != nil && !isAdmin {

		cdSelect = `
		COALESCE(bi.name, m.name) as name,
		COALESCE(bi.factory_code, m.factory_code) as factory_code,
		COALESCE(bi.sku, m.sku) as sku,
		COALESCE(bi.barcode, m.barcode) as barcode,
		COALESCE(bi.specification, m.specification) as specification,
		COALESCE(bi.description, m.description) as description,
		COALESCE(bi.remark, m.remark) as remark,
		COALESCE(bi.tpb_code, m.tpb_code) as tpb_code,
		COALESCE(bi.minimum_stock, m.minimum_stock) as minimum_stock,
		COALESCE(bi.price_sell, iu.price_sell) as price_sell,
		COALESCE(bi.price_buy, iu.price_buy) as price_buy,
		COALESCE(bi.margin, iu.margin) as margin,
		COALESCE(bi.status, m.status) as status,
		COALESCE(bi.expired_at, m.expired_at) as expired_at,
		COALESCE(bi.created_at, m.created_at) as created_at,
		COALESCE(bi.updated_at, m.updated_at) as updated_at,
		COALESCE(bi.deleted_at, m.deleted_at) as deleted_at,

		bi.id as branch_item_id,
		`
	} else {
		cdSelect = `
			m.name, m.factory_code, m.sku, m.barcode, m.specification, m.description, m.remark, m.tpb_code, m.minimum_stock, iu.price_sell, iu.price_buy, iu.margin, m.status, m.expired_at, m.created_at, m.updated_at, m.deleted_at,
		`
	}

	query := `SELECT *
    FROM ( 
        SELECT DISTINCT ON (m.id)
					m.id, m.item_sub_group_id, isg.parent_id as item_group_id, m.item_unit_id, m.code, m.is_all_branch,
					` + cdSelect + `

					m.id as quotation_id,
					u.name as unit_name,
					b.name as branch_name,
					isg.name as item_sub_group_name,
					ig.name as item_group_name,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM quotations m
				LEFT JOIN mix_values isg ON m.item_sub_group_id = isg.id
				LEFT JOIN mix_values ig ON isg.parent_id = ig.id
				LEFT JOIN item_units iu ON iu.id = m.item_unit_id
				LEFT JOIN mix_values u ON iu.unit_id = u.id
				LEFT JOIN branch_items bi ON bi.item_unit_id = iu.id
				LEFT JOIN branches b ON bi.branch_id = b.id
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
    ) AS alias WHERE 1=1`

	var args []interface{}

	i := 1
	query += " AND id = $1"
	args = append(args, params.ID)
	i++

	isDeletedQuery := ` AND deleted_at IS NULL`
	if params.IsDeleted != nil && *params.IsDeleted == 1 {
		isDeletedQuery = " AND deleted_at IS NOT NULL"
	}

	query += isDeletedQuery

	if err := r.sqlDB.Get(&quotation, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return &quotation, nil
}

// BeginTransaction starts a new transaction
func (r *QuotationRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

// Rollback all changes in the transaction
func (r *QuotationRepository) Rollback() *gorm.DB {
	return r.db.Rollback()
}

func (r *QuotationRepository) CreateQuotation(tx *gorm.DB, quotation *models.Quotation, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuotationRepository-CreateQuotation", opentracing.ChildOf(span.Context()))
	if err := tx.Create(quotation).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *QuotationRepository) UpdateQuotation(tx *gorm.DB, quotation *models.Quotation, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuotationRepository-UpdateQuotation", opentracing.ChildOf(span.Context()))

	if err := tx.Select("*").Omit(
		"created_at", "created_by_id", "branch_id",
	).Updates(quotation).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *QuotationRepository) DeleteQuotation(tx *gorm.DB, params *dtos.GetQuotationParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuotationRepository-DeleteQuotation", opentracing.ChildOf(span.Context()))

	if err := tx.Delete(&models.Quotation{}, params.ID).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil

}

func (s *QuotationRepository) RestoreQuotation(tx *gorm.DB, params *dtos.GetQuotationParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuotationRepository-RestoreQuotation", opentracing.ChildOf(span.Context()))

	var quotation models.Quotation
	if err := tx.Unscoped().Model(&quotation).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *QuotationRepository) CreateQuoDts(tx *gorm.DB, quoDts []*models.QuoDt, quotationID uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuoDtRepository-CreateQuoDts", opentracing.ChildOf(span.Context()))

	result := tx.CreateInBatches(quoDts, len(quoDts))

	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return result.Error
	}

	return nil
}

// bulk/batch update quoDts
func (r *QuotationRepository) UpdateQuoDts(tx *gorm.DB, quoDts []*models.QuoDt, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuoDtRepository-UpdateQuoDts", opentracing.ChildOf(span.Context()))

	data := make([]map[string]interface{}, 0)
	for _, quoDt := range quoDts {
		data = append(data, map[string]interface{}{
			"id":           quoDt.ID,
			"quotation_id": quoDt.QuotationID,
			"ref_id":       quoDt.RefID,
			"vat_id":       quoDt.VatID,
			"item_unit_id": quoDt.ItemUnitID,
			// "ref_json":      quoDt.RefJSON,
			"ref_type":      quoDt.RefType,
			"remark":        quoDt.Remark,
			"vat_perc":      quoDt.VatPerc,
			"qty_so":        quoDt.QtySO,
			"qty":           quoDt.Qty,
			"price_sell":    quoDt.PriceSell,
			"subtotal":      quoDt.Subtotal,
			"disc_am":       quoDt.DiscAm,
			"disc_perc":     quoDt.DiscPerc,
			"total_am":      quoDt.TotalAm,
			"updated_by_id": quoDt.UpdatedByID,
			"updated_at":    time.Now(),
		})
	}

	// if err := r.utilRepo.BulkUpdate(tx, "quoDts", "id", data, childSpan); err != nil {
	if err := r.utilRepo.Upsert(tx, "quo_dts", "id", data, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *QuotationRepository) DeleteQuoDtsWhereNotIn(tx *gorm.DB, quotationID uint, quoDtIDs []uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuoDtRepository-DeleteQuoDtsWhereNotIn", opentracing.ChildOf(span.Context()))

	if err := tx.Where("quotation_id = ? AND id NOT IN ?", quotationID, quoDtIDs).Delete(&models.QuoDt{}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

// func (r *QuotationRepository) GetQuoDtsByQuotationIDs(ctx *fiber.Ctx, quotationIDs uint, span opentracing.Span) ([]dtos.QuotationQuoDtListDTO, error) {
func (r *QuotationRepository) GetQuoDtsByQuotationIDs(ctx *fiber.Ctx, quotationIDs []uint, span opentracing.Span) ([]dtos.QuotationQuoDtListDTO, error) {
	childSpan := opentracing.StartSpan("QuotationRepository-GetQuoDtsByQuotationID", opentracing.ChildOf(span.Context()))

	quoDts := []dtos.QuotationQuoDtListDTO{}

	query := `SELECT qd.id, qd.quotation_id, 
		qd.item_unit_id, qd.vat_id, qd.ref_id, qd.item_id, qd.ref_type, qd.item_type, qd.remark, qd.vat_perc, qd.qty_so, qd.qty, qd.price_sell, qd.price_buy, qd.subtotal, qd.disc_am, qd.disc_perc, qd.total_am, qd.vat_perc,
		qd.created_at, qd.updated_at, qd.deleted_at,

		isg.id as item_sub_group_id,
		ig.id as item_group_id,
		isg.name as item_sub_group_name,
		ig.name as item_group_name,
		u.name as unit_name,
		pi.name as item_name,

		cu.name as created_by_name,
		uu.name as updated_by_name

	FROM quo_dts qd
	LEFT JOIN quotations p ON qd.quotation_id = p.id
	LEFT JOIN products pi ON qd.item_id = pi.id
	LEFT JOIN item_units iu ON qd.item_unit_id = iu.id
	LEFT JOIN mix_values u ON iu.unit_id = u.id
	LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
	LEFT JOIN mix_values ig ON isg.parent_id = ig.id
	LEFT JOIN users cu ON qd.created_by_id = cu.id
	LEFT JOIN users uu ON qd.updated_by_id = uu.id
	WHERE qd.deleted_at IS NULL`

	var args []interface{}
	// i := 1

	if len(quotationIDs) > 0 {
		query += " AND qd.quotation_id IN (" + utils.JoinUintsToString(quotationIDs, ",") + ")"
	}

	if err := r.sqlDB.SelectContext(ctx.Context(), &quoDts, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return quoDts, nil
}

func (r *QuotationRepository) CreateQuoDtBoms(tx *gorm.DB, quoDtBoms []*models.QuoDtBom, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuoDtBomRepository-CreateQuoDtBoms", opentracing.ChildOf(span.Context()))

	result := tx.CreateInBatches(quoDtBoms, len(quoDtBoms))

	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return result.Error
	}

	return nil
}

func (r *QuotationRepository) DeleteQuoDtBomsWhereNotIn(tx *gorm.DB, quotationID uint, quoDtIDs []uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuoDtRepository-DeleteQuoDtsWhereNotIn", opentracing.ChildOf(span.Context()))

	if err := tx.Where("quotation_id = ? AND id NOT IN ?", quotationID, quoDtIDs).Delete(&models.QuoDt{}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

// bulk/batch update quoDts
func (r *QuotationRepository) UpdateQuoDtBoms(tx *gorm.DB, quoDtBoms []*models.QuoDtBom, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuoDtRepository-UpdateQuoDtBoms", opentracing.ChildOf(span.Context()))

	data := make([]map[string]interface{}, 0)
	for _, quoDtBom := range quoDtBoms {
		data = append(data, map[string]interface{}{
			"id":           quoDtBom.ID,
			"quotation_id": quoDtBom.QuotationID,
			"quo_dt_id":    quoDtBom.QuoDtID,
			"product_id":   quoDtBom.ProductID,
			"item_id":      quoDtBom.ItemID,
			"item_unit_id": quoDtBom.ItemUnitID,
			// "ref_json":      quoDtBom.RefJSON,
			"remark":        quoDtBom.Remark,
			"qty":           quoDtBom.Qty,
			"price_sell":    quoDtBom.PriceSell,
			"price_buy":     quoDtBom.PriceBuy,
			"subtotal":      quoDtBom.Subtotal,
			"updated_by_id": quoDtBom.UpdatedByID,
			"updated_at":    time.Now(),
		})
	}

	// if err := r.utilRepo.BulkUpdate(tx, "quoDtBoms", "id", data, childSpan); err != nil {
	if err := r.utilRepo.Upsert(tx, "quo_dt_boms", "id", data, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}
