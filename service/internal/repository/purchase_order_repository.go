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

type PurchaseOrderRepository struct {
	db       *gorm.DB
	sqlDB    *sqlx.DB
	utilRepo *UtilRepository
	tracer   opentracing.Tracer
}

func NewPurchaseOrderRepository(db *gorm.DB, sqlDB *sqlx.DB, utilRepo *UtilRepository, tracer opentracing.Tracer) *PurchaseOrderRepository {
	return &PurchaseOrderRepository{
		db:       db,
		sqlDB:    sqlDB,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (r *PurchaseOrderRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *PurchaseOrderRepository) Rollback() *gorm.DB {
	return r.db.Rollback()
}

func (r *PurchaseOrderRepository) CreatePurchaseOrder(tx *gorm.DB, purchaseOrder *models.PurchaseOrder, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-CreatePurchaseOrder", opentracing.ChildOf(span.Context()))
	if err := tx.Create(purchaseOrder).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *PurchaseOrderRepository) CreatePoDts(tx *gorm.DB, poDts []models.PurchaseOrderDt, purchaseOrderID uint, span opentracing.Span) (*gorm.DB, []models.PurchaseOrderDt, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-CreatePoDts", opentracing.ChildOf(span.Context()))

	tx = tx.Create(&poDts)
	if tx.Error != nil {
		utils.LogErrors(childSpan, tx.Error)
		tx.Rollback()
		log.Fatalf("Failed to insert poDts: %v", tx.Error)
	}

	return tx, poDts, nil
}

func (r *PurchaseOrderRepository) CreatePoDtBoms(tx *gorm.DB, poDtBoms []map[string]interface{}, purchaseOrderID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-CreatePoDtBoms", opentracing.ChildOf(span.Context()))

	err := r.utilRepo.Upsert(tx, "purchase_order_dt_boms", "id", poDtBoms, span)
	if err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *PurchaseOrderRepository) GetPurchaseOrderByID(ctx *fiber.Ctx, params *dtos.GetPurchaseOrderParams, tx *gorm.DB, span opentracing.Span) (*dtos.PurchaseOrderDetailDTO, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-GetPurchaseOrderByID", opentracing.ChildOf(span.Context()))
	var purchaseOrder dtos.PurchaseOrderDetailDTO

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (po.id)
					po.id, po.customer_id, po.purchase_type_id, po.currency_id, po.vat_id, po.payment_term_id, po.shipping_term_id, po.pph23_id, po.branch_id, 
					po.po_no, po.po_date, po.delivery_date, po.shipping_destination, po.remark, po.status, po.exchange_rate, po.pph23_percentage,
					po.discount_percentage, po.discount_amount, po.discount_percentage_amount, po.discount_amount_product, po.discount_final_header, po.discount_type,
					po.vat_percentage, po.total_amount_products, po.total_qty, po.subtotal, po.total_discount, po.total_pph23, po.total_vat, po.grand_total, 
					po.created_by_id, po.updated_by_id, po.deleted_by_id, po.created_at, po.updated_at, po.deleted_at,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM purchase_orders po
				LEFT JOIN purchase_order_dts pod ON pod.po_id = po.id
				LEFT JOIN products pi ON pod.product_id = pi.id
				LEFT JOIN item_units iu ON pod.item_unit_id = iu.id
				LEFT JOIN purchase_order_dt_boms podb ON podb.po_dt_id = pod.id
				LEFT JOIN products it ON podb.product_id = it.id

				LEFT JOIN mix_values cur ON po.currency_id = cur.id
				LEFT JOIN mix_values vat ON po.vat_id = vat.id
				LEFT JOIN mix_values pph ON po.pph23_id = pph.id

        LEFT JOIN users cu ON po.created_by_id = cu.id
        LEFT JOIN users uu ON po.updated_by_id = uu.id
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

	isDeletedQuery := ` AND deleted_at IS NULL`
	if params.IsDeleted != nil && *params.IsDeleted == 1 {
		isDeletedQuery = " AND deleted_at IS NOT NULL"
	}

	query += isDeletedQuery

	if err := r.sqlDB.Get(&purchaseOrder, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		childSpan.LogKV("query", query)
		return nil, err
	}

	return &purchaseOrder, nil
}

func (r *PurchaseOrderRepository) GetPoDtsByPurchaseOrderIDs(ctx *fiber.Ctx, tx *gorm.DB, purchaseOrderIDs []uint, span opentracing.Span) ([]dtos.PurchaseOrderPoDtListDTO, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-GetPoDtsByPurchaseOrderIDs", opentracing.ChildOf(span.Context()))

	poDts := []dtos.PurchaseOrderPoDtListDTO{}

	query := `SELECT pod.id, pod.po_id, pod.product_uuid,
		pod.item_unit_id, pod.vat_id, pod.pph23_id, pod.ref_id, pod.product_id, pod.ref_type, pod.product_type, pod.gen_code, pod.remark, 
		pod.need_qty, pod.qty, pod.price, pod.subtotal, pod.discount_amount, pod.discount_percentage, pod.discount_percentage_num, 
		pod.discount_percentage_amount, pod.discount_final, pod.discount_type, pod.is_vat, pod.is_pph23,
		pod.total_amount, pod.created_by_id, pod.updated_by_id, pod.deleted_by_id, pod.created_at, pod.updated_at, pod.deleted_at,

		pod.id as po_dt_id,
		isg.id as item_sub_group_id,
		ig.id as item_group_id,
		isg.name as item_sub_group_name,
		ig.name as item_group_name,
		u.name as unit_name,
		pi.name as item_name,
		pi.code as item_code,

		cu.name as created_by_name,
		uu.name as updated_by_name

	FROM purchase_order_dts pod
	LEFT JOIN purchase_orders po ON pod.po_id = po.id
	LEFT JOIN products pi ON pod.product_id = pi.id
	LEFT JOIN item_units iu ON pod.item_unit_id = iu.id
	LEFT JOIN mix_values u ON iu.unit_id = u.id
	LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
	LEFT JOIN mix_values ig ON isg.parent_id = ig.id
	LEFT JOIN users cu ON pod.created_by_id = cu.id
	LEFT JOIN users uu ON pod.updated_by_id = uu.id
	WHERE pod.deleted_at IS NULL`

	var args []interface{}
	i := 1

	if len(purchaseOrderIDs) > 0 {
		query += " AND pod.po_id = ANY($1)"
		args = append(args, pq.Array(purchaseOrderIDs))
		i++
	}

	if err := r.sqlDB.SelectContext(ctx.Context(), &poDts, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return poDts, nil
}

func (r *PurchaseOrderRepository) GetPoDtsBomByPurchaseOrders(ctx *fiber.Ctx, filters map[string]string, purchaseOrderIDs []uint, span opentracing.Span) ([]dtos.PurchaseOrderPoDtBomListDTO, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-GetPoDtsBomByPurchaseOrders", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	poDtBoms := []dtos.PurchaseOrderPoDtBomListDTO{}

	filterDBColumnKey := []string{
		"po.po_no", "po.remark",
		"pi.name",
		"it.name",
		"pod.remark",
		"pod.gen_code",
		"podb.remark",
		"podb.gen_code",
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
		"status":           "po.status",
		"customer_id":      "po.customer_id",
		"purchase_type_id": "po.purchase_type_id",
		"currency_id":      "po.currency_id",
		"vat_id":           "po.vat_id",
		"payment_term_id":  "po.payment_term_id",
		"shipping_term_id": "po.shipping_term_id",
		"pph23_id":         "po.pph23_id",
	}

	for key, _ := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", value, i)
			args = append(args, value)
			i++
		}
	}

	if len(purchaseOrderIDs) > 0 {
		condition += fmt.Sprintf(" AND podb.po_id = ANY($%d)", i)
		args = append(args, pq.Array(purchaseOrderIDs))
		i++
	}

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (podb.id)
					podb.id, podb.product_uuid, podb.po_id, podb.po_dt_id, podb.product_id, podb.item_unit_id, podb.gen_code, podb.remark, 
					podb.qty, podb.price, podb.subtotal, podb.created_by_id, podb.updated_by_id, podb.deleted_by_id, podb.created_at, podb.updated_at, podb.deleted_at,
					podb.id as po_dt_bom_id,
					it.name as item_name,
					it.code as item_code,
					it.barcode as item_barcode,
					it.sku as item_sku,
					it.factory_code as item_factory_code,
					it.specification as item_specification,
					it.qty_stock as item_qty_stock,
					u.name as unit_name,
					po.branch_id,

					isg.name as item_sub_group_name,
					ig.name as item_group_name,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM purchase_order_dt_boms podb
				LEFT JOIN purchase_order_dts pod ON pod.id = podb.po_dt_id
				LEFT JOIN purchase_orders po ON po.id = pod.po_id
				LEFT JOIN products pi ON pod.product_id = pi.id
				LEFT JOIN item_units iu ON pod.item_unit_id = iu.id
				LEFT JOIN mix_values u ON iu.unit_id = u.id
				LEFT JOIN products it ON podb.product_id = it.id
				LEFT JOIN mix_values isg ON it.item_sub_group_id = isg.id
				LEFT JOIN mix_values ig ON isg.parent_id = ig.id

        LEFT JOIN users cu ON po.created_by_id = cu.id
        LEFT JOIN users uu ON po.updated_by_id = uu.id
				WHERE 1=1` + condition + queryGlobal + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

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

	selectSpan := opentracing.StartSpan("SelectQuery", opentracing.ChildOf(childSpan.Context()))

	err := r.sqlDB.SelectContext(ctx.Context(), &poDtBoms, query, args...)
	if err != nil {
		selectSpan.LogKV("query", query)
		utils.LogErrors(selectSpan, err)
		return nil, err
	}

	return poDtBoms, nil
}

func (r *PurchaseOrderRepository) GetPurchaseOrders(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.PurchaseOrderListDTO, int, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-GetPurchaseOrders", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	purchaseOrders := []dtos.PurchaseOrderListDTO{}

	var total int

	filterDBColumnKey := []string{
		"po.po_no", "po.remark",
		"pi.name",
		"it.name",
		"pod.remark",
		"pod.gen_code",
		"podb.remark",
		"podb.gen_code",
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
		"status":           "po.status",
		"customer_id":      "po.customer_id",
		"purchase_type_id": "po.purchase_type_id",
		"currency_id":      "po.currency_id",
		"vat_id":           "po.vat_id",
		"payment_term_id":  "po.payment_term_id",
		"shipping_term_id": "po.shipping_term_id",
		"pph23_id":         "po.pph23_id",
	}

	for key, _ := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", filterKey[key], i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":      "po.customer_id",
		"purchase_type_ids": "po.purchase_type_id",
		"currency_ids":      "po.currency_id",
		"payment_term_ids":  "po.payment_term_id",
		"shipping_term_ids": "po.shipping_term_id",
		"pph23_ids":         "po.pph23_id",
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
		"vat_ids": []string{"po.vat_id", "pod.vat_id"},
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

	if filters["date_type"] != "" && filters["start_date"] != "" && filters["end_date"] != "" {
		dateField := ""
		switch filters["date_type"] {
		case "po_date":
			dateField = "po.po_date"
		case "delivery_date":
			dateField = "po.delivery_date"
		default:
			dateField = "po.po_date"
		}

		condition += fmt.Sprintf(" AND %s BETWEEN $%d AND $%d", dateField, i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
	}

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (po.id)
					po.id, po.customer_id, po.purchase_type_id, po.currency_id, po.vat_id, po.payment_term_id, po.shipping_term_id, po.pph23_id, po.branch_id, 
					po.po_no, po.po_date, po.delivery_date, po.shipping_destination, po.remark, po.status, po.exchange_rate, po.pph23_percentage,
					po.discount_percentage, po.discount_amount, po.discount_percentage_amount, po.discount_amount_product, po.discount_final_header, po.discount_type,
					po.vat_percentage, po.total_amount_products, po.total_qty, po.subtotal, po.total_discount, po.total_pph23, po.total_vat, po.grand_total, 
					po.created_by_id, po.updated_by_id, po.deleted_by_id, po.created_at, po.updated_at, po.deleted_at,

					pi.id as product_id,
					it.id as item_id,
					pod.vat_id as po_dt_vat_id,

					pi.name as product_name,
					it.name as item_name,
					cur.name as currency_name,
					vat.name as vat_name,
					pph.name as pph23_name,
					pt.name as purchase_type_name,
					c.name as customer_name,
					pterm.name as payment_term_name,
					sterm.name as shipping_term_name,

					pod.remark as po_dt_remark,
					pod.gen_code as po_dt_gen_code,
					podb.remark as po_dt_bom_remark,
					podb.gen_code as po_dt_bom_gen_code,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM purchase_orders po
				LEFT JOIN purchase_order_dts pod ON pod.po_id = po.id
				LEFT JOIN products pi ON pod.product_id = pi.id
				LEFT JOIN item_units iu ON pod.item_unit_id = iu.id
				LEFT JOIN purchase_order_dt_boms podb ON podb.po_dt_id = pod.id
				LEFT JOIN products it ON podb.product_id = it.id

				LEFT JOIN mix_values cur ON po.currency_id = cur.id
				LEFT JOIN mix_values vat ON po.vat_id = vat.id
				LEFT JOIN mix_values pph ON po.pph23_id = pph.id
				LEFT JOIN mix_values pt ON po.purchase_type_id = pt.id
				LEFT JOIN mix_values pterm ON po.payment_term_id = pterm.id
				LEFT JOIN mix_values sterm ON po.shipping_term_id = sterm.id
				LEFT JOIN customers c ON po.customer_id = c.id

        LEFT JOIN users cu ON po.created_by_id = cu.id
        LEFT JOIN users uu ON po.updated_by_id = uu.id
				WHERE 1=1` + condition + queryGlobal + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery

	for key, value := range filters {
		switch key {
		case "po_no", "remark":
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

		err := r.sqlDB.SelectContext(ctx.Context(), &purchaseOrders, query, args...)
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

	return purchaseOrders, total, nil
}

func (r *PurchaseOrderRepository) UpdatePurchaseOrder(tx *gorm.DB, purchaseOrder *models.PurchaseOrder, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-UpdatePurchaseOrder", opentracing.ChildOf(span.Context()))

	if err := tx.Where("id = ?", purchaseOrder.ID).Select("*").Omit(
		"created_at", "created_by_id", "branch_id",
	).Updates(purchaseOrder).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *PurchaseOrderRepository) UpdatePoDts(tx *gorm.DB, poDts []models.PurchaseOrderDt, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-UpdatePoDts", opentracing.ChildOf(span.Context()))

	data := make([]map[string]interface{}, 0)
	for _, poDt := range poDts {
		data = append(data, map[string]interface{}{
			"id":                         poDt.ID,
			"product_uuid":               poDt.ProductUuid,
			"po_id":                      poDt.PoID,
			"item_unit_id":               poDt.ItemUnitID,
			"vat_id":                     poDt.VatID,
			"pph23_id":                   poDt.Pph23ID,
			"ref_id":                     poDt.RefID,
			"product_id":                 poDt.ProductID,
			"is_vat":                     poDt.IsVat,
			"is_pph23":                   poDt.IsPph23,
			"product_type":               poDt.ProductType,
			"ref_type":                   poDt.RefType,
			"gen_code":                   poDt.GenCode,
			"remark":                     poDt.Remark,
			"need_qty":                   poDt.NeedQty,
			"qty":                        poDt.Qty,
			"price":                      poDt.Price,
			"subtotal":                   poDt.Subtotal,
			"discount_amount":            poDt.DiscountAmount,
			"discount_percentage":        poDt.DiscountPercentage,
			"discount_percentage_num":    poDt.DiscountPercentageNum,
			"discount_percentage_amount": poDt.DiscountPercentageAmount,
			"discount_final":             poDt.DiscountFinal,
			"discount_type":              poDt.DiscountType,
			"total_amount":               poDt.TotalAmount,
			"updated_by_id":              poDt.UpdatedByID,
			"updated_at":                 time.Now(),
		})
	}

	if err := r.utilRepo.Upsert(tx, "purchase_order_dts", "id", data, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *PurchaseOrderRepository) DeletePoDtsWhereNotIn(ctx *fiber.Ctx, tx *gorm.DB, purchaseOrderID uint, poDtIDs []uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-DeletePoDtsWhereNotIn", opentracing.ChildOf(span.Context()))

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	now := time.Now()
	deletedAt := gorm.DeletedAt{Time: now, Valid: true}

	query := tx.Model(&models.PurchaseOrderDt{}).Where("po_id = ? AND deleted_at IS NULL", purchaseOrderID)

	if len(poDtIDs) > 0 {
		query = query.Where("id NOT IN (?)", poDtIDs)
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

func (r *PurchaseOrderRepository) GetUpdatedPoDtsByPurchaseOrderIDs(ctx *fiber.Ctx, tx *gorm.DB, purchaseOrderIDs []uint, span opentracing.Span) ([]dtos.PurchaseOrderPoDtListUpdateDTO, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-GetUpdatedPoDtsByPurchaseOrderIDs", opentracing.ChildOf(span.Context()))

	poDts := []dtos.PurchaseOrderPoDtListUpdateDTO{}

	query := `SELECT pod.id, pod.po_id, pod.product_uuid,
		pod.item_unit_id, pod.vat_id, pod.pph23_id, pod.ref_id, pod.product_id, pod.ref_type, pod.product_type, pod.gen_code, pod.remark, 
		pod.need_qty, pod.qty, pod.price, pod.subtotal, pod.discount_amount, pod.discount_percentage, pod.discount_percentage_num, 
		pod.discount_percentage_amount, pod.discount_final, pod.discount_type, pod.is_vat, pod.is_pph23,
		pod.total_amount, pod.created_by_id, pod.updated_by_id, pod.deleted_by_id, pod.created_at, pod.updated_at, pod.deleted_at,

		pod.id as po_dt_id,
		isg.id as item_sub_group_id,
		ig.id as item_group_id,
		isg.name as item_sub_group_name,
		ig.name as item_group_name,
		u.name as unit_name,
		pi.name as item_name,
		pi.code as item_code,

		cu.name as created_by_name,
		uu.name as updated_by_name

	FROM purchase_order_dts pod
	LEFT JOIN purchase_orders po ON pod.po_id = po.id
	LEFT JOIN products pi ON pod.product_id = pi.id
	LEFT JOIN item_units iu ON pod.item_unit_id = iu.id
	LEFT JOIN mix_values u ON iu.unit_id = u.id
	LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
	LEFT JOIN mix_values ig ON isg.parent_id = ig.id
	LEFT JOIN users cu ON pod.created_by_id = cu.id
	LEFT JOIN users uu ON pod.updated_by_id = uu.id
	WHERE pod.deleted_at IS NULL`

	var args []interface{}
	i := 1

	if len(purchaseOrderIDs) > 0 {
		query += " AND pod.po_id = ANY($1)"
		args = append(args, pq.Array(purchaseOrderIDs))
		i++
	}

	if err := tx.Raw(query, args...).Scan(&poDts).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return poDts, nil
}

func (r *PurchaseOrderRepository) UpdatePoDtBoms(tx *gorm.DB, poDtBoms []map[string]interface{}, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-UpdatePoDtBoms", opentracing.ChildOf(span.Context()))

	if err := r.utilRepo.Upsert(tx, "purchase_order_dt_boms", "id", poDtBoms, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *PurchaseOrderRepository) DeletePoDtBomsWhereNotIn(ctx *fiber.Ctx, tx *gorm.DB, purchaseOrderID uint, poDtBomIDs []uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-DeletePoDtBomsWhereNotIn", opentracing.ChildOf(span.Context()))

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	now := time.Now()
	deletedAt := gorm.DeletedAt{Time: now, Valid: true}

	query := tx.Model(&models.PurchaseOrderDtBom{}).Where("po_id = ? AND deleted_at IS NULL", purchaseOrderID)

	if len(poDtBomIDs) > 0 {
		query = query.Where("id NOT IN (?)", poDtBomIDs)
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

func (r *PurchaseOrderRepository) DeletePurchaseOrder(tx *gorm.DB, params *dtos.GetPurchaseOrderParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-DeletePurchaseOrder", opentracing.ChildOf(span.Context()))

	if err := tx.Delete(&models.PurchaseOrder{}, params.ID).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *PurchaseOrderRepository) RestorePurchaseOrder(tx *gorm.DB, params *dtos.GetPurchaseOrderParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-RestorePurchaseOrder", opentracing.ChildOf(span.Context()))

	var purchaseOrder models.PurchaseOrder
	if err := tx.Unscoped().Model(&purchaseOrder).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *PurchaseOrderRepository) DeletePoDtsByPurchaseOrderID(tx *gorm.DB, params *dtos.GetPurchaseOrderParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-DeletePoDtsByPurchaseOrderID", opentracing.ChildOf(span.Context()))

	if err := tx.Where("po_id = ?", params.ID).Delete(&models.PurchaseOrderDt{}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *PurchaseOrderRepository) DeletePoDtBomsByPurchaseOrderID(tx *gorm.DB, params *dtos.GetPurchaseOrderParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-DeletePoDtBomsByPurchaseOrderID", opentracing.ChildOf(span.Context()))

	if err := tx.Where("po_id = ?", params.ID).Delete(&models.PurchaseOrderDtBom{}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *PurchaseOrderRepository) LockPurchaseOrderHeader(ctx *fiber.Ctx, tx *gorm.DB, req dtos.UpdatePurchaseOrderRequest, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-LockPurchaseOrderHeader", opentracing.ChildOf(span.Context()))

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", []uint{req.ID}).Find(&models.PurchaseOrder{}).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *PurchaseOrderRepository) LockPoDts(ctx *fiber.Ctx, tx *gorm.DB, poDtIDs []*uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-LockPoDts", opentracing.ChildOf(span.Context()))

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", poDtIDs).Find(&models.PurchaseOrderDt{}).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *PurchaseOrderRepository) LockPoDtBoms(ctx *fiber.Ctx, tx *gorm.DB, poDtBomIDs []*uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-LockPoDtBoms", opentracing.ChildOf(span.Context()))

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", poDtBomIDs).Find(&models.PurchaseOrderDtBom{}).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}
