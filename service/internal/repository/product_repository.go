package repository

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/middleware"
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

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, 0, fiber.NewError(http.StatusUnauthorized, "Unauthorized")
	}
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	products := []dtos.ProductListDTO{}
	var total int

	// select column
	if branchID != nil && !isAdmin {
	} else {
	}

	query := `SELECT *
    FROM ( 
        SELECT DISTINCT ON (m.id)
					m.id, m.collection_id, m.unit_id, m.branch_id, m.code, m.factory_code, m.name, m.sku, m.barcode, m.specification, m.description, m.remark, m.price_sell, m.price_buy, m.margin, m.expired_at, m.status, m.created_at, m.updated_at, m.deleted_at,

					c.name as collection_name,
					u.name as unit_name,
					bi.branch_id as branch_id, 

        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM products m
				LEFT JOIN mix_values c ON m.collection_id = c.id
				LEFT JOIN mix_values u ON m.unit_id = u.id
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT DISTINCT ON (m.id) 
					m.id, m.collection_id, m.unit_id, m.branch_id, m.code, m.factory_code, m.name, m.sku, m.barcode, m.specification, m.description, m.remark, m.price_sell, m.price_buy, m.margin, m.expired_at, m.status, m.created_at, m.updated_at, m.deleted_at,

					c.name as collection_name,
					u.name as unit_name,
					bi.branch_id as branch_id,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM products m
				LEFT JOIN mix_values c ON m.collection_id = c.id
				LEFT JOIN mix_values u ON m.unit_id = u.id
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
		"collection_id": "collection_id",
		"unit_id":       "unit_id",
		"status":        "status",
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

	query := ` SELECT *
	FROM (
		
		SELECT DISTINCT ON (m.id) 
			m.id, m.collection_id, m.unit_id, m.branch_id, m.code, m.factory_code, m.name, m.sku, m.barcode, m.specification, m.description, m.remark, m.price_sell, m.price_buy, m.margin, m.expired_at, m.status, m.created_at, m.updated_at, m.deleted_at,
			c.name as collection_name,
			u.name as unit_name,
			b.name as branch_name,

			cu.name as created_by_name,
			uu.name as updated_by_name

		FROM products m
		LEFT JOIN mix_values c ON m.collection_id = isg.id
		LEFT JOIN mix_values u ON m.unit_id = u.id
		LEFT JOIN branches b ON m.branch_id = b.id
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
	defer childSpan.Finish()

	data := make([]map[string]interface{}, 0)
	for _, bom := range boms {
		data = append(data, map[string]interface{}{
			"id":            bom.ID,
			"ms_item_id":    bom.MsItemID,
			"item_unit_id":  bom.ItemUnitID,
			"qty":           bom.Qty,
			"remark":        bom.Remark,
			"updated_at":    time.Now(),
			"updated_by_id": bom.UpdatedByID,
		})
	}

	log.Println("data", data)

	if err := r.utilRepo.BulkUpdate(tx, "boms", "id", data, childSpan); err != nil {
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
