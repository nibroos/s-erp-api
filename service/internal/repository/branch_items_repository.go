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

type BranchItemRepository struct {
	db     *gorm.DB
	sqlDB  *sqlx.DB
	tracer opentracing.Tracer
}

func NewBranchItemRepository(db *gorm.DB, sqlDB *sqlx.DB, tracer opentracing.Tracer) *BranchItemRepository {
	return &BranchItemRepository{
		db:     db,
		sqlDB:  sqlDB,
		tracer: tracer,
	}
}

func (r *BranchItemRepository) GetBranchItems(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.BranchItemListDTO, int, error) {
	// Create a child span for the controller
	childSpan := opentracing.StartSpan("BranchItemRepository-GetBranchItems", opentracing.ChildOf(span.Context()))

	branchItems := []dtos.BranchItemListDTO{}
	var total int

	query := `SELECT *
    FROM ( 
        SELECT DISTINCT ON (m.id)
					m.id, m.product_id, m.branch_id, m.code, m.factory_code, m.name, m.sku, m.barcode, m.specification, m.description, m.remark, m.tpb_code, m.minimum_stock, m.price_sell, m.price_buy, m.margin, m.status, m.created_at, m.updated_at, m.deleted_at,
					p.name as product_name,
					u.name as unit_name,
					iu.unit_id,
					b.name as branch_name,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM branch_items m
				LEFT JOIN item_units iu ON m.item_unit_id = iu.id
				LEFT JOIN products p ON iu.product_id = p.id
				LEFT JOIN units u ON iu.unit_id = u.id
				LEFT JOIN branches b ON m.branch_id = b.id
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT DISTINCT ON (m.id) 
					m.id, m.product_id, m.branch_id, m.code, m.factory_code, m.name, m.sku, m.barcode, m.specification, m.description, m.remark, m.tpb_code, m.minimum_stock, m.price_sell, m.price_buy, m.margin, m.status, m.created_at, m.updated_at, m.deleted_at,
					p.name as product_name,
					u.name as unit_name,
					iu.unit_id,
					b.name as branch_name,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM branch_items m
				LEFT JOIN item_units iu ON m.item_unit_id = iu.id
				LEFT JOIN products p ON iu.product_id = p.id
				LEFT JOIN units u ON iu.unit_id = u.id
				LEFT JOIN branches b ON m.branch_id = b.id
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	var args []interface{}

	i := 1

	filterKey := []string{
		"branch_id",
		"product_id",
		"unit_id",
		"status",
	}

	for key := range filterKey {
		if filters[filterKey[key]] != "" {
			query += fmt.Sprintf(" AND %s = $%d", filterKey[key], i)
			countQuery += fmt.Sprintf(" AND %s = $%d", filterKey[key], i)
			args = append(args, filters[filterKey[key]])
			i++
		}
	}

	if value, ok := filters["global"]; ok && value != "" {
		query += fmt.Sprintf(" AND (code ILIKE $%d OR factory_code ILIKE $%d OR name ILIKE $%d OR sku ILIKE $%d OR barcode ILIKE $%d OR specification ILIKE $%d OR description ILIKE $%d OR remark ILIKE $%d OR tpb_code ILIKE $%d)", i, i+1, i+2, i+3, i+4, i+5, i+6, i+7, i+8)
		countQuery += fmt.Sprintf(" AND (code ILIKE $%d OR factory_code ILIKE $%d OR name ILIKE $%d OR sku ILIKE $%d OR barcode ILIKE $%d OR specification ILIKE $%d OR description ILIKE $%d OR remark ILIKE $%d OR tpb_code ILIKE $%d)", i, i+1, i+2, i+3, i+4, i+5, i+6, i+7, i+8)
		args = append(args, "%"+value+"%", "%"+value+"%", "%"+value+"%", "%"+value+"%", "%"+value+"%", "%"+value+"%", "%"+value+"%", "%"+value+"%", "%"+value+"%")
		i += 9
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

		err := r.sqlDB.SelectContext(ctx.Context(), &branchItems, query, args...)
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

	return branchItems, total, nil
}

func (r *BranchItemRepository) GetBranchItemByID(ctx *fiber.Ctx, params *dtos.GetBranchItemParams, span opentracing.Span) (*dtos.BranchItemDetailDTO, error) {
	childSpan := opentracing.StartSpan("BranchItemRepository-GetBranchItemByID", opentracing.ChildOf(span.Context()))
	var branchItem dtos.BranchItemDetailDTO

	query := `
	SELECT DISTINCT ON (m.id) 
			m.id, m.product_id, m.branch_id, m.code, m.factory_code, m.name, m.sku, m.barcode, m.specification, m.description, m.remark, m.tpb_code, m.minimum_stock, m.price_sell, m.price_buy, m.margin, m.status, m.created_at, m.updated_at, m.deleted_at,
			p.name as product_name,
			u.name as unit_name,
			iu.unit_id,
			b.name as branch_name,
				
			cu.name as created_by_name,
			uu.name as updated_by_name
	FROM branch_items m
	LEFT JOIN item_units iu ON m.item_unit_id = iu.id
	LEFT JOIN products p ON iu.product_id = p.id
	LEFT JOIN units u ON iu.unit_id = u.id
	LEFT JOIN branches b ON m.branch_id = b.id
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

	if err := r.sqlDB.Get(&branchItem, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return &branchItem, nil
}

// BeginTransaction starts a new transaction
func (r *BranchItemRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *BranchItemRepository) CreateBranchItem(tx *gorm.DB, branchItem *models.BranchItem, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("BranchItemRepository-CreateBranchItem", opentracing.ChildOf(span.Context()))
	if err := tx.Create(branchItem).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *BranchItemRepository) UpdateBranchItem(tx *gorm.DB, branchItem *models.BranchItem, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("BranchItemRepository-UpdateBranchItem", opentracing.ChildOf(span.Context()))

	if err := tx.Select("*").Omit("created_at", "created_by_id").Updates(branchItem).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil

}

func (r *BranchItemRepository) DeleteBranchItem(tx *gorm.DB, params *dtos.GetBranchItemParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("BranchItemRepository-DeleteBranchItem", opentracing.ChildOf(span.Context()))

	if err := tx.Delete(&models.BranchItem{}, params.ID).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil

}

func (s *BranchItemRepository) RestoreBranchItem(tx *gorm.DB, params *dtos.GetBranchItemParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("BranchItemRepository-RestoreBranchItem", opentracing.ChildOf(span.Context()))

	var branchItem models.BranchItem
	if err := tx.Unscoped().Model(&branchItem).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil

}
