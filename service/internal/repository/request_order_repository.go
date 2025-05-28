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
)

type RequestOrderRepository struct {
	db       *gorm.DB
	sqlDB    *sqlx.DB
	utilRepo *UtilRepository
	tracer   opentracing.Tracer
}

func NewRequestOrderRepository(db *gorm.DB, sqlDB *sqlx.DB, utilRepo *UtilRepository, tracer opentracing.Tracer) *RequestOrderRepository {
	return &RequestOrderRepository{
		db:       db,
		sqlDB:    sqlDB,
		tracer:   tracer,
		utilRepo: utilRepo,
	}
}

func (r *RequestOrderRepository) GetRequestOrders(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RequestOrderListDTO, int, error) {
	childSpan := opentracing.StartSpan("RequestOrderRepository-GetRequestOrders", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	requestOrders := []dtos.RequestOrderListDTO{}

	var total int

	filterDBColumnKey := []string{
		"ro.request_no", "ro.remark", "ro.requested",
		"b.name",
		"w.name",
		"rodt.remark",
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

	if value, ok := filters["status"]; ok && value != "" {
		condition += fmt.Sprintf(" AND ro.status = $%d", i)
		args = append(args, value)
		i++
	}

	filterKey := map[string]string{
		"warehouse_id": "ro.warehouse_id",
		"customer_id":  "so.customer_id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"warehouse_ids": "ro.warehouse_id",
		"customer_ids":  "so.customer_id",
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

	// And one for array conditions with OR
	filterIDsOrArrayKey := map[string][]string{
		"product_ids": {"rodt.item_id", "rodt.product_id"},
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

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (ro.id)
					ro.id, ro.branch_id, ro.warehouse_id,
					ro.request_no, ro.remark, ro.requested, ro.status,
					ro.rev_no, ro.grand_total_order_product_qty, ro.grand_total_order_item_qty, ro.grand_total_wh_qty, ro.grand_total_req_qty, 
					ro.created_by_id, ro.updated_by_id, ro.deleted_by_id, ro.created_at, ro.updated_at, ro.deleted_at,
					TO_CHAR(ro.request_date, 'YYYY-MM-DD') as request_date,

					b.name as branch_name,
					w.name as warehouse_name,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM request_orders ro
				LEFT JOIN request_order_dts rodt ON rodt.request_order_id = ro.id
				LEFT JOIN branches b ON ro.branch_id = b.id
				LEFT JOIN mix_values w ON ro.warehouse_id = w.id
				LEFT JOIN so_dts sodt ON rodt.ref_id = sodt.id AND rodt.ref_type = 'so'
				LEFT JOIN sales_orders so ON so.id = sodt.sales_order_id

        LEFT JOIN users cu ON ro.created_by_id = cu.id
        LEFT JOIN users uu ON ro.updated_by_id = uu.id
				WHERE 1=1` + condition + queryGlobal + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery

	for key, value := range filters {
		switch key {
		case "request_no", "remark", "requested":
			if value != "" {
				query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				countQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

	if startDate, ok := filters["start_date"]; ok && startDate != "" {
		query += fmt.Sprintf(" AND request_date >= $%d", i)
		countQuery += fmt.Sprintf(" AND request_date >= $%d", i)
		args = append(args, startDate)
		i++
	}

	if endDate, ok := filters["end_date"]; ok && endDate != "" {
		query += fmt.Sprintf(" AND request_date <= $%d", i)
		countQuery += fmt.Sprintf(" AND request_date <= $%d", i)
		args = append(args, endDate)
		i++
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "request_date")
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

		err := r.sqlDB.SelectContext(ctx.Context(), &requestOrders, query, args...)
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

	return requestOrders, total, nil
}

func (r *RequestOrderRepository) GetRequestOrderByID(ctx *fiber.Ctx, params *dtos.GetRequestOrderParams, tx *gorm.DB, span opentracing.Span) (*dtos.RequestOrderDetailDTO, error) {
	childSpan := opentracing.StartSpan("RequestOrderRepository-GetRequestOrderByID", opentracing.ChildOf(span.Context()))
	var requestOrder dtos.RequestOrderDetailDTO

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (ro.id)
            ro.id, ro.branch_id, ro.warehouse_id,
            ro.request_no, ro.remark, ro.requested, ro.status, 
            ro.rev_no, ro.grand_total_order_product_qty, ro.grand_total_order_item_qty, ro.grand_total_wh_qty, ro.grand_total_req_qty, 
            ro.created_by_id, ro.updated_by_id, ro.deleted_by_id, ro.created_at, ro.updated_at, ro.deleted_at,
            TO_CHAR(ro.request_date, 'YYYY-MM-DD') as request_date,

            br.company_profile_id,
            w.name as warehouse_name,
						rodt.qty_po,

            cu.name as created_by_name,
            uu.name as updated_by_name

        FROM request_orders ro
        LEFT JOIN request_order_dts rodt ON rodt.request_order_id = ro.id
        LEFT JOIN users cu ON ro.created_by_id = cu.id
        LEFT JOIN users uu ON ro.updated_by_id = uu.id
		LEFT JOIN branches br ON ro.branch_id = br.id
        LEFT JOIN mix_values w ON ro.warehouse_id = w.id
    ) AS alias WHERE 1=1`

	if params.IsDeleted != nil && *params.IsDeleted == 1 {
		baseQuery += " AND deleted_at IS NOT NULL"
	} else {
		baseQuery += " AND deleted_at IS NULL"
	}

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

	if err := r.sqlDB.Get(&requestOrder, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		childSpan.LogKV("query", query)
		return nil, err
	}

	return &requestOrder, nil
}

func (r *RequestOrderRepository) GetUpdatedRequestOrderDts(ctx *fiber.Ctx, requestOrderID uint, span opentracing.Span) ([]dtos.RequestOrderDtListUpdateDTO, error) {
	childSpan := opentracing.StartSpan("RequestOrderRepository-GetUpdatedRequestOrderDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	requestOrderDts := []dtos.RequestOrderDtListUpdateDTO{}

	query := `
	SELECT 
		rodt.id, rodt.product_uuid, rodt.request_order_id, rodt.item_unit_id, 
		rodt.ref_id, rodt.product_id, rodt.item_id, rodt.ref_type, rodt.product_type, rodt.remark, 
		rodt.product_name, rodt.item_name, rodt.unit_name, rodt.price_sell,
		rodt.order_product_qty, rodt.order_item_qty, rodt.wh_qty, rodt.req_qty,
		rodt.created_by_id, rodt.updated_by_id, rodt.deleted_by_id, 
		rodt.created_at, rodt.updated_at, rodt.deleted_at,
		
		rodt.id as request_order_dt_id,
		isg.id as item_sub_group_id,
		ig.id as item_group_id,
		isg.name as item_sub_group_name,
		ig.name as item_group_name,
		u.name as unit_name,
		p.name as item_name,
		p.code as item_code,
		
		cu.name as created_by_name,
		uu.name as updated_by_name
	FROM request_order_dts rodt
	LEFT JOIN products p ON rodt.product_id = p.id
	LEFT JOIN item_units iu ON rodt.item_unit_id = iu.id
	LEFT JOIN mix_values u ON iu.unit_id = u.id
	LEFT JOIN mix_values isg ON p.item_sub_group_id = isg.id
	LEFT JOIN mix_values ig ON isg.parent_id = ig.id
	LEFT JOIN users cu ON rodt.created_by_id = cu.id
	LEFT JOIN users uu ON rodt.updated_by_id = uu.id
	WHERE rodt.request_order_id = $1 AND rodt.deleted_at IS NULL
	ORDER BY rodt.id ASC
	`

	err := r.sqlDB.SelectContext(ctx.Context(), &requestOrderDts, query, requestOrderID)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return requestOrderDts, nil
}

func (r *RequestOrderRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *RequestOrderRepository) Rollback() *gorm.DB {
	return r.db.Rollback()
}

func (r *RequestOrderRepository) CreateRequestOrder(tx *gorm.DB, requestOrder *models.RequestOrder, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("RequestOrderRepository-CreateRequestOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Create(&requestOrder)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *RequestOrderRepository) CreateRequestOrderDts(tx *gorm.DB, requestOrderDts []models.RequestOrderDt, span opentracing.Span) (*gorm.DB, []models.RequestOrderDt, error) {
	childSpan := opentracing.StartSpan("RequestOrderRepository-CreateRequestOrderDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Create(&requestOrderDts)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, nil, result.Error
	}

	return tx, requestOrderDts, nil
}

func (r *RequestOrderRepository) GetRequestOrderDts(ctx *fiber.Ctx, requestOrderID uint, isDeleted *int, span opentracing.Span) ([]dtos.RequestOrderDtListDTO, error) {
	childSpan := opentracing.StartSpan("RequestOrderRepository-GetRequestOrderDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	requestOrderDts := []dtos.RequestOrderDtListDTO{}

	realStockCTE := `
		WITH end_date AS (
				SELECT '` + time.Now().Format("2006-01-02") + `'::date AS closing_date
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
		real_stock AS (
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
		`

	query := realStockCTE + `
	SELECT 
		rodt.id, rodt.product_uuid, rodt.request_order_id, rodt.item_unit_id, 
		rodt.ref_id, rodt.product_id, rodt.item_id, rodt.ref_type, rodt.product_type, rodt.remark, 
		rodt.order_product_qty, rodt.order_item_qty, rodt.wh_qty, rodt.req_qty,
		rodt.created_by_id, rodt.updated_by_id, rodt.deleted_by_id, 
		rodt.created_at, rodt.updated_at, rodt.deleted_at,
		
		p_product.name as product_name,
		
		p_product.code as product_code,
		
		p.name as item_name, 
		p.code as item_code,
		
		rodt.price_sell,
		u.name as unit_name,
		
		cu.name as created_by_name,
		uu.name as updated_by_name,

		CASE WHEN rodt.ref_type = 'so' THEN so.sales_order_no ELSE NULL END as ref_num,
		
		CASE 
			WHEN rodt.ref_type = 'so' AND rodt.product_type = 'item' THEN sodt.sales_order_id
			WHEN rodt.ref_type = 'so' AND rodt.product_type = 'product' THEN sodtb_so.sales_order_id
			ELSE NULL 
		END as sales_order_id
	FROM request_order_dts rodt
	LEFT JOIN request_orders ro ON rodt.request_order_id = ro.id
	LEFT JOIN products p ON rodt.item_id = p.id
	LEFT JOIN products p_product ON rodt.product_id = p_product.id
	LEFT JOIN item_units iu ON rodt.item_unit_id = iu.id
	LEFT JOIN mix_values u ON iu.unit_id = u.id
	LEFT JOIN sales_orders so ON rodt.ref_id = so.id AND rodt.ref_type = 'so'
	LEFT JOIN so_dts sodt ON rodt.ref_id = sodt.id AND rodt.ref_type = 'so' AND rodt.product_type = 'item'
	LEFT JOIN so_dt_boms sodtb ON rodt.ref_id = sodtb.id AND rodt.ref_type = 'so' AND rodt.product_type = 'product'
	LEFT JOIN so_dts sodtb_so ON sodtb.so_dt_id = sodtb_so.id
	LEFT JOIN users cu ON rodt.created_by_id = cu.id
	LEFT JOIN users uu ON rodt.updated_by_id = uu.id
	LEFT JOIN real_stock rs ON rodt.item_id = rs.item_id AND ro.warehouse_id = rs.warehouse_id
	WHERE rodt.request_order_id = $1
	`

	if isDeleted != nil && *isDeleted == 1 {
		query += " AND rodt.deleted_at IS NOT NULL"
	} else {
		query += " AND rodt.deleted_at IS NULL"
	}

	query += " ORDER BY rodt.id ASC"

	err := r.sqlDB.SelectContext(ctx.Context(), &requestOrderDts, query, requestOrderID)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return requestOrderDts, nil
}

func (r *RequestOrderRepository) UpdateRequestOrder(tx *gorm.DB, requestOrder *models.RequestOrder, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("RequestOrderRepository-UpdateRequestOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Model(&models.RequestOrder{}).Where("id = ?", requestOrder.ID).Updates(requestOrder)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *RequestOrderRepository) BulkCreateRequestOrderDts(tx *gorm.DB, requestOrderDts []models.RequestOrderDt, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("RequestOrderRepository-BulkCreateRequestOrderDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(requestOrderDts) == 0 {
		return tx, nil
	}

	result := tx.Create(&requestOrderDts)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *RequestOrderRepository) BulkUpdateRequestOrderDts(tx *gorm.DB, requestOrderDts []models.RequestOrderDt, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("RequestOrderRepository-BulkUpdateRequestOrderDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(requestOrderDts) == 0 {
		return tx, nil
	}

	if err := tx.Save(&requestOrderDts).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return tx, err
	}

	return tx, nil
}

func (r *RequestOrderRepository) DeleteRequestOrderDtsByIDs(tx *gorm.DB, requestOrderDtIDs []uint, userID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("RequestOrderRepository-DeleteRequestOrderDtsByIDs", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(requestOrderDtIDs) == 0 {
		return tx, nil
	}

	result := tx.Model(&models.RequestOrderDt{}).Where("id IN ?", requestOrderDtIDs).Updates(map[string]interface{}{
		"deleted_by_id": userID,
		"deleted_at":    time.Now(),
	})
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *RequestOrderRepository) DeleteRequestOrder(tx *gorm.DB, requestOrderID uint, userID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("RequestOrderRepository-DeleteRequestOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Model(&models.RequestOrder{}).Where("id = ?", requestOrderID).Updates(map[string]interface{}{
		"deleted_by_id": userID,
		"deleted_at":    time.Now(),
	})
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *RequestOrderRepository) RestoreRequestOrder(ctx *fiber.Ctx, params *dtos.GetRequestOrderParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("RequestOrderRepository-RestoreRequestOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var requestOrder models.RequestOrder
	if err := tx.Unscoped().Model(&requestOrder).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *RequestOrderRepository) GetRefSalesOrderDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefSalesOrderForRequestOrderListDTO, int, error) {
	childSpan := opentracing.StartSpan("RequestOrderRepository-GetRefSalesOrderDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	soDts := []dtos.RefSalesOrderForRequestOrderListDTO{}

	var total int

	baseCondition := ""
	if filters["specific_ids"] != "" {
		baseCondition += fmt.Sprintf(" AND (sodt.id IN (%s))", filters["specific_ids"])
	} else {
		baseCondition += " AND sales_orders.status NOT IN ('CANCELLED', 'FINISH')"
	}

	if filters["ids"] != "" {
		baseCondition += fmt.Sprintf(" AND sodt.id IN (%s)", filters["ids"])
	}

	// warehouseCondition := ""
	// if value, ok := filters["warehouse_id"]; ok && value != "" {
	// 	warehouseCondition = fmt.Sprintf(" AND sc.warehouse_id = %s", value)
	// }

	realStockCTE := `
		WITH end_date AS (
				SELECT '` + time.Now().Format("2006-01-02") + `'::date AS closing_date
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
		real_stock AS (
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
		`

	baseQueryItems := `
    SELECT
        sodt.id,
        sodt.sales_order_id,
        sodt.item_unit_id,
        sodt.id as ref_id,
        sodt.item_id,
        sodt.item_id as product_id,
        'so' as ref_type,
        sodt.item_type,
        sodt.gen_code,
        sodt.remark,
        sodt.qty,
        sodt.price_sell,
        sodt.created_by_id,
        sodt.updated_by_id,
        sodt.deleted_by_id,
        sodt.created_at,
        sodt.updated_at,
        sodt.deleted_at,
        sales_orders.customer_id,
        sales_orders.order_type_id,
        sales_orders.remark as head_remark,
        sales_orders.sales_order_no,
        sales_orders.po_buyer_no,
        sales_orders.order_at as order_date,
        sales_orders.shipping_at as shipping_date,
        TO_CHAR(sales_orders.due_at, 'YYYY-MM-DD') as due_at,
        c.name as customer_name,
        ot.name as order_type_name,
        p.name as item_name,
        p.code as item_code,
        p.sku as item_sku,
        NULL as product_name,
        NULL as product_code,
        u.name as unit_name,
        cu.name as created_by_name,
        uu.name as updated_by_name,
        sodt.qty as order_product_qty,
        sodt.qty as order_item_qty,
        COALESCE(rs.end_qty, 0) as wh_qty,
        (sodt.qty - COALESCE(rs.end_qty, 0)) as req_qty,
        sales_orders.branch_id
    FROM so_dts sodt
    LEFT JOIN sales_orders ON sodt.sales_order_id = sales_orders.id
    LEFT JOIN customers c ON sales_orders.customer_id = c.id
    LEFT JOIN mix_values ot ON sales_orders.order_type_id = ot.id
    LEFT JOIN products p ON sodt.item_id = p.id
    LEFT JOIN item_units iu ON sodt.item_unit_id = iu.id
    LEFT JOIN mix_values u ON iu.unit_id = u.id
    LEFT JOIN users cu ON sodt.created_by_id = cu.id
    LEFT JOIN users uu ON sodt.updated_by_id = uu.id
    LEFT JOIN real_stock rs ON rs.item_id = sodt.item_id AND rs.warehouse_id = sales_orders.warehouse_id
    WHERE sodt.item_type = 'item'
    AND sodt.deleted_at IS NULL
    ` + baseCondition

	baseQueryProducts := `
    SELECT
        sodt.id,
        sodt.sales_order_id,
        sodtb.item_unit_id,
        sodtb.id as ref_id,
        sodtb.item_id,
        sodt.item_id as product_id,
        'so' as ref_type,
        sodt.item_type,
        sodt.gen_code,
        sodtb.remark,
        sodtb.qty,
        sodtb.price_sell,
        sodt.created_by_id,
        sodt.updated_by_id,
        sodt.deleted_by_id,
        sodt.created_at,
        sodt.updated_at,
        sodt.deleted_at,
        sales_orders.customer_id,
        sales_orders.order_type_id,
        sales_orders.remark as head_remark,
        sales_orders.sales_order_no,
        sales_orders.po_buyer_no,
        sales_orders.order_at as order_date,
        sales_orders.shipping_at as shipping_date,
        TO_CHAR(sales_orders.due_at, 'YYYY-MM-DD') as due_at,
        c.name as customer_name,
        ot.name as order_type_name,
        p_bom.name as item_name,
        p_bom.code as item_code,
        p_bom.sku as item_sku,
        p.name as product_name,
        p.code as product_code,
        u.name as unit_name,
        cu.name as created_by_name,
        uu.name as updated_by_name,
        sodt.qty as order_product_qty,
        (sodt.qty * sodtb.qty) as order_item_qty,
        COALESCE(rs.end_qty, 0) as wh_qty,
        ((sodt.qty * sodtb.qty) - COALESCE(rs.end_qty, 0)) as req_qty,
        sales_orders.branch_id
    FROM so_dts sodt
    LEFT JOIN so_dt_boms sodtb ON sodt.id = sodtb.so_dt_id
    LEFT JOIN sales_orders ON sodt.sales_order_id = sales_orders.id
    LEFT JOIN customers c ON sales_orders.customer_id = c.id
    LEFT JOIN mix_values ot ON sales_orders.order_type_id = ot.id
    LEFT JOIN products p ON sodt.item_id = p.id
    LEFT JOIN products p_bom ON sodtb.item_id = p_bom.id
    LEFT JOIN item_units iu ON sodtb.item_unit_id = iu.id
    LEFT JOIN mix_values u ON iu.unit_id = u.id
    LEFT JOIN users cu ON sodt.created_by_id = cu.id
    LEFT JOIN users uu ON sodt.updated_by_id = uu.id
    LEFT JOIN real_stock rs ON rs.item_id = sodtb.item_id AND rs.warehouse_id = sales_orders.warehouse_id
    WHERE sodt.item_type = 'product'
    AND sodtb.id IS NOT NULL
    AND sodt.deleted_at IS NULL
    ` + baseCondition

	combinedQuery := "(" + baseQueryItems + ") UNION ALL (" + baseQueryProducts + ")"

	query := realStockCTE + " " + "SELECT * FROM (" + combinedQuery + ") AS combined_result WHERE 1=1"
	countQuery := realStockCTE + " " + "SELECT COUNT(*) FROM (" + combinedQuery + ") AS count_alias WHERE 1=1"

	var args []interface{}
	i := 1

	if value, ok := filters["global"]; ok && value != "" {
		query += " AND (sales_order_no ILIKE $1 OR po_buyer_no ILIKE $1 OR customer_name ILIKE $1 OR item_name ILIKE $1 OR item_code ILIKE $1 OR product_name ILIKE $1 OR product_code ILIKE $1)"
		countQuery += " AND (sales_order_no ILIKE $1 OR po_buyer_no ILIKE $1 OR customer_name ILIKE $1 OR item_name ILIKE $1 OR item_code ILIKE $1 OR product_name ILIKE $1 OR product_code ILIKE $1)"
		args = append(args, "%"+value+"%")
		i++
	}

	if value, ok := filters["sales_order_no"]; ok && value != "" {
		query += fmt.Sprintf(" AND sales_order_no ILIKE $%d", i)
		countQuery += fmt.Sprintf(" AND sales_order_no ILIKE $%d", i)
		args = append(args, "%"+value+"%")
		i++
	} else if value, ok := filters["so_no"]; ok && value != "" {
		query += fmt.Sprintf(" AND sales_order_no ILIKE $%d", i)
		countQuery += fmt.Sprintf(" AND sales_order_no ILIKE $%d", i)
		args = append(args, "%"+value+"%")
		i++
	}

	if value, ok := filters["po_buyer_no"]; ok && value != "" {
		query += fmt.Sprintf(" AND po_buyer_no ILIKE $%d", i)
		countQuery += fmt.Sprintf(" AND po_buyer_no ILIKE $%d", i)
		args = append(args, "%"+value+"%")
		i++
	}

	if value, ok := filters["product_name"]; ok && value != "" {
		query += fmt.Sprintf(" AND (item_name ILIKE $%d OR product_name ILIKE $%d)", i, i)
		countQuery += fmt.Sprintf(" AND (item_name ILIKE $%d OR product_name ILIKE $%d)", i, i)
		args = append(args, "%"+value+"%")
		i++
	}

	if value, ok := filters["item_name"]; ok && value != "" {
		query += fmt.Sprintf(" AND item_name ILIKE $%d", i)
		countQuery += fmt.Sprintf(" AND item_name ILIKE $%d", i)
		args = append(args, "%"+value+"%")
		i++
	}

	if value, ok := filters["product_code"]; ok && value != "" {
		query += fmt.Sprintf(" AND (item_code ILIKE $%d OR product_code ILIKE $%d)", i, i)
		countQuery += fmt.Sprintf(" AND (item_code ILIKE $%d OR product_code ILIKE $%d)", i, i)
		args = append(args, "%"+value+"%")
		i++
	}

	if value, ok := filters["item_type"]; ok && value != "" {
		query += fmt.Sprintf(" AND item_type = $%d", i)
		countQuery += fmt.Sprintf(" AND item_type = $%d", i)
		args = append(args, value)
		i++
	}

	filterKey := map[string]string{
		"customer_id":   "customer_id",
		"order_type_id": "order_type_id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" && value != "null" {
			query += fmt.Sprintf(" AND %s = $%d", col, i)
			countQuery += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":   "customer_id",
		"order_type_ids": "order_type_id",
	}

	for key, valueID := range filterIDsKey {
		if value, ok := filters[key]; ok && value != "" {
			ids := strings.Split(value, ",")
			intIDs, err := utils.SplitStringArrayOfInts(ids)
			if err != nil {
				utils.LogErrors(childSpan, err)
				return nil, 0, err
			}

			query += fmt.Sprintf(" AND %s = ANY($%d)", valueID, i)
			countQuery += fmt.Sprintf(" AND %s = ANY($%d)", valueID, i)
			args = append(args, pq.Array(intIDs))
			i++
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "created_at")
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

		err := r.sqlDB.SelectContext(ctx.Context(), &soDts, query, args...)
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

	return soDts, total, nil
}

func (r *RequestOrderRepository) GetRefProductForRequestOrder(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefProductForRequestOrderListDTO, int, error) {
	childSpan := opentracing.StartSpan("RequestOrderRepository-GetRefProductForRequestOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	products := []dtos.RefProductForRequestOrderListDTO{}

	var total int

	baseCondition := ""
	if filters["specific_ids"] != "" {
		baseCondition += fmt.Sprintf(" AND (p.id IN (%s))", filters["specific_ids"])
	}

	if filters["ids"] != "" {
		baseCondition += fmt.Sprintf(" AND p.id IN (%s)", filters["ids"])
	}

	baseQuerySingleItems := `
    SELECT
        p.id,
        iu.id as item_unit_id,
        p.id as ref_id,
        p.id as item_id,
        p.id as product_id,
        'product' as ref_type,
        'item' as item_type,
        p.remark,
        iu.price_sell,
        p.created_by_id,
        p.updated_by_id,
        p.deleted_by_id,
        p.created_at,
        p.updated_at,
        p.deleted_at,
        p.name as item_name,
        p.code as item_code,
        p.sku as item_sku,
        NULL as product_name,
        NULL as product_code,
        u.name as unit_name,
        cu.name as created_by_name,
        uu.name as updated_by_name,
        0 as order_product_qty,
        0 as order_item_qty,
        0 as wh_qty,
        0 as req_qty
    FROM products p
    LEFT JOIN item_units iu ON p.item_unit_id = iu.id
    LEFT JOIN mix_values u ON iu.unit_id = u.id
    LEFT JOIN users cu ON p.created_by_id = cu.id
    LEFT JOIN users uu ON p.updated_by_id = uu.id
    WHERE p.prod_type = 'single'
    AND p.deleted_at IS NULL
    AND p.status = 1
    ` + baseCondition

	baseQueryProducts := `
    SELECT
        p.id,
        iu_bom.id as item_unit_id,
        b.id as ref_id,
        b.product_item_id as item_id,
        p.id as product_id,
        'product' as ref_type,
        'product' as item_type,
        b.remark,
        iu_bom.price_sell,
        p.created_by_id,
        p.updated_by_id,
        p.deleted_by_id,
        p.created_at,
        p.updated_at,
        p.deleted_at,
        p_bom.name as item_name,
        p_bom.code as item_code,
        p_bom.sku as item_sku,
        p.name as product_name,
        p.code as product_code,
        u_bom.name as unit_name,
        cu.name as created_by_name,
        uu.name as updated_by_name,
        0 as order_product_qty,
        0 as order_item_qty,
        0 as wh_qty,
        0 as req_qty
    FROM products p
    LEFT JOIN boms b ON p.id = b.product_id
    LEFT JOIN products p_bom ON b.product_item_id = p_bom.id
    LEFT JOIN item_units iu_bom ON b.item_unit_id = iu_bom.id
    LEFT JOIN mix_values u_bom ON iu_bom.unit_id = u_bom.id
    LEFT JOIN users cu ON p.created_by_id = cu.id
    LEFT JOIN users uu ON p.updated_by_id = uu.id
    WHERE p.prod_type = 'product'
    AND b.id IS NOT NULL
    AND p.deleted_at IS NULL
    AND p.status = 1
    AND p_bom.deleted_at IS NULL
    ` + baseCondition

	combinedQuery := "(" + baseQuerySingleItems + ") UNION ALL (" + baseQueryProducts + ")"

	query := "SELECT * FROM (" + combinedQuery + ") AS combined_result WHERE 1=1"
	countQuery := "SELECT COUNT(*) FROM (" + combinedQuery + ") AS count_alias WHERE 1=1"

	var args []interface{}
	i := 1

	if value, ok := filters["global"]; ok && value != "" {
		query += " AND (item_name ILIKE $1 OR item_code ILIKE $1 OR product_name ILIKE $1 OR product_code ILIKE $1)"
		countQuery += " AND (item_name ILIKE $1 OR item_code ILIKE $1 OR product_name ILIKE $1 OR product_code ILIKE $1)"
		args = append(args, "%"+value+"%")
		i++
	}

	if value, ok := filters["product_code"]; ok && value != "" {
		query += fmt.Sprintf(" AND (item_code ILIKE $%d OR product_code ILIKE $%d)", i, i)
		countQuery += fmt.Sprintf(" AND (item_code ILIKE $%d OR product_code ILIKE $%d)", i, i)
		args = append(args, "%"+value+"%")
		i++
	}

	if value, ok := filters["product_name"]; ok && value != "" {
		query += fmt.Sprintf(" AND (item_name ILIKE $%d OR product_name ILIKE $%d)", i, i)
		countQuery += fmt.Sprintf(" AND (item_name ILIKE $%d OR product_name ILIKE $%d)", i, i)
		args = append(args, "%"+value+"%")
		i++
	}

	if value, ok := filters["item_code"]; ok && value != "" {
		query += fmt.Sprintf(" AND item_code ILIKE $%d", i)
		countQuery += fmt.Sprintf(" AND item_code ILIKE $%d", i)
		args = append(args, "%"+value+"%")
		i++
	}

	if value, ok := filters["item_name"]; ok && value != "" {
		query += fmt.Sprintf(" AND item_name ILIKE $%d", i)
		countQuery += fmt.Sprintf(" AND item_name ILIKE $%d", i)
		args = append(args, "%"+value+"%")
		i++
	}

	if value, ok := filters["item_type"]; ok && value != "" {
		query += fmt.Sprintf(" AND item_type = $%d", i)
		countQuery += fmt.Sprintf(" AND item_type = $%d", i)
		args = append(args, value)
		i++
	}

	filterKey := map[string]string{
		"item_group_id":     "item_group_id",
		"item_sub_group_id": "item_sub_group_id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" && value != "null" {
			query += fmt.Sprintf(" AND %s = $%d", col, i)
			countQuery += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	if !isAdmin && branchID != nil {
		query += fmt.Sprintf(" AND (branch_id = $%d OR branch_id IS NULL)", i)
		countQuery += fmt.Sprintf(" AND (branch_id = $%d OR branch_id IS NULL)", i)
		args = append(args, branchID)
		i++
	}

	if isAdmin && filters["branch_id"] != "" {
		query += fmt.Sprintf(" AND (branch_id = $%d OR branch_id IS NULL)", i)
		countQuery += fmt.Sprintf(" AND (branch_id = $%d OR branch_id IS NULL)", i)
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "created_at")
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

func (r *RequestOrderRepository) LockRequestOrder(tx *gorm.DB, requestOrderID uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("RequestOrderRepository-LockRequestOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	lockQuery := "SELECT id FROM request_orders WHERE id = ? FOR UPDATE"
	if err := tx.Exec(lockQuery, requestOrderID).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *RequestOrderRepository) GetRequestOrderForUpdate(tx *gorm.DB, requestOrderID uint, span opentracing.Span) (*models.RequestOrder, error) {
	childSpan := opentracing.StartSpan("RequestOrderRepository-GetRequestOrderForUpdate", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var requestOrder models.RequestOrder
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", requestOrderID).First(&requestOrder).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return &requestOrder, nil
}

func (r *RequestOrderRepository) GetRequestOrderCreatedThisMonth(ctx *fiber.Ctx, tx *gorm.DB, span opentracing.Span) (int, error) {
	childSpan := opentracing.StartSpan("RequestOrderRepository-GetRequestOrderCreatedThisMonth", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var count int
	query := `
    SELECT COUNT(*) 
    FROM request_orders 
    WHERE EXTRACT(MONTH FROM created_at) = EXTRACT(MONTH FROM CURRENT_DATE) 
    AND EXTRACT(YEAR FROM created_at) = EXTRACT(YEAR FROM CURRENT_DATE)
    `

	err := tx.Raw(query).Scan(&count).Error
	if err != nil {
		utils.LogErrors(childSpan, err)
		return 0, err
	}

	return count, nil
}

func (r *RequestOrderRepository) GetWidgetRequestOrders(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RequestOrderStatusWidget, int, error) {
	childSpan := opentracing.StartSpan("RequestOrderRepository-GetWidgetRequestOrders", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	var widgets []dtos.RequestOrderStatusWidget
	var total int

	filterDBColumnKey := []string{
		"ro.request_no", "ro.remark", "ro.requested",
		"b.name",
		"w.name",
		"rodt.remark",
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

	for key, value := range filters {
		switch key {
		case "request_no", "remark", "requested":
			if value != "" {
				condition += fmt.Sprintf(" AND ro.%s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

	if !isAdmin && branchID != nil {
		condition += fmt.Sprintf(" AND ro.branch_id = $%d", i)
		args = append(args, branchID)
		i++
	}

	if isAdmin && filters["branch_id"] != "" {
		condition += fmt.Sprintf(" AND ro.branch_id = $%d", i)
		args = append(args, filters["branch_id"])
		i++
	}

	filterKey := map[string]string{
		"warehouse_id": "ro.warehouse_id",
		"customer_id":  "so.customer_id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"warehouse_ids": "ro.warehouse_id",
		"customer_ids":  "so.customer_id",
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

	if filters["start_date"] != "" && filters["end_date"] != "" {
		condition += fmt.Sprintf(" AND (ro.request_date BETWEEN $%d AND $%d)", i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
	}

	query := `
    WITH status_values AS (
        SELECT status, row_number() over () as status_order
        FROM (VALUES
            ('TOTAL', 0),
            ('PENDING', 1),
            ('APPROVED', 2),
            ('CANCELLED', 3)
        ) AS s(status)
    ),
    filtered_requests AS (
        SELECT DISTINCT
            ro.id,
            COALESCE(ro.status, 'PENDING') as ro_status,
            ro.grand_total_req_qty as total_qty,
            0 as grand_total
        FROM request_orders ro
        LEFT JOIN request_order_dts rodt ON rodt.request_order_id = ro.id
				LEFT JOIN so_dts sodt ON rodt.ref_id = sodt.id AND rodt.ref_type = 'so'
				LEFT JOIN sales_orders so ON so.id = sodt.sales_order_id
        WHERE ro.deleted_at IS NULL
        ` + condition + queryGlobal + `
    ),
    request_order_stats AS (
        SELECT
            sv.status,
            sv.status_order,
            COUNT(DISTINCT fr.id) as order_count,
            COALESCE(SUM(fr.total_qty), 0) as total_qty
        FROM status_values sv
        LEFT JOIN filtered_requests fr ON
            (sv.status = fr.ro_status) OR
            (sv.status = 'TOTAL')
        GROUP BY sv.status, sv.status_order
    )
    SELECT
        sv.status,
        order_count,
        total_qty
    FROM request_order_stats sv
    ORDER BY status_order
    `

	var selectErr error
	selectSpan := opentracing.StartSpan("SelectQuery", opentracing.ChildOf(childSpan.Context()))

	err := r.sqlDB.SelectContext(ctx.Context(), &widgets, query, args...)
	if err != nil {
		selectSpan.LogKV("query", query)
		utils.LogErrors(selectSpan, err)
		selectErr = err
	}

	if selectErr != nil {
		defer childSpan.Finish()
	}

	if selectErr != nil {
		return nil, 0, selectErr
	}

	return widgets, total, nil
}

func (r *RequestOrderRepository) Commit(tx *gorm.DB) error {
	return tx.Commit().Error
}
