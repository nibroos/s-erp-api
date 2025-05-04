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
	tracer   opentracing.Tracer
}

func NewInventoryRepository(db *gorm.DB, sqlDB *sqlx.DB, utilRepo *UtilRepository, tracer opentracing.Tracer) *InventoryRepository {
	return &InventoryRepository{
		db:       db,
		sqlDB:    sqlDB,
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
		"vat_ids": []string{"iv.vat_id", "sd.vat_id"},
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
		"vat_ids": []string{"iv.vat_id", "sd.vat_id"},
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
		customCondition += fmt.Sprintf("%s", join)
	}

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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "id")
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

				-- io_type
				ot.options_json->>'io_type' as io_type,

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

	if err := r.sqlDB.Get(&salesOrder, query, args...); err != nil {
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
			"id":               invDt.ID,
			"inventory_id":     invDt.InventoryID,
			"item_unit_id":     invDt.ItemUnitID,
			"vat_id":           invDt.VatID,
			"pph23_id":         invDt.Pph23ID,
			"ref_so_dt_id":     invDt.RefSoDtID,
			"ref_so_dt_bom_id": invDt.RefSoDtBomID,
			"ref_po_dt_id":     invDt.RefPoDtID,
			"ref_po_dt_bom_id": invDt.RefPoDtBomID,
			"ref_inv_dt_id":    invDt.RefInvDtID,
			"ref_product_id":   invDt.RefProductID,
			"item_id":          invDt.ItemID,
			"product_uuid":     invDt.ProductUuid,
			"item_type":        invDt.ItemType,
			"ref_type":         invDt.RefType,
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

		ivd.ref_so_dt_id, ivd.ref_so_dt_bom_id, ivd.ref_po_dt_id, ivd.ref_po_dt_bom_id, ivd.ref_inv_dt_id, ivd.ref_product_id,

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
		 sd.qty_out, sdb.qty_out, ivd_refs.qty_out
		) as qty_out,
		COALESCE(
		 sd.qty, sdb.qty, ivd_refs.qty
		) as ref_qty,

		COALESCE(
			so.po_buyer_no, iv_refs.inventory_no, NULL
		) as ref_num,

		cu.name as created_by_name,
		uu.name as updated_by_name

	FROM inv_dts ivd
	LEFT JOIN so_dts sd ON sd.id = ivd.ref_so_dt_id AND ivd.ref_type = 'so'
	LEFT JOIN so_dt_boms sdb ON sdb.id = ivd.ref_so_dt_bom_id AND ivd.ref_type = 'so'
	LEFT JOIN sales_orders so ON (so.id = sd.sales_order_id OR so.id = sdb.sales_order_id)
	LEFT JOIN inv_dts ivd_refs ON ivd_refs.id = ivd.ref_inv_dt_id
	LEFT JOIN inventories iv_refs ON ivd_refs.inventory_id = iv_refs.id
	LEFT JOIN inventories p ON ivd.inventory_id = p.id
	LEFT JOIN products pi ON ivd.item_id = pi.id
	LEFT JOIN item_units iu ON ivd.item_unit_id = iu.id
	LEFT JOIN mix_values u ON iu.unit_id = u.id
	LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
	LEFT JOIN mix_values ig ON isg.parent_id = ig.id
	LEFT JOIN users cu ON ivd.created_by_id = cu.id
	LEFT JOIN users uu ON ivd.updated_by_id = uu.id
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

	childSpan := opentracing.StartSpan("InventoryRepository-GetRefIndexQuoDts", opentracing.ChildOf(span.Context()))

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
		"vat_ids": []string{"so.vat_id", "sd.vat_id"},
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
		"vat_ids": []string{"q.vat_id", "sd.vat_id"},
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
		query += fmt.Sprintf(" AND id = ANY($1)")
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
		query += fmt.Sprintf(" AND %s = ANY($1)", parentColumnName)
		args = append(args, pq.Array(detailIDs))
		i++
	}

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
		"vat_ids": []string{"iv.vat_id", "sd.vat_id"},
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
		customCondition += fmt.Sprintf("%s", join)
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
				LEFT JOIN so_dts sd ON ivd.ref_so_dt_id = sd.id
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
