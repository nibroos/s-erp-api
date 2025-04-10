package repository

import (
	"fmt"
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
		"so.po_buyer_no", "so.sales_order_no", "so.remark", "so.ship_dest",
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
		condition += fmt.Sprintf(" AND id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"status":        "so.status",
		"customer_id":   "so.customer_id",
		"order_type_id": "so.order_type_id",
		"currency_id":   "so.currency_id",
		"vat_id":        "so.vat_id",
		"payment_id":    "so.payment_id",
		"pph23_id":      "so.pph23_id",
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
		"customer_ids":   "so.customer_id",
		"order_type_ids": "so.order_type_id",
		"currency_ids":   "so.currency_id",
		"payment_ids":    "so.payment_id",
		"pph23_ids":      "so.pph23_id",
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

	joinCondition := ""

	filterKeyJoin := map[string]string{
		"is_task_exists": "JOIN schedules s ON s.sales_order_id = so.id AND s.deleted_at IS NULL JOIN schedule_tasks stp ON stp.schedule_id = s.id AND stp.deleted_at IS NULL AND stp.entity_type = 'steps' JOIN schedule_tasks st ON st.parent_id = stp.id AND st.deleted_at IS NULL AND st.entity_type = 'tasks'",
	}
	for _, join := range filterKeyJoin {
		joinCondition += fmt.Sprintf(" %s", join)
	}

	customCondition := ""
	filterKeyCustom := map[string]string{
		"is_task_exists": " AND st.is_checked = 1",
	}
	for _, join := range filterKeyCustom {
		customCondition += fmt.Sprintf("%s", join)
	}

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (so.id)
					so.id, so.customer_id, so.order_type_id, so.currency_id, so.vat_id, so.payment_id, so.pph23_id, so.branch_id,
					so.po_buyer_no, so.sales_order_no, so.ship_dest, so.remark, 
					so.status, so.exchange_rate, so.pph23_perc, so.markup_perc, so.total_qty, so.subtotal, so.total_discount, so.total_pph23, so.total_vat, so.grand_total, so.created_by_id, so.updated_by_id, so.deleted_by_id, so.created_at, so.updated_at, so.deleted_at,
					TO_CHAR(so.order_at, 'YYYY-MM-DD') as order_at,
					TO_CHAR(so.shipping_at, 'YYYY-MM-DD') as shipping_at,
					TO_CHAR(so.agree_at, 'YYYY-MM-DD') as agree_at,
					TO_CHAR(so.due_at, 'YYYY-MM-DD') as due_at,
					so.vat_perc, so.disc_am, so.disc_perc, so.disc_perc_am, so.disc_final, so.disc_type, so.qty_out, so.si_total_am, so.sa_total_am,

					pi.id as product_id,
					it.id as item_id,
					sd.vat_id as so_dt_vat_id,

					pi.name as product_name,
					it.name as item_name,
					cur.name as currency_name,
					vat.name as vat_name,
					pph.name as pph23_name,

					ot.name as order_type_name,
					c.name as customer_name,

					sd.remark as so_dt_remark,
					sd.gen_code as so_dt_gen_code,
					sdb.remark as so_dt_bom_remark,
					sdb.gen_code as so_dt_bom_gen_code,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM sales_orders so
				LEFT JOIN so_dts sd ON sd.sales_order_id = so.id
				LEFT JOIN products pi ON sd.item_id = pi.id
				LEFT JOIN item_units iu ON sd.item_unit_id = iu.id
				LEFT JOIN so_dt_boms sdb ON sdb.so_dt_id = sd.id
				LEFT JOIN products it ON sdb.item_id = it.id

				LEFT JOIN mix_values cur ON so.currency_id = cur.id
				LEFT JOIN mix_values vat ON so.vat_id = vat.id
				LEFT JOIN mix_values pph ON so.pph23_id = pph.id
				LEFT JOIN mix_values ot ON so.order_type_id = ot.id
				LEFT JOIN customers c ON so.customer_id = c.id

        LEFT JOIN users cu ON so.created_by_id = cu.id
        LEFT JOIN users uu ON so.updated_by_id = uu.id
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "order_at")
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

func (r *SalesOrderRepository) GetSalesOrderByID(ctx *fiber.Ctx, params *dtos.GetSalesOrderParams, tx *gorm.DB, span opentracing.Span) (*dtos.SalesOrderDetailDTO, error) {
	childSpan := opentracing.StartSpan("SalesOrderRepository-GetSalesOrderByID", opentracing.ChildOf(span.Context()))
	var salesOrder dtos.SalesOrderDetailDTO

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	baseQuery := `
    FROM ( 
			SELECT DISTINCT ON (so.id)
				so.id, so.customer_id, so.order_type_id, so.currency_id, so.vat_id, so.payment_id, so.pph23_id, so.branch_id,
				so.po_buyer_no, so.po_buyer_no_ori, so.sales_order_no, so.ship_dest, so.remark, 
				so.status, so.exchange_rate, so.pph23_perc, so.markup_perc, so.total_qty, so.subtotal, so.total_discount, so.total_pph23, so.total_vat, so.grand_total, so.created_by_id, so.updated_by_id, so.deleted_by_id, so.created_at, so.updated_at, so.deleted_at,
				TO_CHAR(so.order_at, 'YYYY-MM-DD') as order_at,
				TO_CHAR(so.shipping_at, 'YYYY-MM-DD') as shipping_at,
				TO_CHAR(so.agree_at, 'YYYY-MM-DD') as agree_at,
				TO_CHAR(so.due_at, 'YYYY-MM-DD') as due_at,
				so.vat_perc, so.disc_am, so.disc_perc, so.disc_perc_am, so.disc_final, so.disc_type, so.qty_out, so.si_total_am, so.sa_total_am,
				so.rev_no,

				cu.name as created_by_name,
				uu.name as updated_by_name

			FROM sales_orders so
			LEFT JOIN so_dts sd ON sd.sales_order_id = so.id
			LEFT JOIN products pi ON sd.item_id = pi.id
			LEFT JOIN item_units iu ON sd.item_unit_id = iu.id
			LEFT JOIN so_dt_boms sdb ON sdb.so_dt_id = sd.id
			LEFT JOIN products it ON sdb.item_id = it.id

			LEFT JOIN mix_values cur ON so.currency_id = cur.id
			LEFT JOIN mix_values vat ON so.vat_id = vat.id
			LEFT JOIN mix_values pph ON so.pph23_id = pph.id
			LEFT JOIN mix_values ot ON so.order_type_id = ot.id
			LEFT JOIN customers c ON so.customer_id = c.id

			LEFT JOIN users cu ON so.created_by_id = cu.id
			LEFT JOIN users uu ON so.updated_by_id = uu.id
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

func (r *SalesOrderRepository) GetScheduleBySalesOrderID(ctx *fiber.Ctx, params *dtos.GetSalesOrderParams, tx *gorm.DB, span opentracing.Span) (*dtos.ScheduleDetailDTO, error) {
	childSpan := opentracing.StartSpan("SalesOrderRepository-GetScheduleBySalesOrderID", opentracing.ChildOf(span.Context()))
	var schedule dtos.ScheduleDetailDTO

	baseQuery := `
    FROM ( 
			SELECT DISTINCT ON (s.id)
				s.id, s.assignee_id, s.sales_order_id, s.uuid, s.steps_id, s.title, s.remark, s.status, s.color, s.created_by_id, s.updated_by_id, s.deleted_by_id, s.deleted_at,

				TO_CHAR(s.start_at, 'YYYY-MM-DD') as start_at,
				TO_CHAR(s.end_at, 'YYYY-MM-DD') as end_at,

				ass.name as assignee_name,
				cu.name as created_by_name,
				uu.name as updated_by_name

			FROM schedules s
			LEFT JOIN sales_orders so ON s.sales_order_id = so.id

			LEFT JOIN users ass ON s.assignee_id = ass.id
			LEFT JOIN users cu ON s.created_by_id = cu.id
			LEFT JOIN users uu ON s.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	var args []interface{}

	i := 1
	query += " AND sales_order_id = $1"
	args = append(args, params.ID)
	i++

	if err := r.sqlDB.Get(&schedule, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		childSpan.LogKV("query", query)
		return nil, err
	}

	return &schedule, nil
}

// BeginTransaction starts a new transaction
func (r *SalesOrderRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

// Rollback all changes in the transaction
func (r *SalesOrderRepository) Rollback() *gorm.DB {
	return r.db.Rollback()
}

func (r *SalesOrderRepository) CreateSalesOrder(tx *gorm.DB, salesOrder *models.SalesOrder, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesOrderRepository-CreateSalesOrder", opentracing.ChildOf(span.Context()))
	if err := tx.Create(salesOrder).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *SalesOrderRepository) UpdateSalesOrder(tx *gorm.DB, salesOrder *models.SalesOrder, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesOrderRepository-UpdateSalesOrder", opentracing.ChildOf(span.Context()))

	if err := tx.Where("id = ?", salesOrder.ID).Select("*").Omit(
		"created_at", "created_by_id", "branch_id",
	).Updates(salesOrder).Error; err != nil {
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

	var salesOrder models.SalesOrder
	if err := tx.Unscoped().Model(&salesOrder).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *SalesOrderRepository) CreateSoDts(tx *gorm.DB, soDts []models.SoDt, salesOrderID uint, span opentracing.Span) (*gorm.DB, []models.SoDt, error) {
	childSpan := opentracing.StartSpan("SoDtRepository-CreateSoDts", opentracing.ChildOf(span.Context()))

	tx = tx.Create(&soDts)
	if tx.Error != nil {
		utils.LogErrors(childSpan, tx.Error)
		tx.Rollback()
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
			"item_unit_id":   soDt.ItemUnitID,
			"vat_id":         soDt.VatID,
			"ref_id":         soDt.RefID,
			"item_id":        soDt.ItemID,
			"item_type":      soDt.ItemType,
			"ref_type":       soDt.RefType,
			// "ref_json":      soDt.RefJSON,
			// "item_json":     soDt.ItemJSON,
			"gen_code":           soDt.GenCode,
			"remark":             soDt.Remark,
			"vat_perc":           soDt.VatPerc,
			"vat_perc_am":        soDt.VatPercAm,
			"pph23_perc":         soDt.Pph23Perc,
			"pph23_perc_am":      soDt.Pph23PercAm,
			"markup_perc":        soDt.MarkupPerc,
			"markup_perc_am":     soDt.MarkupPercAm,
			"is_vat":             soDt.IsVat,
			"is_pph23":           soDt.IsPph23,
			"is_lock_markup":     soDt.IsLockMarkup,
			"is_lock_price_sell": soDt.IsLockPriceSell,
			"qty_out":            soDt.QtyOut,
			"qty":                soDt.Qty,
			"price_sell":         soDt.PriceSell,
			"price_buy":          soDt.PriceBuy,
			"subtotal_sell":      soDt.SubtotalSell,
			"subtotal_buy":       soDt.SubtotalBuy,
			"disc_am":            soDt.DiscAm,
			"disc_perc":          soDt.DiscPerc,
			"disc_perc_num":      soDt.DiscPercNum,
			"disc_perc_am":       soDt.DiscPercAm,
			"disc_final":         soDt.DiscFinal,
			"disc_type":          soDt.DiscType,
			"total_am":           soDt.TotalAm,
			"updated_by_id":      soDt.UpdatedByID,
			"updated_at":         time.Now(),
		})
	}

	// if err := r.utilRepo.BulkUpdate(tx, "soDts", "id", data, childSpan); err != nil {
	if err := r.utilRepo.Upsert(tx, "so_dts", "id", data, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *SalesOrderRepository) DeleteSoDtsWhereNotIn(ctx *fiber.Ctx, tx *gorm.DB, salesOrderID uint, soDtIDs []uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SoDtRepository-DeleteSoDtsWhereNotIn", opentracing.ChildOf(span.Context()))

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	now := time.Now()
	deletedAt := gorm.DeletedAt{Time: now, Valid: true}

	query := tx.Model(&models.SoDt{}).Where("sales_order_id = ? AND deleted_at IS NULL", salesOrderID)

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

// func (r *SalesOrderRepository) GetSoDtsBySalesOrderIDs(ctx *fiber.Ctx, salesOrderIDs uint, span opentracing.Span) ([]dtos.SalesOrderSoDtListDTO, error) {
func (r *SalesOrderRepository) GetSoDtsBySalesOrderIDs(ctx *fiber.Ctx, tx *gorm.DB, salesOrderIDs []uint, span opentracing.Span) ([]dtos.SalesOrderSoDtListDTO, error) {
	childSpan := opentracing.StartSpan("SalesOrderRepository-GetSoDtsBySalesOrderIDs", opentracing.ChildOf(span.Context()))

	soDts := []dtos.SalesOrderSoDtListDTO{}

	query := `SELECT sd.id, sd.sales_order_id, sd.product_uuid,
		sd.item_unit_id, sd.vat_id, sd.ref_id, sd.item_id, sd.ref_type, sd.item_type, sd.gen_code, sd.remark, sd.vat_perc, sd.qty_out, sd.qty, sd.price_sell, sd.price_buy, sd.subtotal_sell, sd.subtotal_buy, sd.disc_am, sd.disc_perc, sd.disc_perc_num, sd.disc_perc_am, sd.disc_final, sd.disc_type, sd.total_am, sd.created_by_id, sd.updated_by_id, sd.deleted_by_id, sd.created_at, sd.updated_at, sd.deleted_at,
		sd.vat_perc, sd.vat_perc_am, sd.pph23_perc, sd.pph23_perc_am, sd.markup_perc, sd.markup_perc_am, sd.is_vat, sd.is_pph23, sd.is_lock_price_sell, sd.is_lock_markup,
		sd.created_at, sd.updated_at, sd.deleted_at,

		p.customer_id,

		sd.id as so_dt_id,
		isg.id as item_sub_group_id,
		ig.id as item_group_id,
		isg.name as item_sub_group_name,
		ig.name as item_group_name,
		u.name as unit_name,
		pi.name as item_name,
		pi.code as item_code,

		q.quo_no as ref_num,

		cu.name as created_by_name,
		uu.name as updated_by_name

	FROM so_dts sd
	LEFT JOIN sales_orders p ON sd.sales_order_id = p.id
	LEFT JOIN products pi ON sd.item_id = pi.id
	LEFT JOIN item_units iu ON sd.item_unit_id = iu.id
	LEFT JOIN mix_values u ON iu.unit_id = u.id
	LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
	LEFT JOIN mix_values ig ON isg.parent_id = ig.id
	LEFT JOIN users cu ON sd.created_by_id = cu.id
	LEFT JOIN users uu ON sd.updated_by_id = uu.id
	LEFT JOIN quo_dts qd ON qd.id = sd.ref_id AND sd.ref_type = 'quotations'
	LEFT JOIN quotations q ON q.id = qd.quotation_id
	WHERE sd.deleted_at IS NULL`

	var args []interface{}
	i := 1

	if len(salesOrderIDs) > 0 {
		query += " AND sd.sales_order_id = ANY($1)"
		args = append(args, pq.Array(salesOrderIDs))
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

// func (r *SalesOrderRepository) GetSoDtsBySalesOrderIDs(ctx *fiber.Ctx, salesOrderIDs uint, span opentracing.Span) ([]dtos.SalesOrderSoDtListDTO, error) {
func (r *SalesOrderRepository) GetUpdatedSoDtsBySalesOrderIDs(ctx *fiber.Ctx, tx *gorm.DB, salesOrderIDs []uint, span opentracing.Span) ([]dtos.SalesOrderSoDtListUpdateDTO, error) {
	childSpan := opentracing.StartSpan("SalesOrderRepository-GetSoDtsBySalesOrderIDs", opentracing.ChildOf(span.Context()))

	soDts := []dtos.SalesOrderSoDtListUpdateDTO{}

	query := `SELECT sd.id, sd.sales_order_id, sd.product_uuid,
		sd.item_unit_id, sd.vat_id, sd.ref_id, sd.item_id, sd.ref_type, sd.item_type, sd.gen_code, sd.remark, sd.qty_out, sd.qty, sd.price_sell, sd.price_buy, sd.subtotal_sell, sd.subtotal_buy, sd.disc_am, sd.disc_perc, sd.disc_perc_num, sd.disc_perc_am, sd.disc_final, sd.disc_type, sd.total_am, sd.created_by_id, sd.updated_by_id, sd.deleted_by_id, sd.created_at, sd.updated_at, sd.deleted_at,
		sd.vat_perc, sd.vat_perc_am, sd.pph23_perc, sd.pph23_perc_am, sd.markup_perc, sd.markup_perc_am, sd.is_vat, sd.is_pph23, sd.is_lock_price_sell, sd.is_lock_markup,
		sd.created_at, sd.updated_at, sd.deleted_at,

		sd.id as so_dt_id,
		isg.id as item_sub_group_id,
		ig.id as item_group_id,
		isg.name as item_sub_group_name,
		ig.name as item_group_name,
		u.name as unit_name,
		pi.name as item_name,
		pi.code as item_code,

		cu.name as created_by_name,
		uu.name as updated_by_name

	FROM so_dts sd
	LEFT JOIN sales_orders p ON sd.sales_order_id = p.id
	LEFT JOIN products pi ON sd.item_id = pi.id
	LEFT JOIN item_units iu ON sd.item_unit_id = iu.id
	LEFT JOIN mix_values u ON iu.unit_id = u.id
	LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
	LEFT JOIN mix_values ig ON isg.parent_id = ig.id
	LEFT JOIN users cu ON sd.created_by_id = cu.id
	LEFT JOIN users uu ON sd.updated_by_id = uu.id
	WHERE sd.deleted_at IS NULL`

	var args []interface{}
	i := 1

	if len(salesOrderIDs) > 0 {
		query += " AND sd.sales_order_id = ANY($1)"
		args = append(args, pq.Array(salesOrderIDs))
		i++

	}

	// GORM Raw
	if err := tx.Raw(query, args...).Scan(&soDts).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return soDts, nil
}

func (r *SalesOrderRepository) CreateSoDtBoms(tx *gorm.DB, soDtBoms []map[string]interface{}, salesOrderID uint, span opentracing.Span) (*gorm.DB, error) {
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

func (r *SalesOrderRepository) DeleteSoDtBomsWhereNotIn(ctx *fiber.Ctx, tx *gorm.DB, salesOrderID uint, soDtBomIDs []uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SoDtRepository-DeleteSoDtBomsWhereNotIn", opentracing.ChildOf(span.Context()))

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	now := time.Now()
	deletedAt := gorm.DeletedAt{Time: now, Valid: true}

	query := tx.Model(&models.SoDtBom{}).Where("sales_order_id = ? AND deleted_at IS NULL", salesOrderID)

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

func (r *SalesOrderRepository) GetSoDtsBomBySalesOrders(ctx *fiber.Ctx, filters map[string]string, salesOrderIDs []uint, span opentracing.Span) ([]dtos.SalesOrderSoDtBomListDTO, error) {
	childSpan := opentracing.StartSpan("SalesOrderRepository-GetSoDtsBomBySalesOrders", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	soDtBoms := []dtos.SalesOrderSoDtBomListDTO{}

	filterDBColumnKey := []string{
		"so.po_buyer_no", "so.sales_order_no", "so.remark",
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
		condition += fmt.Sprintf(" AND id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"status":        "so.status",
		"customer_id":   "so.customer_id",
		"order_type_id": "so.order_type_id",
		"currency_id":   "so.currency_id",
		"vat_id":        "so.vat_id",
		"payment_id":    "so.payment_id",
		"pph23_id":      "so.pph23_id",
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

	if len(salesOrderIDs) > 0 {
		condition += fmt.Sprintf(" AND sdb.sales_order_id = ANY($%d)", i)
		// countQuery += fmt.Sprintf(" AND so.sales_order_id = ANY($%d)", i)
		args = append(args, pq.Array(salesOrderIDs))
		i++
	}

	filterKeyLike := map[string]string{
		"po_buyer_no":    "so.po_buyer_no",
		"sales_order_no": "so.sales_order_no",
		"remark":         "so.remark",
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
				LEFT JOIN so_dts sd ON sd.id = sdb.so_dt_id
				LEFT JOIN sales_orders so ON so.id = sd.sales_order_id
				LEFT JOIN products pi ON sd.item_id = pi.id
				LEFT JOIN item_units iu ON sd.item_unit_id = iu.id
				LEFT JOIN mix_values u ON iu.unit_id = u.id
				LEFT JOIN products it ON sdb.item_id = it.id
				LEFT JOIN mix_values isg ON it.item_sub_group_id = isg.id
				LEFT JOIN mix_values ig ON isg.parent_id = ig.id

        LEFT JOIN users cu ON so.created_by_id = cu.id
        LEFT JOIN users uu ON so.updated_by_id = uu.id
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

	// orderColumn := utils.GetStringOrDefault(filters["order_column"], "po_buyer_no")
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

func (r *SalesOrderRepository) GetRefIndexQuoDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefIndexQuoDtListDTO, int, error) {

	childSpan := opentracing.StartSpan("SalesOrderRepository-GetRefIndexQuoDts", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	products := []dtos.RefIndexQuoDtListDTO{}

	var total int

	filterDBColumnKey := []string{
		"q.quo_no", "q.title", "q.remark",
		"pi.name",
		"it.name",
		"qd.remark",
		"qd.gen_code",
		"qdb.remark",
		"qdb.gen_code",
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
		"product_id":    "qd.item_id",
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
		"product_ids":        "qd.item_id",
		"quotation_ids":      "q.id",
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
		"vat_ids": []string{"q.vat_id", "qd.vat_id"},
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
        SELECT DISTINCT ON (qd.id)
					qd.id, qd.quotation_id, qd.product_uuid,
					qd.item_unit_id, qd.vat_id, qd.ref_id, qd.item_id, qd.item_unit_id, qd.ref_type, qd.item_type, qd.gen_code, qd.remark, qd.vat_perc, qd.qty_so, qd.qty, qd.price_sell, qd.price_buy, qd.subtotal_sell, qd.subtotal_buy, qd.disc_am, qd.disc_perc, qd.disc_perc_num, qd.disc_perc_am, qd.disc_final, qd.disc_type, qd.total_am, qd.created_by_id, qd.updated_by_id, qd.deleted_by_id,
					qd.vat_perc, qd.vat_perc_am, qd.pph23_perc, qd.pph23_perc_am, qd.markup_perc, qd.markup_perc_am, qd.is_vat, qd.is_pph23, qd.is_lock_price_sell, qd.is_lock_markup,
					qd.created_at, qd.updated_at, qd.deleted_at,
					q.vat_id as head_vat_id,
					q.pph23_id as head_pph23_id,
					q.vat_perc as head_vat_perc,
					q.pph23_perc as head_pph23_perc,
					q.disc_am as head_disc_am,
					q.disc_perc as head_disc_perc,
					q.markup_perc as head_markup_perc,
					q.remark as head_remark,
					q.is_vat as head_is_vat,
					q.quo_no as ref_num,

					q.quo_no,
					q.due_at,
					q.customer_id,
					q.order_type_id,
					q.currency_id,
					q.exchange_rate,

					qd.id as quo_dt_id,
					isg.id as item_sub_group_id,
					ig.id as item_group_id,
					isg.name as item_sub_group_name,
					ig.name as item_group_name,
					u.name as unit_name,
					c.name as customer_name,
					pi.name as item_name,
					pi.code as item_code,
					pi.sku as item_sku,

					cu.name as created_by_name,
					uu.name as updated_by_name

				FROM quo_dts qd
				LEFT JOIN quotations q ON qd.quotation_id = q.id
				LEFT JOIN products pi ON qd.item_id = pi.id
				LEFT JOIN item_units iu ON qd.item_unit_id = iu.id
				LEFT JOIN quo_dt_boms qdb ON qdb.quo_dt_id = qd.id
				LEFT JOIN products it ON qdb.item_id = it.id
				LEFT JOIN customers c ON q.customer_id = c.id
				LEFT JOIN so_dts sd ON qd.id = sd.ref_id AND sd.ref_type = 'quotations'

				LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
				LEFT JOIN mix_values ig ON isg.parent_id = ig.id
				LEFT JOIN mix_values u ON iu.unit_id = u.id

        LEFT JOIN users cu ON q.created_by_id = cu.id
        LEFT JOIN users uu ON q.updated_by_id = uu.id
				WHERE 1=1 AND (q.status = 'WAITING')` + condition + queryGlobal + `
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "quo_no")
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

func (r *SalesOrderRepository) GetRefQuoDtsBomByQuoDtIDs(ctx *fiber.Ctx, filters map[string]string, quotationIDs []uint, span opentracing.Span) ([]dtos.QuotationQuoDtBomListDTO, error) {
	childSpan := opentracing.StartSpan("QuotationRepository-GetQuoDtsBomByQuotations", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	quoDtBoms := []dtos.QuotationQuoDtBomListDTO{}

	filterDBColumnKey := []string{
		"q.quo_no", "q.title", "q.remark",
		"pi.name",
		"it.name",
		"qd.remark",
		"qd.gen_code",
		"qdb.remark",
		"qdb.gen_code",
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
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":        "q.customer_id",
		"order_type_ids":      "q.order_type_id",
		"currency_ids":        "q.currency_id",
		"payment_ids":         "q.payment_id",
		"pph23_ids":           "q.pph23_id",
		"quo_dt_ref_ids":      "qd.ref_id",
		"quo_dt_bom_item_ids": "qdb.item_id",
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
		"vat_ids": []string{"q.vat_id", "qd.vat_id"},
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
		condition += fmt.Sprintf(" AND qdb.quotation_id = ANY($%d)", i)
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
        SELECT DISTINCT ON (qdb.id)
					qdb.id, qdb.product_uuid, qdb.quotation_id, qdb.quo_dt_id, qdb.product_id, qdb.item_id, qdb.item_unit_id, qdb.gen_code, qdb.remark, qdb.qty, qdb.price_sell, qdb.price_buy, qdb.subtotal_sell, qdb.subtotal_buy, qdb.created_by_id, qdb.updated_by_id, qdb.deleted_by_id, qdb.created_at, qdb.updated_at, qdb.deleted_at,
					qdb.id as quo_dt_bom_id,
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

        FROM quo_dt_boms qdb
				LEFT JOIN quo_dts qd ON qd.id = qdb.quo_dt_id
				LEFT JOIN quotations q ON q.id = qd.quotation_id
				LEFT JOIN products pi ON qd.item_id = pi.id
				LEFT JOIN item_units iu ON qd.item_unit_id = iu.id
				LEFT JOIN mix_values u ON iu.unit_id = u.id
				LEFT JOIN products it ON qdb.item_id = it.id
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

func (r *SalesOrderRepository) GetQuoDtQtyUpdate(ctx *fiber.Ctx, tx *gorm.DB, filters map[string]string, span opentracing.Span) ([]dtos.GetQuoDtQtyUpdateDTO, error) {

	childSpan := opentracing.StartSpan("SalesOrderRepository-GetQuoDtQtyUpdate", opentracing.ChildOf(span.Context()))

	products := []dtos.GetQuoDtQtyUpdateDTO{}

	var args []interface{}

	// i := 1
	condition := ""

	condition += fmt.Sprintf(" AND qd.id IN (%s)", filters["ids"])

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (qd.id)
					qd.id, qd.id as quo_dt_id, qd.qty_so,
					q.id as quotation_id

				FROM quo_dts qd
				LEFT JOIN quotations q ON qd.quotation_id = q.id
				WHERE 1=1` + condition + `
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

func (r *SalesOrderRepository) BulkUpdateQuoDtsQty(ctx *fiber.Ctx, tx *gorm.DB, quoDts []map[string]interface{}, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesOrderRepository-BulkUpdateQuoDtsQty", opentracing.ChildOf(span.Context()))

	if err := r.utilRepo.Upsert(tx, "quo_dts", "id", quoDts, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

// UpdateQuoStatus
func (r *SalesOrderRepository) UpdateQuoStatus(ctx *fiber.Ctx, tx *gorm.DB, params dtos.UpdateQuotationStatusRequest, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesOrderRepository-UpdateQuoStatus", opentracing.ChildOf(span.Context()))

	updateParam := map[string]interface{}{
		"id":     params.ID,
		"status": params.Status,
	}

	// upsert
	if err := r.utilRepo.Upsert(tx, "quotations", "id", []map[string]interface{}{updateParam}, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

// GetQuotationID
func (r *SalesOrderRepository) GetQuotationIDBySalesOrderID(ctx *fiber.Ctx, tx *gorm.DB, params map[string]string, span opentracing.Span) (uint, error) {
	childSpan := opentracing.StartSpan("SalesOrderRepository-GetQuotationID", opentracing.ChildOf(span.Context()))

	var quotationID uint

	baseQuery := `
		FROM (
			SELECT DISTINCT ON (q.id)
				q.id
			FROM quotations q
			LEFT JOIN quo_dts qd ON q.id = qd.quotation_id
			LEFT JOIN so_dts sd ON qd.id = sd.ref_id AND sd.ref_type = 'quotations'
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

// GetCustomerSalesOrderCreatedThisMonth
func (r *SalesOrderRepository) GetCustomerSalesOrderCreatedThisMonth(ctx *fiber.Ctx, tx *gorm.DB, customerID uint, span opentracing.Span) (int, error) {
	childSpan := opentracing.StartSpan("SalesOrderRepository-GetCustomerSalesOrderCreatedThisMonth", opentracing.ChildOf(span.Context()))

	var total int

	baseQuery := `
		FROM (
			SELECT COUNT(*) as total
			FROM sales_orders so
			WHERE so.customer_id = $1 AND so.created_at >= date_trunc('month', CURRENT_DATE)
			AND so.deleted_at IS NULL
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

// CreateSchedule
func (r *SalesOrderRepository) CreateSchedule(ctx *fiber.Ctx, tx *gorm.DB, schedule *models.Schedule, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesOrderRepository-CreateSchedule", opentracing.ChildOf(span.Context()))

	if err := tx.Create(schedule).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)

		return nil, err
	}

	return tx, nil
}

// CreateScheduleSteps
func (r *SalesOrderRepository) CreateScheduleSteps(ctx *fiber.Ctx, tx *gorm.DB, scheduleSteps []*models.ScheduleTask, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesOrderRepository-CreateScheduleSteps", opentracing.ChildOf(span.Context()))

	if err := tx.Create(&scheduleSteps).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)

		return nil, err
	}

	return tx, nil
}

// CreateScheduleTasks
func (r *SalesOrderRepository) CreateScheduleTasks(ctx *fiber.Ctx, tx *gorm.DB, scheduleTasks []*models.ScheduleTask, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesOrderRepository-CreateScheduleTasks", opentracing.ChildOf(span.Context()))

	if err := tx.Create(&scheduleTasks).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)

		return nil, err
	}

	return tx, nil
}

// GetScheduleTasksByScheduleID
func (r *SalesOrderRepository) GetScheduleTasksByScheduleID(ctx *fiber.Ctx, filters map[string]string, scheduleIDs []uint, span opentracing.Span) ([]dtos.ScheduleTaskListDTO, error) {
	childSpan := opentracing.StartSpan("SalesOrderRepository-GetScheduleTasksByScheduleID", opentracing.ChildOf(span.Context()))

	scheduleTasks := []dtos.ScheduleTaskListDTO{}

	filterDBColumnKey := []string{
		"st.title", "st.color", "st.remark",
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
		"schedule_id": "st.schedule_id",
		"entity_id":   "st.entity_id",
		"assignee_id": "st.assignee_id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		// "customer_ids":       "so.customer_id",
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
		// "vat_ids": []string{"so.vat_id", "sd.vat_id"},
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

	if len(scheduleIDs) > 0 {
		condition += fmt.Sprintf(" AND st.schedule_id = ANY($%d)", i)
		args = append(args, pq.Array(scheduleIDs))
		i++
	}

	filterKeyLike := map[string]string{
		"title":  "st.title",
		"color":  "st.color",
		"remark": "st.remark",
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
        SELECT DISTINCT ON (st.id)
					st.id, st.schedule_id, st.assignee_id, st.parent_id, st.entity_id, st.entity_type, st.uuid, st.parent_uuid, st.title, st.remark, st.order_item, st.color, st.is_checked,
					
					TO_CHAR(st.start_at, 'YYYY-MM-DD') as start_at,
					TO_CHAR(st.end_at, 'YYYY-MM-DD') as end_at,

					st.created_by_id, st.updated_by_id, st.deleted_by_id, st.created_at, st.updated_at, st.deleted_at,

					cu.name as created_by_name,
					uu.name as updated_by_name

				FROM schedule_tasks st

        LEFT JOIN users cu ON st.created_by_id = cu.id
        LEFT JOIN users uu ON st.updated_by_id = uu.id
				WHERE 1=1` + condition + queryGlobal + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	selectSpan := opentracing.StartSpan("SelectQuery", opentracing.ChildOf(childSpan.Context()))

	err := r.sqlDB.SelectContext(ctx.Context(), &scheduleTasks, query, args...)
	if err != nil {
		selectSpan.LogKV("query", query)
		utils.LogErrors(selectSpan, err)
		return nil, err
	}

	return scheduleTasks, nil
}

func (r *SalesOrderRepository) UpdateSalesOrderSchedule(tx *gorm.DB, salesOrderSchedule *models.Schedule, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesOrderRepository-UpdateSalesOrderSchedule", opentracing.ChildOf(span.Context()))

	if err := tx.Where("id = ?", salesOrderSchedule.ID).Select("*").Omit(
		"created_at", "created_by_id",
	).Updates(salesOrderSchedule).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *SalesOrderRepository) DeleteScheduleStepsWhereNotIn(ctx *fiber.Ctx, tx *gorm.DB, scheduleID uint, scheduleStepIDs []uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SoDtRepository-DeleteScheduleStepsWhereNotIn", opentracing.ChildOf(span.Context()))

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	now := time.Now()
	deletedAt := gorm.DeletedAt{Time: now, Valid: true}

	query := tx.Model(&models.ScheduleTask{}).Where("schedule_id = ? AND entity_type = 'steps' AND deleted_at IS NULL", scheduleID)

	if len(scheduleStepIDs) > 0 {
		query = query.Where("id NOT IN (?)", scheduleStepIDs)
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

// bulk/batch update schedule steps
func (r *SalesOrderRepository) UpdateScheduleSteps(ctx *fiber.Ctx, tx *gorm.DB, scheduleSteps []*models.ScheduleTask, userID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SoDtRepository-UpdateScheduleSteps", opentracing.ChildOf(span.Context()))

	data := make([]map[string]interface{}, 0)
	for _, scheduleStep := range scheduleSteps {
		data = append(data, map[string]interface{}{
			"id":            scheduleStep.ID,
			"schedule_id":   scheduleStep.ScheduleID,
			"assignee_id":   scheduleStep.AssigneeID,
			"parent_id":     scheduleStep.ParentID,
			"entity_id":     scheduleStep.EntityID,
			"entity_type":   scheduleStep.EntityType,
			"uuid":          scheduleStep.UUID,
			"parent_uuid":   scheduleStep.ParentUUID,
			"title":         scheduleStep.Title,
			"remark":        scheduleStep.Remark,
			"order_item":    scheduleStep.OrderItem,
			"color":         scheduleStep.Color,
			"is_checked":    scheduleStep.IsChecked,
			"start_at":      scheduleStep.StartAt,
			"end_at":        scheduleStep.EndAt,
			"updated_by_id": userID,
			"updated_at":    time.Now(),
		})
	}

	// if err := r.utilRepo.BulkUpdate(tx, "scheduleSteps", "id", data, childSpan); err != nil {
	if err := r.utilRepo.Upsert(tx, "schedule_tasks", "id", data, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *SalesOrderRepository) GetUpdatedScheduleStepsBySalesOrderScheduleIDs(ctx *fiber.Ctx, tx *gorm.DB, salesOrderScheduleIDs []uint, span opentracing.Span) ([]dtos.UpdatedScheduleStepListDTO, error) {
	childSpan := opentracing.StartSpan("SalesOrderRepository-GetSoDtsBySalesOrderIDs", opentracing.ChildOf(span.Context()))

	steps := []dtos.UpdatedScheduleStepListDTO{}

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (st.id)
					st.id, st.schedule_id, st.assignee_id, st.parent_id, st.entity_id, st.entity_type, st.uuid, st.parent_uuid, st.title, st.remark, st.order_item, st.color, st.is_checked,
					
					TO_CHAR(st.start_at, 'YYYY-MM-DD') as start_at,
					TO_CHAR(st.end_at, 'YYYY-MM-DD') as end_at,

					st.created_by_id, st.updated_by_id, st.deleted_by_id, st.created_at, st.updated_at, st.deleted_at,

					cu.name as created_by_name,
					uu.name as updated_by_name

				FROM schedule_tasks st

        LEFT JOIN users cu ON st.created_by_id = cu.id
        LEFT JOIN users uu ON st.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	var args []interface{}
	i := 1

	if len(salesOrderScheduleIDs) > 0 {
		query += " AND schedule_id = ANY($1)"
		args = append(args, pq.Array(salesOrderScheduleIDs))
		i++

	}

	// GORM Raw
	if err := tx.Raw(query, args...).Scan(&steps).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return steps, nil
}

func (r *SalesOrderRepository) DeleteScheduleTasksWhereNotIn(ctx *fiber.Ctx, tx *gorm.DB, salesOrderID uint, soDtBomIDs []uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SoDtRepository-DeleteScheduleTasksWhereNotIn", opentracing.ChildOf(span.Context()))

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	now := time.Now()
	deletedAt := gorm.DeletedAt{Time: now, Valid: true}

	query := tx.Model(&models.ScheduleTask{}).Where("schedule_id = ? AND entity_type = 'tasks' AND deleted_at IS NULL", salesOrderID)

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
func (r *SalesOrderRepository) UpdateScheduleTasks(tx *gorm.DB, tasks []map[string]interface{}, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SoDtRepository-UpdateScheduleTasks", opentracing.ChildOf(span.Context()))

	if err := r.utilRepo.Upsert(tx, "schedule_tasks", "id", tasks, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}
