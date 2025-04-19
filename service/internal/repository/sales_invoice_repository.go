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

type SalesInvoiceRepository struct {
	db       *gorm.DB
	sqlDB    *sqlx.DB
	utilRepo *UtilRepository
	tracer   opentracing.Tracer
}

func NewSalesInvoiceRepository(db *gorm.DB, sqlDB *sqlx.DB, utilRepo *UtilRepository, tracer opentracing.Tracer) *SalesInvoiceRepository {
	return &SalesInvoiceRepository{
		db:       db,
		sqlDB:    sqlDB,
		tracer:   tracer,
		utilRepo: utilRepo,
	}
}

func (r *SalesInvoiceRepository) GetSalesInvoices(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.SalesInvoiceListDTO, int, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-GetSalesInvoices", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	salesInvoices := []dtos.SalesInvoiceListDTO{}

	var total int

	filterDBColumnKey := []string{
		"si.invoice_no", "si.remark", "si.status",
		"c.name",
		"sidt.remark",
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
		condition += fmt.Sprintf(" AND si.id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"customer_id":     "si.customer_id",
		"currency_id":     "si.currency_id",
		"payment_term_id": "si.payment_term_id",
		"vat_id":          "si.vat_id",
		"pph23_id":        "si.pph23_id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":     "si.customer_id",
		"currency_ids":     "si.currency_id",
		"payment_term_ids": "si.payment_term_id",
		"pph23_ids":        "si.pph23_id",
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
		"vat_ids": []string{"si.vat_id", "sidt.vat_id"},
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
        SELECT DISTINCT ON (si.id)
					si.id, si.customer_id, si.currency_id, si.payment_term_id, si.vat_id, si.pph23_id, si.branch_id, si.bank_id,
					si.invoice_no, si.remark, si.status, 
					si.exchange_rate, si.pph23_percentage, si.vat_percentage, si.total_qty, si.subtotal, si.total_discount, si.total_pph23, si.total_vat, si.grand_total, si.created_by_id, si.updated_by_id, si.deleted_by_id, si.created_at, si.updated_at, si.deleted_at,
					TO_CHAR(si.invoice_date, 'YYYY-MM-DD') as invoice_date,
					si.discount_amount, si.discount_percentage, si.discount_percentage_amount, si.discount_final, si.discount_type, si.total_amount_products, si.total_dp_products, si.total_balance_products,

					c.name as customer_name,
					cur.name as currency_name,
					pt.name as payment_term_name,
					vat.name as vat_name,
					pph.name as pph23_name,
					b.name as branch_name,

					sidt.remark as sales_invoice_dt_remark,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM sales_invoices si
				LEFT JOIN sales_invoice_dts sidt ON sidt.sales_invoice_id = si.id
				LEFT JOIN customers c ON si.customer_id = c.id
				LEFT JOIN mix_values cur ON si.currency_id = cur.id
				LEFT JOIN mix_values pt ON si.payment_term_id = pt.id
				LEFT JOIN mix_values vat ON si.vat_id = vat.id
				LEFT JOIN mix_values pph ON si.pph23_id = pph.id
				LEFT JOIN branches b ON si.branch_id = b.id
				LEFT JOIN bank_informations bk ON si.bank_id = bk.id

        LEFT JOIN users cu ON si.created_by_id = cu.id
        LEFT JOIN users uu ON si.updated_by_id = uu.id
				WHERE 1=1` + condition + queryGlobal + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery

	for key, value := range filters {
		switch key {
		case "invoice_no", "remark", "status":
			if value != "" {
				query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				countQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

	if startDate, ok := filters["start_date"]; ok && startDate != "" {
		query += fmt.Sprintf(" AND invoice_date >= $%d", i)
		countQuery += fmt.Sprintf(" AND invoice_date >= $%d", i)
		args = append(args, startDate)
		i++
	}

	if endDate, ok := filters["end_date"]; ok && endDate != "" {
		query += fmt.Sprintf(" AND invoice_date <= $%d", i)
		countQuery += fmt.Sprintf(" AND invoice_date <= $%d", i)
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "invoice_date")
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

		err := r.sqlDB.SelectContext(ctx.Context(), &salesInvoices, query, args...)
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

	return salesInvoices, total, nil
}

func (r *SalesInvoiceRepository) GetSalesInvoiceByID(ctx *fiber.Ctx, params *dtos.GetSalesInvoiceParams, tx *gorm.DB, span opentracing.Span) (*dtos.SalesInvoiceDetailDTO, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-GetSalesInvoiceByID", opentracing.ChildOf(span.Context()))
	var salesInvoice dtos.SalesInvoiceDetailDTO

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (si.id)
            si.id, si.customer_id, si.currency_id, si.payment_term_id, si.vat_id, si.pph23_id, si.branch_id, si.bank_id,
            si.invoice_no, si.remark, si.status, 
            si.exchange_rate, si.pph23_percentage, si.vat_percentage, si.total_qty, si.subtotal, si.total_discount, si.total_pph23, si.total_vat, si.grand_total, si.created_by_id, si.updated_by_id, si.deleted_by_id, si.created_at, si.updated_at, si.deleted_at,
            TO_CHAR(si.invoice_date, 'YYYY-MM-DD') as invoice_date,
            si.discount_amount, si.discount_percentage, si.discount_percentage_amount, si.discount_final, si.discount_type, si.total_amount_products, si.total_dp_products, si.total_balance_products, si.rev_no,

            cu.name as created_by_name,
            uu.name as updated_by_name

        FROM sales_invoices si
        LEFT JOIN sales_invoice_dts sidt ON sidt.sales_invoice_id = si.id
        LEFT JOIN users cu ON si.created_by_id = cu.id
        LEFT JOIN users uu ON si.updated_by_id = uu.id
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

	if err := r.sqlDB.Get(&salesInvoice, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		childSpan.LogKV("query", query)
		return nil, err
	}

	return &salesInvoice, nil
}

func (r *SalesInvoiceRepository) GetUpdatedSalesInvoiceDts(ctx *fiber.Ctx, salesInvoiceID uint, span opentracing.Span) ([]dtos.SalesInvoiceDtListUpdateDTO, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-GetUpdatedSalesInvoiceDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	salesInvoiceDts := []dtos.SalesInvoiceDtListUpdateDTO{}

	query := `
	SELECT 
		sidt.id, sidt.product_uuid, sidt.sales_invoice_id, sidt.item_unit_id, sidt.vat_id, sidt.pph23_id, 
				sidt.ref_id, sidt.ref_dt_id, sidt.product_id, sidt.ref_type, sidt.product_type, sidt.remark, 
		sidt.is_vat, sidt.is_pph23, sidt.qty, sidt.price, sidt.subtotal,
		sidt.discount, sidt.total_amount, sidt.total_dp, sidt.total_balance, sidt.created_by_id, sidt.updated_by_id, sidt.deleted_by_id, 
		sidt.created_at, sidt.updated_at, sidt.deleted_at,
		
		sidt.id as sales_invoice_dt_id,
		isg.id as item_sub_group_id,
		ig.id as item_group_id,
		isg.name as item_sub_group_name,
		ig.name as item_group_name,
		u.name as unit_name,
		p.name as item_name,
		p.code as item_code,
		
		cu.name as created_by_name,
		uu.name as updated_by_name
	FROM sales_invoice_dts sidt
	LEFT JOIN products p ON sidt.product_id = p.id
	LEFT JOIN item_units iu ON sidt.item_unit_id = iu.id
	LEFT JOIN mix_values u ON iu.unit_id = u.id
	LEFT JOIN mix_values isg ON p.item_sub_group_id = isg.id
	LEFT JOIN mix_values ig ON isg.parent_id = ig.id
	LEFT JOIN users cu ON sidt.created_by_id = cu.id
	LEFT JOIN users uu ON sidt.updated_by_id = uu.id
	WHERE sidt.sales_invoice_id = $1 AND sidt.deleted_at IS NULL
	ORDER BY sidt.id ASC
	`

	err := r.sqlDB.SelectContext(ctx.Context(), &salesInvoiceDts, query, salesInvoiceID)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return salesInvoiceDts, nil
}

func (r *SalesInvoiceRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *SalesInvoiceRepository) Rollback() *gorm.DB {
	return r.db.Rollback()
}

func (r *SalesInvoiceRepository) CreateSalesInvoice(tx *gorm.DB, salesInvoice *models.SalesInvoice, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-CreateSalesInvoice", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if salesInvoice.InvoiceNo != nil {
		var count int64
		if err := tx.Model(&models.SalesInvoice{}).Where("invoice_no = ?", *salesInvoice.InvoiceNo).Count(&count).Error; err != nil {
			utils.LogErrors(childSpan, err)
			return tx, err
		}

		if count > 0 {
			err := fmt.Errorf("invoice number %s already exists", *salesInvoice.InvoiceNo)
			utils.LogErrors(childSpan, err)
			return tx, err
		}
	}

	result := tx.Create(&salesInvoice)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *SalesInvoiceRepository) CreateSalesInvoiceDts(tx *gorm.DB, salesInvoiceDts []models.SalesInvoiceDt, span opentracing.Span) (*gorm.DB, []models.SalesInvoiceDt, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-CreateSalesInvoiceDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Create(&salesInvoiceDts)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, nil, result.Error
	}

	return tx, salesInvoiceDts, nil
}

func (r *SalesInvoiceRepository) GetSalesInvoiceDts(ctx *fiber.Ctx, salesInvoiceID uint, isDeleted *int, span opentracing.Span) ([]dtos.SalesInvoiceDtListDTO, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-GetSalesInvoiceDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	salesInvoiceDts := []dtos.SalesInvoiceDtListDTO{}

	query := `
	SELECT 
		sidt.id, sidt.product_uuid, sidt.sales_invoice_id, sidt.item_unit_id, sidt.vat_id, sidt.pph23_id, 
		sidt.ref_id, sidt.ref_dt_id, sidt.product_id, sidt.ref_type, sidt.product_type, sidt.remark, 
		sidt.is_vat, sidt.is_pph23, sidt.qty, sidt.price, sidt.subtotal,
		sidt.discount, sidt.total_amount, sidt.total_dp, sidt.total_balance, sidt.created_by_id, sidt.updated_by_id, sidt.deleted_by_id, 
		sidt.created_at, sidt.updated_at, sidt.deleted_at,
		
		p.name as item_name, p.code as item_code,
		u.name as unit_name,
		v.name as vat_name,
		pph.name as pph23_name,
		
		cu.name as created_by_name,
		uu.name as updated_by_name,

		CASE WHEN sidt.ref_type = 'so' THEN so.sales_order_no ELSE NULL END as ref_num
	FROM sales_invoice_dts sidt
	LEFT JOIN products p ON sidt.product_id = p.id
	LEFT JOIN item_units iu ON sidt.item_unit_id = iu.id
	LEFT JOIN mix_values u ON iu.unit_id = u.id
	LEFT JOIN mix_values v ON sidt.vat_id = v.id
	LEFT JOIN mix_values pph ON sidt.pph23_id = pph.id
	LEFT JOIN sales_orders so ON sidt.ref_id = so.id AND sidt.ref_type = 'so'
	LEFT JOIN users cu ON sidt.created_by_id = cu.id
	LEFT JOIN users uu ON sidt.updated_by_id = uu.id
	WHERE sidt.sales_invoice_id = $1
	`

	if isDeleted != nil && *isDeleted == 1 {
		query += " AND sidt.deleted_at IS NOT NULL"
	} else {
		query += " AND sidt.deleted_at IS NULL"
	}

	query += " ORDER BY sidt.id ASC"

	err := r.sqlDB.SelectContext(ctx.Context(), &salesInvoiceDts, query, salesInvoiceID)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return salesInvoiceDts, nil
}

func (r *SalesInvoiceRepository) UpdateSalesInvoice(tx *gorm.DB, salesInvoice *models.SalesInvoice, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-UpdateSalesInvoice", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Model(&models.SalesInvoice{}).Where("id = ?", salesInvoice.ID).Updates(salesInvoice)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *SalesInvoiceRepository) BulkCreateSalesInvoiceDts(tx *gorm.DB, salesInvoiceDts []models.SalesInvoiceDt, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-BulkCreateSalesInvoiceDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(salesInvoiceDts) == 0 {
		return tx, nil
	}

	result := tx.Create(&salesInvoiceDts)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *SalesInvoiceRepository) BulkUpdateSalesInvoiceDts(tx *gorm.DB, salesInvoiceDts []models.SalesInvoiceDt, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-BulkUpdateSalesInvoiceDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(salesInvoiceDts) == 0 {
		return tx, nil
	}

	if err := tx.Save(&salesInvoiceDts).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return tx, err
	}

	return tx, nil
}

func (r *SalesInvoiceRepository) DeleteSalesInvoiceDtsByIDs(tx *gorm.DB, salesInvoiceDtIDs []uint, userID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-DeleteSalesInvoiceDtsByIDs", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(salesInvoiceDtIDs) == 0 {
		return tx, nil
	}

	result := tx.Model(&models.SalesInvoiceDt{}).Where("id IN ?", salesInvoiceDtIDs).Updates(map[string]interface{}{
		"deleted_by_id": userID,
		"deleted_at":    time.Now(),
	})
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *SalesInvoiceRepository) DeleteSalesInvoice(tx *gorm.DB, salesInvoiceID uint, userID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-DeleteSalesInvoice", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Model(&models.SalesInvoice{}).Where("id = ?", salesInvoiceID).Updates(map[string]interface{}{
		"deleted_by_id": userID,
		"deleted_at":    time.Now(),
	})
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *SalesInvoiceRepository) RestoreSalesInvoice(ctx *fiber.Ctx, params *dtos.GetSalesInvoiceParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-RestoreSalesInvoice", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var salesInvoice models.SalesInvoice
	if err := tx.Unscoped().Model(&salesInvoice).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *SalesInvoiceRepository) GetRefSalesOrderDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefSalesOrderForInvoiceListDTO, int, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-GetRefSalesOrderDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	soDts := []dtos.RefSalesOrderForInvoiceListDTO{}

	var total int

	filterDBColumnKey := []string{
		"so.sales_order_no", "so.po_buyer_no", "so.remark",
		"c.name",
		"p.name", "p.code",
		"sodt.remark",
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

	if value, ok := filters["so_no"]; ok && value != "" {
		condition += fmt.Sprintf(" AND so.sales_order_no ILIKE $%d", i)
		args = append(args, "%"+value+"%")
		i++
	}

	if value, ok := filters["po_buyer_no"]; ok && value != "" {
		condition += fmt.Sprintf(" AND so.po_buyer_no ILIKE $%d", i)
		args = append(args, "%"+value+"%")
		i++
	}

	if value, ok := filters["product_name"]; ok && value != "" {
		condition += fmt.Sprintf(" AND p.name ILIKE $%d", i)
		args = append(args, "%"+value+"%")
		i++
	}

	if value, ok := filters["product_code"]; ok && value != "" {
		condition += fmt.Sprintf(" AND p.code ILIKE $%d", i)
		args = append(args, "%"+value+"%")
		i++
	}

	if value, ok := filters["item_type"]; ok && value != "" {
		condition += fmt.Sprintf(" AND sodt.item_type = $%d", i)
		args = append(args, value)
		i++
	}

	filterKey := map[string]string{
		"customer_id":   "so.customer_id",
		"currency_id":   "so.currency_id",
		"order_type_id": "so.order_type_id",
		"vat_id":        "so.vat_id",
		"pph23_id":      "so.pph23_id",
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
		"currency_ids":   "so.currency_id",
		"order_type_ids": "so.order_type_id",
		"pph23_ids":      "so.pph23_id",
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
		"vat_ids": []string{"so.vat_id", "sodt.vat_id"},
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
		SELECT DISTINCT ON (sodt.id)
				sodt.id, sodt.product_uuid, sodt.sales_order_id, sodt.item_unit_id, sodt.vat_id, sodt.pph23_id, 
				sodt.ref_id, sodt.item_id, sodt.ref_type, sodt.item_type, sodt.gen_code, sodt.remark, 
				sodt.vat_perc, sodt.vat_perc_am, sodt.pph23_perc, sodt.pph23_perc_am, 
				sodt.markup_perc, sodt.markup_perc_am, sodt.is_vat, sodt.is_pph23, 
				sodt.is_lock_markup, sodt.is_lock_price_sell, sodt.qty, sodt.qty_out, 
				sodt.price_sell, sodt.price_buy, sodt.subtotal_sell, sodt.subtotal_buy, 
				sodt.disc_am, sodt.disc_perc, sodt.disc_perc_num, sodt.disc_perc_am, 
				sodt.disc_final, sodt.disc_type, sodt.total_am, 
				COALESCE(sodt.total_dp, 0) as total_dp,
				(sodt.total_am - COALESCE(sodt.total_dp, 0)) as total_balance,
				sodt.created_by_id, 
				sodt.updated_by_id, sodt.deleted_by_id, sodt.created_at, sodt.updated_at, sodt.deleted_at,

				so.customer_id, so.order_type_id, so.currency_id, so.vat_id as head_vat_id, 
				so.pph23_id as head_pph23_id, so.vat_perc as head_vat_perc, 
				so.pph23_perc as head_pph23_perc, so.disc_am as head_disc_am, 
				so.disc_perc as head_disc_perc, so.markup_perc as head_markup_perc, 
				so.remark as head_remark, so.exchange_rate, so.sales_order_no, so.po_buyer_no, 
				so.order_at as order_date, so.shipping_at as shipping_date,
				TO_CHAR(so.due_at, 'YYYY-MM-DD') as due_at,

				c.name as customer_name,
				ot.name as order_type_name,
				p.name as item_name, p.code as item_code, p.sku as item_sku,
				u.name as unit_name,
				v.name as vat_name,
				pph.name as pph23_name,
				
				cu.name as created_by_name,
				uu.name as updated_by_name

		FROM so_dts sodt
				LEFT JOIN sales_orders so ON sodt.sales_order_id = so.id
				LEFT JOIN so_dt_boms sodtb ON sodtb.so_dt_id = sodt.id
				LEFT JOIN customers c ON so.customer_id = c.id
				LEFT JOIN mix_values ot ON so.order_type_id = ot.id
				LEFT JOIN products p ON sodt.item_id = p.id
				LEFT JOIN item_units iu ON sodt.item_unit_id = iu.id
				LEFT JOIN mix_values u ON iu.unit_id = u.id
				LEFT JOIN mix_values v ON sodt.vat_id = v.id
				LEFT JOIN mix_values pph ON sodt.pph23_id = pph.id

		LEFT JOIN users cu ON sodt.created_by_id = cu.id
		LEFT JOIN users uu ON sodt.updated_by_id = uu.id
				WHERE 1=1 AND so.status IN ('PROCESS', 'DELIVERY', 'SCHEDULE', 'INVOICE')` + condition + queryGlobal + `
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

func (r *SalesInvoiceRepository) GetSoDtBoms(ctx *fiber.Ctx, soDtIDs []uint, span opentracing.Span) ([]dtos.SalesOrderSoDtBomListDTO, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-GetSoDtBoms", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(soDtIDs) == 0 {
		return []dtos.SalesOrderSoDtBomListDTO{}, nil
	}

	soDtBoms := []dtos.SalesOrderSoDtBomListDTO{}

	query := `
    SELECT 
        sodtb.id, sodtb.product_uuid, sodtb.sales_order_id, sodtb.so_dt_id, 
        sodtb.product_id, sodtb.item_id, sodtb.item_unit_id, sodtb.gen_code, sodtb.remark, 
        sodtb.qty, sodtb.qty_out, sodtb.price_sell, sodtb.price_buy, 
        sodtb.subtotal_sell, sodtb.subtotal_buy, sodtb.created_by_id, 
        sodtb.updated_by_id, sodtb.deleted_by_id, sodtb.created_at, sodtb.updated_at, sodtb.deleted_at,
        
        p.name as item_name, p.code as item_code, p.sku as item_sku,
        p.barcode as item_barcode, p.factory_code as item_factory_code,
        p.specification as item_specification,
        u.name as unit_name,
        
        cu.name as created_by_name,
        uu.name as updated_by_name
    FROM so_dt_boms sodtb
    LEFT JOIN products p ON sodtb.item_id = p.id
    LEFT JOIN item_units iu ON sodtb.item_unit_id = iu.id
    LEFT JOIN mix_values u ON iu.unit_id = u.id
    LEFT JOIN users cu ON sodtb.created_by_id = cu.id
    LEFT JOIN users uu ON sodtb.updated_by_id = uu.id
    WHERE sodtb.so_dt_id IN (?) AND sodtb.deleted_at IS NULL
    ORDER BY sodtb.id ASC
    `

	query = r.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Raw(query, soDtIDs)
	})

	err := r.sqlDB.SelectContext(ctx.Context(), &soDtBoms, query)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return soDtBoms, nil
}

func (r *SalesInvoiceRepository) BulkUpdateSalesOrdersStatus(tx *gorm.DB, salesOrderIDs []uint, status string, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-BulkUpdateSalesOrdersStatus", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(salesOrderIDs) == 0 {
		return tx, nil
	}

	query := `
        UPDATE sales_orders 
        SET 
            status = CASE WHEN status != 'INVOICE' THEN ? ELSE status END,
            history_status = CASE WHEN status != 'INVOICE' THEN status ELSE history_status END
        WHERE id IN ?
    `

	result := tx.Exec(query, status, salesOrderIDs)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *SalesInvoiceRepository) RestoreSalesOrdersStatus(tx *gorm.DB, salesOrderIDs []uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-RestoreSalesOrdersStatus", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(salesOrderIDs) == 0 {
		return tx, nil
	}

	query := `
        UPDATE sales_orders 
        SET 
            status = history_status,
            history_status = status
        WHERE id IN ? AND status = 'INVOICE'
    `

	result := tx.Exec(query, salesOrderIDs)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *SalesInvoiceRepository) LockSalesOrders(tx *gorm.DB, salesOrderIDs []uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-LockSalesOrders", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(salesOrderIDs) == 0 {
		return nil
	}

	lockQuery := "SELECT id FROM sales_orders WHERE id IN ? FOR UPDATE"
	if err := tx.Exec(lockQuery, salesOrderIDs).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *SalesInvoiceRepository) LockSalesInvoice(tx *gorm.DB, salesInvoiceID uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-LockSalesInvoice", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	lockQuery := "SELECT id FROM sales_invoices WHERE id = ? FOR UPDATE"
	if err := tx.Exec(lockQuery, salesInvoiceID).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *SalesInvoiceRepository) GetSalesInvoiceForUpdate(tx *gorm.DB, salesInvoiceID uint, span opentracing.Span) (*models.SalesInvoice, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-GetSalesInvoiceForUpdate", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var salesInvoice models.SalesInvoice
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", salesInvoiceID).First(&salesInvoice).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return &salesInvoice, nil
}

func (r *SalesInvoiceRepository) GetSalesInvoiceCreatedThisMonth(ctx *fiber.Ctx, tx *gorm.DB, customerID uint, span opentracing.Span) (int, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceRepository-GetCustomerSalesInvoiceCreatedThisMonth", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var count int
	query := `
    SELECT COUNT(*) 
    FROM sales_invoices 
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

func (r *SalesInvoiceRepository) Commit(tx *gorm.DB) error {
	return tx.Commit().Error
}
