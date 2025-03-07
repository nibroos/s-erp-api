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

type ProductRepository struct {
	db       *gorm.DB
	sqlDB    *sqlx.DB
	utilRepo *UtilRepository
	tracer   opentracing.Tracer
}

func NewProductRepository(db *gorm.DB, sqlDB *sqlx.DB, utilRepo *UtilRepository, tracer opentracing.Tracer) *ProductRepository {
	return &ProductRepository{
		db:       db,
		sqlDB:    sqlDB,
		tracer:   tracer,
		utilRepo: utilRepo,
	}
}

func (r *ProductRepository) GetProducts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.ProductListDTO, int, error) {
	// Create a child span for the controller
	childSpan := opentracing.StartSpan("ProductRepository-GetProducts", opentracing.ChildOf(span.Context()))

	// Simulate an error for testing Jaeger tracing
	if filters["simulate_error"] == "true" {
		utils.LogErrors(childSpan, fmt.Errorf("simulated error"))

		return nil, 0, fmt.Errorf("simulated error")
	}

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	products := []dtos.ProductListDTO{}

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

					m.id as product_id,
					u.name as unit_name,
					isg.name as item_sub_group_name,
					ig.name as item_group_name,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM products m
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

					m.id as product_id,
					u.name as unit_name,
					isg.name as item_sub_group_name,
					ig.name as item_group_name,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM products m
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

		err := r.sqlDB.SelectContext(ctx.Context(), &products, query, args...)
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

	return products, total, nil
}

func (r *ProductRepository) GetProductByID(ctx *fiber.Ctx, params *dtos.GetProductParams, span opentracing.Span) (*dtos.ProductDetailDTO, error) {
	childSpan := opentracing.StartSpan("ProductRepository-GetProductByID", opentracing.ChildOf(span.Context()))
	var product dtos.ProductDetailDTO

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

					m.id as product_id,
					u.name as unit_name,
					b.name as branch_name,
					isg.name as item_sub_group_name,
					ig.name as item_group_name,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM products m
				LEFT JOIN mix_values isg ON m.item_sub_group_id = isg.id
				LEFT JOIN mix_values ig ON isg.parent_id = ig.id
				LEFT JOIN item_units iu ON iu.id = m.item_unit_id
				LEFT JOIN mix_values u ON iu.unit_id = u.id
				LEFT JOIN branch_items bi ON bi.item_unit_id = iu.id
				LEFT JOIN branches b ON bi.branch_id = b.id
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
    ) AS alias WHERE 1=1`

	// query := ` SELECT *
	// FROM (

	// 	SELECT DISTINCT ON (m.id)
	// 		m.id, m.code, m.factory_code, m.name, m.sku, m.barcode, m.specification, m.description, m.remark, iu.price_sell, iu.price_buy, iu.margin, m.expired_at, m.status, m.created_at, m.updated_at, m.deleted_at,

	// 		m.id as product_id,
	// 		m.name as product_name,
	// 		u.name as unit_name,

	// 		cu.name as created_by_name,
	// 		uu.name as updated_by_name

	// 	FROM products m
	// 	LEFT JOIN mix_values isg ON m.item_sub_group_id = isg.id
	// 	LEFT JOIN mix_values ig ON isg.parent_id = ig.id
	// 	LEFT JOIN item_units iu ON iu.id = m.item_unit_id
	// 	LEFT JOIN mix_values u ON iu.unit_id = u.id
	// 	LEFT JOIN users cu ON m.created_by_id = cu.id
	// 	LEFT JOIN users uu ON m.updated_by_id = uu.id
	// ) AS alias WHERE 1=1`

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

	if err := r.sqlDB.Get(&product, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return &product, nil
}

// BeginTransaction starts a new transaction
func (r *ProductRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

// Rollback all changes in the transaction
func (r *ProductRepository) Rollback() *gorm.DB {
	return r.db.Rollback()
}

func (r *ProductRepository) CreateProduct(tx *gorm.DB, product *models.Product, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ProductRepository-CreateProduct", opentracing.ChildOf(span.Context()))
	if err := tx.Create(product).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *ProductRepository) UpdateProduct(tx *gorm.DB, product *models.Product, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ProductRepository-UpdateProduct", opentracing.ChildOf(span.Context()))

	if err := tx.Select("*").Omit(
		"created_at", "created_by_id", "branch_id",
	).Updates(product).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *ProductRepository) DeleteProduct(tx *gorm.DB, params *dtos.GetProductParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ProductRepository-DeleteProduct", opentracing.ChildOf(span.Context()))

	if err := tx.Delete(&models.Product{}, params.ID).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil

}

func (s *ProductRepository) RestoreProduct(tx *gorm.DB, params *dtos.GetProductParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ProductRepository-RestoreProduct", opentracing.ChildOf(span.Context()))

	var product models.Product
	if err := tx.Unscoped().Model(&product).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *ProductRepository) CreateBoms(tx *gorm.DB, boms []*models.Bom, productID uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("BomRepository-CreateBoms", opentracing.ChildOf(span.Context()))

	result := tx.CreateInBatches(boms, len(boms))

	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return result.Error
	}

	return nil
}

// bulk/batch update boms
func (r *ProductRepository) UpdateBoms(tx *gorm.DB, boms []*models.Bom, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("BomRepository-UpdateBoms", opentracing.ChildOf(span.Context()))

	data := make([]map[string]interface{}, 0)
	for _, bom := range boms {
		data = append(data, map[string]interface{}{
			"id":              bom.ID,
			"product_id":      bom.ProductID,
			"product_item_id": bom.ProductItemID,
			"item_unit_id":    bom.ItemUnitID,
			"qty":             bom.Qty,
			"remark":          bom.Remark,
			"updated_by_id":   bom.UpdatedByID,
			"updated_at":      time.Now(),
		})
	}

	// if err := r.utilRepo.BulkUpdate(tx, "boms", "id", data, childSpan); err != nil {
	if err := r.utilRepo.Upsert(tx, "boms", "id", data, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *ProductRepository) DeleteBomsWhereNotIn(tx *gorm.DB, productID uint, bomIDs []uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("BomRepository-DeleteBomsWhereNotIn", opentracing.ChildOf(span.Context()))

	if err := tx.Where("product_id = ? AND id NOT IN ?", productID, bomIDs).Delete(&models.Bom{}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *ProductRepository) GetBomsByProductID(ctx *fiber.Ctx, productID uint, span opentracing.Span) ([]dtos.ProductBomListDTO, error) {
	childSpan := opentracing.StartSpan("ProductRepository-GetBomsByProductID", opentracing.ChildOf(span.Context()))

	boms := []dtos.ProductBomListDTO{}

	query := `SELECT b.id, b.product_id, b.product_item_id, b.item_unit_id, b.qty, b.remark, b.created_at, b.updated_at, b.deleted_at,
		b.id as bom_id,
		pi.name as product_item_name, 
		u.name as item_unit_name,

		isg.id as item_sub_group_id,
		ig.id as item_group_id,
		isg.name as item_sub_group_name,
		ig.name as item_group_name,

		cu.name as created_by_name,
		uu.name as updated_by_name

	FROM boms b
	LEFT JOIN products p ON b.product_id = p.id
	LEFT JOIN products pi ON b.product_item_id = pi.id
	LEFT JOIN item_units iu ON b.item_unit_id = iu.id
	LEFT JOIN mix_values u ON iu.unit_id = u.id
	LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
	LEFT JOIN mix_values ig ON isg.parent_id = ig.id
	LEFT JOIN users cu ON b.created_by_id = cu.id
	LEFT JOIN users uu ON b.updated_by_id = uu.id
	WHERE b.product_id = $1 AND b.deleted_at IS NULL`

	if err := r.sqlDB.SelectContext(ctx.Context(), &boms, query, productID); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return boms, nil
}

func (r *ProductRepository) CreateItemUnits(tx *gorm.DB, itemUnits []*models.ItemUnit, productID uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ProductRepository-CreateItemUnits", opentracing.ChildOf(span.Context()))
	// bulk insert

	result := tx.CreateInBatches(itemUnits, len(itemUnits))

	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return result.Error
	}

	return nil
}

// func (r *ProductRepository) GetItemUnitIDBySelectedItemID(ctx *fiber.Ctx, params *dtos.GetProductItemUnitParams, span opentracing.Span) (*dtos.ItemUnitDetailDTO, error) {
func (r *ProductRepository) GetItemUnitIDBySelectedItemID(ctx *fiber.Ctx, tx *gorm.DB, params *dtos.GetProductItemUnitParams, span opentracing.Span) (*dtos.ItemUnitDetailDTO, error) {
	childSpan := opentracing.StartSpan("ProductRepository-GetItemUnitIDBySelectedItemID", opentracing.ChildOf(span.Context()))
	var msItem dtos.ItemUnitDetailDTO

	query := ` SELECT *
		FROM (
			SELECT DISTINCT ON (m.id)
				m.id, m.product_id, m.unit_id

        FROM item_units m
				LEFT JOIN products mi ON m.product_id = mi.id
    ) AS alias WHERE 1=1`

	var args []interface{}

	i := 1
	query += " AND product_id = $1"
	args = append(args, params.ProductID)
	i++

	// unit_id
	query += " AND unit_id = $2"
	args = append(args, params.UnitID)
	i++

	// if err := r.sqlDB.Get(&msItem, query, args...); err != nil {
	if err := tx.Raw(query, args...).Scan(&msItem).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return &msItem, nil
}
