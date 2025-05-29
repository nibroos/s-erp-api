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
		tracer:   tracer,
		utilRepo: utilRepo,
	}
}

func (r *PurchaseOrderRepository) GetPurchaseOrders(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.PurchaseOrderListDTO, int, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-GetPurchaseOrders", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	purchaseOrders := []dtos.PurchaseOrderListDTO{}

	var total int

	filterDBColumnKey := []string{
		"po.po_no", "po.remark", "po.shipping_destination",
		"pi.name",
		"pd.remark",
		"pd.gen_code",
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

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
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
		"vat_ids": {"po.vat_id", "pd.vat_id"},
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
		"product_ids": {"pi.id"},
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

	if filters["date_type"] != "" && filters["start_date"] != "" && filters["end_date"] != "" {
		dateColumn := ""
		switch filters["date_type"] {
		case "1":
			dateColumn = "po.po_date"
		case "2":
			dateColumn = "po.delivery_date"
		default:
			dateColumn = "po.po_date"
		}

		condition += fmt.Sprintf(" AND %s BETWEEN $%d AND $%d", dateColumn, i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
	}

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (po.id)
					po.id, po.customer_id, po.purchase_type_id, po.currency_id, po.vat_id, po.payment_term_id, po.shipping_term_id, po.pph23_id, po.branch_id,
					po.po_no, po.shipping_destination, po.remark, 
					po.status, po.exchange_rate, po.pph23_percentage, po.vat_percentage, po.total_qty, po.subtotal, po.total_discount, po.total_pph23, po.total_vat, po.grand_total, po.created_by_id, po.updated_by_id, po.deleted_by_id, po.created_at, po.updated_at, po.deleted_at,
					TO_CHAR(po.po_date, 'YYYY-MM-DD') as po_date,
					TO_CHAR(po.delivery_date, 'YYYY-MM-DD') as delivery_date,
					po.discount_amount, po.discount_percentage, po.discount_percentage_amount, po.discount_final_header, po.discount_amount_product, po.discount_type, po.total_amount_products,

					pi.id as product_id,
					pd.vat_id as po_dt_vat_id,

					pi.name as product_name,
					cur.name as currency_name,
					vat.name as vat_name,
					pph.name as pph23_name,
					pt.name as purchase_type_name,

					pd.remark as po_dt_remark,
					pd.gen_code as po_dt_gen_code,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM purchase_orders po
				LEFT JOIN purchase_order_dts pd ON pd.po_id = po.id
				LEFT JOIN products pi ON pd.product_id = pi.id
				LEFT JOIN item_units iu ON pd.item_unit_id = iu.id

				LEFT JOIN mix_values cur ON po.currency_id = cur.id
				LEFT JOIN mix_values vat ON po.vat_id = vat.id
				LEFT JOIN mix_values pph ON po.pph23_id = pph.id
				LEFT JOIN mix_values pt ON po.purchase_type_id = pt.id
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
		case "po_no", "shipping_destination", "remark":
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "po_date")
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

func (r *PurchaseOrderRepository) GetPurchaseOrderByID(ctx *fiber.Ctx, params *dtos.GetPurchaseOrderParams, tx *gorm.DB, span opentracing.Span) (*dtos.PurchaseOrderDetailDTO, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-GetPurchaseOrderByID", opentracing.ChildOf(span.Context()))
	var purchaseOrder dtos.PurchaseOrderDetailDTO

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	baseQuery := `
    FROM ( 
			SELECT DISTINCT ON (po.id)
				po.id, po.customer_id, po.purchase_type_id, po.currency_id, po.vat_id, po.payment_term_id, po.shipping_term_id, po.pph23_id, po.branch_id, po.payment_id,
				po.po_no, po.po_no_ori, po.rev_no, po.shipping_destination, po.remark, 
				po.status, po.exchange_rate, po.pph23_percentage, po.vat_percentage, po.total_qty, po.subtotal, po.total_discount, po.total_pph23, po.total_vat, po.grand_total, po.created_by_id, po.updated_by_id, po.deleted_by_id, po.created_at, po.updated_at, po.deleted_at,
				TO_CHAR(po.po_date, 'YYYY-MM-DD') as po_date,
				TO_CHAR(po.delivery_date, 'YYYY-MM-DD') as delivery_date,
				po.discount_amount, po.discount_percentage, po.discount_percentage_amount, po.discount_final_header, po.discount_amount_product, po.discount_type, po.total_amount_products,
				po.is_vat,

				br.company_profile_id,
				c.name as customer_name,
				c.phone as phone,
				c.pic as pic,
				c.address as address,
				py.account_name,
				py.name as bank_name,

				cu.name as created_by_name,
				uu.name as updated_by_name

			FROM purchase_orders po
			LEFT JOIN purchase_order_dts pd ON pd.po_id = po.id
			LEFT JOIN products pi ON pd.product_id = pi.id
			LEFT JOIN item_units iu ON pd.item_unit_id = iu.id

			LEFT JOIN mix_values cur ON po.currency_id = cur.id
			LEFT JOIN mix_values vat ON po.vat_id = vat.id
			LEFT JOIN mix_values pph ON po.pph23_id = pph.id
			LEFT JOIN mix_values pt ON po.purchase_type_id = pt.id
			LEFT JOIN bank_informations py ON po.payment_id = py.id
			LEFT JOIN customers c ON po.customer_id = c.id
			LEFT JOIN branches br ON po.branch_id = br.id

			LEFT JOIN users cu ON po.created_by_id = cu.id
			LEFT JOIN users uu ON po.updated_by_id = uu.id
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

	if !isAdmin && branchID != nil {
		query += fmt.Sprintf(" AND (branch_id = $%d)", i+1)
		args = append(args, branchID)
	}

	err := r.sqlDB.GetContext(ctx.Context(), &purchaseOrder, query, args...)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	poDts, err := r.GetPurchaseOrderPoDts(ctx, tx, &dtos.GetPurchaseOrderPoDtParams{
		PurchaseOrderID: purchaseOrder.ID,
		IsDeleted:       params.IsDeleted,
	}, span)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	purchaseOrder.PoDts = poDts

	return &purchaseOrder, nil
}

func (r *PurchaseOrderRepository) GetPurchaseOrderPoDts(ctx *fiber.Ctx, tx *gorm.DB, params *dtos.GetPurchaseOrderPoDtParams, span opentracing.Span) ([]dtos.PurchaseOrderPoDtListDTO, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-GetPurchaseOrderPoDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	poDts := []dtos.PurchaseOrderPoDtListDTO{}

	query := `
	SELECT DISTINCT ON (pd.id)
		pd.id, pd.product_uuid, pd.po_id, pd.item_unit_id, pd.vat_id, pd.pph23_id, pd.ref_id, pd.product_id, pd.bom_id, pd.ref_so_dt_id, pd.ref_so_dt_bom_id,pd.ref_ro_dt_id, pd.ref_product_id, pd.ref_product_bom_id,
		pd.product_type, pd.product_json, pd.ref_type, pd.ref_json, pd.gen_code, pd.remark,
		pd.need_qty, pd.qty, pd.price, pd.subtotal, pd.discount_amount, pd.discount_percentage, pd.discount_percentage_num,
		pd.discount_percentage_amount, pd.discount_final, pd.discount_type, pd.is_vat, pd.is_pph23, pd.total_amount,
		pd.created_by_id, pd.updated_by_id, pd.deleted_by_id, pd.created_at, pd.updated_at, pd.deleted_at,

		pd.discount_percentage_amount + pd.discount_amount as sub_discount,
		
		p.name as product_name,
		p.code as product_code,
		mv.name as unit_name,

		p.name as item_name,

		COALESCE(
		 sd.qty_po, rod.qty_po, 0
		) as qty_po,
		COALESCE(
		 sd.qty, rod.req_qty, 0
		) as ref_qty,
		
		COALESCE(
			so.po_buyer_no, NULL
		) as ref_num,

		cu.name as created_by_name,
		uu.name as updated_by_name
	FROM purchase_order_dts pd
	LEFT JOIN so_dts sd ON sd.id = pd.ref_so_dt_id AND pd.ref_type = 'so'
	LEFT JOIN sales_orders so ON (so.id = sd.sales_order_id)
	LEFT JOIN request_order_dts rod ON pd.ref_ro_dt_id = rod.id AND pd.ref_type = 'ro'
	LEFT JOIN products p ON pd.product_id = p.id
	LEFT JOIN item_units iu ON pd.item_unit_id = iu.id
	LEFT JOIN mix_values mv ON iu.unit_id = mv.id
	LEFT JOIN users cu ON pd.created_by_id = cu.id
	LEFT JOIN users uu ON pd.updated_by_id = uu.id
	WHERE pd.po_id = $1
	`

	if params.IsDeleted != nil && *params.IsDeleted == 0 {
		query += " AND pd.deleted_at IS NULL"
	}

	query += " ORDER BY pd.id ASC"

	err := r.sqlDB.SelectContext(ctx.Context(), &poDts, query, params.PurchaseOrderID)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return poDts, nil
}

func (r *PurchaseOrderRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *PurchaseOrderRepository) Rollback(tx *gorm.DB) *gorm.DB {
	return tx.Rollback()
}

func (r *PurchaseOrderRepository) Commit(tx *gorm.DB) *gorm.DB {
	return tx.Commit()
}

func (r *PurchaseOrderRepository) CreatePurchaseOrder(tx *gorm.DB, purchaseOrder *models.PurchaseOrder, span opentracing.Span) (*models.PurchaseOrder, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-CreatePurchaseOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if err := tx.Create(purchaseOrder).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return purchaseOrder, nil
}

func (r *PurchaseOrderRepository) CreatePoDts(tx *gorm.DB, poDts []models.PoDt, span opentracing.Span) ([]models.PoDt, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-CreatePoDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if err := tx.Create(&poDts).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return poDts, nil
}

func (r *PurchaseOrderRepository) UpdatePurchaseOrder(tx *gorm.DB, purchaseOrder *models.PurchaseOrder, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-UpdatePurchaseOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if err := tx.Model(&models.PurchaseOrder{}).Where("id = ?", purchaseOrder.ID).Updates(map[string]interface{}{
		"customer_id":                purchaseOrder.CustomerID,
		"purchase_type_id":           purchaseOrder.PurchaseTypeID,
		"currency_id":                purchaseOrder.CurrencyID,
		"vat_id":                     purchaseOrder.VatID,
		"is_vat":                     purchaseOrder.IsVat,
		"payment_term_id":            purchaseOrder.PaymentTermID,
		"payment_id":                 purchaseOrder.PaymentID,
		"branch_id":                  purchaseOrder.BranchID,
		"shipping_term_id":           purchaseOrder.ShippingTermID,
		"pph23_id":                   purchaseOrder.Pph23ID,
		"po_no":                      purchaseOrder.PoNo,
		"rev_no":                     purchaseOrder.RevNo,
		"po_no_ori":                  purchaseOrder.PoNoOri,
		"po_date":                    purchaseOrder.PoDate,
		"delivery_date":              purchaseOrder.DeliveryDate,
		"shipping_destination":       purchaseOrder.ShippingDestination,
		"remark":                     purchaseOrder.Remark,
		"status":                     purchaseOrder.Status,
		"exchange_rate":              purchaseOrder.ExchangeRate,
		"discount_percentage":        purchaseOrder.DiscountPercentage,
		"discount_amount":            purchaseOrder.DiscountAmount,
		"discount_percentage_amount": purchaseOrder.DiscountPercentageAmount,
		"discount_final_header":      purchaseOrder.DiscountFinalHeader,
		"discount_amount_product":    purchaseOrder.DiscountAmountProduct,
		"discount_type":              purchaseOrder.DiscountType,
		"pph23_percentage":           purchaseOrder.Pph23Percentage,
		"vat_percentage":             purchaseOrder.VatPercentage,
		"total_amount_products":      purchaseOrder.TotalAmountProducts,
		"subtotal":                   purchaseOrder.Subtotal,
		"total_qty":                  purchaseOrder.TotalQty,
		"total_discount":             purchaseOrder.TotalDiscount,
		"total_pph23":                purchaseOrder.TotalPph23,
		"total_vat":                  purchaseOrder.TotalVat,
		"grand_total":                purchaseOrder.GrandTotal,
		"updated_by_id":              purchaseOrder.UpdatedByID,
		"updated_at":                 time.Now(),
	}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *PurchaseOrderRepository) DeletePoDtsWhereNotIn(ctx *fiber.Ctx, tx *gorm.DB, purchaseOrderID uint, poDtIDs []*uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-DeletePoDtsWhereNotIn", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	query := tx.Model(&models.PoDt{}).Where("po_id = ? AND deleted_at IS NULL", purchaseOrderID)

	if len(poDtIDs) > 0 {
		query = query.Where("id NOT IN (?)", poDtIDs)
	}

	if err := query.Updates(map[string]interface{}{
		"deleted_by_id": userID,
		"deleted_at":    time.Now(),
	}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *PurchaseOrderRepository) UpdatePoDts(tx *gorm.DB, poDts []models.PoDt, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-UpdatePoDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	for _, poDt := range poDts {
		if poDt.ID == 0 {
			if err := tx.Create(&poDt).Error; err != nil {
				utils.LogErrors(childSpan, err)
				return err
			}
		} else {
			if err := tx.Model(&models.PoDt{}).Where("id = ?", poDt.ID).Updates(map[string]interface{}{
				"product_uuid":               poDt.ProductUuid,
				"po_id":                      poDt.PoID,
				"item_unit_id":               poDt.ItemUnitID,
				"vat_id":                     poDt.VatID,
				"pph23_id":                   poDt.Pph23ID,
				"ref_id":                     poDt.RefID,
				"ref_so_dt_id":               poDt.RefSoDtID,
				"ref_so_dt_bom_id":           poDt.RefSoDtBomID,
				"ref_product_id":             poDt.RefProductID,
				"product_id":                 poDt.ProductID,
				"bom_id":                     poDt.BomID,
				"product_type":               poDt.ProductType,
				"product_json":               poDt.ProductJSON,
				"ref_type":                   poDt.RefType,
				"ref_json":                   poDt.RefJSON,
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
				"vat_perc":                   poDt.VatPerc,
				"vat_perc_am":                poDt.VatPercAm,
				"pph23_perc":                 poDt.Pph23Perc,
				"pph23_perc_am":              poDt.Pph23PercAm,
				"is_vat":                     poDt.IsVat,
				"is_pph23":                   poDt.IsPph23,
				"total_amount":               poDt.TotalAmount,
				"updated_by_id":              poDt.UpdatedByID,
				"updated_at":                 time.Now(),
				"deleted_at":                 nil,
			}).Error; err != nil {
				utils.LogErrors(childSpan, err)
				return err
			}
		}
	}

	return nil
}

func (r *PurchaseOrderRepository) DeletePurchaseOrder(ctx *fiber.Ctx, id uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-DeletePurchaseOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&models.PurchaseOrder{}).Where("id = ?", id).Updates(map[string]interface{}{
		"deleted_by_id": userID,
		"deleted_at":    time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)
		return err
	}

	if err := tx.Model(&models.PoDt{}).Where("po_id = ?", id).Updates(map[string]interface{}{
		"deleted_by_id": userID,
		"deleted_at":    time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)
		return err
	}

	if err := tx.Commit().Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *PurchaseOrderRepository) UpdatePurchaseOrderStatus(ctx *fiber.Ctx, id uint, status string, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-UpdatePurchaseOrderStatus", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&models.PurchaseOrder{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        status,
		"updated_by_id": userID,
		"updated_at":    time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)
		return err
	}

	if err := tx.Commit().Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *PurchaseOrderRepository) RestorePurchaseOrder(tx *gorm.DB, params *dtos.GetPurchaseOrderParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-RestorePurchaseOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var purchaseOrder models.PurchaseOrder
	if err := tx.Unscoped().Model(&purchaseOrder).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	if err := tx.Unscoped().Model(&models.PoDt{}).Where("po_id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *PurchaseOrderRepository) GetWidgetPurchaseOrders(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.PurchaseOrderStatusWidget, int, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-GetWidgetPurchaseOrders", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	purchaseOrders := []dtos.PurchaseOrderStatusWidget{}

	var total int

	filterDBColumnKey := []string{
		"po.po_no", "po.remark", "po.shipping_destination",
		"pi.name",
		"pd.remark",
		"pd.gen_code",
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

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
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
		"vat_ids": {"po.vat_id", "pd.vat_id"},
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
		dateColumn := ""
		switch filters["date_type"] {
		case "1":
			dateColumn = "po.po_date"
		case "2":
			dateColumn = "po.delivery_date"
		default:
			dateColumn = "po.po_date"
		}

		condition += fmt.Sprintf(" AND %s BETWEEN $%d AND $%d", dateColumn, i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
	}

	filterLikeKeys := map[string]string{
		"po_no":                "po.po_no",
		"shipping_destination": "po.shipping_destination",
		"remark":               "po.remark",
	}

	for key, value := range filters {
		if value != "" {
			if column, ok := filterLikeKeys[key]; ok {
				condition += fmt.Sprintf(" AND %s ILIKE $%d", column, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

	if !isAdmin && branchID != nil {
		condition += fmt.Sprintf(" AND (branch_id = $%d)", i)
		args = append(args, branchID)
		i++
	}

	if isAdmin && filters["branch_id"] != "" {
		condition += fmt.Sprintf(" AND (branch_id = $%d)", i)
		args = append(args, filters["branch_id"])
		i++
	}

	query := `
	WITH status_values AS (
        SELECT status, row_number() over () as status_order
        FROM (VALUES 
            ('TOTAL', 0),      -- Set TOTAL with order 0 to appear first
            ('PROCESS', 1),
            ('FINISH', 2),
            ('PARTIAL', 3),
            ('CANCELED', 4)
        ) AS s(status)
    ),
    filtered_orders AS (
        SELECT 
            po.id,             
            po.status as po_status,
            po.total_qty,
            po.grand_total
        FROM purchase_orders po
        LEFT JOIN purchase_order_dts pd ON pd.po_id = po.id
        LEFT JOIN products pi ON pd.product_id = pi.id
        LEFT JOIN item_units iu ON pd.item_unit_id = iu.id
        WHERE po.deleted_at IS NULL
        ` + condition + queryGlobal + `
    ),
    purchase_order_stats AS (
        SELECT 
            sv.status,
            sv.status_order,
            COUNT(DISTINCT fo.id) as order_count,
            COALESCE(SUM(fo.total_qty), 0) as total_qty,
            COALESCE(SUM(fo.grand_total), 0) as grand_total
        FROM status_values sv
        LEFT JOIN filtered_orders fo ON sv.status = fo.po_status OR sv.status = 'TOTAL'
        GROUP BY sv.status, sv.status_order
    )
    SELECT 
        sv.status,
        order_count,
        total_qty,
        grand_total
    FROM purchase_order_stats sv
    ORDER BY status_order`

	var selectErr error
	selectSpan := opentracing.StartSpan("SelectQuery", opentracing.ChildOf(childSpan.Context()))

	err := r.sqlDB.SelectContext(ctx.Context(), &purchaseOrders, query, args...)
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

	return purchaseOrders, total, nil
}

// GetCustomerSalesOrderCreatedThisMonth
func (r *PurchaseOrderRepository) GetCustomerPurchaseOrderCreatedThisMonth(ctx *fiber.Ctx, tx *gorm.DB, customerID uint, span opentracing.Span) (int, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-GetCustomerSalesOrderCreatedThisMonth", opentracing.ChildOf(span.Context()))

	var total int

	baseQuery := `
		FROM (
			SELECT COUNT(*) as total
			FROM purchase_orders po
			WHERE po.customer_id = $1 AND po.created_at >= date_trunc('month', CURRENT_DATE)
			AND po.deleted_at IS NULL
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

func (r *PurchaseOrderRepository) GetRefIndexSoDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefPoIndexSoDtListDTO, int, error) {

	childSpan := opentracing.StartSpan("PurchaseOrderRepository-GetRefIndexSoDts", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	products := []dtos.RefPoIndexSoDtListDTO{}

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
						sd.qty_po,
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
						COALESCE(sd.qty, 0) - COALESCE(sd.qty_po, 0) as balance, 
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
						sdb.qty_po,
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
						COALESCE(sdb.qty, 0) - COALESCE(sdb.qty_po, 0) as balance, 
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

			iu.unit_id as item_unit_unit_id,
			iu.price_sell as price_sell,
			iu.price_buy as price_buy,
			iu.conversion as item_unit_conversion,

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

func (r *PurchaseOrderRepository) GetRefIndexRoDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefPoIndexRoDtListDTO, int, error) {

	childSpan := opentracing.StartSpan("PurchaseOrderRepository-GetRefIndexRoDts", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	products := []dtos.RefPoIndexRoDtListDTO{}

	var total int

	filterDBColumnKey := []string{
		"so.request_no",
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
		condition += fmt.Sprintf(" AND so.id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"status":      "so.status",
		"customer_id": "so2.customer_id",
		"product_id":  "sd.item_id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":       "so2.customer_id",
		"warehouse_ids":      "so.warehouse_id",
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
		"request_no": "so.request_no",
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
						sd.id as ref_ro_dt_id,
						sd.request_order_id, 
						sd.ref_type,
						sd.ref_id, 
						sd.item_id, 
						sd.item_unit_id, 
						sd.qty_po,
						sd.req_qty AS ref_qty,
						COALESCE(sd.req_qty, 0) - COALESCE(sd.qty_po, 0) as balance, 

						sd.remark
					FROM request_order_dts sd
    ) AS sd 
		LEFT JOIN request_orders so ON sd.request_order_id = so.id
		LEFT JOIN products pi ON sd.item_id = pi.id
		LEFT JOIN item_units iu ON sd.item_unit_id = iu.id
		LEFT JOIN so_dts sd2 ON sd2.id = sd.ref_id AND sd.ref_type = 'so'
		LEFT JOIN sales_orders so2 ON sd2.sales_order_id = so2.id
		LEFT JOIN customers c ON so2.customer_id = c.id

		LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
		LEFT JOIN mix_values ig ON isg.parent_id = ig.id
		LEFT JOIN mix_values u ON iu.unit_id = u.id
		WHERE 1=1 AND COALESCE(sd.balance, 0) > 0`

	query := `SELECT sd.*,
			TO_CHAR(so.request_date, 'YYYY-MM-DD') as request_date,
			so.request_no as ref_num,
			isg.name as item_sub_group_name,
			ig.name as item_group_name,
			u.name as unit_name,
			c.name as customer_name,
			pi.name as item_name,
			pi.code as item_code,
			pi.sku as item_sku,
			iu.unit_id as item_unit_unit_id,
			iu.price_sell as price_sell,
			iu.price_buy as price_buy,
			iu.conversion as item_unit_conversion,
			'so' as ref_type
		` + baseQuery + condition + queryGlobal

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery + condition + queryGlobal

	if filters["start_date"] != "" && filters["end_date"] != "" {
		query += fmt.Sprintf(" AND (so.request_date BETWEEN $%d AND $%d)", i, i+1)
		countQuery += fmt.Sprintf(" AND (so.request_date BETWEEN $%d AND $%d)", i, i+1)
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
		"request_date":  "so.request_date",
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "so.request_date")
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

func (r *PurchaseOrderRepository) GetRefPoDtByRefDtID(ctx *fiber.Ctx, tx *gorm.DB, tableName string, parentColumnName string, detailIDs []uint, span opentracing.Span) ([]map[string]interface{}, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-GetRefPoDtByRefDtID", opentracing.ChildOf(span.Context()))

	var details []map[string]interface{}

	baseQuery := fmt.Sprintf(`
		FROM (
			SELECT DISTINCT ON (sd.id)
				sd.id, sd.qty_po, sd.%s
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

	err := tx.Raw(query, args...).Scan(&details).Error
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return details, nil
}

func (r *PurchaseOrderRepository) BulkUpdateReverseInvRefDtsQty(ctx *fiber.Ctx, tx *gorm.DB, refDts []map[string]interface{}, tableName string, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-BulkUpdateReverseInvRefDtsQty", opentracing.ChildOf(span.Context()))

	if err := r.utilRepo.Upsert(tx, tableName, "id", refDts, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}
