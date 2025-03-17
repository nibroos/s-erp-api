package repository

import (
	"fmt"
	"log"
	"strings"
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
	"gorm.io/gorm/clause"
)

type SalesOrderRepository struct {
	db       *gorm.DB
	sqlDB    *sqlx.DB
	utilRepo *UtilRepository
	tracer   opentracing.Tracer
}

func NewSalesOrderRepository(db *gorm.DB, sqlDB *sqlx.DB, utilRepo *UtilRepository, tracer opentracing.Tracer) *SalesOrderRepository {
	return &SalesOrderRepository{
		db:       db,
		sqlDB:    sqlDB,
		tracer:   tracer,
		utilRepo: utilRepo,
	}
}

func (r *SalesOrderRepository) GetSalesOrders(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.SalesOrderListDTO, int, error) {
	childSpan := opentracing.StartSpan("SalesOrderRepository-GetSalesOrders", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	products := []dtos.SalesOrderListDTO{}

	var total int

	filterDBColumnKey := []string{
		"q.sales_order_no", "q.title", "q.remark",
		"pi.name",
		"it.name",
		"sod.remark",
		"sod.gen_code",
		"sdb.remark",
		"sdb.gen_code",
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
		condition += fmt.Sprintf(" AND id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"status":        "q.status",
		"customer_id":   "q.customer_id",
		"order_type_id": "q.order_type_id",
		"currency_id":   "q.currency_id",
		"vat_id":        "q.vat_id",
		"payment_id":    "q.payment_id",
		"pph23_id":      "q.pph23_id",
		"expired_at":    "q.expired_at",
		"due_at":        "q.due_at",
		"is_approve":    "q.is_approve",
	}

	for key, _ := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", value, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":   "q.customer_id",
		"order_type_ids": "q.order_type_id",
		"currency_ids":   "q.currency_id",
		"payment_ids":    "q.payment_id",
		"pph23_ids":      "q.pph23_id",
	}

	for key, valueID := range filterIDsKey {
		if value, ok := filters[key]; ok && value != "" {
			// Split the string into an array of integers
			ids := strings.Split(value, ",")
			intIDs, err := utils.SplitStringArrayOfInts(ids)
			if err != nil {
				utils.LogErrors(childSpan, err)
				return nil, 0, err
			}

			condition += fmt.Sprintf(" AND %s = ANY($%d)", valueID, i)
			args = append(args, pq.Array(intIDs)) // Use pq.Array to pass the array to PostgreSQL
			i++
		}
	}

	filterIDsOrKey := map[string][]string{
		"vat_ids": []string{"q.vat_id", "sod.vat_id"},
	}

	for key, valueIDs := range filterIDsOrKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += " AND ("
			for idx, valueID := range valueIDs {
				if idx > 0 {
					condition += " OR"
				}
				condition += fmt.Sprintf(" %s IN ($%d)", valueID, i)
				args = append(args, value)
			}
			condition += ")"
		}
	}

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (q.id)
					q.id, q.customer_id, q.order_type_id, q.currency_id, q.vat_id, q.payment_id, q.pph23_id, q.branch_id, q.sales_order_no, q.title, q.remark, q.status, q.is_approved, q.exchange_rate, q.pph23_perc, q.total_qty, q.subtotal, q.total_discount, q.total_pph23, q.total_vat, q.grand_total, q.created_by_id, q.updated_by_id, q.deleted_by_id, q.created_at, q.updated_at, q.deleted_at,
					TO_CHAR(q.due_at, 'YYYY-MM-DD') as due_at,
					TO_CHAR(q.expired_at, 'YYYY-MM-DD') as expired_at,
					q.vat_perc, q.disc_am, q.disc_perc, q.disc_perc_am, q.disc_final, q.disc_type,

					pi.id as product_id,
					it.id as item_id,
					sod.vat_id as so_dt_vat_id,

					pi.name as product_name,
					it.name as item_name,
					cur.name as currency_name,
					vat.name as vat_name,
					pph.name as pph23_name,

					ot.name as order_type_name,
					c.name as customer_name,

					sod.remark as so_dt_remark,
					sod.gen_code as so_dt_gen_code,
					sdb.remark as so_dt_bom_remark,
					sdb.gen_code as so_dt_bom_gen_code,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM sales_orders q
				LEFT JOIN so_dts sod ON sod.sales_order_id = q.id
				LEFT JOIN products pi ON sod.item_id = pi.id
				LEFT JOIN item_units iu ON sod.item_unit_id = iu.id
				LEFT JOIN so_dt_boms sdb ON sdb.so_dt_id = sod.id
				LEFT JOIN products it ON sdb.item_id = it.id

				LEFT JOIN mix_values cur ON q.currency_id = cur.id
				LEFT JOIN mix_values vat ON q.vat_id = vat.id
				LEFT JOIN mix_values pph ON q.pph23_id = pph.id
				LEFT JOIN mix_values ot ON q.order_type_id = ot.id
				LEFT JOIN customers c ON q.customer_id = c.id

        LEFT JOIN users cu ON q.created_by_id = cu.id
        LEFT JOIN users uu ON q.updated_by_id = uu.id
				WHERE 1=1` + condition + queryGlobal + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery

	for key, value := range filters {
		switch key {
		case "sales_order_no", "title", "remark":
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "sales_order_no")
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

		err := r.sqlDB.SelectContext(ctx.Context(), &products, query, args...)
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

	return products, total, nil
}

func (r *SalesOrderRepository) GetSalesOrderByID(ctx *fiber.Ctx, params *dtos.GetSalesOrderParams, tx *gorm.DB, span opentracing.Span) (*dtos.SalesOrderDetailDTO, error) {
	childSpan := opentracing.StartSpan("SalesOrderRepository-GetSalesOrderByID", opentracing.ChildOf(span.Context()))
	var sales_order dtos.SalesOrderDetailDTO

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (q.id)
					q.id, q.customer_id, q.order_type_id, q.currency_id, q.vat_id, q.payment_id, q.pph23_id, q.branch_id, q.sales_order_no, q.title, q.remark, q.status, q.is_approved, q.exchange_rate, q.pph23_perc, q.total_qty, q.subtotal, q.total_discount, q.total_pph23, q.total_vat, q.grand_total, q.created_by_id, q.updated_by_id, q.deleted_by_id, q.created_at, q.updated_at, q.deleted_at,
					TO_CHAR(q.due_at, 'YYYY-MM-DD') as due_at,
					TO_CHAR(q.expired_at, 'YYYY-MM-DD') as expired_at,
					q.vat_perc, q.disc_am, q.disc_perc, q.disc_perc_am, q.disc_final, q.disc_type,
					-- q.sales_order_id,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM sales_orders q
				LEFT JOIN so_dts sod ON sod.sales_order_id = q.id
				LEFT JOIN products pi ON sod.item_id = pi.id
				LEFT JOIN item_units iu ON sod.item_unit_id = iu.id
				LEFT JOIN so_dt_boms sdb ON sdb.so_dt_id = sod.id
				LEFT JOIN products it ON sdb.item_id = it.id

				LEFT JOIN mix_values cur ON q.currency_id = cur.id
				LEFT JOIN mix_values vat ON q.vat_id = vat.id
				LEFT JOIN mix_values pph ON q.pph23_id = pph.id

        LEFT JOIN users cu ON q.created_by_id = cu.id
        LEFT JOIN users uu ON q.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	var args []interface{}

	i := 1
	query += " AND id = $1"
	args = append(args, params.ID)
	i++

	if !isAdmin && branchID != nil {
		query += fmt.Sprintf(" AND (branch_id = $%d)", i)
		args = append(args, branchID)
		i++
	}

	// if isAdmin && filters["branch_id"] != "" {
	// 	query += fmt.Sprintf(" AND (branch_id = $%d)", i)
	// 	args = append(args, filters["branch_id"])
	// 	i++
	// }

	isDeletedQuery := ` AND deleted_at IS NULL`
	if params.IsDeleted != nil && *params.IsDeleted == 1 {
		isDeletedQuery = " AND deleted_at IS NOT NULL"
	}

	query += isDeletedQuery

	if err := r.sqlDB.Get(&sales_order, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		childSpan.LogKV("query", query)
		return nil, err
	}

	return &sales_order, nil
}

// BeginTransaction starts a new transaction
func (r *SalesOrderRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

// Rollback all changes in the transaction
func (r *SalesOrderRepository) Rollback() *gorm.DB {
	return r.db.Rollback()
}

func (r *SalesOrderRepository) CreateSalesOrder(tx *gorm.DB, sales_order *models.SalesOrder, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesOrderRepository-CreateSalesOrder", opentracing.ChildOf(span.Context()))
	if err := tx.Create(sales_order).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *SalesOrderRepository) UpdateSalesOrder(tx *gorm.DB, sales_order *models.SalesOrder, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesOrderRepository-UpdateSalesOrder", opentracing.ChildOf(span.Context()))

	if err := tx.Select("*").Omit(
		"created_at", "created_by_id", "branch_id",
	).Updates(sales_order).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *SalesOrderRepository) DeleteSalesOrder(tx *gorm.DB, params *dtos.GetSalesOrderParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesOrderRepository-DeleteSalesOrder", opentracing.ChildOf(span.Context()))

	if err := tx.Delete(&models.SalesOrder{}, params.ID).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil

}

func (s *SalesOrderRepository) RestoreSalesOrder(tx *gorm.DB, params *dtos.GetSalesOrderParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesOrderRepository-RestoreSalesOrder", opentracing.ChildOf(span.Context()))

	var sales_order models.SalesOrder
	if err := tx.Unscoped().Model(&sales_order).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *SalesOrderRepository) CreateSoDts(tx *gorm.DB, soDts []models.SoDt, sales_orderID uint, span opentracing.Span) (*gorm.DB, []models.SoDt, error) {
	childSpan := opentracing.StartSpan("SoDtRepository-CreateSoDts", opentracing.ChildOf(span.Context()))

	tx = tx.Create(&soDts)
	if tx.Error != nil {
		utils.LogErrors(childSpan, tx.Error)
		tx.Rollback()
		log.Fatalf("Failed to insert soDts: %v", tx.Error)
	}

	return tx, soDts, nil
}

// bulk/batch update soDts
func (r *SalesOrderRepository) UpdateSoDts(tx *gorm.DB, soDts []models.SoDt, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SoDtRepository-UpdateSoDts", opentracing.ChildOf(span.Context()))

	data := make([]map[string]interface{}, 0)
	for _, soDt := range soDts {
		data = append(data, map[string]interface{}{
			"id":             soDt.ID,
			"product_uuid":   soDt.ProductUuid,
			"sales_order_id": soDt.SalesOrderID,
			"ref_id":         soDt.RefID,
			"vat_id":         soDt.VatID,
			"item_unit_id":   soDt.ItemUnitID,
			"item_id":        soDt.ItemID,
			"ref_type":       soDt.RefType,
			"item_type":      soDt.ItemType,
			// "ref_json":      soDt.RefJSON,
			// "item_json":     soDt.ItemJSON,
			"gen_code":      soDt.GenCode,
			"remark":        soDt.Remark,
			"vat_perc":      soDt.VatPerc,
			"vat_perc_am":   soDt.VatPercAm,
			"qty_so":        soDt.QtySO,
			"qty":           soDt.Qty,
			"price_sell":    soDt.PriceSell,
			"price_buy":     soDt.PriceBuy,
			"subtotal_sell": soDt.SubtotalSell,
			"subtotal_buy":  soDt.SubtotalBuy,
			"disc_am":       soDt.DiscAm,
			"disc_perc":     soDt.DiscPerc,
			"disc_perc_num": soDt.DiscPercNum,
			"disc_perc_am":  soDt.DiscPercAm,
			"disc_final":    soDt.DiscFinal,
			"disc_type":     soDt.DiscType,
			"total_am":      soDt.TotalAm,
			"updated_by_id": soDt.UpdatedByID,
			"updated_at":    time.Now(),
		})
	}

	// if err := r.utilRepo.BulkUpdate(tx, "soDts", "id", data, childSpan); err != nil {
	if err := r.utilRepo.Upsert(tx, "so_dts", "id", data, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *SalesOrderRepository) DeleteSoDtsWhereNotIn(ctx *fiber.Ctx, tx *gorm.DB, sales_orderID uint, soDtIDs []uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SoDtRepository-DeleteSoDtsWhereNotIn", opentracing.ChildOf(span.Context()))

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	now := time.Now()
	deletedAt := gorm.DeletedAt{Time: now, Valid: true}

	query := tx.Model(&models.SoDt{}).Where("sales_order_id = ? AND deleted_at IS NULL", sales_orderID)

	if len(soDtIDs) > 0 {
		query = query.Where("id NOT IN (?)", soDtIDs)
	}

	if err := query.Updates(map[string]interface{}{
		"deleted_by_id": userID,
		"deleted_at":    deletedAt,
	}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return tx, err
	}

	return tx, nil
}

// func (r *SalesOrderRepository) GetSoDtsBySalesOrderIDs(ctx *fiber.Ctx, sales_orderIDs uint, span opentracing.Span) ([]dtos.SalesOrderSoDtListDTO, error) {
func (r *SalesOrderRepository) GetSoDtsBySalesOrderIDs(ctx *fiber.Ctx, tx *gorm.DB, sales_orderIDs []uint, span opentracing.Span) ([]dtos.SalesOrderSoDtListDTO, error) {
	childSpan := opentracing.StartSpan("SalesOrderRepository-GetSoDtsBySalesOrderIDs", opentracing.ChildOf(span.Context()))

	soDts := []dtos.SalesOrderSoDtListDTO{}

	query := `SELECT sod.id, sod.sales_order_id, sod.product_uuid,
		sod.item_unit_id, sod.vat_id, sod.ref_id, sod.item_id, sod.ref_type, sod.item_type, sod.gen_code, sod.remark, sod.vat_perc, sod.qty_so, sod.qty, sod.price_sell, sod.price_buy, sod.subtotal_sell, sod.subtotal_buy, sod.vat_perc, sod.vat_perc_am, sod.disc_am, sod.disc_perc, sod.disc_perc_num, sod.disc_perc_am, sod.disc_final, sod.disc_type, sod.total_am, sod.created_by_id, sod.updated_by_id, sod.deleted_by_id, sod.created_at, sod.updated_at, sod.deleted_at,
		sod.created_at, sod.updated_at, sod.deleted_at,

		sod.id as so_dt_id,
		isg.id as item_sub_group_id,
		ig.id as item_group_id,
		isg.name as item_sub_group_name,
		ig.name as item_group_name,
		u.name as unit_name,
		pi.name as item_name,
		pi.code as item_code,

		cu.name as created_by_name,
		uu.name as updated_by_name

	FROM so_dts sod
	LEFT JOIN sales_orders p ON sod.sales_order_id = p.id
	LEFT JOIN products pi ON sod.item_id = pi.id
	LEFT JOIN item_units iu ON sod.item_unit_id = iu.id
	LEFT JOIN mix_values u ON iu.unit_id = u.id
	LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
	LEFT JOIN mix_values ig ON isg.parent_id = ig.id
	LEFT JOIN users cu ON sod.created_by_id = cu.id
	LEFT JOIN users uu ON sod.updated_by_id = uu.id
	WHERE sod.deleted_at IS NULL`

	var args []interface{}
	i := 1

	if len(sales_orderIDs) > 0 {
		query += " AND sod.sales_order_id = ANY($1)"
		args = append(args, pq.Array(sales_orderIDs))
		i++

	}

	// // Get the underlying *sql.DB from GORM transaction
	// sqlDB, err := tx.DB()
	// if err != nil {
	// 	utils.LogErrors(childSpan, err)
	// 	return nil, err
	// }

	// // Convert *sql.Tx to *sqlx.Tx using sqlx.NewTx
	// sqlxTx := sqlx.NewDb(sqlDB, "postgres")

	// // Use sqlx transaction to execute the query
	// if err := sqlxTx.SelectContext(ctx.Context(), &soDts, query, args...); err != nil {
	// 	utils.LogErrors(childSpan, err)
	// 	return nil, err
	// }

	if err := r.sqlDB.SelectContext(ctx.Context(), &soDts, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return soDts, nil
}

// func (r *SalesOrderRepository) GetSoDtsBySalesOrderIDs(ctx *fiber.Ctx, sales_orderIDs uint, span opentracing.Span) ([]dtos.SalesOrderSoDtListDTO, error) {
func (r *SalesOrderRepository) GetUpdatedSoDtsBySalesOrderIDs(ctx *fiber.Ctx, tx *gorm.DB, sales_orderIDs []uint, span opentracing.Span) ([]dtos.SalesOrderSoDtListUpdateDTO, error) {
	childSpan := opentracing.StartSpan("SalesOrderRepository-GetSoDtsBySalesOrderIDs", opentracing.ChildOf(span.Context()))

	soDts := []dtos.SalesOrderSoDtListUpdateDTO{}

	query := `SELECT sod.id, sod.sales_order_id, sod.product_uuid,
		sod.item_unit_id, sod.vat_id, sod.ref_id, sod.item_id, sod.ref_type, sod.item_type, sod.gen_code, sod.remark, sod.vat_perc, sod.qty_so, sod.qty, sod.price_sell, sod.price_buy, sod.subtotal_sell, sod.subtotal_buy, sod.vat_perc, sod.vat_perc_am, sod.disc_am, sod.disc_perc, sod.disc_perc_num, sod.disc_perc_am, sod.disc_final, sod.disc_type, sod.total_am, sod.created_by_id, sod.updated_by_id, sod.deleted_by_id, sod.created_at, sod.updated_at, sod.deleted_at,
		sod.created_at, sod.updated_at, sod.deleted_at,

		sod.id as so_dt_id,
		isg.id as item_sub_group_id,
		ig.id as item_group_id,
		isg.name as item_sub_group_name,
		ig.name as item_group_name,
		u.name as unit_name,
		pi.name as item_name,
		pi.code as item_code,

		cu.name as created_by_name,
		uu.name as updated_by_name

	FROM so_dts sod
	LEFT JOIN sales_orders p ON sod.sales_order_id = p.id
	LEFT JOIN products pi ON sod.item_id = pi.id
	LEFT JOIN item_units iu ON sod.item_unit_id = iu.id
	LEFT JOIN mix_values u ON iu.unit_id = u.id
	LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
	LEFT JOIN mix_values ig ON isg.parent_id = ig.id
	LEFT JOIN users cu ON sod.created_by_id = cu.id
	LEFT JOIN users uu ON sod.updated_by_id = uu.id
	WHERE sod.deleted_at IS NULL`

	var args []interface{}
	i := 1

	if len(sales_orderIDs) > 0 {
		query += " AND sod.sales_order_id = ANY($1)"
		args = append(args, pq.Array(sales_orderIDs))
		i++

	}

	// GORM Raw
	if err := tx.Raw(query, args...).Scan(&soDts).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return soDts, nil
}

func (r *SalesOrderRepository) CreateSoDtBoms(tx *gorm.DB, soDtBoms []map[string]interface{}, sales_orderID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SoDtRepository-CreateSoDtBoms", opentracing.ChildOf(span.Context()))

	err := r.utilRepo.Upsert(tx, "so_dt_boms", "id", soDtBoms, span)
	// err := tx.Create(&soDtBoms).Error
	if err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *SalesOrderRepository) DeleteSoDtBomsWhereNotIn(ctx *fiber.Ctx, tx *gorm.DB, sales_orderID uint, soDtBomIDs []uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SoDtRepository-DeleteSoDtBomsWhereNotIn", opentracing.ChildOf(span.Context()))

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	now := time.Now()
	deletedAt := gorm.DeletedAt{Time: now, Valid: true}

	query := tx.Model(&models.SoDtBom{}).Where("sales_order_id = ? AND deleted_at IS NULL", sales_orderID)

	if len(soDtBomIDs) > 0 {
		query = query.Where("id NOT IN (?)", soDtBomIDs)
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

// bulk/batch update soDts
func (r *SalesOrderRepository) UpdateSoDtBoms(tx *gorm.DB, soDtBoms []map[string]interface{}, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SoDtRepository-UpdateSoDtBoms", opentracing.ChildOf(span.Context()))

	// data := make([]map[string]interface{}, 0)
	// for _, soDtBom := range soDtBoms {
	// 	itemJson := "{}"
	// 	genCode := ""

	// 	data = append(data, map[string]interface{}{
	// 		"id":            soDtBom["id"],
	// 		"sales_order_id":  soDtBom["sales_order_id"],
	// 		"so_dt_id":     soDtBom["so_dt_id"],
	// 		"product_id":    soDtBom["product_id"],
	// 		"item_id":       soDtBom["item_id"],
	// 		"item_unit_id":  soDtBom["item_unit_id"],
	// 		"item_json":     itemJson,
	// 		"gen_code":      genCode,
	// 		"remark":        soDtBom["remark"],
	// 		"qty":           soDtBom["qty"],
	// 		"price_sell":    soDtBom["price_sell"],
	// 		"price_buy":     soDtBom["price_buy"],
	// 		"subtotal_sell": soDtBom["subtotal_sell"],
	// 		"subtotal_buy":  soDtBom["subtotal_buy"],
	// 		"updated_by_id": soDtBom["updated_by_id"],
	// 		"updated_at":    time.Now(),
	// 	})
	// }

	// log.Println("soDtBoms", soDtBoms)

	// if err := r.utilRepo.BulkUpdate(tx, "soDtBoms", "id", data, childSpan); err != nil {
	if err := r.utilRepo.Upsert(tx, "so_dt_boms", "id", soDtBoms, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *SalesOrderRepository) DeleteSoDtBomsBySalesOrderID(tx *gorm.DB, params *dtos.GetSalesOrderParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SoDtRepository-DeleteSoDtBomsBySalesOrderID", opentracing.ChildOf(span.Context()))

	if err := tx.Where("sales_order_id = ?", params.ID).Delete(&models.SoDtBom{}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *SalesOrderRepository) DeleteSoDtsBySalesOrderID(tx *gorm.DB, params *dtos.GetSalesOrderParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SoDtRepository-DeleteSoDtsBySalesOrderID", opentracing.ChildOf(span.Context()))

	if err := tx.Where("sales_order_id = ?", params.ID).Delete(&models.SoDt{}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *SalesOrderRepository) GetSoDtsBomBySalesOrders(ctx *fiber.Ctx, filters map[string]string, sales_orderIDs []uint, span opentracing.Span) ([]dtos.SalesOrderSoDtBomListDTO, error) {
	childSpan := opentracing.StartSpan("SalesOrderRepository-GetSoDtsBomBySalesOrders", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	soDtBoms := []dtos.SalesOrderSoDtBomListDTO{}

	filterDBColumnKey := []string{
		"q.sales_order_no", "q.title", "q.remark",
		"pi.name",
		"it.name",
		"sod.remark",
		"sod.gen_code",
		"sdb.remark",
		"sdb.gen_code",
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
		condition += fmt.Sprintf(" AND id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"status":        "q.status",
		"customer_id":   "q.customer_id",
		"order_type_id": "q.order_type_id",
		"currency_id":   "q.currency_id",
		"vat_id":        "q.vat_id",
		"payment_id":    "q.payment_id",
		"pph23_id":      "q.pph23_id",
		"expired_at":    "q.expired_at",
		"due_at":        "q.due_at",
		"is_approve":    "q.is_approve",
	}

	for key, _ := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", value, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":       "q.customer_id",
		"order_type_ids":     "q.order_type_id",
		"currency_ids":       "q.currency_id",
		"payment_ids":        "q.payment_id",
		"pph23_ids":          "q.pph23_id",
		"so_dt_ref_ids":      "sod.ref_id",
		"so_dt_bom_item_ids": "sdb.item_id",
	}

	for key, valueID := range filterIDsKey {
		if value, ok := filters[key]; ok && value != "" {
			// Split the string into an array of integers
			ids := strings.Split(value, ",")
			intIDs, err := utils.SplitStringArrayOfInts(ids)
			if err != nil {
				utils.LogErrors(childSpan, err)
				return nil, err
			}

			condition += fmt.Sprintf(" AND %s = ANY($%d)", valueID, i)
			args = append(args, pq.Array(intIDs)) // Use pq.Array to pass the array to PostgreSQL
			i++
		}
	}

	filterIDsOrKey := map[string][]string{
		"vat_ids": []string{"q.vat_id", "sod.vat_id"},
	}

	for key, valueIDs := range filterIDsOrKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += " AND ("
			for idx, valueID := range valueIDs {
				if idx > 0 {
					condition += " OR"
				}
				condition += fmt.Sprintf(" %s IN ($%d)", valueID, i)
				args = append(args, value)
			}
			condition += ")"
		}
	}

	if len(sales_orderIDs) > 0 {
		condition += fmt.Sprintf(" AND sdb.sales_order_id = ANY($%d)", i)
		// countQuery += fmt.Sprintf(" AND q.sales_order_id = ANY($%d)", i)
		args = append(args, pq.Array(sales_orderIDs))
		i++
	}

	filterKeyLike := map[string]string{
		"sales_order_no": "q.sales_order_no",
		"title":          "q.title",
		"remark":         "q.remark",
	}

	for key, _ := range filterKeyLike {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s ILIKE $%d", value, i)
			// countQuery += fmt.Sprintf(" AND %s ILIKE $%d", value, i)
			args = append(args, "%"+value+"%")
			i++
		}
	}

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (sdb.id)
					sdb.id, sdb.product_uuid, sdb.sales_order_id, sdb.so_dt_id, sdb.product_id, sdb.item_id, sdb.item_unit_id, sdb.gen_code, sdb.remark, sdb.qty, sdb.price_sell, sdb.price_buy, sdb.subtotal_sell, sdb.subtotal_buy, sdb.created_by_id, sdb.updated_by_id, sdb.deleted_by_id, sdb.created_at, sdb.updated_at, sdb.deleted_at,
					sdb.id as so_dt_bom_id,
					it.name as item_name,
					it.code as item_code,
					it.barcode as item_barcode,
					it.sku as item_sku,
					it.factory_code as item_factory_code,
					it.specification as item_specification,
					it.qty_stock as item_qty_stock,
					u.name as unit_name,

					isg.name as item_sub_group_name,
					ig.name as item_group_name,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM so_dt_boms sdb
				LEFT JOIN so_dts sod ON sod.id = sdb.so_dt_id
				LEFT JOIN sales_orders q ON q.id = sod.sales_order_id
				LEFT JOIN products pi ON sod.item_id = pi.id
				LEFT JOIN item_units iu ON sod.item_unit_id = iu.id
				LEFT JOIN mix_values u ON iu.unit_id = u.id
				LEFT JOIN products it ON sdb.item_id = it.id
				LEFT JOIN mix_values isg ON it.item_sub_group_id = isg.id
				LEFT JOIN mix_values ig ON isg.parent_id = ig.id

        LEFT JOIN users cu ON q.created_by_id = cu.id
        LEFT JOIN users uu ON q.updated_by_id = uu.id
				WHERE 1=1` + condition + queryGlobal + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	// countQuery := `SELECT COUNT(*) as total
	// 	` + baseQuery

	if !isAdmin && branchID != nil {
		query += fmt.Sprintf(" AND (branch_id = $%d)", i)
		// countQuery += fmt.Sprintf(" AND (branch_id = $%d)", i)
		args = append(args, branchID)
		i++
	}

	if isAdmin && filters["branch_id"] != "" {
		query += fmt.Sprintf(" AND (branch_id = $%d)", i)
		// countQuery += fmt.Sprintf(" AND (branch_id = $%d)", i)
		args = append(args, filters["branch_id"])
		i++
	}

	// orderColumn := utils.GetStringOrDefault(filters["order_column"], "sales_order_no")
	// orderDirection := utils.GetStringOrDefault(filters["order_direction"], "asc")
	// query += fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)

	// perPage := utils.GetIntOrDefault(filters["per_page"], 10)
	// currentPage := utils.GetIntOrDefault(filters["page"], 1)

	// if filters["is_csv"] != "1" {
	// 	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
	// 	args = append(args, perPage, (currentPage-1)*perPage)
	// }

	selectSpan := opentracing.StartSpan("SelectQuery", opentracing.ChildOf(childSpan.Context()))

	err := r.sqlDB.SelectContext(ctx.Context(), &soDtBoms, query, args...)
	if err != nil {
		selectSpan.LogKV("query", query)
		utils.LogErrors(selectSpan, err)
		return nil, err
	}

	return soDtBoms, nil
}

// Lock SalesOrder Header
func (r *SalesOrderRepository) LockSalesOrderHeader(ctx *fiber.Ctx, tx *gorm.DB, req dtos.UpdateSalesOrderRequest, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesOrderRepository-LockSalesOrderHeader", opentracing.ChildOf(span.Context()))

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", []uint{req.ID}).Find(&models.SalesOrder{}).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)

		return err
	}

	return nil
}

func (r *SalesOrderRepository) LockSoDts(ctx *fiber.Ctx, tx *gorm.DB, soDtIDs []*uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesOrderRepository-LockSoDts", opentracing.ChildOf(span.Context()))

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", soDtIDs).Find(&models.SoDt{}).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)

		return err
	}

	return nil
}

func (r *SalesOrderRepository) LockSoDtBoms(ctx *fiber.Ctx, tx *gorm.DB, soDtBomIDs []*uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesOrderRepository-LockSoDtBoms", opentracing.ChildOf(span.Context()))

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", soDtBomIDs).Find(&models.SoDtBom{}).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)

		return err
	}

	return nil
}
