package repository

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/nibroos/s-erp-api/service/internal/auth"
	"github.com/nibroos/s-erp-api/service/internal/config"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InventoryRepository struct {
	db       *gorm.DB
	sqlDB    *sqlx.DB
	utilRepo *UtilRepository
	rabbitmq *config.RabbitMQ
	tracer   opentracing.Tracer
}

func NewInventoryRepository(db *gorm.DB, sqlDB *sqlx.DB, utilRepo *UtilRepository, rabbitmq *config.RabbitMQ, tracer opentracing.Tracer) *InventoryRepository {
	return &InventoryRepository{
		db:       db,
		sqlDB:    sqlDB,
		rabbitmq: rabbitmq,
		tracer:   tracer,
		utilRepo: utilRepo,
	}
}

func (r *InventoryRepository) GetInventories(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.InventoryListDTO, int, error) {
	childSpan := opentracing.StartSpan("InventoryRepository-GetInventories", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	products := []dtos.InventoryListDTO{}

	var total int

	filterDBColumnKey := []string{
		"iv.inventory_no", "iv.do_no", "iv.surat_jalan_no", "iv.invoice_no", "iv.remark", "iv.ship_dest",
		"pi.name",
		"so.po_buyer_no",
		"so.sales_order_no",
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
		condition += fmt.Sprintf(" AND iv.id IN (%s)", filters["ids"])
	}

	if filters["io_type"] != "" {
		condition += fmt.Sprintf(" AND ot.options_json->>'io_type' = '%s'", filters["io_type"])
	}

	filterKey := map[string]string{
		"status":          "iv.status",
		"customer_id":     "iv.customer_id",
		"io_type_id":      "iv.io_type_id",
		"currency_id":     "iv.currency_id",
		"vat_id":          "iv.vat_id",
		"payment_term_id": "iv.payment_term_id",
		"pph23_id":        "iv.pph23_id",
		"warehouse_id":    "iv.warehouse_id",
		"due_at":          "iv.due_at",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":     "iv.customer_id",
		"io_type_ids":      "iv.io_type_id",
		"currency_ids":     "iv.currency_id",
		"payment_term_ids": "iv.payment_term_id",
		"pph23_ids":        "iv.pph23_id",
		"warehouse_ids":    "iv.warehouse_id",
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
		"vat_ids": {"iv.vat_id", "sd.vat_id"},
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

	// And one for array conditions with OR
	filterIDsOrArrayKey := map[string][]string{
		"product_ids": {"ivd.ref_product_id", "ivd.item_id"},
	}

	// Handle array OR conditions
	for key, valueIDs := range filterIDsOrArrayKey {
		if value, ok := filters[key]; ok && value != "" {
			// Split the string into an array of integers
			ids := strings.Split(value, ",")
			intIDs, err := utils.SplitStringArrayOfInts(ids)
			if err != nil {
				utils.LogErrors(childSpan, err)
				return nil, 0, err
			}

			condition += " AND ("
			for idx, valueID := range valueIDs {
				if idx > 0 {
					condition += " OR"
				}
				condition += fmt.Sprintf(" %s = ANY($%d)", valueID, i)
				args = append(args, pq.Array(intIDs))
				i++
			}
			condition += ")"
		}
	}

	filterDBColumnLikeKey := map[string]string{
		"do_no":          "iv.do_no",
		"invoice_no":     "iv.invoice_no",
		"surat_jalan_no": "iv.surat_jalan_no",
		"inventory_no":   "iv.inventory_no",
		"po_buyer_no":    "so.po_buyer_no",
		"sales_order_no": "so.sales_order_no",
		"ship_dest":      "iv.ship_dest",
		"remark":         "iv.remark",
		"item_name":      "pi.name",
	}

	for key, value := range filters {
		if value != "" {
			for keyLike, column := range filterDBColumnLikeKey {
				if filters[key] != "" && key == keyLike {
					condition += fmt.Sprintf(" AND %s ILIKE $%d", column, i)
					args = append(args, "%"+value+"%")
					i++
				}
			}
		}
	}

	// if date_type, start_date, end_date filled
	if filters["date_type"] != "" && filters["start_date"] != "" && filters["end_date"] != "" {

		filterDateTypeKey := map[string]string{
			"ingoing_at":  "iv.ingoing_at",
			"invoice_at":  "iv.invoice_at",
			"do_at":       "iv.do_at",
			"shipping_at": "so.shipping_at",
			"expired_at":  "ivd.expired_at",
		}

		dateTypeColumn := filterDateTypeKey[filters["date_type"]]
		condition += fmt.Sprintf(" AND (%s BETWEEN $%d AND $%d)", dateTypeColumn, i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
	}

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (iv.id)
					iv.id, iv.customer_id, iv.io_type_id, iv.currency_id, iv.vat_id, iv.payment_term_id, iv.pph23_id, iv.warehouse_id, iv.branch_id,
					iv.inventory_no, iv.surat_jalan_no, iv.do_no, iv.invoice_no, iv.ship_dest, iv.remark, 
					iv.status, iv.exchange_rate, iv.pph23_perc, iv.total_qty, iv.subtotal, iv.total_pph23, iv.total_vat, iv.grand_total, iv.created_by_id, iv.updated_by_id, iv.deleted_by_id, iv.created_at, iv.updated_at, iv.deleted_at,
					TO_CHAR(iv.ingoing_at, 'YYYY-MM-DD') as ingoing_at,
					TO_CHAR(iv.do_at, 'YYYY-MM-DD') as do_at,
					TO_CHAR(iv.invoice_at, 'YYYY-MM-DD') as invoice_at,

					ot.name as io_type_name,
					c.name as customer_name,
					cur.name as currency_name,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM inventories iv
				LEFT JOIN inv_dts ivd ON ivd.inventory_id = iv.id
				LEFT JOIN products pi ON ivd.item_id = pi.id
				LEFT JOIN item_units iu ON ivd.item_unit_id = iu.id

				LEFT JOIN mix_values cur ON iv.currency_id = cur.id
				LEFT JOIN mix_values vat ON iv.vat_id = vat.id
				LEFT JOIN mix_values pph ON iv.pph23_id = pph.id
				LEFT JOIN mix_values ot ON iv.io_type_id = ot.id
				LEFT JOIN customers c ON iv.customer_id = c.id
				LEFT JOIN so_dts sd ON ivd.ref_so_dt_id = sd.id
				LEFT JOIN sales_orders so ON sd.sales_order_id = so.id

        LEFT JOIN users cu ON iv.created_by_id = cu.id
        LEFT JOIN users uu ON iv.updated_by_id = uu.id
				WHERE 1=1` + condition + queryGlobal + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery

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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "ingoing_at")
	orderDirection := utils.GetStringOrDefault(filters["order_direction"], "desc")
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

func (r *InventoryRepository) GetStocks(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.StockListDTO, int, error) {
	childSpan := opentracing.StartSpan("InventoryRepository-GetStocks", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	products := []dtos.StockListDTO{}

	var total int

	filterDBColumnKey := []string{
		"pi.name",
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
		condition += fmt.Sprintf(" AND iv.id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"item_id":      "st.item_id",
		"warehouse_id": "st.warehouse_id",
		"branch_id":    "st.branch_id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"warehouse_ids": "st.warehouse_id",
		"item_ids":      "st.item_id",
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
		"vat_ids": {"iv.vat_id", "sd.vat_id"},
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

	joinCondition := ""

	customCondition := ""
	filterKeyCustom := map[string]string{
		// "is_task_exists": " AND st.is_checked = 1",
	}
	for _, join := range filterKeyCustom {
		customCondition += join
	}

	// orderColumnKeys := map[string]string{
	// 	"id":             "st.id",
	// 	"item_id":        "st.item_id",
	// 	"item_name":      "pi.name",
	//
	// 	"warehouse_id":   "st.warehouse_id",
	// 	"branch_id":      "st.branch_id",
	// 	"qty":            "st.qty",
	// }

	// orderColumn := utils.GetStringOrDefault(filters["order_column"], "id")
	// orderDirection := utils.GetStringOrDefault(filters["order_direction"], "desc")

	// for key, value := range orderColumnKeys {
	// 	if key == orderColumn {
	// 		log.Println("orderColumn", orderColumn, key, value)
	// 		orderColumn = value
	// 		log.Println("orderColumn2", orderColumn, key, value)
	// 		break
	// 	}
	// 	log.Println("orderColumn0", orderColumn, key, value)
	// }

	// // orderQuery := fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)
	// orderQuery := fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (st.id)
					st.id, st.item_id, st.warehouse_id, st.branch_id,
					st.qty,
					st.created_at, st.updated_at, st.deleted_at,

					pi.name as item_name,
					w.name as warehouse_name,
					b.name as branch_name,
					u.name as unit_name

        FROM stocks st
				LEFT JOIN products pi ON st.item_id = pi.id
				LEFT JOIN item_units iu ON pi.item_unit_id = iu.id
				LEFT JOIN branches b ON b.id = st.branch_id
				LEFT JOIN mix_values u ON u.id = iu.unit_id
				LEFT JOIN mix_values w ON w.id = st.warehouse_id
				` + joinCondition + `
				WHERE 1=1` + condition + queryGlobal + customCondition + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery

	for key, value := range filters {
		switch key {
		case "po_buyer_no", "sales_order_no", "ship_dest", "remark":
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "id")
	orderDirection := utils.GetStringOrDefault(filters["order_direction"], "desc")
	query += fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)

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

func (r *InventoryRepository) GetInventoryByID(ctx *fiber.Ctx, params *dtos.GetInventoryParams, tx *gorm.DB, span opentracing.Span) (*dtos.InventoryDetailDTO, error) {
	childSpan := opentracing.StartSpan("InventoryRepository-GetInventoryByID", opentracing.ChildOf(span.Context()))
	var salesOrder dtos.InventoryDetailDTO

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	baseQuery := `
    FROM ( 
			SELECT DISTINCT ON (iv.id)
				iv.id, iv.rev_no, iv.customer_id, iv.io_type_id, iv.currency_id, iv.vat_id, iv.payment_term_id, iv.pph23_id, iv.warehouse_id, iv.branch_id,
				iv.inventory_no, iv.surat_jalan_no, iv.do_no, iv.invoice_no, iv.ship_dest, iv.remark, 
				iv.status, iv.exchange_rate, iv.pph23_perc, iv.total_qty, iv.subtotal, iv.total_pph23, iv.total_vat, iv.grand_total, iv.created_by_id, iv.updated_by_id, iv.deleted_by_id, iv.created_at, iv.updated_at, iv.deleted_at,
				TO_CHAR(iv.ingoing_at, 'YYYY-MM-DD') as ingoing_at,
				TO_CHAR(iv.do_at, 'YYYY-MM-DD') as do_at,
				TO_CHAR(iv.invoice_at, 'YYYY-MM-DD') as invoice_at,

				br.company_profile_id,
				c.name as customer_name,
				c.code as customer_code,
				c.phone as phone,
				c.address as address,
				cur.name as currency_name,
				vat.name as vat_name,
				pph.name as pph23_name,
				w.name as warehouse_name,

				-- io_type
				ot.options_json->>'io_type' as io_type,
				-- io_type_short
				CASE WHEN ot.options_json->>'io_type' = 'INVENTORY_IN' THEN 'IN' ELSE 'OUT' END as io_type_short,
				ivd.qty_out as qty_out_on_in,

				cu.name as created_by_name,
				uu.name as updated_by_name

			FROM inventories iv
			LEFT JOIN inv_dts ivd ON ivd.inventory_id = iv.id
			LEFT JOIN products pi ON ivd.item_id = pi.id
			LEFT JOIN item_units iu ON ivd.item_unit_id = iu.id

			LEFT JOIN mix_values cur ON iv.currency_id = cur.id
			LEFT JOIN mix_values vat ON iv.vat_id = vat.id
			LEFT JOIN mix_values pph ON iv.pph23_id = pph.id
			LEFT JOIN mix_values ot ON iv.io_type_id = ot.id
			LEFT JOIN mix_values w ON iv.warehouse_id = w.id
			LEFT JOIN customers c ON iv.customer_id = c.id
			LEFT JOIN branches br ON iv.branch_id = br.id

			LEFT JOIN users cu ON iv.created_by_id = cu.id
			LEFT JOIN users uu ON iv.updated_by_id = uu.id
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

	if err := tx.Raw(query, args...).Scan(&salesOrder).Error; err != nil {
		utils.LogErrors(childSpan, err)
		childSpan.LogKV("query", query)
		return nil, err
	}

	return &salesOrder, nil
}

// BeginTransaction starts a new transaction
func (r *InventoryRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

// Rollback all changes in the transaction
func (r *InventoryRepository) Rollback() *gorm.DB {
	return r.db.Rollback()
}

func (r *InventoryRepository) CreateInventory(tx *gorm.DB, salesOrder *models.Inventory, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InventoryRepository-CreateInventory", opentracing.ChildOf(span.Context()))
	if err := tx.Create(salesOrder).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *InventoryRepository) UpdateInventory(tx *gorm.DB, salesOrder *models.Inventory, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InventoryRepository-UpdateInventory", opentracing.ChildOf(span.Context()))

	if err := tx.Where("id = ?", salesOrder.ID).Select("*").Omit(
		"created_at", "created_by_id", "branch_id",
	).Updates(salesOrder).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *InventoryRepository) DeleteInventory(tx *gorm.DB, params *dtos.GetInventoryParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InventoryRepository-DeleteInventory", opentracing.ChildOf(span.Context()))

	if err := tx.Delete(&models.Inventory{}, params.ID).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil

}

func (s *InventoryRepository) RestoreInventory(tx *gorm.DB, params *dtos.GetInventoryParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InventoryRepository-RestoreInventory", opentracing.ChildOf(span.Context()))

	var salesOrder models.Inventory
	if err := tx.Unscoped().Model(&salesOrder).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *InventoryRepository) CreateInvDts(tx *gorm.DB, invDts []models.InvDt, salesOrderID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvDtRepository-CreateInvDts", opentracing.ChildOf(span.Context()))

	if err := tx.Model(&models.InvDt{}).Create(&invDts).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return tx, err
	}

	return tx, nil
}

// bulk/batch update invDts
func (r *InventoryRepository) UpdateInvDts(tx *gorm.DB, invDts []models.InvDt, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvDtRepository-UpdateInvDts", opentracing.ChildOf(span.Context()))

	data := make([]map[string]interface{}, 0)
	for _, invDt := range invDts {
		data = append(data, map[string]interface{}{
			"id":                 invDt.ID,
			"inventory_id":       invDt.InventoryID,
			"item_unit_id":       invDt.ItemUnitID,
			"vat_id":             invDt.VatID,
			"pph23_id":           invDt.Pph23ID,
			"ref_so_dt_id":       invDt.RefSoDtID,
			"ref_so_dt_bom_id":   invDt.RefSoDtBomID,
			"ref_po_dt_id":       invDt.RefPoDtID,
			"ref_po_dt_bom_id":   invDt.RefPoDtBomID,
			"ref_inv_dt_id":      invDt.RefInvDtID,
			"ref_product_id":     invDt.RefProductID,
			"ref_product_bom_id": invDt.RefProductBomID,
			"item_id":            invDt.ItemID,
			"product_uuid":       invDt.ProductUuid,
			"item_type":          invDt.ItemType,
			"ref_type":           invDt.RefType,
			// "ref_json":      invDt.RefJSON,
			// "item_json":     invDt.ItemJSON,
			"gen_code":      invDt.GenCode,
			"remark":        invDt.Remark,
			"vat_perc":      invDt.VatPerc,
			"vat_perc_am":   invDt.VatPercAm,
			"pph23_perc":    invDt.Pph23Perc,
			"pph23_perc_am": invDt.Pph23PercAm,
			"is_vat":        invDt.IsVat,
			"is_pph23":      invDt.IsPph23,
			"qty_out":       invDt.QtyOut,
			"qty":           invDt.Qty,
			"price_sell":    invDt.PriceSell,
			"price_buy":     invDt.PriceBuy,
			"subtotal_sell": invDt.SubtotalSell,
			"subtotal_buy":  invDt.SubtotalBuy,
			"total_am":      invDt.TotalAm,
			"expired_at":    invDt.ExpiredAt,
			"updated_by_id": invDt.UpdatedByID,
			"updated_at":    time.Now(),
		})
	}

	// if err := r.utilRepo.BulkUpdate(tx, "invDts", "id", data, childSpan); err != nil {
	if err := r.utilRepo.Upsert(tx, "inv_dts", "id", data, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *InventoryRepository) DeleteInvDtsWhereNotIn(ctx *fiber.Ctx, tx *gorm.DB, inventoryID uint, invDtIDs []uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvDtRepository-DeleteInvDtsWhereNotIn", opentracing.ChildOf(span.Context()))

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	now := time.Now()
	deletedAt := gorm.DeletedAt{Time: now, Valid: true}

	query := tx.Model(&models.InvDt{}).Where("inventory_id = ? AND deleted_at IS NULL", inventoryID)

	if len(invDtIDs) > 0 {
		query = query.Where("id NOT IN (?)", invDtIDs)
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

// func (r *InventoryRepository) GetInvDtsByInventoryIDs(ctx *fiber.Ctx, salesOrderIDs uint, span opentracing.Span) ([]dtos.InventoryInvDtListDTO, error) {
func (r *InventoryRepository) GetInvDtsByInventoryIDs(ctx *fiber.Ctx, tx *gorm.DB, inventoryIDs []uint, span opentracing.Span) ([]dtos.InventoryInvDtListDTO, error) {
	childSpan := opentracing.StartSpan("InventoryRepository-GetInvDtsByInventoryIDs", opentracing.ChildOf(span.Context()))

	invDts := []dtos.InventoryInvDtListDTO{}

	query := `SELECT ivd.id, ivd.inventory_id, ivd.product_uuid,
		ivd.item_unit_id, ivd.vat_id, ivd.item_id, ivd.ref_type, ivd.item_type, ivd.gen_code, ivd.remark, ivd.vat_perc, ivd.qty_out, ivd.qty_invoice, ivd.qty, ivd.price_sell, ivd.price_buy, ivd.subtotal_sell, ivd.subtotal_buy, ivd.total_am, ivd.created_by_id, ivd.updated_by_id, ivd.deleted_by_id, ivd.created_at, ivd.updated_at, ivd.deleted_at,
		ivd.vat_perc, ivd.vat_perc_am, ivd.pph23_perc, ivd.pph23_perc_am, ivd.is_vat, ivd.is_pph23,
		ivd.created_at, ivd.updated_at, ivd.deleted_at,
		TO_CHAR(ivd.expired_at, 'YYYY-MM-DD') as expired_at,

		ivd.ref_so_dt_id, ivd.ref_so_dt_bom_id, ivd.ref_ro_dt_id, ivd.ref_po_dt_id, ivd.ref_po_dt_bom_id, ivd.ref_inv_dt_id, ivd.ref_product_id, ivd.ref_product_bom_id,

		p.customer_id,

		ivd.id as inv_dt_id,
		isg.id as item_sub_group_id,
		ig.id as item_group_id,
		isg.name as item_sub_group_name,
		ig.name as item_group_name,
		u.name as unit_name,
		pi.name as item_name,
		pi.code as item_code,
		COALESCE(
		 sd.qty_out, sdb.qty_out, ivd_refs.qty_out, rd.qty_out, 0
		) as qty_out,
		COALESCE(
		 ivd.qty_out, 0
		) as qty_out_on_in,
		ivd.qty as qty_init,
		COALESCE(
		 pd.qty_in, 0
		) as qty_in,
		COALESCE(
		 sd.qty, sdb.qty, ivd_refs.qty, pd.qty, 0
		) as ref_qty,

		COALESCE(
			so.po_buyer_no, iv_refs.inventory_no, po.po_no, ro.request_no, NULL
		) as ref_num,

		cu.name as created_by_name,
		uu.name as updated_by_name

	FROM inv_dts ivd
	LEFT JOIN so_dts sd ON sd.id = ivd.ref_so_dt_id AND ivd.ref_type = 'so'
	LEFT JOIN so_dt_boms sdb ON sdb.id = ivd.ref_so_dt_bom_id AND ivd.ref_type = 'so'
	LEFT JOIN request_order_dts rd ON rd.id = ivd.ref_ro_dt_id
	LEFT JOIN request_orders ro ON ro.id = rd.request_order_id
	LEFT JOIN sales_orders so ON (so.id = sd.sales_order_id OR so.id = sdb.sales_order_id)
	LEFT JOIN purchase_order_dts pd ON pd.id               = ivd.ref_po_dt_id
	LEFT JOIN purchase_orders po ON po.id                  = pd.po_id
	LEFT JOIN inv_dts ivd_refs ON ivd_refs.id              = ivd.ref_inv_dt_id
	LEFT JOIN inventories iv_refs ON ivd_refs.inventory_id = iv_refs.id
	LEFT JOIN inventories p ON ivd.inventory_id            = p.id
	LEFT JOIN products pi ON ivd.item_id                   = pi.id
	LEFT JOIN item_units iu ON ivd.item_unit_id            = iu.id
	LEFT JOIN mix_values u ON iu.unit_id                   = u.id
	LEFT JOIN mix_values isg ON pi.item_sub_group_id       = isg.id
	LEFT JOIN mix_values ig ON isg.parent_id               = ig.id
	LEFT JOIN users cu ON ivd.created_by_id                = cu.id
	LEFT JOIN users uu ON ivd.updated_by_id                = uu.id
	WHERE ivd.deleted_at IS NULL`

	var args []interface{}
	i := 1

	if len(inventoryIDs) > 0 {
		query += " AND ivd.inventory_id = ANY($1)"
		args = append(args, pq.Array(inventoryIDs))
		i++
	}

	if err := r.sqlDB.SelectContext(ctx.Context(), &invDts, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return invDts, nil
}

func (r *InventoryRepository) DeleteInvDtsByInventoryID(tx *gorm.DB, params *dtos.GetInventoryParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InvDtRepository-DeleteInvDtsByInventoryID", opentracing.ChildOf(span.Context()))

	if err := tx.Where("inventory_id = ?", params.ID).Delete(&models.InvDt{}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

// Lock Inventory Header
func (r *InventoryRepository) LockInventoryHeader(ctx *fiber.Ctx, tx *gorm.DB, req dtos.FormInventoryRequest, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InventoryRepository-LockInventoryHeader", opentracing.ChildOf(span.Context()))

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", []uint{*req.ID}).Find(&models.Inventory{}).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)

		return err
	}

	return nil
}

func (r *InventoryRepository) LockInvDts(ctx *fiber.Ctx, tx *gorm.DB, invDtIDs []*uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InventoryRepository-LockInvDts", opentracing.ChildOf(span.Context()))

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", invDtIDs).Find(&models.InvDt{}).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)

		return err
	}

	return nil
}

func (r *InventoryRepository) GetRefIndexSoDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefInvIndexSoDtListDTO, int, error) {

	childSpan := opentracing.StartSpan("InventoryRepository-GetRefIndexSoDts", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	products := []dtos.RefInvIndexSoDtListDTO{}

	var total int

	filterDBColumnKey := []string{
		"so.po_buyer_no",
		"pi.name",
		"sd.remark",
		"sd.gen_code",
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
		condition += fmt.Sprintf(" AND so.id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"status":        "so.status",
		"customer_id":   "so.customer_id",
		"order_type_id": "so.order_type_id",
		"currency_id":   "so.currency_id",
		"vat_id":        "so.vat_id",
		"payment_id":    "so.payment_id",
		"pph23_id":      "so.pph23_id",
		"product_id":    "sd.item_id",
		"expired_at":    "so.expired_at",
		"due_at":        "so.due_at",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":       "so.customer_id",
		"order_type_ids":     "so.order_type_id",
		"currency_ids":       "so.currency_id",
		"payment_ids":        "so.payment_id",
		"pph23_ids":          "so.pph23_id",
		"product_ids":        "sd.item_id",
		"quotation_ids":      "so.id",
		"item_group_ids":     "ig.id",
		"item_sub_group_ids": "pi.item_sub_group_id",
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
		"vat_ids": {"so.vat_id", "sd.vat_id"},
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

	filterKeyLike := map[string]string{
		"po_buyer_no": "so.po_buyer_no",
	}

	for key, valColumn := range filterKeyLike {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s ILIKE $%d", valColumn, i)
			// countQuery += fmt.Sprintf(" AND %s ILIKE $%d", value, i)
			args = append(args, "%"+value+"%")
			i++
		}
	}

	baseQuery := `
    FROM ( 
					SELECT
						-- sd.id, 
						sd.id as ref_so_dt_id,
						null as ref_so_dt_bom_id,
						'item' as item_type,
						sd.sales_order_id, 
						sd.product_uuid,
						sd.item_id, 
						sd.item_unit_id, 
						sd.qty_out,
						sd.qty AS ref_qty, 
						sd.price_sell, 
						sd.price_buy, 
						sd.subtotal_sell, 
						sd.subtotal_buy,
						sd.gen_code, 
						sd.vat_perc,
						sd.vat_perc_am,
						sd.pph23_perc,
						sd.pph23_perc_am,
						sd.is_vat,
						sd.is_pph23,
						COALESCE(sd.qty, 0) * COALESCE(sd.price_buy, 0) as subtotal_buy, 
						COALESCE(sd.qty, 0) * COALESCE(sd.price_sell, 0) as subtotal_sell, 
						COALESCE(sd.qty, 0) - COALESCE(sd.qty_out, 0) as balance, 
						sd.remark
					FROM so_dts sd
					WHERE sd.item_type = 'item'

					UNION ALL

					SELECT 
						-- sdb.id,
						null as ref_so_dt_id,
						sdb.id as ref_so_dt_bom_id,
						'bom' as item_type,
						sdb.sales_order_id,
						sdb.product_uuid,
						sdb.item_id,
						sdb.item_unit_id,
						sdb.qty_out,
						sdb.qty AS ref_qty,
						sdb.price_sell,
						sdb.price_buy,
						sdb.subtotal_sell,
						sdb.subtotal_buy,
						sdb.gen_code,
						sd1.vat_perc,
						sd1.vat_perc_am,
						sd1.pph23_perc,
						sd1.pph23_perc_am,
						sd1.is_vat,
						sd1.is_pph23,
						COALESCE(sdb.qty, 0) * COALESCE(sdb.price_buy, 0) as subtotal_buy, 
						COALESCE(sdb.qty, 0) * COALESCE(sdb.price_sell, 0) as subtotal_sell, 
						COALESCE(sdb.qty, 0) - COALESCE(sdb.qty_out, 0) as balance, 
						sdb.remark
					FROM so_dt_boms sdb
					LEFT JOIN so_dts sd1 ON sd1.id = sdb.so_dt_id
					WHERE sdb.deleted_at IS NULL
    ) AS sd 
		LEFT JOIN sales_orders so ON sd.sales_order_id = so.id
		LEFT JOIN products pi ON sd.item_id = pi.id
		LEFT JOIN item_units iu ON sd.item_unit_id = iu.id
		LEFT JOIN customers c ON so.customer_id = c.id

		LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
		LEFT JOIN mix_values ig ON isg.parent_id = ig.id
		LEFT JOIN mix_values u ON iu.unit_id = u.id
		WHERE 1=1 AND COALESCE(sd.balance, 0) > 0`

	query := `SELECT sd.*,
			TO_CHAR(so.order_at, 'YYYY-MM-DD') as order_at,
			TO_CHAR(so.shipping_at, 'YYYY-MM-DD') as shipping_at,
			TO_CHAR(so.agree_at, 'YYYY-MM-DD') as agree_at,
			TO_CHAR(so.due_at, 'YYYY-MM-DD') as due_at,
			so.customer_id as customer_id,
			so.po_buyer_no as ref_num,
			isg.name as item_sub_group_name,
			ig.name as item_group_name,
			u.name as unit_name,
			c.name as customer_name,
			pi.name as item_name,
			pi.code as item_code,
			pi.sku as item_sku,
			'so' as ref_type
		` + baseQuery + condition + queryGlobal

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery + condition + queryGlobal

	if filters["date_type"] != "" && filters["start_date"] != "" && filters["end_date"] != "" {
		query += fmt.Sprintf(" AND (so.%s BETWEEN $%d AND $%d)", filters["date_type"], i, i+1)
		countQuery += fmt.Sprintf(" AND (so.%s BETWEEN $%d AND $%d)", filters["date_type"], i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
	}

	if !isAdmin && branchID != nil {
		query += fmt.Sprintf(" AND (so.branch_id = $%d)", i)
		countQuery += fmt.Sprintf(" AND (so.branch_id = $%d)", i)
		args = append(args, branchID)
		i++
	}

	if isAdmin && filters["branch_id"] != "" {
		query += fmt.Sprintf(" AND (so.branch_id = $%d)", i)
		countQuery += fmt.Sprintf(" AND (so.branch_id = $%d)", i)
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

	columnIDsKey := map[string]string{
		"ref_num":       "so.po_buyer_no",
		"order_at":      "so.order_at",
		"customer_name": "c.name",
		"item_code":     "pi.code",
		"item_name":     "pi.name",
		"item_sku":      "pi.sku",
		"unit_name":     "u.name",
		"ref_qty":       "sd.ref_qty",
		"balance":       "sd.balance",
		"qty_out":       "sd.qty_out",
		"remark":        "sd.remark",
	}

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "so.order_at")
	orderDirection := utils.GetStringOrDefault(filters["order_direction"], "desc")
	for key, valueID := range columnIDsKey {
		if value, ok := filters["order_column"]; ok && value != "" && key == value {
			orderColumn = valueID
		}
	}

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

func (r *InventoryRepository) GetRefIndexRoDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefInvIndexRoDtListDTO, int, error) {

	childSpan := opentracing.StartSpan("InventoryRepository-GetRefIndexRoDts", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	products := []dtos.RefInvIndexRoDtListDTO{}

	var total int

	filterDBColumnKey := []string{
		"ro.request_no",
		"pi.name",
		"sd.remark",
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
		condition += fmt.Sprintf(" AND ro.id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"status":        "ro.status",
		"warehouse_id":  "ro.warehouse_id",
		"order_type_id": "ro.order_type_id",
		"currency_id":   "ro.currency_id",
		"vat_id":        "ro.vat_id",
		"payment_id":    "ro.payment_id",
		"pph23_id":      "ro.pph23_id",
		"product_id":    "sd.item_id",
		"expired_at":    "ro.expired_at",
		"due_at":        "ro.due_at",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"warehouse_ids":      "ro.warehouse_id",
		"order_type_ids":     "ro.order_type_id",
		"currency_ids":       "ro.currency_id",
		"payment_ids":        "ro.payment_id",
		"pph23_ids":          "ro.pph23_id",
		"product_ids":        "sd.item_id",
		"quotation_ids":      "ro.id",
		"item_group_ids":     "ig.id",
		"item_sub_group_ids": "pi.item_sub_group_id",
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
		"vat_ids": {"ro.vat_id", "sd.vat_id"},
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

	filterKeyLike := map[string]string{
		"request_no": "ro.request_no",
	}

	for key, valColumn := range filterKeyLike {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s ILIKE $%d", valColumn, i)
			// countQuery += fmt.Sprintf(" AND %s ILIKE $%d", value, i)
			args = append(args, "%"+value+"%")
			i++
		}
	}

	baseQuery := `
    FROM ( 
					SELECT
						-- sd.id, 
						sd.id as ref_ro_dt_id,
						null as ref_ro_dt_bom_id,
						'item' as item_type,
						sd.request_order_id, 
						sd.product_uuid,
						sd.item_id, 
						sd.item_unit_id, 
						sd.qty_out,
						sd.req_qty AS ref_qty, 
						sd.price_sell,
						COALESCE(sd.req_qty, 0) * COALESCE(sd.price_sell, 0) as subtotal_sell, 
						COALESCE(sd.req_qty, 0) - COALESCE(sd.qty_out, 0) as balance, 
						sd.remark
					FROM request_order_dts sd
    ) AS sd 
		LEFT JOIN request_orders ro ON sd.request_order_id = ro.id
		LEFT JOIN products pi ON sd.item_id = pi.id
		LEFT JOIN item_units iu ON sd.item_unit_id = iu.id
		LEFT JOIN mix_values c ON ro.warehouse_id = c.id

		LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
		LEFT JOIN mix_values ig ON isg.parent_id = ig.id
		LEFT JOIN mix_values u ON iu.unit_id = u.id
		WHERE 1=1 AND COALESCE(sd.balance, 0) > 0`

	query := `SELECT sd.*,
			TO_CHAR(ro.request_date, 'YYYY-MM-DD') as request_date,
			ro.warehouse_id as warehouse_id,
			ro.request_no as ref_num,
			isg.name as item_sub_group_name,
			ig.name as item_group_name,
			u.name as unit_name,
			c.name as warehouse_name,
			pi.name as item_name,
			pi.code as item_code,
			pi.sku as item_sku,
			'ro' as ref_type
		` + baseQuery + condition + queryGlobal

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery + condition + queryGlobal

	if filters["start_date"] != "" && filters["end_date"] != "" {
		query += fmt.Sprintf(" AND (ro.request_date BETWEEN $%d AND $%d)", i, i+1)
		countQuery += fmt.Sprintf(" AND (ro.request_date BETWEEN $%d AND $%d)", i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
	}

	if !isAdmin && branchID != nil {
		query += fmt.Sprintf(" AND (ro.branch_id = $%d)", i)
		countQuery += fmt.Sprintf(" AND (ro.branch_id = $%d)", i)
		args = append(args, branchID)
		i++
	}

	if isAdmin && filters["branch_id"] != "" {
		query += fmt.Sprintf(" AND (ro.branch_id = $%d)", i)
		countQuery += fmt.Sprintf(" AND (ro.branch_id = $%d)", i)
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

	columnIDsKey := map[string]string{
		"ref_num":       "ro.po_buyer_no",
		"order_at":      "ro.order_at",
		"customer_name": "c.name",
		"item_code":     "pi.code",
		"item_name":     "pi.name",
		"item_sku":      "pi.sku",
		"unit_name":     "u.name",
		"ref_qty":       "sd.ref_qty",
		"balance":       "sd.balance",
		"qty_out":       "sd.qty_out",
		"remark":        "sd.remark",
	}

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "ro.order_at")
	orderDirection := utils.GetStringOrDefault(filters["order_direction"], "desc")
	for key, valueID := range columnIDsKey {
		if value, ok := filters["order_column"]; ok && value != "" && key == value {
			orderColumn = valueID
		}
	}

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

func (r *InventoryRepository) GetRefIndexPoDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefInvIndexPoDtListDTO, int, error) {

	childSpan := opentracing.StartSpan("InventoryRepository-GetRefIndexPoDts", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	products := []dtos.RefInvIndexPoDtListDTO{}

	var total int

	filterDBColumnKey := []string{
		"so.po_no",
		"pi.name",
		"sd.remark",
		"sd.gen_code",
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
		condition += fmt.Sprintf(" AND so.id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"status":        "so.status",
		"customer_id":   "so.customer_id",
		"order_type_id": "so.order_type_id",
		"currency_id":   "so.currency_id",
		"vat_id":        "so.vat_id",
		"payment_id":    "so.payment_id",
		"pph23_id":      "so.pph23_id",
		"product_id":    "sd.item_id",
		"expired_at":    "so.expired_at",
		"due_at":        "so.due_at",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":       "so.customer_id",
		"order_type_ids":     "so.order_type_id",
		"currency_ids":       "so.currency_id",
		"payment_ids":        "so.payment_id",
		"pph23_ids":          "so.pph23_id",
		"product_ids":        "sd.item_id",
		"quotation_ids":      "so.id",
		"item_group_ids":     "ig.id",
		"item_sub_group_ids": "pi.item_sub_group_id",
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
		"vat_ids": {"so.vat_id", "sd.vat_id"},
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

	filterKeyLike := map[string]string{
		"po_buyer_no": "so.po_buyer_no",
	}

	for key, valColumn := range filterKeyLike {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s ILIKE $%d", valColumn, i)
			// countQuery += fmt.Sprintf(" AND %s ILIKE $%d", value, i)
			args = append(args, "%"+value+"%")
			i++
		}
	}

	baseQuery := `
    FROM ( 
					SELECT DISTINCT ON (sd.id)
						-- sd.id, 
						sd.id as ref_po_dt_id,
						null as ref_po_dt_bom_id,
						'item' as item_type,
						sd.po_id, 
						sd.product_uuid,
						sd.product_id as item_id, 
						sd.item_unit_id, 
						sd.qty_in,
						sd.qty AS ref_qty,  
						sd.price as price_buy, 
						sd.subtotal as subtotal_buy,
						sd.gen_code, 
						sd.vat_perc,
						sd.vat_perc_am,
						sd.pph23_perc,
						sd.pph23_perc_am,
						sd.is_pph23,
						COALESCE(sd.qty, 0) * COALESCE(sd.price, 0) as subtotal_buy, 
						0 as subtotal_sell, 
						COALESCE(sd.qty, 0) - COALESCE(sd.qty_in, 0) as balance, 
						sd.remark
					FROM purchase_order_dts sd
					WHERE sd.deleted_at IS NULL
    ) AS sd 
		LEFT JOIN purchase_orders so ON sd.po_id = so.id
		LEFT JOIN products pi ON sd.item_id = pi.id
		LEFT JOIN item_units iu ON sd.item_unit_id = iu.id
		LEFT JOIN customers c ON so.customer_id = c.id

		LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
		LEFT JOIN mix_values ig ON isg.parent_id = ig.id
		LEFT JOIN mix_values u ON iu.unit_id = u.id
		WHERE 1=1 AND COALESCE(sd.balance, 0) > 0`

	query := `SELECT sd.*,
			TO_CHAR(so.po_date, 'YYYY-MM-DD') as po_date,
			TO_CHAR(so.delivery_date, 'YYYY-MM-DD') as delivery_date,
			so.customer_id as customer_id,
			so.vat_id as vat_id,
			so.pph23_id as pph23_id,
			so.is_vat as is_vat,
			so.currency_id as currency_id,
			so.payment_term_id as payment_term_id,
			so.exchange_rate as exchange_rate,
			so.po_no as ref_num,
			so.shipping_destination as ship_dest,
			isg.name as item_sub_group_name,
			ig.name as item_group_name,
			u.name as unit_name,
			c.name as customer_name,
			pi.name as item_name,
			pi.code as item_code,
			pi.sku as item_sku,
			'po' as ref_type
		` + baseQuery + condition + queryGlobal

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery + condition + queryGlobal

	if filters["date_type"] != "" && filters["start_date"] != "" && filters["end_date"] != "" {
		query += fmt.Sprintf(" AND (so.%s BETWEEN $%d AND $%d)", filters["date_type"], i, i+1)
		countQuery += fmt.Sprintf(" AND (so.%s BETWEEN $%d AND $%d)", filters["date_type"], i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
	}

	if !isAdmin && branchID != nil {
		query += fmt.Sprintf(" AND (so.branch_id = $%d)", i)
		countQuery += fmt.Sprintf(" AND (so.branch_id = $%d)", i)
		args = append(args, branchID)
		i++
	}

	if isAdmin && filters["branch_id"] != "" {
		query += fmt.Sprintf(" AND (so.branch_id = $%d)", i)
		countQuery += fmt.Sprintf(" AND (so.branch_id = $%d)", i)
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

	columnIDsKey := map[string]string{
		"ref_num": "so.po_buyer_no",
		// "order_at":      "so.order_at",
		"customer_name": "c.name",
		"item_code":     "pi.code",
		"item_name":     "pi.name",
		"item_sku":      "pi.sku",
		"unit_name":     "u.name",
		"ref_qty":       "sd.ref_qty",
		"balance":       "sd.balance",
		"qty_out":       "sd.qty_out",
		"remark":        "sd.remark",
	}

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "so.po_date")
	orderDirection := utils.GetStringOrDefault(filters["order_direction"], "desc")
	for key, valueID := range columnIDsKey {
		if value, ok := filters["order_column"]; ok && value != "" && key == value {
			orderColumn = valueID
		}
	}

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

func (r *InventoryRepository) GetRefSoDtsBomByQuoDtIDs(ctx *fiber.Ctx, filters map[string]string, quotationIDs []uint, span opentracing.Span) ([]dtos.RefInvIndexSoDtListDTO, error) {
	childSpan := opentracing.StartSpan("QuotationRepository-GetRefSoDtsBomByQuoDtIDs", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	quoDtBoms := []dtos.RefInvIndexSoDtListDTO{}

	filterDBColumnKey := []string{
		"q.quo_no", "q.title", "q.remark",
		"pi.name",
		"it.name",
		"sd.remark",
		"sd.gen_code",
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
		condition += fmt.Sprintf(" AND q.id IN (%s)", filters["ids"])
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
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
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
		"so_dt_ref_ids":      "sd.ref_id",
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
		"vat_ids": {"q.vat_id", "sd.vat_id"},
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

	if len(quotationIDs) > 0 {
		condition += fmt.Sprintf(" AND sdb.quotation_id = ANY($%d)", i)
		// countQuery += fmt.Sprintf(" AND q.quotation_id = ANY($%d)", i)
		args = append(args, pq.Array(quotationIDs))
		i++
	}

	filterKeyLike := map[string]string{
		"quo_no": "q.quo_no",
		"title":  "q.title",
		"remark": "q.remark",
	}

	for key, valColumn := range filterKeyLike {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s ILIKE $%d", valColumn, i)
			// countQuery += fmt.Sprintf(" AND %s ILIKE $%d", value, i)
			args = append(args, "%"+value+"%")
			i++
		}
	}

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (sdb.id)
					sdb.id, sdb.product_uuid, sdb.quotation_id, sdb.so_dt_id, sdb.product_id, sdb.item_id, sdb.item_unit_id, sdb.gen_code, sdb.remark, sdb.qty, sdb.price_sell, sdb.price_buy, sdb.subtotal_sell, sdb.subtotal_buy, sdb.created_by_id, sdb.updated_by_id, sdb.deleted_by_id, sdb.created_at, sdb.updated_at, sdb.deleted_at,
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
					ig.name as item_group_name

        FROM so_dt_boms sdb
				LEFT JOIN so_dts sd ON sd.id = sdb.so_dt_id
				LEFT JOIN quotations q ON q.id = sd.quotation_id
				LEFT JOIN products pi ON sd.item_id = pi.id
				LEFT JOIN item_units iu ON sd.item_unit_id = iu.id
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

	// orderColumn := utils.GetStringOrDefault(filters["order_column"], "quo_no")
	// orderDirection := utils.GetStringOrDefault(filters["order_direction"], "asc")
	// query += fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)

	// perPage := utils.GetIntOrDefault(filters["per_page"], 10)
	// currentPage := utils.GetIntOrDefault(filters["page"], 1)

	// if filters["is_csv"] != "1" {
	// 	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
	// 	args = append(args, perPage, (currentPage-1)*perPage)
	// }

	selectSpan := opentracing.StartSpan("SelectQuery", opentracing.ChildOf(childSpan.Context()))

	err := r.sqlDB.SelectContext(ctx.Context(), &quoDtBoms, query, args...)
	if err != nil {
		selectSpan.LogKV("query", query)
		utils.LogErrors(selectSpan, err)
		return nil, err
	}

	return quoDtBoms, nil
}

func (r *InventoryRepository) GetSoDtQtyUpdate(ctx *fiber.Ctx, tx *gorm.DB, filters map[string]string, span opentracing.Span) ([]dtos.GetInvSoDtQtyUpdateDTO, error) {

	childSpan := opentracing.StartSpan("InventoryRepository-GetSoDtQtyUpdate", opentracing.ChildOf(span.Context()))

	products := []dtos.GetInvSoDtQtyUpdateDTO{}

	var args []interface{}

	// i := 1
	condition := ""

	condition += fmt.Sprintf(" AND sd.id IN (%s)", filters["ids"])

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (sd.id)
					sd.id, sd.id as so_dt_id, sd.qty_out,
					sd.sales_order_id as sales_order_id

				FROM so_dts sd
				WHERE sd.deleted_at IS NULL` + condition + `
    ) AS alias WHERE 1=1`

	query := `SELECT *
		` + baseQuery

	var wg sync.WaitGroup
	var selectErr error

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

	if selectErr != nil {
		return nil, selectErr
	}

	return products, nil
}

func (r *InventoryRepository) GetSoDtBomQtyUpdate(ctx *fiber.Ctx, tx *gorm.DB, filters map[string]string, span opentracing.Span) ([]dtos.GetInvSoDtQtyUpdateDTO, error) {

	childSpan := opentracing.StartSpan("InventoryRepository-GetSoDtBomQtyUpdate", opentracing.ChildOf(span.Context()))

	products := []dtos.GetInvSoDtQtyUpdateDTO{}

	var args []interface{}

	// i := 1
	condition := ""

	condition += fmt.Sprintf(" AND sdb.id IN (%s)", filters["ids"])

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (sdb.id)
					sdb.id, sdb.id as so_dt_bom_id, sdb.qty_out,
					sdb.sales_order_id as sales_order_id

				FROM so_dt_boms sdb
				WHERE sdb.deleted_at IS NULL ` + condition + `
    ) AS alias WHERE 1=1`

	query := `SELECT *
		` + baseQuery

	var wg sync.WaitGroup
	var selectErr error

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

	if selectErr != nil {
		return nil, selectErr
	}

	return products, nil
}

func (r *InventoryRepository) BulkUpdateInvSoDtsQty(ctx *fiber.Ctx, tx *gorm.DB, soDts []map[string]interface{}, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InventoryRepository-BulkUpdateInvSoDtsQty", opentracing.ChildOf(span.Context()))

	if err := r.utilRepo.Upsert(tx, "so_dts", "id", soDts, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *InventoryRepository) BulkUpdateReverseInvRefDtsQty(ctx *fiber.Ctx, tx *gorm.DB, refDts []map[string]interface{}, tableName string, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InventoryRepository-BulkUpdateReverseInvRefDtsQty", opentracing.ChildOf(span.Context()))

	if err := r.utilRepo.Upsert(tx, tableName, "id", refDts, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *InventoryRepository) BulkUpdateInvSoDtBomsQty(ctx *fiber.Ctx, tx *gorm.DB, soDts []map[string]interface{}, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InventoryRepository-BulkUpdateInvSoDtBomsQty", opentracing.ChildOf(span.Context()))

	if err := r.utilRepo.Upsert(tx, "so_dt_boms", "id", soDts, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

// UpdateQuoStatus
func (r *InventoryRepository) UpdateSoStatus(ctx *fiber.Ctx, tx *gorm.DB, params dtos.UpdateInvSalesOrderStatusRequest, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InventoryRepository-UpdateQuoStatus", opentracing.ChildOf(span.Context()))

	updateParam := map[string]interface{}{
		"id":     params.ID,
		"status": params.Status,
	}

	// upsert
	if err := r.utilRepo.Upsert(tx, "sales_orders", "id", []map[string]interface{}{updateParam}, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

// GetSoID
func (r *InventoryRepository) GetSoIDByInventoryID(ctx *fiber.Ctx, tx *gorm.DB, params map[string]string, span opentracing.Span) (uint, error) {
	childSpan := opentracing.StartSpan("InventoryRepository-GetSoIDByInventoryID", opentracing.ChildOf(span.Context()))

	var quotationID uint

	baseQuery := `
		FROM (
			SELECT DISTINCT ON (q.id)
				q.id
			FROM quotations q
			LEFT JOIN so_dts sd ON q.id = sd.quotation_id
			LEFT JOIN so_dts sd ON sd.id = sd.ref_id AND sd.ref_type = 'quotations'
			LEFT JOIN sales_orders so ON sd.sales_order_id = so.id
			WHERE so.id = $1
		) AS alias WHERE 1=1`

	query := `SELECT *
		` + baseQuery

	err := tx.Raw(query, params["sales_order_id"]).Scan(&quotationID).Error
	if err != nil {
		utils.LogErrors(childSpan, err)
		return 0, err
	}

	return quotationID, nil
}

// GetCustomerInventoryCreatedThisMonth
func (r *InventoryRepository) GetCustomerInventoryCreatedThisMonth(ctx *fiber.Ctx, tx *gorm.DB, customerID uint, span opentracing.Span) (int, error) {
	childSpan := opentracing.StartSpan("InventoryRepository-GetCustomerInventoryCreatedThisMonth", opentracing.ChildOf(span.Context()))

	total := 0

	log.Println("customerID check")
	if customerID == 0 {
		log.Println("customerID is 0", customerID)
		return 0, nil
	}

	baseQuery := `
		FROM (
			SELECT COUNT(*) as total
			FROM inventories invs
			WHERE invs.customer_id = $1 AND invs.created_at >= date_trunc('month', CURRENT_DATE)
			AND invs.deleted_at IS NULL
		) AS alias WHERE 1=1`

	query := `SELECT *
		` + baseQuery

	err := tx.Raw(query, customerID).Scan(&total).Error
	if err != nil {
		utils.LogErrors(childSpan, err)
		return 0, err
	}

	return total, nil
}

// create/update stock
func (r *InventoryRepository) CreateOrUpdateStockOut(ctx *fiber.Ctx, tx *gorm.DB, req dtos.FormInventoryRequest, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InventoryRepository-CreateOrUpdateStockOut", opentracing.ChildOf(span.Context()))

	branchID := utils.GetDefaultBranchID(ctx)

	for _, invDt := range req.InvDts {
		// First try to find existing stock
		var stock models.Stock
		stockQuery := tx.Model(&models.Stock{}).Where(&models.Stock{
			WarehouseID: req.WarehouseID,
			ItemID:      invDt.ItemID,
			BranchID:    branchID,
		})

		result := stockQuery.First(&stock)

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// Create new stock if not found
				qty := -(*invDt.Qty)
				stock = models.Stock{
					WarehouseID: req.WarehouseID,
					ItemID:      invDt.ItemID,
					BranchID:    branchID,
					Qty:         &qty,
				}
				if err := tx.Model(&models.Stock{}).Create(&stock).Error; err != nil {
					utils.LogErrors(childSpan, err)
					return err
				}

				continue
				// return &newStock, nil
			}
			utils.LogErrors(childSpan, result.Error)
			return result.Error
		}

		// Update existing stock qty
		var qty float64
		if stock.Qty != nil {
			qty = *stock.Qty - *invDt.Qty
		} else {
			qty = *invDt.Qty
		}
		stock.Qty = &qty

		if err := tx.Model(&models.Stock{}).Where("id = ?", stock.ID).Save(&stock).Error; err != nil {
			utils.LogErrors(childSpan, err)
			return err
		}
	}

	return nil
}

// create/update stock
func (r *InventoryRepository) CreateOrUpdateStockIn(ctx *fiber.Ctx, tx *gorm.DB, req dtos.FormInventoryRequest, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InventoryRepository-CreateOrUpdateStockIn", opentracing.ChildOf(span.Context()))

	branchID := utils.GetDefaultBranchID(ctx)

	for _, invDt := range req.InvDts {
		// First try to find existing stock
		var stock models.Stock
		result := tx.Where(&models.Stock{
			WarehouseID: req.WarehouseID,
			ItemID:      invDt.ItemID,
			BranchID:    branchID,
		}).First(&stock)

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// Create new stock if not found
				qty := invDt.Qty
				stock := models.Stock{
					WarehouseID: req.WarehouseID,
					ItemID:      invDt.ItemID,
					BranchID:    branchID,
					Qty:         qty,
				}
				if err := tx.Create(&stock).Error; err != nil {
					utils.LogErrors(childSpan, err)
					return err
				}

				continue
				// return &newStock, nil
			} else {
				utils.LogErrors(childSpan, result.Error)
				return result.Error
			}
		}

		// Update existing stock qty
		var qty float64
		if stock.Qty != nil {
			qty = *stock.Qty + *invDt.Qty
		} else {
			qty = *invDt.Qty
		}
		stock.Qty = &qty

		if err := tx.Save(&stock).Error; err != nil {
			utils.LogErrors(childSpan, err)
			return err
		}
	}

	return nil
}

// create/update stock
func (r *InventoryRepository) ResetCreateOrUpdateStockOut(ctx *fiber.Ctx, tx *gorm.DB, req dtos.FormInventoryRequest, oldInvDts []dtos.InventoryInvDtListDTO, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InventoryRepository-ResetCreateOrUpdateStockOut", opentracing.ChildOf(span.Context()))

	branchID := utils.GetDefaultBranchID(ctx)

	for _, invDt := range oldInvDts {
		// First try to find existing stock
		var stock models.Stock
		result := tx.Where(&models.Stock{
			WarehouseID: req.WarehouseID,
			ItemID:      invDt.ItemID,
			BranchID:    branchID,
		}).First(&stock)

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// Create new stock if not found
				qty := -(*invDt.Qty)
				stock := models.Stock{
					WarehouseID: req.WarehouseID,
					ItemID:      invDt.ItemID,
					BranchID:    branchID,
					Qty:         &qty,
				}
				if err := tx.Create(&stock).Error; err != nil {
					utils.LogErrors(childSpan, err)
					return err
				}

				continue
				// return &newStock, nil
			} else {
				utils.LogErrors(childSpan, result.Error)
				return result.Error
			}
		}

		// Update existing stock qty
		var qty float64
		if stock.Qty != nil {
			qty = *stock.Qty + *invDt.Qty
		} else {
			qty = *invDt.Qty
		}
		stock.Qty = &qty

		if err := tx.Save(&stock).Error; err != nil {
			utils.LogErrors(childSpan, err)
			return err
		}
	}

	return nil
}

// create/update stock
func (r *InventoryRepository) ResetCreateOrUpdateStockIn(ctx *fiber.Ctx, tx *gorm.DB, req dtos.FormInventoryRequest, oldInvDts []dtos.InventoryInvDtListDTO, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InventoryRepository-ResetCreateOrUpdateStockIn", opentracing.ChildOf(span.Context()))

	branchID := utils.GetDefaultBranchID(ctx)

	for _, invDt := range oldInvDts {
		// First try to find existing stock
		var stock models.Stock
		result := tx.Where(&models.Stock{
			WarehouseID: req.WarehouseID,
			ItemID:      invDt.ItemID,
			BranchID:    branchID,
		}).First(&stock)

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// Create new stock if not found
				qty := invDt.Qty
				stock := models.Stock{
					WarehouseID: req.WarehouseID,
					ItemID:      invDt.ItemID,
					BranchID:    branchID,
					Qty:         qty,
				}
				if err := tx.Create(&stock).Error; err != nil {
					utils.LogErrors(childSpan, err)
					return err
				}

				continue
				// return &newStock, nil
			} else {
				utils.LogErrors(childSpan, result.Error)
				return result.Error
			}
		}

		// Update existing stock qty
		var qty float64
		if stock.Qty != nil {
			qty = *stock.Qty - *invDt.Qty
		} else {
			qty = *invDt.Qty
		}
		stock.Qty = &qty
		log.Println("ResetCreateOrUpdateStockIn-invDt", *invDt.Qty, *stock.Qty, qty)

		if err := tx.Save(&stock).Error; err != nil {
			utils.LogErrors(childSpan, err)
			return err
		}
	}

	return nil
}

func (r *InventoryRepository) GetRefOutDtByRefDtID(ctx *fiber.Ctx, tx *gorm.DB, tableName string, parentColumnName string, detailIDs []uint, span opentracing.Span) ([]map[string]interface{}, error) {
	childSpan := opentracing.StartSpan("InventoryRepository-GetRefOutDtBySoDtID", opentracing.ChildOf(span.Context()))

	var details []map[string]interface{}

	baseQuery := fmt.Sprintf(`
		FROM (
			SELECT DISTINCT ON (sd.id)
				sd.id, sd.qty_out, sd.%s
			FROM %s sd
		) AS alias WHERE 1=1`, parentColumnName, tableName)

	query := `SELECT *
		` + baseQuery

	var args []interface{}
	i := 1

	if len(detailIDs) > 0 {
		query += " AND id = ANY($1)"
		args = append(args, pq.Array(detailIDs))
		i++
	}

	log.Println("GetRefOutDtBySoDtID-query", query)
	log.Println("GetRefOutDtBySoDtID-parentColumnName", parentColumnName, tableName)
	log.Println("GetRefOutDtBySoDtID-detailIDs", detailIDs)

	err := tx.Raw(query, args...).Scan(&details).Error
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return details, nil
}

func (r *InventoryRepository) GetRefInDtByRefDtID(ctx *fiber.Ctx, tx *gorm.DB, tableName string, parentColumnName string, detailIDs []uint, span opentracing.Span) ([]map[string]interface{}, error) {
	childSpan := opentracing.StartSpan("InventoryRepository-GetRefOutDtBySoDtID", opentracing.ChildOf(span.Context()))

	var details []map[string]interface{}

	baseQuery := fmt.Sprintf(`
		FROM (
			SELECT DISTINCT ON (sd.id)
				sd.id, sd.qty_in, sd.%s
			FROM %s sd
		) AS alias WHERE 1=1`, parentColumnName, tableName)

	query := `SELECT *
		` + baseQuery

	var args []interface{}
	i := 1

	if len(detailIDs) > 0 {
		query += " AND id = ANY($1)"
		args = append(args, pq.Array(detailIDs))
		i++
	}

	log.Println("GetRefInDtByRefDtID-query", query)

	err := tx.Raw(query, args...).Scan(&details).Error
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return details, nil
}

func (r *InventoryRepository) GetRefInHeadByRefDtID(ctx *fiber.Ctx, tx *gorm.DB, tableName string, parentTableName string, parentColumnName string, detailIDs []uint, span opentracing.Span) ([]map[string]interface{}, error) {
	childSpan := opentracing.StartSpan("InventoryRepository-GetRefInHeadByRefDtID", opentracing.ChildOf(span.Context()))

	var details []map[string]interface{}

	baseQuery := fmt.Sprintf(`
    FROM (
        SELECT
            h.id, 
            SUM(sd.qty_in) as qty_in, 
            sd.%s, 
						MIN(sd.id) as dt_id,
            h.total_qty, 
            h.status
        FROM %s sd
        LEFT JOIN %s h ON sd.%s = h.id
        WHERE sd.deleted_at IS NULL
				GROUP BY h.id, sd.%s, h.total_qty, h.status
    ) AS alias WHERE 1=1`, parentColumnName, tableName, parentTableName, parentColumnName, parentColumnName)

	query := `SELECT *
		` + baseQuery

	var args []interface{}
	i := 1

	if len(detailIDs) > 0 {
		query += " AND dt_id = ANY($1)"
		args = append(args, pq.Array(detailIDs))
		i++
	}

	log.Println("GetRefInHeadByRefDtID-query", query)

	err := tx.Raw(query, args...).Scan(&details).Error
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return details, nil
}

func (r *InventoryRepository) GetRefIndexInvDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefInvIndexInvDtListDTO, int, error) {
	childSpan := opentracing.StartSpan("InventoryRepository-GetRefIndexInvDts", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	products := []dtos.RefInvIndexInvDtListDTO{}

	var total int

	filterDBColumnKey := []string{
		"iv.inventory_no", "iv.do_no", "iv.surat_jalan_no", "iv.invoice_no", "iv.remark", "iv.ship_dest",
		"pi.name",
		"so.po_buyer_no",
		"so.sales_order_no",
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
		condition += fmt.Sprintf(" AND iv.id IN (%s)", filters["ids"])
	}

	if filters["io_type"] != "" {
		condition += fmt.Sprintf(" AND ot.options_json->>'io_type' = '%s'", filters["io_type"])
	}

	filterKey := map[string]string{
		"status":          "iv.status",
		"customer_id":     "iv.customer_id",
		"io_type_id":      "iv.io_type_id",
		"currency_id":     "iv.currency_id",
		"vat_id":          "iv.vat_id",
		"payment_term_id": "iv.payment_term_id",
		"pph23_id":        "iv.pph23_id",
		"due_at":          "iv.due_at",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":     "iv.customer_id",
		"io_type_ids":      "iv.io_type_id",
		"currency_ids":     "iv.currency_id",
		"payment_term_ids": "iv.payment_term_id",
		"pph23_ids":        "iv.pph23_id",
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
		"vat_ids": {"iv.vat_id", "sd.vat_id"},
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

	joinCondition := ""

	if filters["is_schedule_not_exists"] == "1" {
		condition += " AND s.id IS NULL"
	}

	customCondition := ""
	filterKeyCustom := map[string]string{
		// "is_task_exists": " AND st.is_checked = 1",
	}
	for _, join := range filterKeyCustom {
		customCondition += join
	}

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (ivd.id)
					ivd.id, iv.branch_id,
					iv.inventory_no, iv.inventory_no as ref_num, iv.surat_jalan_no, iv.do_no, iv.invoice_no, 
					iv.created_at, iv.updated_at, iv.deleted_at,
					TO_CHAR(iv.ingoing_at, 'YYYY-MM-DD') as ingoing_at,
					TO_CHAR(iv.do_at, 'YYYY-MM-DD') as do_at,
					TO_CHAR(iv.invoice_at, 'YYYY-MM-DD') as invoice_at,

					ivd.id as inv_dt_id,
					ivd.item_id as item_id,
					ivd.item_unit_id as item_unit_id,
					ivd.id as ref_inv_dt_id,
					ivd.qty as ref_qty, 
					ivd.qty_out, 
					ivd.price_sell, 
					ivd.price_buy,
					ivd.subtotal_sell,
					ivd.subtotal_buy,

					ivd.vat_perc,
					ivd.vat_perc_am,
					ivd.vat_id,
					ivd.is_vat,
					ivd.pph23_perc,
					ivd.pph23_perc_am,
					ivd.pph23_id,
					ivd.is_pph23,
					COALESCE(ivd.qty - COALESCE(ivd.qty_out, 0), 0) as balance,

					ivd.expired_at,
					pi.name as item_name,
					pi.code as item_code,
					pi.sku as item_sku,

					u.name as unit_name,
					ot.name as io_type_name,
					isg.name as item_sub_group_name,
					ig.name as item_group_name,
					c.name as customer_name

        FROM inv_dts ivd
				LEFT JOIN inventories iv ON ivd.inventory_id = iv.id
				LEFT JOIN products pi ON ivd.item_id = pi.id
				LEFT JOIN item_units iu ON ivd.item_unit_id = iu.id

				LEFT JOIN mix_values cur ON iv.currency_id = cur.id
				LEFT JOIN mix_values vat ON iv.vat_id = vat.id
				LEFT JOIN mix_values pph ON iv.pph23_id = pph.id
				LEFT JOIN mix_values ot ON iv.io_type_id = ot.id
				LEFT JOIN mix_values u ON iu.unit_id = u.id
				LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
				LEFT JOIN mix_values ig ON isg.parent_id = ig.id
				LEFT JOIN customers c ON iv.customer_id = c.id
				LEFT JOIN so_dts sd ON ivd.ref_po_dt_id = sd.id
				LEFT JOIN sales_orders so ON sd.sales_order_id = so.id

        LEFT JOIN users cu ON iv.created_by_id = cu.id
        LEFT JOIN users uu ON iv.updated_by_id = uu.id
				` + joinCondition + `
				WHERE 1=1` + condition + queryGlobal + customCondition + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery

	for key, value := range filters {
		switch key {
		case "do_no", "invoice_no", "surat_jalan_no", "inventory_no":
			if value != "" {
				query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				countQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

	// if date_type, start_date, end_date filled
	if filters["date_type"] != "" && filters["start_date"] != "" && filters["end_date"] != "" {
		query += fmt.Sprintf(" AND (%s BETWEEN $%d AND $%d)", filters["date_type"], i, i+1)
		countQuery += fmt.Sprintf(" AND (%s BETWEEN $%d AND $%d)", filters["date_type"], i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "ingoing_at")
	orderDirection := utils.GetStringOrDefault(filters["order_direction"], "desc")
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

func (r *InventoryRepository) GetStockClosings(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.StockClosingListDTO, int, error) {
	childSpan := opentracing.StartSpan("InventoryRepository-GetStockClosings", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	products := []dtos.StockClosingListDTO{}
	var total int

	var args []interface{}
	queryGlobal := ""
	i := 1

	condition := ""

	// Handle date filters
	dateCondition := ""
	if endAt, ok := filters["end_at"]; ok && endAt != "" {
		if startAt, ok := filters["start_at"]; ok && startAt != "" {
			// Case: Both start_at and end_at are provided
			dateCondition = `
				WITH date_range AS (
						SELECT 
								$` + fmt.Sprint(i) + `::date AS start_date,
								$` + fmt.Sprint(i+1) + `::date AS end_date
				),
				all_products AS (
						SELECT p.id as item_id FROM products p WHERE p.deleted_at IS NULL
				),
				all_warehouses AS (
						SELECT mv.id as warehouse_id 
						FROM mix_values mv
						LEFT JOIN groups g ON mv.group_id = g.id
						WHERE g.name = 'warehouses'
						AND mv.deleted_at IS NULL
				),
				product_warehouse_combinations AS (
						SELECT ap.item_id, aw.warehouse_id
						FROM all_products ap
						CROSS JOIN all_warehouses aw
				),
				latest_closings AS (
						SELECT 
								sc.item_id,
								sc.warehouse_id,
								sc.closing_at,
								sc.end_qty AS begin_qty
						FROM stock_closings sc
						WHERE sc.closing_at = (
								SELECT MAX(sc2.closing_at)
								FROM stock_closings sc2
								WHERE sc2.closing_at < (SELECT start_date FROM date_range)
								AND sc2.item_id = sc.item_id
								AND sc2.warehouse_id = sc.warehouse_id
								AND sc2.deleted_at IS NULL
						)
						AND sc.deleted_at IS NULL
				),
				-- New CTE to calculate movements between last closing and start date
				gap_movements AS (
						SELECT 
								id.item_id,
								i.warehouse_id,
								SUM(CASE 
										WHEN iot.options_json->>'io_type' = 'INVENTORY_IN' THEN COALESCE(id.qty, 0)
										ELSE 0 
								END) AS gap_in_qty,
								SUM(CASE 
										WHEN iot.options_json->>'io_type' = 'INVENTORY_OUT' THEN COALESCE(id.qty, 0)
										ELSE 0 
								END) AS gap_out_qty
						FROM inv_dts id
						INNER JOIN inventories i ON id.inventory_id = i.id 
						INNER JOIN mix_values iot ON i.io_type_id = iot.id
						WHERE i.ingoing_at > (
								SELECT COALESCE(
										(SELECT MAX(sc.closing_at) 
										FROM stock_closings sc 
										WHERE sc.item_id = id.item_id 
										AND sc.warehouse_id = i.warehouse_id 
										AND sc.closing_at < (SELECT start_date FROM date_range)
										AND sc.deleted_at IS NULL
										),
										'1970-01-01'
								)
						)
						AND i.ingoing_at < (SELECT start_date FROM date_range)
						AND i.deleted_at IS NULL 
						AND id.deleted_at IS NULL
						GROUP BY id.item_id, i.warehouse_id
				),
				inventory_movements AS (
						SELECT 
								id.item_id,
								i.warehouse_id,
								SUM(CASE 
										WHEN iot.options_json->>'io_type' = 'INVENTORY_IN' THEN COALESCE(id.qty, 0)
										ELSE 0 
								END) AS in_qty,
								SUM(CASE 
										WHEN iot.options_json->>'io_type' = 'INVENTORY_OUT' THEN COALESCE(id.qty, 0)
										ELSE 0 
								END) AS out_qty,
								MAX(CASE 
										WHEN iot.options_json->>'io_type' = 'INVENTORY_IN' THEN id.price_buy
										ELSE NULL 
								END) AS price_buy,
								MAX(CASE 
										WHEN iot.options_json->>'io_type' = 'INVENTORY_IN' THEN id.price_sell
										ELSE NULL 
								END) AS price_sell
						FROM inv_dts id
						INNER JOIN inventories i ON id.inventory_id = i.id 
						INNER JOIN mix_values iot ON i.io_type_id = iot.id
						WHERE i.ingoing_at >= (SELECT start_date FROM date_range)
						AND i.ingoing_at <= (SELECT end_date FROM date_range)
						AND i.deleted_at IS NULL 
						AND id.deleted_at IS NULL
						GROUP BY id.item_id, i.warehouse_id
				),
				range_closings AS (
						SELECT 
								sc.item_id,
								sc.warehouse_id,
								sc.closing_at,
								sc.last_closing_at,
								sc.begin_qty,
								sc.in_qty,
								sc.out_qty,
								sc.adjustment_qty,
								sc.end_qty,
								sc.price_sell,
								sc.price_buy,
								sc.total_value_sell,
								sc.total_value_buy
						FROM stock_closings sc
						WHERE sc.closing_at = (SELECT end_date FROM date_range)
						AND sc.deleted_at IS NULL
				),
				combined_data AS (
						SELECT 
								pwc.item_id,
								pwc.warehouse_id,
								TO_CHAR((SELECT end_date FROM date_range), 'YYYY-MM-DD') AS closing_at,
								COALESCE(
										TO_CHAR(lc.closing_at, 'YYYY-MM-DD'), 
										TO_CHAR((SELECT start_date FROM date_range), 'YYYY-MM-DD')
								) AS last_closing_at,
								-- Calculate adjusted begin_qty by including gap movements
								COALESCE(lc.begin_qty, 0) + COALESCE(gm.gap_in_qty, 0) - COALESCE(gm.gap_out_qty, 0) AS begin_qty,
								COALESCE(im.in_qty, 0) AS in_qty,
								COALESCE(im.out_qty, 0) AS out_qty,
								COALESCE(rc.adjustment_qty, 0) AS adjustment_qty,
								-- Calculate end_qty including gap movements
								COALESCE(lc.begin_qty, 0) + COALESCE(gm.gap_in_qty, 0) - COALESCE(gm.gap_out_qty, 0) + 
										COALESCE(im.in_qty, 0) - COALESCE(im.out_qty, 0) + COALESCE(rc.adjustment_qty, 0) AS end_qty,
								COALESCE(rc.price_sell, im.price_sell) AS price_sell,
								COALESCE(rc.price_buy, im.price_buy) AS price_buy,
								(COALESCE(lc.begin_qty, 0) + COALESCE(gm.gap_in_qty, 0) - COALESCE(gm.gap_out_qty, 0) + 
										COALESCE(im.in_qty, 0) - COALESCE(im.out_qty, 0) + COALESCE(rc.adjustment_qty, 0)) * 
										COALESCE(rc.price_sell, im.price_sell) AS total_value_sell,
								(COALESCE(lc.begin_qty, 0) + COALESCE(gm.gap_in_qty, 0) - COALESCE(gm.gap_out_qty, 0) + 
										COALESCE(im.in_qty, 0) - COALESCE(im.out_qty, 0) + COALESCE(rc.adjustment_qty, 0)) * 
										COALESCE(rc.price_buy, im.price_buy) AS total_value_buy
						FROM product_warehouse_combinations pwc
						LEFT JOIN latest_closings lc ON pwc.item_id = lc.item_id AND pwc.warehouse_id = lc.warehouse_id
						LEFT JOIN gap_movements gm ON pwc.item_id = gm.item_id AND pwc.warehouse_id = gm.warehouse_id
						LEFT JOIN inventory_movements im ON pwc.item_id = im.item_id AND pwc.warehouse_id = im.warehouse_id
						LEFT JOIN range_closings rc ON pwc.item_id = rc.item_id AND pwc.warehouse_id = rc.warehouse_id
				)
				SELECT 
						cd.*,
						pi.name as item_name,
						w.name as warehouse_name,
						u.name as unit_name,
						cu.name as created_by_name,
						uu.name as updated_by_name
				FROM combined_data cd
				LEFT JOIN products pi ON cd.item_id = pi.id
				LEFT JOIN item_units iu ON pi.item_unit_id = iu.id
				LEFT JOIN mix_values u ON u.id = iu.unit_id
				LEFT JOIN mix_values w ON w.id = cd.warehouse_id
				LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
				LEFT JOIN mix_values ig ON isg.parent_id = ig.id
				LEFT JOIN users cu ON 1 = cu.id 
				LEFT JOIN users uu ON 1 = uu.id 
				WHERE 1=1`
			args = append(args, startAt, endAt)
			i += 2
		} else {
			// Case: Only end_at is provided
			dateCondition = `
                WITH end_date AS (
                    SELECT $` + fmt.Sprint(i) + `::date AS closing_date
                ),
                all_products AS (
                    SELECT p.id as item_id FROM products p WHERE p.deleted_at IS NULL
                ),
                all_warehouses AS (
                    SELECT mv.id as warehouse_id 
                    FROM mix_values mv
                    LEFT JOIN groups g ON mv.group_id = g.id
                    WHERE g.name = 'warehouses'
                    AND mv.deleted_at IS NULL
                ),
                product_warehouse_combinations AS (
                    SELECT ap.item_id, aw.warehouse_id
                    FROM all_products ap
                    CROSS JOIN all_warehouses aw
                ),
                existing_closings AS (
                    SELECT 
                        st.id, st.item_id, st.warehouse_id,
                        TO_CHAR(st.closing_at, 'YYYY-MM-DD') as closing_at,
                        TO_CHAR(st.last_closing_at, 'YYYY-MM-DD') as last_closing_at,
                        st.begin_qty, st.in_qty, st.out_qty, st.adjustment_qty, st.end_qty,
                        st.price_sell, st.price_buy,
                        st.total_value_sell, st.total_value_buy,
                        st.created_at, st.updated_at, st.deleted_at
                    FROM stock_closings st
                    WHERE st.closing_at = (SELECT closing_date FROM end_date)
                    AND st.deleted_at IS NULL
                ),
                latest_closings AS (
                    SELECT 
                        sc.item_id,
                        sc.warehouse_id,
                        sc.closing_at as last_closing_at,
                        sc.end_qty as begin_qty
                    FROM stock_closings sc
                    WHERE sc.closing_at = (
                        SELECT MAX(sc2.closing_at)
                        FROM stock_closings sc2
                        WHERE sc2.closing_at < (SELECT closing_date FROM end_date)
                        AND sc2.item_id = sc.item_id
                        AND sc2.warehouse_id = sc.warehouse_id
                        AND sc2.deleted_at IS NULL
                    )
                    AND sc.deleted_at IS NULL
                ),
                inventory_movements AS (
                    SELECT 
                        id.item_id,
                        i.warehouse_id,
                        SUM(CASE 
                            WHEN iot.options_json->>'io_type' = 'INVENTORY_IN' THEN COALESCE(id.qty, 0)
                            ELSE 0 
                        END) as in_qty,
                        SUM(CASE 
                            WHEN iot.options_json->>'io_type' = 'INVENTORY_OUT' THEN COALESCE(id.qty, 0)
                            ELSE 0 
                        END) as out_qty,
                        MAX(id.price_buy) as price_buy,
                        MAX(id.price_sell) as price_sell
                    FROM inv_dts id
                    INNER JOIN inventories i ON id.inventory_id = i.id 
                    INNER JOIN mix_values iot ON i.io_type_id = iot.id
                    WHERE i.ingoing_at > (
                        SELECT COALESCE(
                            (SELECT MAX(sc.closing_at) 
                            FROM stock_closings sc 
                            WHERE sc.item_id = id.item_id 
                            AND sc.warehouse_id = i.warehouse_id 
                            AND sc.closing_at < (SELECT closing_date FROM end_date)
                            AND sc.deleted_at IS NULL
                            ),
                            '1970-01-01'
                        )
                    )
                    AND i.ingoing_at <= (SELECT closing_date FROM end_date)
                    AND i.deleted_at IS NULL 
                    AND id.deleted_at IS NULL
                    GROUP BY id.item_id, i.warehouse_id
                ),
                combined_data AS (
										SELECT 
												COALESCE(ec.id, 0) as id,
												pwc.item_id,
												pwc.warehouse_id,
												COALESCE(
														ec.closing_at, 
														TO_CHAR((SELECT closing_date FROM end_date), 'YYYY-MM-DD')
												) as closing_at,
												COALESCE(
														TO_CHAR(lc.last_closing_at, 'YYYY-MM-DD'),
														'1970-01-01'
												) as last_closing_at,
                        COALESCE(lc.begin_qty, 0) as begin_qty,
                        COALESCE(im.in_qty, 0) as in_qty,
                        COALESCE(im.out_qty, 0) as out_qty,
                        COALESCE(ec.adjustment_qty, 0) as adjustment_qty,
                        COALESCE(lc.begin_qty, 0) + COALESCE(im.in_qty, 0) - COALESCE(im.out_qty, 0) + COALESCE(ec.adjustment_qty, 0) as end_qty,
                        COALESCE(ec.price_sell, im.price_sell) as price_sell,
                        COALESCE(ec.price_buy, im.price_buy) as price_buy,
                        (COALESCE(lc.begin_qty, 0) + COALESCE(im.in_qty, 0) - COALESCE(im.out_qty, 0) + COALESCE(ec.adjustment_qty, 0)) * 
                            COALESCE(ec.price_sell, im.price_sell) as total_value_sell,
                        (COALESCE(lc.begin_qty, 0) + COALESCE(im.in_qty, 0) - COALESCE(im.out_qty, 0) + COALESCE(ec.adjustment_qty, 0)) * 
                            COALESCE(ec.price_buy, im.price_buy) as total_value_buy,
                        CURRENT_TIMESTAMP as created_at,
                        NULL as updated_at,
                        NULL as deleted_at
                    FROM product_warehouse_combinations pwc
                    LEFT JOIN existing_closings ec ON pwc.item_id = ec.item_id AND pwc.warehouse_id = ec.warehouse_id
                    LEFT JOIN latest_closings lc ON pwc.item_id = lc.item_id AND pwc.warehouse_id = lc.warehouse_id
                    LEFT JOIN inventory_movements im ON pwc.item_id = im.item_id AND pwc.warehouse_id = im.warehouse_id
                )
                SELECT 
                    cd.*,
                    pi.name as item_name,
                    w.name as warehouse_name,
                    u.name as unit_name,
                    cu.name as created_by_name,
                    uu.name as updated_by_name
                FROM combined_data cd
                LEFT JOIN products pi ON cd.item_id = pi.id
                LEFT JOIN item_units iu ON pi.item_unit_id = iu.id
                LEFT JOIN mix_values u ON u.id = iu.unit_id
                LEFT JOIN mix_values w ON w.id = cd.warehouse_id
								LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
								LEFT JOIN mix_values ig ON isg.parent_id = ig.id
                LEFT JOIN users cu ON 1 = cu.id 
                LEFT JOIN users uu ON 1 = uu.id 
                WHERE 1=1`
			args = append(args, endAt)
			i++
		}
	}

	filterDBColumnKey := []string{
		"pi.name",
	}

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

	if filters["ids"] != "" {
		condition += fmt.Sprintf(" AND iv.id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"item_id":           "cd.item_id",
		"warehouse_id":      "cd.warehouse_id",
		"item_group_id":     "ig.id",
		"item_sub_group_id": "isg.id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"warehouse_ids":      "cd.warehouse_id",
		"item_ids":           "cd.item_id",
		"item_group_ids":     "ig.id",
		"item_sub_group_ids": "isg.id",
	}

	for key, valueID := range filterIDsKey {
		if value, ok := filters[key]; ok && value != "" {
			ids := strings.Split(value, ",")
			intIDs, err := utils.SplitStringArrayOfInts(ids)
			if err != nil {
				utils.LogErrors(childSpan, err)
				return nil, 0, err
			}

			condition += fmt.Sprintf(" AND %s = ANY($%d)", valueID, i)
			args = append(args, pq.Array(intIDs))
			i++
		}
	}

	filterIDsOrKey := map[string][]string{
		"vat_ids": {"iv.vat_id", "sd.vat_id"},
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

	log.Println("dateCondition", dateCondition)

	baseQuery := ""
	if dateCondition != "" {
		baseQuery = dateCondition
	} else {
		log.Println("baseQuery2", baseQuery)
		baseQuery = `
		FROM ( 
			SELECT DISTINCT ON (st.id)
				st.id, st.item_id, st.warehouse_id,
				TO_CHAR(st.closing_at, 'YYYY-MM-DD') as closing_at,
				TO_CHAR(st.last_closing_at, 'YYYY-MM-DD') as last_closing_at,
				st.begin_qty, st.in_qty, st.out_qty, st.adjustment_qty, st.end_qty,
				st.price_sell, st.price_buy,
				st.total_value_sell, st.total_value_buy,
				st.created_at, st.updated_at, st.deleted_at,
				pi.name as item_name,
				w.name as warehouse_name,
				u.name as unit_name,
				cu.name as created_by_name,
				uu.name as updated_by_name
			FROM stock_closings st
			LEFT JOIN products pi ON st.item_id = pi.id
			LEFT JOIN item_units iu ON pi.item_unit_id = iu.id
			LEFT JOIN mix_values u ON u.id = iu.unit_id
			LEFT JOIN mix_values w ON w.id = st.warehouse_id
			LEFT JOIN users cu ON st.created_by_id = cu.id
			LEFT JOIN users uu ON st.updated_by_id = uu.id
			WHERE 1=1` + condition + queryGlobal + `
		) AS alias WHERE 1=1 AND deleted_at IS NULL`
	}

	query := ``
	countQuery := ``

	if dateCondition != "" {
		// For the main query
		query = dateCondition + condition + queryGlobal

		// Add additional filters
		for key, value := range filters {
			switch key {
			case "item_name", "warehouse_name", "unit_name":
				if value != "" {
					query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
					args = append(args, "%"+value+"%")
					i++
				}
			}
		}

		// Add ordering
		orderColumn := utils.GetStringOrDefault(filters["order_column"], "closing_at")
		orderDirection := utils.GetStringOrDefault(filters["order_direction"], "desc")
		query += fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)

		// For count query, we need to wrap the existing query
		countQuery = `SELECT COUNT(*) as total FROM (` + query + `) AS count_subquery`
	} else {
		// Original query structure when no date filters
		query = `SELECT * ` + baseQuery
		countQuery = `SELECT COUNT(*) as total ` + baseQuery

		// Add additional filters to both queries
		for key, value := range filters {
			switch key {
			case "item_name", "warehouse_name", "unit_name":
				if value != "" {
					query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
					countQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
					args = append(args, "%"+value+"%")
					i++
				}
			}
		}

		// Add ordering only to main query
		orderColumn := utils.GetStringOrDefault(filters["order_column"], "id")
		orderDirection := utils.GetStringOrDefault(filters["order_direction"], "desc")
		query += fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)
	}
	countArgs := append([]interface{}{}, args...)

	var wg sync.WaitGroup
	var countErr, selectErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		if filters["is_csv"] != "1" {
			countSpan := opentracing.StartSpan("CountQuery", opentracing.ChildOf(childSpan.Context()))
			defer countSpan.Finish()

			// err := r.db.Raw(countQuery, countArgs...).Scan(&total).Error
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
		defer selectSpan.Finish()

		err := r.sqlDB.SelectContext(ctx.Context(), &products, query, args...)
		if err != nil {
			selectSpan.LogKV("query", query)
			utils.LogErrors(selectSpan, err)
			selectErr = err
		}
	}()

	wg.Wait()

	if countErr != nil {
		return nil, 0, countErr
	}

	if selectErr != nil {
		return nil, 0, selectErr
	}

	return products, total, nil
}

func (r *InventoryRepository) GetStockClosingsBackup(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.StockClosingListDTO, int, error) {
	childSpan := opentracing.StartSpan("InventoryRepository-GetStockClosings", opentracing.ChildOf(span.Context()))

	products := []dtos.StockClosingListDTO{}

	var total int

	filterDBColumnKey := []string{
		"pi.name",
	}

	var args []interface{}

	queryGlobal := ""

	i := 1
	// filters["start_at"], filters["end_at"]

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
		condition += fmt.Sprintf(" AND iv.id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"item_id":      "st.item_id",
		"warehouse_id": "st.warehouse_id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"warehouse_ids": "st.warehouse_id",
		"item_ids":      "st.item_id",
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
		"vat_ids": {"iv.vat_id", "sd.vat_id"},
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
        SELECT DISTINCT ON (st.id)
					st.id, st.item_id, st.warehouse_id,
					TO_CHAR(st.closing_at, 'YYYY-MM-DD') as closing_at,
					TO_CHAR(st.last_closing_at, 'YYYY-MM-DD') as last_closing_at,
					st.begin_qty, st.in_qty, st.out_qty, st.adjustment_qty, st.end_qty,
					st.price_sell, st.price_buy,
					st.total_value_sell, st.total_value_buy,
					st.created_at, st.updated_at, st.deleted_at,

					pi.name as item_name,
					w.name as warehouse_name,
					u.name as unit_name,
					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM stock_closings st
				LEFT JOIN products pi ON st.item_id = pi.id
				LEFT JOIN item_units iu ON pi.item_unit_id = iu.id
				LEFT JOIN mix_values u ON u.id = iu.unit_id
				LEFT JOIN mix_values w ON w.id = st.warehouse_id
				LEFT JOIN users cu ON st.created_by_id = cu.id
				LEFT JOIN users uu ON st.updated_by_id = uu.id
				WHERE 1=1` + condition + queryGlobal + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery

	for key, value := range filters {
		switch key {
		// case "po_buyer_no", "sales_order_no", "ship_dest", "remark":
		case "item_name", "warehouse_name", "unit_name":
			if value != "" {
				query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				countQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "id")
	orderDirection := utils.GetStringOrDefault(filters["order_direction"], "desc")
	query += fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)

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

func (r *InventoryRepository) CreateStockClosings(ctx *fiber.Ctx, tx *gorm.DB, date string, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InventoryRepository-CreateStockClosings", opentracing.ChildOf(span.Context()))

	log.Printf("[INFO] Processing stock closing for %s...", date)
	adminID := 1
	query := `WITH date_params AS (
			SELECT $1::date as closing_date
		),
		all_warehouses AS (
				SELECT mv.id as warehouse_id
				FROM mix_values mv
				LEFT JOIN groups g ON mv.group_id = g.id
				WHERE g.name = 'warehouses'
				AND mv.deleted_at IS NULL
		),
		latest_closing AS (
				SELECT 
						item_id, 
						warehouse_id, 
						closing_at as last_closing_date,
						end_qty as prev_qty
				FROM stock_closings sc
				WHERE sc.closing_at = (
						SELECT MAX(closing_at)
						FROM stock_closings sc2
						WHERE sc2.closing_at < (SELECT closing_date FROM date_params)
						AND sc2.item_id = sc.item_id
						AND sc2.warehouse_id = sc.warehouse_id
						AND sc2.deleted_at IS NULL
				)
				AND sc.deleted_at IS NULL
		),
		inventory_movements AS (
				SELECT 
						id.item_id,
						i.warehouse_id,
						SUM(CASE 
								WHEN iot.options_json->>'io_type' = 'INVENTORY_IN' THEN COALESCE(id.qty, 0)
								ELSE 0 
						END) as qty_in,
						SUM(CASE 
								WHEN iot.options_json->>'io_type' = 'INVENTORY_OUT' THEN COALESCE(id.qty, 0)
								ELSE 0 
						END) as qty_out,
						MAX(CASE 
								WHEN iot.options_json->>'io_type' = 'INVENTORY_IN' THEN id.price_buy
								ELSE NULL 
						END) as last_buy_price,
						MAX(CASE 
								WHEN iot.options_json->>'io_type' = 'INVENTORY_IN' THEN id.price_sell
								ELSE NULL 
						END) as last_sell_price
				FROM inv_dts id
				INNER JOIN inventories i ON id.inventory_id = i.id 
				INNER JOIN mix_values iot ON i.io_type_id = iot.id
				WHERE i.ingoing_at > (
						SELECT COALESCE(
								(SELECT MAX(closing_at) 
								FROM stock_closings sc 
								WHERE sc.item_id = id.item_id 
								AND sc.warehouse_id = i.warehouse_id 
								AND sc.closing_at < (SELECT closing_date FROM date_params)
								AND sc.deleted_at IS NULL
								),
								'1970-01-01'
						)
				)
				AND i.ingoing_at <= (SELECT closing_date FROM date_params)
				AND i.deleted_at IS NULL 
				AND id.deleted_at IS NULL
				GROUP BY id.item_id, i.warehouse_id
		),
		all_products AS (
				SELECT 
						p.id as item_id,
						p.item_unit_id,
						iu.price_buy as default_price_buy,
						iu.price_sell as default_price_sell
				FROM products p
				LEFT JOIN item_units iu ON p.item_unit_id = iu.id 
				WHERE p.deleted_at IS NULL
		),
		product_warehouse_combinations AS (
				SELECT 
						ap.item_id,
						aw.warehouse_id,
						ap.default_price_buy,
						ap.default_price_sell
				FROM all_products ap
				CROSS JOIN all_warehouses aw
		),
		calculated_stock AS (
				SELECT
						pwc.item_id,
						pwc.warehouse_id,
						COALESCE(lc.last_closing_date, (SELECT closing_date FROM date_params)) as last_closing_at,
						COALESCE(lc.prev_qty, 0) as begin_qty,
						COALESCE(im.qty_in, 0) as in_qty,
						COALESCE(im.qty_out, 0) as out_qty,
						0 as adjustment_qty,
						COALESCE(lc.prev_qty, 0) + COALESCE(im.qty_in, 0) - COALESCE(im.qty_out, 0) as end_qty,
						COALESCE(im.last_sell_price, pwc.default_price_sell, 0) as price_sell,
						COALESCE(im.last_buy_price, pwc.default_price_buy, 0) as price_buy
				FROM product_warehouse_combinations pwc
				LEFT JOIN latest_closing lc ON pwc.item_id = lc.item_id 
						AND pwc.warehouse_id = lc.warehouse_id
				LEFT JOIN inventory_movements im ON pwc.item_id = im.item_id 
						AND pwc.warehouse_id = im.warehouse_id
		)
		INSERT INTO stock_closings (
				item_id,
				warehouse_id, 
				closing_at,
				last_closing_at,
				begin_qty,
				in_qty,
				out_qty,
				adjustment_qty,
				end_qty,
				price_sell,
				price_buy,
				total_value_sell,
				total_value_buy,
				created_at,
				created_by_id
		)
		SELECT
				item_id,
				warehouse_id,
				(SELECT closing_date FROM date_params),
				last_closing_at,
				begin_qty,
				in_qty, 
				out_qty,
				adjustment_qty,
				end_qty,
				price_sell,
				price_buy,
				(end_qty * price_sell) as total_value_sell,
				(end_qty * price_buy) as total_value_buy,
				CURRENT_TIMESTAMP,
				$2
		FROM calculated_stock;`

	var productCount, warehouseCount int
	tx.Raw("SELECT COUNT(*) FROM products WHERE deleted_at IS NULL").Scan(&productCount)
	tx.Raw("SELECT COUNT(*) FROM mix_values WHERE group_id = (SELECT id FROM groups WHERE name = 'warehouses') AND deleted_at IS NULL").Scan(&warehouseCount)

	log.Printf("[DEBUG] Found %d active products and %d active warehouses", productCount, warehouseCount)

	err := tx.Error

	result := tx.Exec(query, date, adminID)

	if result.Error != nil {
		defer childSpan.Finish()
		log.Printf("[ERROR] Error creating stock closing: %v", err)
		utils.LogErrors(childSpan, err)
		tx.Rollback()
		return nil, err
	}

	log.Printf("[SUCCESS] Stock closing for %s created successfully.", date)

	return tx, nil
}

func (r *InventoryRepository) DeleteStockClosingByDate(ctx *fiber.Ctx, tx *gorm.DB, date string, userID uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InvDtRepository-DeleteStockClosingByDate", opentracing.ChildOf(span.Context()))

	// now := time.Now()
	// deletedAt := gorm.DeletedAt{Time: now, Valid: true}

	// query := tx.Model(&models.StockClosings{}).Where("closing_at = ?", date)

	// if err := query.Updates(map[string]interface{}{
	// 	"deleted_by_id": userID,
	// 	"deleted_at":    deletedAt,
	// }).Error; err != nil {
	// 	defer childSpan.Finish()
	// 	log.Printf("[ERROR]Error deleting stock closing: %v", err)
	// 	utils.LogErrors(childSpan, err)
	// 	return err
	// }

	// Use Unscoped() to bypass soft delete and Delete() for permanent deletion
	if err := tx.Unscoped().Where("closing_at = ?", date).Delete(&models.StockClosings{}).Error; err != nil {
		defer childSpan.Finish()
		log.Printf("[ERROR]Error deleting stock closing: %v", err)
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *InventoryRepository) GetInventoriesStatus(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.InventoryStatusDTO, int, error) {
	childSpan := opentracing.StartSpan("InventoryRepository-GetInventoriesStatus", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	products := []dtos.InventoryStatusDTO{}

	var total int

	filterDBColumnKey := []string{
		"iv.inventory_no", "iv.do_no", "iv.surat_jalan_no", "iv.invoice_no", "iv.remark", "iv.ship_dest",
		"pi.name",
		"so.po_buyer_no",
		"so.sales_order_no",
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
		condition += fmt.Sprintf(" AND iv.id IN (%s)", filters["ids"])
	}

	if filters["io_type"] != "" {
		condition += fmt.Sprintf(" AND ot.options_json->>'io_type' = '%s'", filters["io_type"])
	}

	filterKey := map[string]string{
		"status":          "iv.status",
		"customer_id":     "iv.customer_id",
		"io_type_id":      "iv.io_type_id",
		"currency_id":     "iv.currency_id",
		"warehouse_id":    "iv.warehouse_id",
		"branch_id":       "iv.branch_id",
		"vat_id":          "iv.vat_id",
		"payment_term_id": "iv.payment_term_id",
		"pph23_id":        "iv.pph23_id",
		"due_at":          "iv.due_at",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":       "iv.customer_id",
		"io_type_ids":        "iv.io_type_id",
		"currency_ids":       "iv.currency_id",
		"payment_term_ids":   "iv.payment_term_id",
		"pph23_ids":          "iv.pph23_id",
		"warehouse_ids":      "iv.warehouse_id",
		"item_group_ids":     "ig.id",
		"item_sub_group_ids": "isg.id",
		"item_ids":           "ivd.item_id",
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
		"vat_ids": {"iv.vat_id", "sd.vat_id"},
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

	// And one for array conditions with OR
	filterIDsOrArrayKey := map[string][]string{
		"product_ids": {"ivd.ref_product_id", "ivd.item_id"},
	}

	// Handle array OR conditions
	for key, valueIDs := range filterIDsOrArrayKey {
		if value, ok := filters[key]; ok && value != "" {
			// Split the string into an array of integers
			ids := strings.Split(value, ",")
			intIDs, err := utils.SplitStringArrayOfInts(ids)
			if err != nil {
				utils.LogErrors(childSpan, err)
				return nil, 0, err
			}

			condition += " AND ("
			for idx, valueID := range valueIDs {
				if idx > 0 {
					condition += " OR"
				}
				condition += fmt.Sprintf(" %s = ANY($%d)", valueID, i)
				args = append(args, pq.Array(intIDs))
				i++
			}
			condition += ")"
		}
	}

	filterDBColumnLikeKey := map[string]string{
		"do_no":          "iv.do_no",
		"invoice_no":     "iv.invoice_no",
		"surat_jalan_no": "iv.surat_jalan_no",
		"inventory_no":   "iv.inventory_no",
		"po_buyer_no":    "so.po_buyer_no",
		"sales_order_no": "so.sales_order_no",
		"ship_dest":      "iv.ship_dest",
		"remark":         "iv.remark",
		"item_name":      "pi.name",
		"customer_name":  "c.name",
	}

	for key, value := range filters {
		if value != "" {
			for keyLike, column := range filterDBColumnLikeKey {
				if filters[key] != "" && key == keyLike {
					condition += fmt.Sprintf(" AND %s ILIKE $%d", column, i)
					args = append(args, "%"+value+"%")
					i++
				}
			}
		}
	}

	// if date_type, start_date, end_date filled
	if filters["date_type"] != "" && filters["start_date"] != "" && filters["end_date"] != "" {

		filterDateTypeKey := map[string]string{
			"ingoing_at":  "iv.ingoing_at",
			"invoice_at":  "iv.invoice_at",
			"do_at":       "iv.do_at",
			"shipping_at": "so.shipping_at",
		}

		dateTypeColumn := filterDateTypeKey[filters["date_type"]]
		condition += fmt.Sprintf(" AND (%s BETWEEN $%d AND $%d)", dateTypeColumn, i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
	}

	columnIDsKey := map[string]string{
		"ingoing_at":  "iv.ingoing_at",
		"invoice_at":  "iv.invoice_at",
		"do_at":       "iv.do_at",
		"shipping_at": "so.shipping_at",
	}

	if filters["order_column"] != "" {
		if _, ok := columnIDsKey[filters["order_column"]]; !ok {
			filters["order_column"] = "iv.ingoing_at"
		}
	}

	// orderColumn := utils.GetStringOrDefault(filters["order_column"], "iv.ingoing_at")
	// orderDirection := utils.GetStringOrDefault(filters["order_direction"], "desc")
	// for key, valueID := range columnIDsKey {
	// 	if value, ok := filters["order_column"]; ok && value != "" && key == value {
	// 		orderColumn = valueID
	// 	}
	// }

	baseQuery := `
    FROM ( 
        SELECT 
					sub.item_id,
					sub.ingoing_at,
					sub.branch_id
        FROM (
            SELECT DISTINCT pi.id,
							pi.id,
							pi.id as item_id,
							pi.name as item_name,
							iv.branch_id,
							iv.ingoing_at,
							iv.created_at,
							iv.updated_at,
							ROW_NUMBER() OVER (PARTITION BY pi.id ORDER BY iv.ingoing_at DESC, pi.id) as rn
            FROM products pi
            JOIN inv_dts ivd ON ivd.item_id = pi.id
            JOIN inventories iv ON ivd.inventory_id = iv.id AND iv.deleted_at IS NULL
            LEFT JOIN item_units iu ON ivd.item_unit_id = iu.id
            LEFT JOIN mix_values cur ON iv.currency_id = cur.id
            LEFT JOIN mix_values vat ON iv.vat_id = vat.id
            LEFT JOIN mix_values pph ON iv.pph23_id = pph.id
            LEFT JOIN mix_values ot ON iv.io_type_id = ot.id
            LEFT JOIN customers c ON iv.customer_id = c.id
            LEFT JOIN so_dts sd ON ivd.ref_so_dt_id = sd.id
            LEFT JOIN sales_orders so ON sd.sales_order_id = so.id
            LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
            LEFT JOIN mix_values ig ON isg.parent_id = ig.id
            WHERE ivd.deleted_at IS NULL
            ` + condition + queryGlobal + `
        ) sub
        WHERE rn = 1
        ORDER BY ingoing_at desc, updated_at desc, created_at desc, item_name
    ) AS alias WHERE 1=1`
	// baseQuery := `
	//   FROM (
	//       SELECT
	// 				pi.id as id,
	// 				pi.id as item_id,
	// 				iv.branch_id

	//       FROM products pi
	// 			JOIN inv_dts ivd ON ivd.item_id = pi.id
	// 			JOIN inventories iv ON ivd.inventory_id = iv.id AND iv.deleted_at IS NULL
	// 			LEFT JOIN item_units iu ON ivd.item_unit_id = iu.id

	// 			LEFT JOIN mix_values cur ON iv.currency_id = cur.id
	// 			LEFT JOIN mix_values vat ON iv.vat_id = vat.id
	// 			LEFT JOIN mix_values pph ON iv.pph23_id = pph.id
	// 			LEFT JOIN mix_values ot ON iv.io_type_id = ot.id
	// 			LEFT JOIN customers c ON iv.customer_id = c.id
	// 			LEFT JOIN so_dts sd ON ivd.ref_so_dt_id = sd.id
	// 			LEFT JOIN sales_orders so ON sd.sales_order_id = so.id
	// 			LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
	// 			LEFT JOIN mix_values ig ON isg.parent_id = ig.id

	// 			WHERE ivd.deleted_at IS NULL
	// 			` + condition + queryGlobal + `
	// 			ORDER BY ` + orderColumn + ` ` + orderDirection + `, pi.id
	//   ) AS alias WHERE 1=1`

	query := `SELECT *
		` + baseQuery

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery

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

func (r *InventoryRepository) GetInventoriesStatusDt(ctx *fiber.Ctx, filters map[string]string, params dtos.GetInventoriesStatusDtParams, span opentracing.Span) ([]dtos.InventoryStatusDtDTO, int, error) {
	childSpan := opentracing.StartSpan("InventoryRepository-GetInventoriesStatusDt", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	products := []dtos.InventoryStatusDtDTO{}

	var total int

	filterDBColumnKey := []string{
		"iv.inventory_no", "iv.do_no", "iv.surat_jalan_no", "iv.invoice_no", "iv.remark", "iv.ship_dest",
		"pi.name",
		"so.po_buyer_no",
		"so.sales_order_no",
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
		condition += fmt.Sprintf(" AND iv.id IN (%s)", filters["ids"])
	}

	if filters["item_ids"] != "" {
		condition += fmt.Sprintf(" AND ivd.item_id IN (%s)", filters["item_ids"])
	}

	if len(params.ItemIDs) > 0 {
		condition += fmt.Sprintf(" AND ivd.item_id = ANY($%d)", i)
		args = append(args, pq.Array(params.ItemIDs))
		i++
	}

	if filters["io_type"] != "" {
		condition += fmt.Sprintf(" AND ot.options_json->>'io_type' = '%s'", filters["io_type"])
	}

	filterKey := map[string]string{
		"status":          "iv.status",
		"customer_id":     "iv.customer_id",
		"io_type_id":      "iv.io_type_id",
		"currency_id":     "iv.currency_id",
		"warehouse_id":    "iv.warehouse_id",
		"branch_id":       "iv.branch_id",
		"vat_id":          "iv.vat_id",
		"payment_term_id": "iv.payment_term_id",
		"pph23_id":        "iv.pph23_id",
		"due_at":          "iv.due_at",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":     "iv.customer_id",
		"io_type_ids":      "iv.io_type_id",
		"currency_ids":     "iv.currency_id",
		"payment_term_ids": "iv.payment_term_id",
		"pph23_ids":        "iv.pph23_id",
		"warehouse_ids":    "iv.warehouse_id",
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
		"vat_ids": {"iv.vat_id", "sd.vat_id"},
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

	filterDBColumnLikeKey := map[string]string{
		"do_no":          "iv.do_no",
		"invoice_no":     "iv.invoice_no",
		"surat_jalan_no": "iv.surat_jalan_no",
		"inventory_no":   "iv.inventory_no",
		"po_buyer_no":    "so.po_buyer_no",
		"sales_order_no": "so.sales_order_no",
		"ship_dest":      "iv.ship_dest",
		"remark":         "iv.remark",
		"item_name":      "pi.name",
		"customer_name":  "c.name",
	}

	for key, value := range filters {
		if value != "" {
			for keyLike, column := range filterDBColumnLikeKey {
				if filters[key] != "" && key == keyLike {
					condition += fmt.Sprintf(" AND %s ILIKE $%d", column, i)
					args = append(args, "%"+value+"%")
					i++
				}
			}
		}
	}

	// if date_type, start_date, end_date filled
	if filters["date_type"] != "" && filters["start_date"] != "" && filters["end_date"] != "" {

		filterDateTypeKey := map[string]string{
			"ingoing_at":  "iv.ingoing_at",
			"invoice_at":  "iv.invoice_at",
			"do_at":       "iv.do_at",
			"shipping_at": "so.shipping_at",
		}

		dateTypeColumn := filterDateTypeKey[filters["date_type"]]
		condition += fmt.Sprintf(" AND (%s BETWEEN $%d AND $%d)", dateTypeColumn, i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
	}

	columnIDsKey := map[string]string{
		"ingoing_at":  "iv.ingoing_at",
		"invoice_at":  "iv.invoice_at",
		"do_at":       "iv.do_at",
		"shipping_at": "so.shipping_at",
	}

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "iv.ingoing_at")
	orderDirection := utils.GetStringOrDefault(filters["order_direction"], "desc")
	for key, valueID := range columnIDsKey {
		if value, ok := filters["order_column"]; ok && value != "" && key == value {
			orderColumn = valueID
		}
	}

	baseQuery := `
    FROM ( 
        SELECT
					ivd.id as id,
					ivd.item_id as item_id,
					TO_CHAR(iv.ingoing_at, 'YYYY-MM-DD') as ingoing_at,

					iot.options_json->>'io_type' as io_type,
					isg.name as item_sub_group_name,
					ig.name as item_group_name,
					w.name as warehouse_name,
					iot.name as io_type_name,
					c.name as customer_name,
					pi.code as item_code,
					pi.name as item_name,
					u.name as unit_name,
					cur.name as currency_name,
					ivd.price_sell as price_sell,
					ivd.price_buy as price_buy,
					ivd.qty as qty

        FROM inv_dts ivd
				LEFT JOIN inventories iv ON ivd.inventory_id = iv.id
				LEFT JOIN products pi ON ivd.item_id = pi.id
				LEFT JOIN item_units iu ON ivd.item_unit_id = iu.id

				LEFT JOIN mix_values cur ON iv.currency_id = cur.id
				LEFT JOIN mix_values vat ON iv.vat_id = vat.id
				LEFT JOIN mix_values pph ON iv.pph23_id = pph.id
				LEFT JOIN mix_values ot ON iv.io_type_id = ot.id
				LEFT JOIN customers c ON iv.customer_id = c.id
				LEFT JOIN so_dts sd ON ivd.ref_so_dt_id = sd.id
				LEFT JOIN sales_orders so ON sd.sales_order_id = so.id
				LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
				LEFT JOIN mix_values ig ON isg.parent_id = ig.id
				LEFT JOIN mix_values iot ON iv.io_type_id = iot.id
				LEFT JOIN mix_values w ON iv.warehouse_id = w.id
				LEFT JOIN mix_values u ON iu.unit_id = u.id

        LEFT JOIN users cu ON iv.created_by_id = cu.id
        LEFT JOIN users uu ON iv.updated_by_id = uu.id
				WHERE iv.deleted_at IS NULL AND ivd.deleted_at IS NULL
				` + condition + queryGlobal + `
				ORDER BY ivd.item_id, ` + orderColumn + ` ` + orderDirection + `, ivd.id desc
    ) AS alias WHERE 1=1`

	query := `SELECT *
		` + baseQuery

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery

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

func (r *InventoryRepository) GetLatestInventory(ctx *fiber.Ctx, tx *gorm.DB, filters map[string]string, span opentracing.Span) (*dtos.InventoryListDTO, error) {
	childSpan := opentracing.StartSpan("InventoryRepository-GetLatestInventory", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	products := dtos.InventoryListDTO{}

	var args []interface{}

	queryGlobal := ""

	i := 1

	condition := ""

	baseQuery := `
    FROM ( 
        SELECT
					iv.id, iv.branch_id,
					TO_CHAR(iv.ingoing_at, 'YYYY-MM-DD') as ingoing_at,
					TO_CHAR(iv.do_at, 'YYYY-MM-DD') as do_at,
					TO_CHAR(iv.invoice_at, 'YYYY-MM-DD') as invoice_at

        FROM inventories iv

        LEFT JOIN users cu ON iv.created_by_id = cu.id
        LEFT JOIN users uu ON iv.updated_by_id = uu.id
				WHERE iv.deleted_at IS NULL
				` + condition + queryGlobal + `
    ) AS alias WHERE 1=1`

	query := `SELECT *
		` + baseQuery

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

	var wg sync.WaitGroup
	var selectErr error

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "ingoing_at")
	orderDirection := utils.GetStringOrDefault(filters["order_direction"], "desc")
	query += fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)
	query += " LIMIT 1"

	wg.Add(1)
	go func() {
		defer wg.Done()
		selectSpan := opentracing.StartSpan("SelectQuery", opentracing.ChildOf(childSpan.Context()))

		err := tx.Raw(query, args...).Scan(&products).Error
		if err != nil {
			selectSpan.LogKV("query", query)
			utils.LogErrors(selectSpan, err)
			selectErr = err
		}
	}()

	wg.Wait()
	defer childSpan.Finish()

	if selectErr != nil {
		return nil, selectErr
	}

	return &products, nil
}
