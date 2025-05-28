package repository

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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
		COALESCE(bi.qty_stock, m.qty_stock) as qty_stock,
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
	// if product_bom_ids filled add cdSelect
	if filters["product_bom_ids"] != "" {
		cdSelect += `p2.name as product_bom_name,`
	}

	// "name", "code", "factory_code", "sku", "barcode", "specification", "description", "remark", "item_name", "item_code", "item_factory_code", "item_sku", "item_barcode", "item_specification", "item_description", "item_remark":
	filterDBColumnKey := []string{
		"m.name",
		"m.code",
		"m.factory_code",
		"m.sku",
		"m.barcode",
		"m.specification",
		"m.description",
		"m.remark",
		"m.tpb_code",
		"pi.name",
		"pi.code",
		"pi.factory_code",
		"pi.sku",
		"pi.barcode",
		"pi.specification",
		"pi.description",
		"pi.remark",
	}

	var args []interface{}

	queryGlobal := ""

	i := 1

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

	condition := ""

	if filters["ids"] != "" {
		condition += fmt.Sprintf(" AND m.id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"unit_id": "iu.unit_id",
		"status":  "m.status",
		// "prod_type":         "m.prod_type",
		"item_sub_group_id": "m.item_sub_group_id",
		"item_group_id":     "isg.parent_id",
	}

	for key, columnName := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", columnName, i)
			// countQuery += fmt.Sprintf(" AND %s = $%d", value, i)
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

	filterIDsKey := map[string]string{
		"item_sub_group_ids": "m.item_sub_group_id",
		"item_group_ids":     "isg.parent_id",
		"product_bom_ids":    "bo2.product_id",
	}

	for key, valueID := range filterIDsKey {
		if value, ok := filters[key]; ok && value != "" {
			// log.Println("product_bom_ids", value)
			condition += fmt.Sprintf(" AND %s IN (%s)", valueID, value)
		}
	}

	joinCondition := ""

	filterKeyJoin := map[string]string{
		// "is_task_exists":         "JOIN schedules s ON s.sales_order_id = so.id AND s.deleted_at IS NULL LEFT JOIN schedule_tasks stp ON stp.schedule_id = s.id AND stp.deleted_at IS NULL AND stp.entity_type = 'steps' LEFT JOIN schedule_tasks st ON st.parent_id = stp.id AND st.deleted_at IS NULL AND st.entity_type = 'tasks'",
		"product_bom_ids": "JOIN boms bo2 ON m.id = bo2.product_item_id AND bo2.deleted_at IS NULL JOIN products p2 ON bo2.product_id = p2.id AND p2.deleted_at IS NULL",
	}
	for key, join := range filterKeyJoin {
		if filters[key] != "" {
			joinCondition += fmt.Sprintf(" %s", join)
		}
	}

	customConditionKey := map[string]string{
		"group_type": ` AND ig.name ILIKE $%d `,
	}
	for key, custValue := range customConditionKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(custValue, i)
			args = append(args, value)
			i++
		}
	}

	query := `SELECT *
    FROM ( 
        SELECT DISTINCT ON (m.id)
					m.id, m.item_sub_group_id, isg.parent_id as item_group_id, m.item_unit_id, m.code, m.is_all_branch,
					` + cdSelect + `
					pi.name as item_name, pi.code as item_code, pi.factory_code as item_factory_code, pi.sku as item_sku, pi.barcode as item_barcode, pi.specification as item_specification, pi.description as item_description, pi.remark as item_remark, pi.tpb_code as item_tpb_code,

					m.id as product_id,
					m.id as ref_id,
					m.id as ref_product_id,
					m.is_pph23,
					m.is_vat,
					m.prod_type,
					u.name as unit_name,
					isg.name as item_sub_group_name,
					ig.name as item_group_name,
					'products' as ref_type,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM products m
				LEFT JOIN boms bo ON m.id = bo.product_id
				LEFT JOIN products pi ON bo.product_item_id = pi.id
				LEFT JOIN mix_values isg ON m.item_sub_group_id = isg.id
				LEFT JOIN mix_values ig ON isg.parent_id = ig.id
				LEFT JOIN item_units iu ON iu.id = m.item_unit_id
				LEFT JOIN mix_values u ON iu.unit_id = u.id
				LEFT JOIN branch_items bi ON bi.item_unit_id = iu.id
				LEFT JOIN branches b ON bi.branch_id = b.id
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
				` + joinCondition + `
				WHERE 1=1` + queryGlobal + condition + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT DISTINCT ON (m.id) 
					m.id, m.item_sub_group_id, isg.parent_id as item_group_id, m.item_unit_id, m.code, m.is_all_branch,
					` + cdSelect + `
					pi.name as item_name, pi.code as item_code, pi.factory_code as item_factory_code, pi.sku as item_sku, pi.barcode as item_barcode, pi.specification as item_specification, pi.description as item_description, pi.remark as item_remark, pi.tpb_code as item_tpb_code,

					m.id as product_id,
					m.id as ref_id,
					m.prod_type,
					u.name as unit_name,
					isg.name as item_sub_group_name,
					ig.name as item_group_name,
					'products' as ref_type,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM products m
				LEFT JOIN boms bo ON m.id = bo.product_id
				LEFT JOIN products pi ON bo.product_item_id = pi.id
				LEFT JOIN mix_values isg ON m.item_sub_group_id = isg.id
				LEFT JOIN mix_values ig ON isg.parent_id = ig.id
				LEFT JOIN item_units iu ON iu.id = m.item_unit_id AND iu.deleted_at IS NOT NULL
				LEFT JOIN mix_values u ON iu.unit_id = u.id
				LEFT JOIN branch_items bi ON bi.item_unit_id = iu.id
				LEFT JOIN branches b ON bi.branch_id = b.id
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
				` + joinCondition + `
				WHERE 1=1` + queryGlobal + condition + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	for key, value := range filters {
		switch key {
		case "name", "code", "factory_code", "sku", "barcode", "specification", "description", "remark", "tpb_code", "item_name", "item_code", "item_factory_code", "item_sku", "item_barcode", "item_specification", "item_description", "item_remark":
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "item_group_name")
	orderDirection := utils.GetStringOrDefault(filters["order_direction"], "desc")
	query += fmt.Sprintf(" ORDER BY item_group_name, item_sub_group_name, %s %s", orderColumn, orderDirection)

	// another order col & dir
	orderColumns := map[string]string{
		"updated_at": "desc",
		"created_at": "desc",
		"name":       "asc",
	}

	for key, value := range orderColumns {
		if filters[key] != "" {
			query += fmt.Sprintf(", %s %s", key, value)
		}
	}

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
		COALESCE(bi.qty_stock, iu.qty_stock) as qty_stock,
		COALESCE(iu.price_sell, 0) as price_sell,
		COALESCE(iu.price_buy, 0) as price_buy,
		COALESCE(bi.margin, iu.margin) as margin,
		COALESCE(bi.status, m.status) as status,
		TO_CHAR(COALESCE(bi.expired_at, m.expired_at), 'YYYY-MM-DD') as expired_at,
		COALESCE(bi.created_at, m.created_at) as created_at,
		COALESCE(bi.updated_at, m.updated_at) as updated_at,
		COALESCE(bi.deleted_at, m.deleted_at) as deleted_at,

		bi.id as branch_item_id,
		`
	} else {
		cdSelect = `
			m.name, m.factory_code, m.sku, m.barcode, m.specification, m.description, m.remark, m.tpb_code, m.minimum_stock, iu.price_sell, iu.price_buy, iu.margin, m.status, 
			TO_CHAR(m.expired_at, 'YYYY-MM-DD') as expired_at,
			m.created_at, m.updated_at, m.deleted_at,
		`
	}

	query := `SELECT *
    FROM ( 
        SELECT DISTINCT ON (m.id)
					m.id, m.item_sub_group_id, isg.parent_id as item_group_id, m.item_unit_id as ori_item_unit_id, m.code, m.is_all_branch, m.is_pph23, m.is_vat,
					` + cdSelect + `

					iu.unit_id as item_unit_id,
					m.id as product_id,
					m.prod_type,
					u.name as unit_name,
					b.name as branch_name,
					isg.name as item_sub_group_name,
					ig.name as item_group_name,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM products m
				LEFT JOIN mix_values isg ON m.item_sub_group_id = isg.id
				LEFT JOIN mix_values ig ON isg.parent_id = ig.id
				LEFT JOIN item_units iu ON iu.id = m.item_unit_id AND iu.deleted_at IS NULL
				LEFT JOIN mix_values u ON iu.unit_id = u.id
				LEFT JOIN branch_items bi ON bi.item_unit_id = iu.id AND bi.deleted_at IS NULL
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
		u.name as item_unit_name,
		pi.name,
		pi.code,
		b.product_item_id as ref_id,

		iu.price_sell,
		iu.price_buy,

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

func (r *ProductRepository) GetItemUnitsByProductID(ctx *fiber.Ctx, productID uint, span opentracing.Span) ([]dtos.ItemUnitListDTO, error) {
	childSpan := opentracing.StartSpan("ProductRepository-GetItemUnitsByProductID", opentracing.ChildOf(span.Context()))

	itemUnits := []dtos.ItemUnitListDTO{}

	query := `SELECT *
    FROM ( 
        SELECT DISTINCT ON (m.id)
					m.product_id, m.unit_id, m.conversion, m.price_sell, m.price_buy, m.margin, m.status, m.created_at, m.updated_at, m.deleted_at,
					p.name as product_name,
					u.name as unit_name,
					u.name,

					m.unit_id as id,
					m.id item_unit_id,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM item_units m
				LEFT JOIN products p ON m.product_id = p.id
				LEFT JOIN mix_values u ON m.unit_id = u.id
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL AND product_id = $1`

	if err := r.sqlDB.SelectContext(ctx.Context(), &itemUnits, query, productID); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return itemUnits, nil
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

func (r *ProductRepository) UpdateItemUnits(tx *gorm.DB, itemUnits []map[string]interface{}, productID uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ProductRepository-CreateItemUnits", opentracing.ChildOf(span.Context()))
	// bulk update

	if err := r.utilRepo.Upsert(tx, "item_units", "id", itemUnits, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *ProductRepository) DeleteItemUnitsWhereNotIn(ctx *fiber.Ctx, tx *gorm.DB, productID uint, itemUnitIDs []uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ProductRepository-DeleteItemUnitsWhereNotIn", opentracing.ChildOf(span.Context()))

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	now := time.Now()
	deletedAt := gorm.DeletedAt{Time: now, Valid: true}

	query := tx.Model(&models.ItemUnit{}).Where("product_id = ? AND deleted_at IS NULL", productID)

	if len(itemUnitIDs) > 0 {
		query = query.Where("id NOT IN (?)", itemUnitIDs)
	}

	if err := query.Updates(map[string]interface{}{
		"deleted_by_id": userID,
		"deleted_at":    deletedAt,
	}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
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
				WHERE m.deleted_at IS NULL
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

// GetBomsByProductIDs(ctx, productIDs, childSpan)
func (r *ProductRepository) GetBomsByProductIDs(ctx *fiber.Ctx, filters map[string]string, productIDs []uint, span opentracing.Span) ([]dtos.ProductBomListDTO, error) {
	childSpan := opentracing.StartSpan("ProductRepository-GetBomsByProductIDs", opentracing.ChildOf(span.Context()))

	boms := []dtos.ProductBomListDTO{}

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	// select column
	cdSelect := `p.name, p.specification, p.description, p.tpb_code,`
	if branchID != nil && !isAdmin {

		cdSelect = `
		COALESCE(bi.name, p.name) as name,
		COALESCE(bi.factory_code, p.factory_code) as factory_code,
		COALESCE(bi.sku, p.sku) as sku,
		COALESCE(bi.barcode, p.barcode) as barcode,
		COALESCE(bi.specification, p.specification) as specification,
		COALESCE(bi.description, p.description) as description,
		COALESCE(bi.tpb_code, p.tpb_code) as tpb_code,

		bi.id as branch_item_id,
		`
	} else {
		cdSelect = `
			p.name, p.factory_code, p.sku, p.barcode, p.specification, p.description, p.tpb_code,
		`
	}

	query := `SELECT * 
	FROM (
			SELECT DISTINCT ON (b.id)
			b.id, b.product_id, b.product_item_id, b.item_unit_id, b.qty, b.remark as remark, b.created_at, b.updated_at, b.deleted_at,
			b.id as bom_id,

			` + cdSelect + `
			p.code, 
			pi.name as item_name, pi.code as item_code, pi.factory_code as item_factory_code, pi.sku as item_sku, pi.barcode as item_barcode, pi.specification as item_specification, pi.description as item_description, pi.remark as item_remark, pi.tpb_code as item_tpb_code,

			iu.price_buy, iu.price_sell, iu.margin, 

			u.name as item_unit_name,

			isg.id as item_sub_group_id,
			ig.id as item_group_id,
			isg.name as item_sub_group_name,
			ig.name as item_group_name,
			br.name as branch_name,

			cu.name as created_by_name,
			uu.name as updated_by_name

		FROM boms b
		LEFT JOIN products p ON b.product_id = p.id
		LEFT JOIN products pi ON b.product_item_id = pi.id
		LEFT JOIN item_units iu ON b.item_unit_id = iu.id

		LEFT JOIN item_units p_item_unit ON p.item_unit_id = p_item_unit.id
		LEFT JOIN branch_items bi ON bi.item_unit_id = iu.id
		LEFT JOIN branches br ON bi.branch_id = b.id

		LEFT JOIN mix_values u ON iu.unit_id = u.id
		LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
		LEFT JOIN mix_values ig ON isg.parent_id = ig.id
		LEFT JOIN users cu ON b.created_by_id = cu.id
		LEFT JOIN users uu ON b.updated_by_id = uu.id
		WHERE b.deleted_at IS NULL
	) AS alias WHERE 1=1`

	var args []interface{}

	if len(productIDs) > 0 {
		query += " AND product_id IN (" + utils.JoinUintsToString(productIDs, ",") + ")"
	}

	i := 1
	for key, value := range filters {
		switch key {
		case "name", "code", "factory_code", "sku", "barcode", "specification", "description", "remark", "item_name", "item_code", "item_factory_code", "item_sku", "item_barcode", "item_specification", "item_description", "item_remark":
			if value != "" {
				query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

	if !isAdmin && branchID != nil {
		query += fmt.Sprintf(" AND (branch_id = $%d)", i)
		args = append(args, branchID)
		i++
	}

	if isAdmin && filters["branch_id"] != "" {
		query += fmt.Sprintf(" AND (branch_id = $%d)", i)
		args = append(args, filters["branch_id"])
		i++
	}

	filterKey := map[string]string{
		"unit_id": "unit_id",
		"status":  "status",
	}

	for key := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			query += fmt.Sprintf(" AND %s = $%d", value, i)
			args = append(args, value)
			i++
		}
	}

	if filters["ids"] != "" {
		query += fmt.Sprintf(" AND id IN (%s)", filters["ids"])
	}
	if value, ok := filters["global"]; ok && value != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR code ILIKE $%d OR factory_code ILIKE $%d OR sku ILIKE $%d OR barcode ILIKE $%d OR specification ILIKE $%d OR description ILIKE $%d OR remark ILIKE $%d OR item_name ILIKE $%d OR item_code ILIKE $%d OR item_factory_code ILIKE $%d OR item_sku ILIKE $%d OR item_barcode ILIKE $%d OR item_specification ILIKE $%d OR item_description ILIKE $%d OR item_remark ILIKE $%d)", i, i+1, i+2, i+3, i+4, i+5, i+6, i+7, i+8, i+9, i+10, i+11, i+12, i+13, i+14, i+15)
		for j := 0; j < 16; j++ {
			args = append(args, "%"+value+"%")
		}
		i += 16
	}
	orderColumn := utils.GetStringOrDefault(filters["order_column"], "name")
	orderDirection := utils.GetStringOrDefault(filters["order_direction"], "asc")
	query += fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)

	if err := r.sqlDB.SelectContext(ctx.Context(), &boms, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return boms, nil
}

func (r *ProductRepository) GetProductBom(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.ProductListDTO, int, error) {
	// Create a child span for the controller
	childSpan := opentracing.StartSpan("ProductRepository-GetProductBom", opentracing.ChildOf(span.Context()))

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
	// cdSelect := `mp.name, mp.specification, mp.description, mp.tpb_code, iu.price_sell, iu.price_buy, mp.minimum_stock,`
	// if branchID != nil && !isAdmin {

	// 	cdSelect = `
	// 	COALESCE(bi.name, mp.name) as name,
	// 	COALESCE(bi.factory_code, mp.factory_code) as factory_code,
	// 	COALESCE(bi.sku, mp.sku) as sku,
	// 	COALESCE(bi.barcode, mp.barcode) as barcode,
	// 	COALESCE(bi.specification, mp.specification) as specification,
	// 	COALESCE(bi.description, mp.description) as description,
	// 	COALESCE(bi.remark, mp.remark) as remark,
	// 	COALESCE(bi.tpb_code, mp.tpb_code) as tpb_code,
	// 	COALESCE(bi.qty_stock, mp.qty_stock) as qty_stock,
	// 	COALESCE(bi.minimum_stock, mp.minimum_stock) as minimum_stock,
	// 	COALESCE(bi.price_sell, iu.price_sell) as price_sell,
	// 	COALESCE(bi.price_buy, iu.price_buy) as price_buy,
	// 	COALESCE(bi.margin, iu.margin) as margin,
	// 	COALESCE(bi.status, mp.status) as status,
	// 	COALESCE(bi.expired_at, mp.expired_at) as expired_at,
	// 	COALESCE(bi.created_at, mp.created_at) as created_at,
	// 	COALESCE(bi.updated_at, mp.updated_at) as updated_at,
	// 	COALESCE(bi.deleted_at, mp.deleted_at) as deleted_at,

	// 	bi.id as branch_item_id,
	// 	`
	// } else {
	// 	cdSelect = `
	// 		mp.name, mp.factory_code, mp.sku, mp.barcode, mp.specification, mp.description, mp.remark, mp.tpb_code, mp.minimum_stock, iu.price_sell, iu.price_buy, iu.margin, mp.status, mp.expired_at, mp.created_at, mp.updated_at, mp.deleted_at,
	// 	`
	// }

	// "name", "code", "factory_code", "sku", "barcode", "specification", "description", "remark", "item_name", "item_code", "item_factory_code", "item_sku", "item_barcode", "item_specification", "item_description", "item_remark":
	filterDBColumnKey := []string{
		"mp.product_bom_name",
		"mp.name",
		"mp.code",
		"mp.factory_code",
		"mp.sku",
		"mp.barcode",
		"mp.specification",
		"mp.description",
		"mp.remark",
		"mp.tpb_code",
	}

	var args []interface{}

	queryGlobal := ""

	i := 1

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

	condition := ""
	conditionCombine := ""

	if filters["ids"] != "" {
		conditionCombine += fmt.Sprintf(" AND combined.id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"unit_id": "iu.unit_id",
		"status":  "combined.status",
		// "prod_type":         "m.prod_type",
		"item_sub_group_id": "m.item_sub_group_id",
		"item_group_id":     "isg.parent_id",
	}

	for key := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			conditionCombine += fmt.Sprintf(" AND %s = $%d", value, i)
			// countQuery += fmt.Sprintf(" AND %s = $%d", value, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"item_sub_group_ids": "combined.item_sub_group_id",
		"item_group_ids":     "ig.id",
		"product_bom_ids":    "combined.product_bom_id",
	}

	for key, valueID := range filterIDsKey {
		if value, ok := filters[key]; ok && value != "" {
			conditionCombine += fmt.Sprintf(" AND %s IN (%s)", valueID, value)
		}
	}

	cdSelect := ``
	// // if product_bom_ids filled add cdSelect
	// if filters["product_bom_ids"] != "" {
	// 	cdSelect += `p2.name as product_bom_name,`
	// }

	baseQuery := `
    FROM (
        SELECT DISTINCT ON (combined.id, combined.item_type) 
            combined.*,
            isg.parent_id as item_group_id,
            u.name as unit_name,
            isg.name as item_sub_group_name,
						COALESCE(bi.price_sell, iu.price_sell) as price_sell,
						COALESCE(bi.price_buy, iu.price_buy) as price_buy,
            ig.name as item_group_name,
            cu.name as created_by_name,
            uu.name as updated_by_name
        FROM (
            SELECT
                m.id, m.item_sub_group_id, m.item_unit_id,
                m.code, m.factory_code, m.name, m.sku, m.barcode, m.specification, 
                m.description, m.remark, m.tpb_code, m.minimum_stock, m.status, 
								TO_CHAR(m.expired_at, 'YYYY-MM-DD') as expired_at,
                m.id as product_id,
                m.id as ref_id,
								m.id as ref_product_id,
								NULL as ref_product_bom_id,
								NULL as product_bom_id,
                m.is_pph23,
                m.is_vat,
                m.created_by_id,
                m.updated_by_id,
                m.is_all_branch,
                m.created_at,
                m.updated_at,
                m.deleted_at,
                '' as product_bom_name,
                'item' as item_type
            FROM products m
						LEFT JOIN boms bo ON m.id = bo.product_item_id
            WHERE m.prod_type = 'single' AND m.deleted_at IS NULL

						UNION ALL

						SELECT
								b.id, i.item_sub_group_id, b.item_unit_id,
								i.code, i.factory_code, i.name, i.sku, i.barcode, i.specification,
								i.description, i.remark, i.tpb_code, i.minimum_stock, i.status, 
								TO_CHAR(i.expired_at, 'YYYY-MM-DD') as expired_at,
								i.id as product_id,
								b.id as ref_id,
								NULL as ref_product_id,
								b.id as ref_product_bom_id,
								b.product_id as product_bom_id,
								i.is_pph23,
								i.is_vat,
								i.created_by_id,
								i.updated_by_id,
								i.is_all_branch,
								i.created_at,
								i.updated_at,
								i.deleted_at,
								p.name as product_bom_name,
								'bom' as item_type
						FROM boms b
						LEFT JOIN products i ON b.product_item_id = i.id
						LEFT JOIN products p ON b.product_id = p.id
						WHERE b.deleted_at IS NULL
        ) combined
        LEFT JOIN mix_values isg ON combined.item_sub_group_id = isg.id
        LEFT JOIN mix_values ig ON isg.parent_id = ig.id
        LEFT JOIN item_units iu ON iu.id = combined.item_unit_id
        LEFT JOIN mix_values u ON iu.unit_id = u.id
        LEFT JOIN branch_items bi ON bi.item_unit_id = iu.id
        LEFT JOIN branches b ON bi.branch_id = b.id
        LEFT JOIN users cu ON combined.created_by_id = cu.id
        LEFT JOIN users uu ON combined.updated_by_id = uu.id
				WHERE combined.deleted_at IS NULL AND combined.status = 1 ` + conditionCombine + `
        ORDER BY combined.id, combined.item_type, combined.updated_at DESC
    ) AS mp`

	// UNION ALL

	// SELECT
	//     b.id, i.item_sub_group_id, b.item_unit_id,
	//     i.code, i.factory_code, i.name, i.sku, i.barcode, i.specification,
	//     i.description, i.remark, i.tpb_code, i.minimum_stock, i.status, i.expired_at,
	//     i.id as product_id,
	//     i.id as ref_id,
	//     i.is_pph23,
	//     i.is_vat,
	//     i.created_by_id,
	//     i.updated_by_id,
	//     i.is_all_branch,
	//     i.created_at,
	//     i.updated_at,
	//     i.deleted_at,
	//     p.name as product_bom_name,
	//     'bom' as item_type
	// FROM boms b
	// LEFT JOIN products i ON b.product_item_id = i.id
	// LEFT JOIN products p ON b.product_id = p.id
	// WHERE b.deleted_at IS NULL

	query := `SELECT 
    mp.*,
		` + cdSelect + `
    mp.item_group_id,
		mp.product_id as item_id,
    mp.unit_name,
    mp.item_sub_group_name,
    mp.item_group_name,
    'products' as ref_type,
    mp.created_by_name,
    mp.updated_by_name
    ` + baseQuery + `
    WHERE 1=1 AND mp.deleted_at IS NULL ` + queryGlobal + condition

	countQuery := `SELECT COUNT(*) ` + baseQuery + `
    WHERE 1=1 AND mp.deleted_at IS NULL ` + queryGlobal + condition
	for key, value := range filters {
		switch key {
		case "name", "code", "factory_code", "sku", "barcode", "specification", "description", "remark", "tpb_code", "item_name", "item_code", "item_factory_code", "item_sku", "item_barcode", "item_specification", "item_description", "item_remark":
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "updated_at")
	orderDirection := utils.GetStringOrDefault(filters["order_direction"], "desc")
	query += fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)

	// another order col & dir
	orderColumns := map[string]string{
		"updated_at": "desc",
		"created_at": "desc",
		"name":       "asc",
	}

	for key, value := range orderColumns {
		if filters[key] != "" {
			query += fmt.Sprintf(", %s %s", key, value)
		}
	}

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

// func (r *ProductRepository) GetItemUnitIDBySelectedItemID(ctx *fiber.Ctx, params *dtos.GetProductItemUnitParams, span opentracing.Span) (*dtos.ItemUnitDetailDTO, error) {
func (r *ProductRepository) GetItemUnitsByProductIDsAndUnitIDs(ctx *fiber.Ctx, tx *gorm.DB, params *dtos.GetProductItemUnitParams, unitIDs []uint, span opentracing.Span) ([]dtos.ItemUnitDetailDTO, error) {
	childSpan := opentracing.StartSpan("ProductRepository-GetItemUnitsByProductIDsAndUnitIDs", opentracing.ChildOf(span.Context()))
	var msItem []dtos.ItemUnitDetailDTO

	log.Println("ProductRepository-GetItemUnitsByProductIDsAndUnitIDs-query1", unitIDs)

	query := ` SELECT *
		FROM (
			SELECT DISTINCT ON (m.id)
				m.id, m.product_id, m.unit_id

        FROM item_units m
				LEFT JOIN products mi ON m.product_id = mi.id
				WHERE m.deleted_at IS NULL
    ) AS alias WHERE 1=1`

	var args []interface{}

	i := 1
	query += " AND product_id = $1"
	args = append(args, params.ProductID)
	i++

	log.Println("ProductRepository-GetItemUnitsByProductIDsAndUnitIDs-query2", unitIDs)

	// unit_id
	query += " AND unit_id = ANY($2)"
	args = append(args, pq.Array(unitIDs))
	i++

	log.Println("ProductRepository-GetItemUnitsByProductIDsAndUnitIDs-query", query)

	// if err := r.sqlDB.Get(&msItem, query, args...); err != nil {
	if err := tx.Raw(query, args...).Scan(&msItem).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return msItem, nil
}
