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

type InvoiceAdjustmentRepository struct {
	db       *gorm.DB
	sqlDB    *sqlx.DB
	utilRepo *UtilRepository
	tracer   opentracing.Tracer
}

func NewInvoiceAdjustmentRepository(db *gorm.DB, sqlDB *sqlx.DB, utilRepo *UtilRepository, tracer opentracing.Tracer) *InvoiceAdjustmentRepository {
	return &InvoiceAdjustmentRepository{
		db:       db,
		sqlDB:    sqlDB,
		tracer:   tracer,
		utilRepo: utilRepo,
	}
}

func (r *InvoiceAdjustmentRepository) GetInvoiceAdjustments(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.InvoiceAdjustmentListDTO, int, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentRepository-GetInvoiceAdjustments", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	invoiceAdjustments := []dtos.InvoiceAdjustmentListDTO{}

	var total int

	filterDBColumnKey := []string{
		"ia.invoice_no", "ia.reference", "ia.remark",
		"c.name",
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
		condition += fmt.Sprintf(" AND ia.id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"customer_id": "ia.customer_id",
		"currency_id": "ia.currency_id",
		"bank_id":     "ia.bank_id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids": "ia.customer_id",
		"currency_ids": "ia.currency_id",
		"bank_ids":     "ia.bank_id",
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

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (ia.id)
            ia.id, ia.customer_id, ia.currency_id, ia.branch_id, ia.bank_id,
            ia.invoice_no, ia.reference, ia.remark, ia.rev_no,
            ia.exchange_rate, ia.total_invoice, ia.total_adjustment, ia.total_balance, ia.total_admin_bank, ia.grand_total, 
            ia.created_by_id, ia.updated_by_id, ia.deleted_by_id, ia.created_at, ia.updated_at, ia.deleted_at,
            TO_CHAR(ia.payment_date, 'YYYY-MM-DD') as payment_date,
            TO_CHAR(ia.ref_start_date, 'YYYY-MM-DD') as ref_start_date,
            TO_CHAR(ia.ref_end_date, 'YYYY-MM-DD') as ref_end_date,
            ia.payment_amount,

            c.name as customer_name,
            cur.name as currency_name,
            b.name as branch_name,
            bi.name as bank_name,

            cu.name as created_by_name,
            uu.name as updated_by_name

        FROM invoice_adjustments ia
        LEFT JOIN customers c ON ia.customer_id = c.id
        LEFT JOIN mix_values cur ON ia.currency_id = cur.id
        LEFT JOIN branches b ON ia.branch_id = b.id
        LEFT JOIN bank_informations bi ON ia.bank_id = bi.id

        LEFT JOIN users cu ON ia.created_by_id = cu.id
        LEFT JOIN users uu ON ia.updated_by_id = uu.id
        WHERE 1=1` + condition + queryGlobal + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
        ` + baseQuery

	countQuery := `SELECT COUNT(*) as total
        ` + baseQuery

	for key, value := range filters {
		switch key {
		case "invoice_no", "reference", "remark":
			if value != "" {
				query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				countQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

	if startDate, ok := filters["start_date"]; ok && startDate != "" {
		query += fmt.Sprintf(" AND payment_date >= $%d", i)
		countQuery += fmt.Sprintf(" AND payment_date >= $%d", i)
		args = append(args, startDate)
		i++
	}

	if endDate, ok := filters["end_date"]; ok && endDate != "" {
		query += fmt.Sprintf(" AND payment_date <= $%d", i)
		countQuery += fmt.Sprintf(" AND payment_date <= $%d", i)
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "payment_date")
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

		err := r.sqlDB.SelectContext(ctx.Context(), &invoiceAdjustments, query, args...)
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

	return invoiceAdjustments, total, nil
}

func (r *InvoiceAdjustmentRepository) GetInvoiceAdjustmentByID(ctx *fiber.Ctx, params *dtos.GetInvoiceAdjustmentParams, tx *gorm.DB, span opentracing.Span) (*dtos.InvoiceAdjustmentDetailDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentRepository-GetInvoiceAdjustmentByID", opentracing.ChildOf(span.Context()))
	var invoiceAdjustment dtos.InvoiceAdjustmentDetailDTO

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (ia.id)
            ia.id, ia.customer_id, ia.currency_id, ia.branch_id, ia.bank_id,
            ia.invoice_no, ia.reference, ia.remark, ia.rev_no,
            ia.exchange_rate, ia.total_invoice, ia.total_adjustment, ia.total_balance, ia.total_admin_bank, ia.grand_total, 
            ia.created_by_id, ia.updated_by_id, ia.deleted_by_id, ia.created_at, ia.updated_at, ia.deleted_at,
            TO_CHAR(ia.payment_date, 'YYYY-MM-DD') as payment_date,
            TO_CHAR(ia.ref_start_date, 'YYYY-MM-DD') as ref_start_date,
            TO_CHAR(ia.ref_end_date, 'YYYY-MM-DD') as ref_end_date,
            ia.payment_amount,

            c.name as customer_name,
            cur.name as currency_name,
            b.name as branch_name,
            bi.name as bank_name,

            cu.name as created_by_name,
            uu.name as updated_by_name

        FROM invoice_adjustments ia
        LEFT JOIN customers c ON ia.customer_id = c.id
        LEFT JOIN mix_values cur ON ia.currency_id = cur.id
        LEFT JOIN branches b ON ia.branch_id = b.id
        LEFT JOIN bank_informations bi ON ia.bank_id = bi.id

        LEFT JOIN users cu ON ia.created_by_id = cu.id
        LEFT JOIN users uu ON ia.updated_by_id = uu.id
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

	if err := r.sqlDB.Get(&invoiceAdjustment, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		childSpan.LogKV("query", query)
		return nil, err
	}

	return &invoiceAdjustment, nil
}

func (r *InvoiceAdjustmentRepository) GetInvoiceAdjustmentDts(ctx *fiber.Ctx, invoiceAdjustmentID uint, isDeleted *int, span opentracing.Span) ([]dtos.InvoiceAdjustmentDtListDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentRepository-GetInvoiceAdjustmentDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	adjustmentDts := []dtos.InvoiceAdjustmentDtListDTO{}

	query := `
	SELECT 
		iadt.id, iadt.invoice_uuid, iadt.invoice_adjustment_id, iadt.ref_id, iadt.ref_type, 
		iadt.ref_json, iadt.invoice_no, iadt.invoice_amount, iadt.total_adjustment, 
		iadt.balance_amount, iadt.adjustment_amount, iadt.admin_bank, iadt.total_amount, 
		iadt.created_by_id, iadt.updated_by_id, iadt.deleted_by_id, 
		iadt.created_at, iadt.updated_at, iadt.deleted_at,
		TO_CHAR(iadt.invoice_date, 'YYYY-MM-DD') as invoice_date,
		
		cu.name as created_by_name,
		uu.name as updated_by_name
	FROM invoice_adjustment_dts iadt
	LEFT JOIN users cu ON iadt.created_by_id = cu.id
	LEFT JOIN users uu ON iadt.updated_by_id = uu.id
	WHERE iadt.invoice_adjustment_id = $1
	`

	if isDeleted != nil && *isDeleted == 1 {
		query += " AND iadt.deleted_at IS NOT NULL"
	} else {
		query += " AND iadt.deleted_at IS NULL"
	}

	query += " ORDER BY iadt.invoice_date ASC"

	err := r.sqlDB.SelectContext(ctx.Context(), &adjustmentDts, query, invoiceAdjustmentID)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return adjustmentDts, nil
}

func (r *InvoiceAdjustmentRepository) GetUpdatedInvoiceAdjustmentDts(ctx *fiber.Ctx, invoiceAdjustmentID uint, span opentracing.Span) ([]dtos.InvoiceAdjustmentDtListDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentRepository-GetUpdatedInvoiceAdjustmentDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	adjustmentDts := []dtos.InvoiceAdjustmentDtListDTO{}

	query := `
	SELECT 
		iadt.id, iadt.invoice_uuid, iadt.invoice_adjustment_id, iadt.ref_id, iadt.ref_type, 
		iadt.ref_json, iadt.invoice_no, iadt.invoice_amount, iadt.total_adjustment, 
		iadt.balance_amount, iadt.adjustment_amount, iadt.admin_bank, iadt.total_amount, 
		iadt.created_by_id, iadt.updated_by_id, iadt.deleted_by_id, 
		iadt.created_at, iadt.updated_at, iadt.deleted_at,
		TO_CHAR(iadt.invoice_date, 'YYYY-MM-DD') as invoice_date,
		
		cu.name as created_by_name,
		uu.name as updated_by_name
	FROM invoice_adjustment_dts iadt
	LEFT JOIN users cu ON iadt.created_by_id = cu.id
	LEFT JOIN users uu ON iadt.updated_by_id = uu.id
	WHERE iadt.invoice_adjustment_id = $1 AND iadt.deleted_at IS NULL
	ORDER BY iadt.id ASC
	`

	err := r.sqlDB.SelectContext(ctx.Context(), &adjustmentDts, query, invoiceAdjustmentID)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return adjustmentDts, nil
}

func (r *InvoiceAdjustmentRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *InvoiceAdjustmentRepository) Rollback() *gorm.DB {
	return r.db.Rollback()
}

func (r *InvoiceAdjustmentRepository) CreateInvoiceAdjustment(tx *gorm.DB, invoiceAdjustment *models.InvoiceAdjustment, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentRepository-CreateInvoiceAdjustment", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Create(&invoiceAdjustment)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceAdjustmentRepository) CreateInvoiceAdjustmentDts(tx *gorm.DB, adjustmentDts []models.InvoiceAdjustmentDt, span opentracing.Span) (*gorm.DB, []models.InvoiceAdjustmentDt, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentRepository-CreateInvoiceAdjustmentDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Create(&adjustmentDts)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, nil, result.Error
	}

	return tx, adjustmentDts, nil
}

func (r *InvoiceAdjustmentRepository) UpdateInvoiceAdjustment(tx *gorm.DB, invoiceAdjustment *models.InvoiceAdjustment, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentRepository-UpdateInvoiceAdjustment", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Model(&models.InvoiceAdjustment{}).Where("id = ?", invoiceAdjustment.ID).Updates(invoiceAdjustment)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceAdjustmentRepository) BulkCreateInvoiceAdjustmentDts(tx *gorm.DB, adjustmentDts []models.InvoiceAdjustmentDt, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentRepository-BulkCreateInvoiceAdjustmentDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(adjustmentDts) == 0 {
		return tx, nil
	}

	result := tx.Create(&adjustmentDts)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceAdjustmentRepository) BulkUpdateInvoiceAdjustmentDts(tx *gorm.DB, adjustmentDts []models.InvoiceAdjustmentDt, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentRepository-BulkUpdateInvoiceAdjustmentDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(adjustmentDts) == 0 {
		return tx, nil
	}

	if err := tx.Save(&adjustmentDts).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return tx, err
	}

	return tx, nil
}

func (r *InvoiceAdjustmentRepository) DeleteInvoiceAdjustmentDtsByIDs(tx *gorm.DB, adjustmentDtIDs []uint, userID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentRepository-DeleteInvoiceAdjustmentDtsByIDs", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(adjustmentDtIDs) == 0 {
		return tx, nil
	}

	result := tx.Model(&models.InvoiceAdjustmentDt{}).Where("id IN ?", adjustmentDtIDs).Updates(map[string]interface{}{
		"deleted_by_id": userID,
		"deleted_at":    time.Now(),
	})
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceAdjustmentRepository) DeleteInvoiceAdjustment(tx *gorm.DB, invoiceAdjustmentID uint, userID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentRepository-DeleteInvoiceAdjustment", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Model(&models.InvoiceAdjustment{}).Where("id = ?", invoiceAdjustmentID).Updates(map[string]interface{}{
		"deleted_by_id": userID,
		"deleted_at":    time.Now(),
	})
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceAdjustmentRepository) RestoreInvoiceAdjustment(ctx *fiber.Ctx, params *dtos.GetInvoiceAdjustmentParams, tx *gorm.DB, span opentracing.Span) (*dtos.InvoiceAdjustmentDetailDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentRepository-RestoreInvoiceAdjustment", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceAdjustment, err := r.GetInvoiceAdjustmentByID(ctx, params, tx, childSpan)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	var invoiceAdjustmentModel models.InvoiceAdjustment
	if err := tx.Unscoped().Model(&invoiceAdjustmentModel).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	if err := tx.Unscoped().Model(&models.InvoiceAdjustmentDt{}).Where("invoice_adjustment_id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return invoiceAdjustment, nil
}

func (r *InvoiceAdjustmentRepository) GetReferenceInvoices(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.ReferenceInvoiceListDTO, int, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentRepository-GetReferenceInvoices", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	referenceInvoices := []dtos.ReferenceInvoiceListDTO{}

	var total int

	var args []interface{}
	i := 1

	salesInvoiceQuery := `
	SELECT 
		si.id, 
		CONCAT('si_', si.id) as invoice_uuid,
		si.id as ref_id,
		'sales_invoice' as ref_type,
		si.customer_id,
		c.name as customer_name,
		si.currency_id,
		cur.name as currency_name,
		si.branch_id,
		b.name as branch_name,
		si.bank_id,
		bi.name as bank_name,
		si.invoice_no,
		TO_CHAR(si.invoice_date, 'YYYY-MM-DD') as invoice_date,
		si.grand_total as invoice_amount,
		COALESCE(si.total_adjustment, 0) as total_adjustment,
		(si.grand_total - COALESCE(si.total_adjustment, 0)) as balance_amount,
		si.remark,
		si.status,
		TO_CHAR(si.created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at,
		TO_CHAR(si.updated_at, 'YYYY-MM-DD HH24:MI:SS') as updated_at,
		si.invoice_date as raw_invoice_date,
		2 as sort_order
	FROM sales_invoices si
	LEFT JOIN customers c ON si.customer_id = c.id
	LEFT JOIN mix_values cur ON si.currency_id = cur.id
	LEFT JOIN branches b ON si.branch_id = b.id
	LEFT JOIN bank_informations bi ON si.bank_id = bi.id
	WHERE si.deleted_at IS NULL AND si.status = 'UNPAID'
	`

	if customerID, ok := filters["customer_id"]; ok && customerID != "" {
		salesInvoiceQuery += fmt.Sprintf(" AND si.customer_id = $%d", i)
		args = append(args, customerID)
		i++
	}

	if startDate, ok := filters["ref_start_date"]; ok && startDate != "" {
		salesInvoiceQuery += fmt.Sprintf(" AND si.invoice_date >= $%d", i)
		args = append(args, startDate)
		i++
	}

	if endDate, ok := filters["ref_end_date"]; ok && endDate != "" {
		salesInvoiceQuery += fmt.Sprintf(" AND si.invoice_date <= $%d", i)
		args = append(args, endDate)
		i++
	}

	if !isAdmin && branchID != nil {
		salesInvoiceQuery += fmt.Sprintf(" AND si.branch_id = $%d", i)
		args = append(args, branchID)
		i++
	}

	if isAdmin && filters["branch_id"] != "" {
		salesInvoiceQuery += fmt.Sprintf(" AND si.branch_id = $%d", i)
		args = append(args, filters["branch_id"])
		i++
	}

	salesInvoiceQuery += " AND (si.grand_total - COALESCE(si.total_adjustment, 0)) > 0"

	invoiceDpQuery := `
	SELECT 
		idp.id, 
		CONCAT('idp_', idp.id) as invoice_uuid,
		idp.id as ref_id,
		'invoice_dp' as ref_type,
		idp.customer_id,
		c.name as customer_name,
		idp.currency_id,
		cur.name as currency_name,
		idp.branch_id,
		b.name as branch_name,
		idp.bank_id,
		bi.name as bank_name,
		idp.invoice_no,
		TO_CHAR(idp.invoice_date, 'YYYY-MM-DD') as invoice_date,
		idp.grand_total as invoice_amount,
		COALESCE(idp.total_adjustment, 0) as total_adjustment,
		(idp.grand_total - COALESCE(idp.total_adjustment, 0)) as balance_amount,
		idp.remark,
		idp.status,
		TO_CHAR(idp.created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at,
		TO_CHAR(idp.updated_at, 'YYYY-MM-DD HH24:MI:SS') as updated_at,
		idp.invoice_date as raw_invoice_date,
		1 as sort_order
	FROM invoice_dps idp
	LEFT JOIN customers c ON idp.customer_id = c.id
	LEFT JOIN mix_values cur ON idp.currency_id = cur.id
	LEFT JOIN branches b ON idp.branch_id = b.id
	LEFT JOIN bank_informations bi ON idp.bank_id = bi.id
	WHERE idp.deleted_at IS NULL AND idp.status = 'UNPAID'
	`

	dpArgs := []interface{}{}
	dpIndex := 1

	if customerID, ok := filters["customer_id"]; ok && customerID != "" {
		invoiceDpQuery += fmt.Sprintf(" AND idp.customer_id = $%d", dpIndex)
		dpArgs = append(dpArgs, customerID)
		dpIndex++
	}

	if startDate, ok := filters["ref_start_date"]; ok && startDate != "" {
		invoiceDpQuery += fmt.Sprintf(" AND idp.invoice_date >= $%d", dpIndex)
		dpArgs = append(dpArgs, startDate)
		dpIndex++
	}

	if endDate, ok := filters["ref_end_date"]; ok && endDate != "" {
		invoiceDpQuery += fmt.Sprintf(" AND idp.invoice_date <= $%d", dpIndex)
		dpArgs = append(dpArgs, endDate)
		dpIndex++
	}

	if !isAdmin && branchID != nil {
		invoiceDpQuery += fmt.Sprintf(" AND idp.branch_id = $%d", dpIndex)
		dpArgs = append(dpArgs, branchID)
		dpIndex++
	}

	if isAdmin && filters["branch_id"] != "" {
		invoiceDpQuery += fmt.Sprintf(" AND idp.branch_id = $%d", dpIndex)
		dpArgs = append(dpArgs, filters["branch_id"])
		dpIndex++
	}

	invoiceDpQuery += " AND (idp.grand_total - COALESCE(idp.total_adjustment, 0)) > 0"

	invoiceMaintenanceQuery := `
	SELECT 
		im.id, 
		CONCAT('im_', im.id) as invoice_uuid,
		im.id as ref_id,
		'invoice_maintenance' as ref_type,
		im.customer_id,
		c.name as customer_name,
		im.currency_id,
		cur.name as currency_name,
		im.branch_id,
		b.name as branch_name,
		im.bank_id,
		bi.name as bank_name,
		im.invoice_no,
		TO_CHAR(im.invoice_date, 'YYYY-MM-DD') as invoice_date,
		im.grand_total as invoice_amount,
		COALESCE(im.total_adjustment, 0) as total_adjustment,
		(im.grand_total - COALESCE(im.total_adjustment, 0)) as balance_amount,
		im.remark,
		im.status,
		TO_CHAR(im.created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at,
		TO_CHAR(im.updated_at, 'YYYY-MM-DD HH24:MI:SS') as updated_at,
		im.invoice_date as raw_invoice_date,
		3 as sort_order
	FROM invoice_maintenances im
	LEFT JOIN customers c ON im.customer_id = c.id
	LEFT JOIN mix_values cur ON im.currency_id = cur.id
	LEFT JOIN branches b ON im.branch_id = b.id
	LEFT JOIN bank_informations bi ON im.bank_id = bi.id
	WHERE im.deleted_at IS NULL AND im.status = 'UNPAID' AND im.approved_status = 'APPROVED'
	`

	imArgs := []interface{}{}
	imIndex := 1

	if customerID, ok := filters["customer_id"]; ok && customerID != "" {
		invoiceMaintenanceQuery += fmt.Sprintf(" AND im.customer_id = $%d", imIndex)
		imArgs = append(imArgs, customerID)
		imIndex++
	}

	if startDate, ok := filters["ref_start_date"]; ok && startDate != "" {
		invoiceMaintenanceQuery += fmt.Sprintf(" AND im.invoice_date >= $%d", imIndex)
		imArgs = append(imArgs, startDate)
		imIndex++
	}

	if endDate, ok := filters["ref_end_date"]; ok && endDate != "" {
		invoiceMaintenanceQuery += fmt.Sprintf(" AND im.invoice_date <= $%d", imIndex)
		imArgs = append(imArgs, endDate)
		imIndex++
	}

	if !isAdmin && branchID != nil {
		invoiceMaintenanceQuery += fmt.Sprintf(" AND im.branch_id = $%d", imIndex)
		imArgs = append(imArgs, branchID)
		imIndex++
	}

	if isAdmin && filters["branch_id"] != "" {
		invoiceMaintenanceQuery += fmt.Sprintf(" AND im.branch_id = $%d", imIndex)
		imArgs = append(imArgs, filters["branch_id"])
		imIndex++
	}

	invoiceMaintenanceQuery += " AND (im.grand_total - COALESCE(im.total_adjustment, 0)) > 0"

	var finalQuery string
	var finalArgs []interface{}

	if refType, ok := filters["ref_type"]; ok && refType != "" {
		if refType == "sales_invoice" {
			finalQuery = salesInvoiceQuery
			finalArgs = args
		} else if refType == "invoice_dp" {
			finalQuery = invoiceDpQuery
			finalArgs = dpArgs
		} else if refType == "invoice_maintenance" {
			finalQuery = invoiceMaintenanceQuery
			finalArgs = imArgs
		} else {
			combinedArgs := []interface{}{}
			paramIndex := 1

			siQuery := `
			SELECT 
				si.id, 
				CONCAT('si_', si.id) as invoice_uuid,
				si.id as ref_id,
				'sales_invoice' as ref_type,
				si.customer_id,
				c.name as customer_name,
				si.currency_id,
				cur.name as currency_name,
				si.branch_id,
				b.name as branch_name,
				si.bank_id,
           		bi.name as bank_name,
				si.invoice_no,
				TO_CHAR(si.invoice_date, 'YYYY-MM-DD') as invoice_date,
				si.grand_total as invoice_amount,
				COALESCE(si.total_adjustment, 0) as total_adjustment,
				(si.grand_total - COALESCE(si.total_adjustment, 0)) as balance_amount,
				si.remark,
				si.status,
				TO_CHAR(si.created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at,
				TO_CHAR(si.updated_at, 'YYYY-MM-DD HH24:MI:SS') as updated_at,
				si.invoice_date as raw_invoice_date,
				2 as sort_order
			FROM sales_invoices si
			LEFT JOIN customers c ON si.customer_id = c.id
			LEFT JOIN mix_values cur ON si.currency_id = cur.id
			LEFT JOIN branches b ON si.branch_id = b.id
			LEFT JOIN bank_informations bi ON si.bank_id = bi.id
			WHERE si.deleted_at IS NULL AND si.status = 'UNPAID'
			`

			if customerID, ok := filters["customer_id"]; ok && customerID != "" {
				siQuery += fmt.Sprintf(" AND si.customer_id = $%d", paramIndex)
				combinedArgs = append(combinedArgs, customerID)
				paramIndex++
			}

			if startDate, ok := filters["ref_start_date"]; ok && startDate != "" {
				siQuery += fmt.Sprintf(" AND si.invoice_date >= $%d", paramIndex)
				combinedArgs = append(combinedArgs, startDate)
				paramIndex++
			}

			if endDate, ok := filters["ref_end_date"]; ok && endDate != "" {
				siQuery += fmt.Sprintf(" AND si.invoice_date <= $%d", paramIndex)
				combinedArgs = append(combinedArgs, endDate)
				paramIndex++
			}

			if !isAdmin && branchID != nil {
				siQuery += fmt.Sprintf(" AND si.branch_id = $%d", paramIndex)
				combinedArgs = append(combinedArgs, branchID)
				paramIndex++
			}

			if isAdmin && filters["branch_id"] != "" {
				siQuery += fmt.Sprintf(" AND si.branch_id = $%d", paramIndex)
				combinedArgs = append(combinedArgs, filters["branch_id"])
				paramIndex++
			}

			siQuery += " AND (si.grand_total - COALESCE(si.total_adjustment, 0)) > 0"

			idpQuery := `
			SELECT 
				idp.id, 
				CONCAT('idp_', idp.id) as invoice_uuid,
				idp.id as ref_id,
				'invoice_dp' as ref_type,
				idp.customer_id,
				c.name as customer_name,
				idp.currency_id,
				cur.name as currency_name,
				idp.branch_id,
				b.name as branch_name,
				idp.bank_id,
            	bi.name as bank_name,
				idp.invoice_no,
				TO_CHAR(idp.invoice_date, 'YYYY-MM-DD') as invoice_date,
				idp.grand_total as invoice_amount,
				COALESCE(idp.total_adjustment, 0) as total_adjustment,
				(idp.grand_total - COALESCE(idp.total_adjustment, 0)) as balance_amount,
				idp.remark,
				idp.status,
				TO_CHAR(idp.created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at,
				TO_CHAR(idp.updated_at, 'YYYY-MM-DD HH24:MI:SS') as updated_at,
				idp.invoice_date as raw_invoice_date,
				1 as sort_order
			FROM invoice_dps idp
			LEFT JOIN customers c ON idp.customer_id = c.id
			LEFT JOIN mix_values cur ON idp.currency_id = cur.id
			LEFT JOIN branches b ON idp.branch_id = b.id
			LEFT JOIN bank_informations bi ON idp.bank_id = bi.id
			WHERE idp.deleted_at IS NULL AND idp.status = 'UNPAID'
			`

			if customerID, ok := filters["customer_id"]; ok && customerID != "" {
				idpQuery += fmt.Sprintf(" AND idp.customer_id = $%d", paramIndex)
				combinedArgs = append(combinedArgs, customerID)
				paramIndex++
			}

			if startDate, ok := filters["ref_start_date"]; ok && startDate != "" {
				idpQuery += fmt.Sprintf(" AND idp.invoice_date >= $%d", paramIndex)
				combinedArgs = append(combinedArgs, startDate)
				paramIndex++
			}

			if endDate, ok := filters["ref_end_date"]; ok && endDate != "" {
				idpQuery += fmt.Sprintf(" AND idp.invoice_date <= $%d", paramIndex)
				combinedArgs = append(combinedArgs, endDate)
				paramIndex++
			}

			if !isAdmin && branchID != nil {
				idpQuery += fmt.Sprintf(" AND idp.branch_id = $%d", paramIndex)
				combinedArgs = append(combinedArgs, branchID)
				paramIndex++
			}

			if isAdmin && filters["branch_id"] != "" {
				idpQuery += fmt.Sprintf(" AND idp.branch_id = $%d", paramIndex)
				combinedArgs = append(combinedArgs, filters["branch_id"])
				paramIndex++
			}

			idpQuery += " AND (idp.grand_total - COALESCE(idp.total_adjustment, 0)) > 0"

			imQuery := `
			SELECT 
				im.id, 
				CONCAT('im_', im.id) as invoice_uuid,
				im.id as ref_id,
				'invoice_maintenance' as ref_type,
				im.customer_id,
				c.name as customer_name,
				im.currency_id,
				cur.name as currency_name,
				im.branch_id,
				b.name as branch_name,
				im.bank_id,
				bi.name as bank_name,
				im.invoice_no,
				TO_CHAR(im.invoice_date, 'YYYY-MM-DD') as invoice_date,
				im.grand_total as invoice_amount,
				COALESCE(im.total_adjustment, 0) as total_adjustment,
				(im.grand_total - COALESCE(im.total_adjustment, 0)) as balance_amount,
				im.remark,
				im.status,
				TO_CHAR(im.created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at,
				TO_CHAR(im.updated_at, 'YYYY-MM-DD HH24:MI:SS') as updated_at,
				im.invoice_date as raw_invoice_date,
				3 as sort_order
			FROM invoice_maintenances im
			LEFT JOIN customers c ON im.customer_id = c.id
			LEFT JOIN mix_values cur ON im.currency_id = cur.id
			LEFT JOIN branches b ON im.branch_id = b.id
			LEFT JOIN bank_informations bi ON im.bank_id = bi.id
			WHERE im.deleted_at IS NULL AND im.status = 'UNPAID' AND im.approved_status = 'APPROVED'
			`

			if customerID, ok := filters["customer_id"]; ok && customerID != "" {
				imQuery += fmt.Sprintf(" AND im.customer_id = $%d", paramIndex)
				combinedArgs = append(combinedArgs, customerID)
				paramIndex++
			}

			if startDate, ok := filters["ref_start_date"]; ok && startDate != "" {
				imQuery += fmt.Sprintf(" AND im.invoice_date >= $%d", paramIndex)
				combinedArgs = append(combinedArgs, startDate)
				paramIndex++
			}

			if endDate, ok := filters["ref_end_date"]; ok && endDate != "" {
				imQuery += fmt.Sprintf(" AND im.invoice_date <= $%d", paramIndex)
				combinedArgs = append(combinedArgs, endDate)
				paramIndex++
			}

			if !isAdmin && branchID != nil {
				imQuery += fmt.Sprintf(" AND im.branch_id = $%d", paramIndex)
				combinedArgs = append(combinedArgs, branchID)
				paramIndex++
			}

			if isAdmin && filters["branch_id"] != "" {
				imQuery += fmt.Sprintf(" AND im.branch_id = $%d", paramIndex)
				combinedArgs = append(combinedArgs, filters["branch_id"])
				paramIndex++
			}

			imQuery += " AND (im.grand_total - COALESCE(im.total_adjustment, 0)) > 0"

			finalQuery = siQuery + " UNION ALL " + idpQuery + " UNION ALL " + imQuery
			finalArgs = combinedArgs
		}
	} else {
		combinedArgs := []interface{}{}
		paramIndex := 1

		siQuery := `
		SELECT 
			si.id, 
			CONCAT('si_', si.id) as invoice_uuid,
			si.id as ref_id,
			'sales_invoice' as ref_type,
			si.customer_id,
			c.name as customer_name,
			si.currency_id,
			cur.name as currency_name,
			si.branch_id,
			b.name as branch_name,
			si.bank_id,
			bi.name as bank_name,
			si.invoice_no,
			TO_CHAR(si.invoice_date, 'YYYY-MM-DD') as invoice_date,
			si.grand_total as invoice_amount,
			COALESCE(si.total_adjustment, 0) as total_adjustment,
			(si.grand_total - COALESCE(si.total_adjustment, 0)) as balance_amount,
			si.remark,
			si.status,
			TO_CHAR(si.created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at,
			TO_CHAR(si.updated_at, 'YYYY-MM-DD HH24:MI:SS') as updated_at,
			si.invoice_date as raw_invoice_date,
			2 as sort_order
		FROM sales_invoices si
		LEFT JOIN customers c ON si.customer_id = c.id
		LEFT JOIN mix_values cur ON si.currency_id = cur.id
		LEFT JOIN branches b ON si.branch_id = b.id
		LEFT JOIN bank_informations bi ON si.bank_id = bi.id
		WHERE si.deleted_at IS NULL AND si.status = 'UNPAID'
		`

		if customerID, ok := filters["customer_id"]; ok && customerID != "" {
			siQuery += fmt.Sprintf(" AND si.customer_id = $%d", paramIndex)
			combinedArgs = append(combinedArgs, customerID)
			paramIndex++
		}

		if startDate, ok := filters["ref_start_date"]; ok && startDate != "" {
			siQuery += fmt.Sprintf(" AND si.invoice_date >= $%d", paramIndex)
			combinedArgs = append(combinedArgs, startDate)
			paramIndex++
		}

		if endDate, ok := filters["ref_end_date"]; ok && endDate != "" {
			siQuery += fmt.Sprintf(" AND si.invoice_date <= $%d", paramIndex)
			combinedArgs = append(combinedArgs, endDate)
			paramIndex++
		}

		if !isAdmin && branchID != nil {
			siQuery += fmt.Sprintf(" AND si.branch_id = $%d", paramIndex)
			combinedArgs = append(combinedArgs, branchID)
			paramIndex++
		}

		if isAdmin && filters["branch_id"] != "" {
			siQuery += fmt.Sprintf(" AND si.branch_id = $%d", paramIndex)
			combinedArgs = append(combinedArgs, filters["branch_id"])
			paramIndex++
		}

		siQuery += " AND (si.grand_total - COALESCE(si.total_adjustment, 0)) > 0"

		idpQuery := `
		SELECT 
			idp.id, 
			CONCAT('idp_', idp.id) as invoice_uuid,
			idp.id as ref_id,
			'invoice_dp' as ref_type,
			idp.customer_id,
			c.name as customer_name,
			idp.currency_id,
			cur.name as currency_name,
			idp.branch_id,
			b.name as branch_name,
			idp.bank_id,
			bi.name as bank_name,
			idp.invoice_no,
			TO_CHAR(idp.invoice_date, 'YYYY-MM-DD') as invoice_date,
			idp.grand_total as invoice_amount,
			COALESCE(idp.total_adjustment, 0) as total_adjustment,
			(idp.grand_total - COALESCE(idp.total_adjustment, 0)) as balance_amount,
			idp.remark,
			idp.status,
			TO_CHAR(idp.created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at,
			TO_CHAR(idp.updated_at, 'YYYY-MM-DD HH24:MI:SS') as updated_at,
			idp.invoice_date as raw_invoice_date,
			1 as sort_order
		FROM invoice_dps idp
		LEFT JOIN customers c ON idp.customer_id = c.id
		LEFT JOIN mix_values cur ON idp.currency_id = cur.id
		LEFT JOIN branches b ON idp.branch_id = b.id
		LEFT JOIN bank_informations bi ON idp.bank_id = bi.id
		WHERE idp.deleted_at IS NULL AND idp.status = 'UNPAID'
		`

		if customerID, ok := filters["customer_id"]; ok && customerID != "" {
			idpQuery += fmt.Sprintf(" AND idp.customer_id = $%d", paramIndex)
			combinedArgs = append(combinedArgs, customerID)
			paramIndex++
		}

		if startDate, ok := filters["ref_start_date"]; ok && startDate != "" {
			idpQuery += fmt.Sprintf(" AND idp.invoice_date >= $%d", paramIndex)
			combinedArgs = append(combinedArgs, startDate)
			paramIndex++
		}

		if endDate, ok := filters["ref_end_date"]; ok && endDate != "" {
			idpQuery += fmt.Sprintf(" AND idp.invoice_date <= $%d", paramIndex)
			combinedArgs = append(combinedArgs, endDate)
			paramIndex++
		}

		if !isAdmin && branchID != nil {
			idpQuery += fmt.Sprintf(" AND idp.branch_id = $%d", paramIndex)
			combinedArgs = append(combinedArgs, branchID)
			paramIndex++
		}

		if isAdmin && filters["branch_id"] != "" {
			idpQuery += fmt.Sprintf(" AND idp.branch_id = $%d", paramIndex)
			combinedArgs = append(combinedArgs, filters["branch_id"])
			paramIndex++
		}

		idpQuery += " AND (idp.grand_total - COALESCE(idp.total_adjustment, 0)) > 0"

		imQuery := `
		SELECT 
			im.id, 
			CONCAT('im_', im.id) as invoice_uuid,
			im.id as ref_id,
			'invoice_maintenance' as ref_type,
			im.customer_id,
			c.name as customer_name,
			im.currency_id,
			cur.name as currency_name,
			im.branch_id,
			b.name as branch_name,
			im.bank_id,
			bi.name as bank_name,
			im.invoice_no,
			TO_CHAR(im.invoice_date, 'YYYY-MM-DD') as invoice_date,
			im.grand_total as invoice_amount,
			COALESCE(im.total_adjustment, 0) as total_adjustment,
			(im.grand_total - COALESCE(im.total_adjustment, 0)) as balance_amount,
			im.remark,
			im.status,
			TO_CHAR(im.created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at,
			TO_CHAR(im.updated_at, 'YYYY-MM-DD HH24:MI:SS') as updated_at,
			im.invoice_date as raw_invoice_date,
			3 as sort_order
		FROM invoice_maintenances im
		LEFT JOIN customers c ON im.customer_id = c.id
		LEFT JOIN mix_values cur ON im.currency_id = cur.id
		LEFT JOIN branches b ON im.branch_id = b.id
		LEFT JOIN bank_informations bi ON im.bank_id = bi.id
		WHERE im.deleted_at IS NULL AND im.status = 'UNPAID' AND im.approved_status = 'APPROVED'
		`

		if customerID, ok := filters["customer_id"]; ok && customerID != "" {
			imQuery += fmt.Sprintf(" AND im.customer_id = $%d", paramIndex)
			combinedArgs = append(combinedArgs, customerID)
			paramIndex++
		}

		if startDate, ok := filters["ref_start_date"]; ok && startDate != "" {
			imQuery += fmt.Sprintf(" AND im.invoice_date >= $%d", paramIndex)
			combinedArgs = append(combinedArgs, startDate)
			paramIndex++
		}

		if endDate, ok := filters["ref_end_date"]; ok && endDate != "" {
			imQuery += fmt.Sprintf(" AND im.invoice_date <= $%d", paramIndex)
			combinedArgs = append(combinedArgs, endDate)
			paramIndex++
		}

		if !isAdmin && branchID != nil {
			imQuery += fmt.Sprintf(" AND im.branch_id = $%d", paramIndex)
			combinedArgs = append(combinedArgs, branchID)
			paramIndex++
		}

		if isAdmin && filters["branch_id"] != "" {
			imQuery += fmt.Sprintf(" AND im.branch_id = $%d", paramIndex)
			combinedArgs = append(combinedArgs, filters["branch_id"])
			paramIndex++
		}

		imQuery += " AND (im.grand_total - COALESCE(im.total_adjustment, 0)) > 0"

		finalQuery = siQuery + " UNION ALL " + idpQuery + " UNION ALL " + imQuery
		finalArgs = combinedArgs
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) as count_query", finalQuery)

	err := r.sqlDB.GetContext(ctx.Context(), &total, countQuery, finalArgs...)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, 0, err
	}

	finalQuery += " ORDER BY raw_invoice_date ASC, sort_order ASC"

	perPage := utils.GetIntOrDefault(filters["per_page"], 10)
	currentPage := utils.GetIntOrDefault(filters["page"], 1)

	finalQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(finalArgs)+1, len(finalArgs)+2)
	finalArgs = append(finalArgs, perPage, (currentPage-1)*perPage)

	err = r.sqlDB.SelectContext(ctx.Context(), &referenceInvoices, finalQuery, finalArgs...)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, 0, err
	}

	return referenceInvoices, total, nil
}

func (r *InvoiceAdjustmentRepository) LockInvoiceAdjustment(tx *gorm.DB, invoiceAdjustmentID uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentRepository-LockInvoiceAdjustment", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	lockQuery := "SELECT id FROM invoice_adjustments WHERE id = ? FOR UPDATE"
	if err := tx.Exec(lockQuery, invoiceAdjustmentID).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *InvoiceAdjustmentRepository) LockReferenceInvoices(tx *gorm.DB, refIDs []uint, refType string, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentRepository-LockReferenceInvoices", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(refIDs) == 0 {
		return nil
	}

	var lockQuery string
	if refType == "sales_invoice" {
		lockQuery = "SELECT id FROM sales_invoices WHERE id IN ? FOR UPDATE"
	} else if refType == "invoice_dp" {
		lockQuery = "SELECT id FROM invoice_dps WHERE id IN ? FOR UPDATE"
	} else if refType == "invoice_maintenance" {
		lockQuery = "SELECT id FROM invoice_maintenances WHERE id IN ? FOR UPDATE"
	} else {
		return fmt.Errorf("invalid reference type: %s", refType)
	}

	if err := tx.Exec(lockQuery, refIDs).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *InvoiceAdjustmentRepository) GetInvoiceAdjustmentForUpdate(tx *gorm.DB, invoiceAdjustmentID uint, span opentracing.Span) (*models.InvoiceAdjustment, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentRepository-GetInvoiceAdjustmentForUpdate", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var invoiceAdjustment models.InvoiceAdjustment
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", invoiceAdjustmentID).First(&invoiceAdjustment).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return &invoiceAdjustment, nil
}

func (r *InvoiceAdjustmentRepository) GetInvoiceAdjustmentCreatedThisMonth(ctx *fiber.Ctx, tx *gorm.DB, customerID uint, span opentracing.Span) (int, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentRepository-GetCustomerInvoiceAdjustmentCreatedThisMonth", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var count int
	query := `
    SELECT COUNT(*) 
    FROM invoice_adjustments 
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

func (r *InvoiceAdjustmentRepository) UpdateReferenceInvoiceAdjustmentAmount(tx *gorm.DB, refID uint, refType string, adjustmentAmount float64, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentRepository-UpdateReferenceInvoiceAdjustmentAmount", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if refType == "sales_invoice" {
		var query string
		if adjustmentAmount < 0 {
			query = `
			UPDATE sales_invoices 
			SET history_status = status,
			    history_total_adjustment = total_adjustment,
			    total_adjustment = NULL,
			    status = 'UNPAID'
			WHERE id = ?
			`
		} else {
			query = `
			UPDATE sales_invoices 
			SET total_adjustment = COALESCE(total_adjustment, 0) + ?,
			    status = CASE 
			        WHEN (grand_total - (COALESCE(total_adjustment, 0) + ?)) <= 0 THEN 'PAID' 
			        ELSE status 
			    END
			WHERE id = ?
			`
			result := tx.Exec(query, adjustmentAmount, adjustmentAmount, refID)
			if result.Error != nil {
				utils.LogErrors(childSpan, result.Error)
				return tx, result.Error
			}
			return tx, nil
		}

		result := tx.Exec(query, refID)
		if result.Error != nil {
			utils.LogErrors(childSpan, result.Error)
			return tx, result.Error
		}
	} else if refType == "invoice_dp" {
		var query string
		if adjustmentAmount < 0 {
			query = `
			UPDATE invoice_dps 
			SET history_status = status,
			    history_total_adjustment = total_adjustment,
			    total_adjustment = NULL,
			    status = 'UNPAID'
			WHERE id = ?
			`
		} else {
			query = `
			UPDATE invoice_dps 
			SET total_adjustment = COALESCE(total_adjustment, 0) + ?,
			    status = CASE 
			        WHEN (grand_total - (COALESCE(total_adjustment, 0) + ?)) <= 0 THEN 'PAID' 
			        ELSE status 
			    END
			WHERE id = ?
			`
			result := tx.Exec(query, adjustmentAmount, adjustmentAmount, refID)
			if result.Error != nil {
				utils.LogErrors(childSpan, result.Error)
				return tx, result.Error
			}
			return tx, nil
		}

		result := tx.Exec(query, refID)
		if result.Error != nil {
			utils.LogErrors(childSpan, result.Error)
			return tx, result.Error
		}
	} else if refType == "invoice_maintenance" {
		var query string
		if adjustmentAmount < 0 {
			query = `
			UPDATE invoice_maintenances 
			SET history_status = status,
			    history_total_adjustment = total_adjustment,
			    total_adjustment = NULL,
			    status = 'UNPAID'
			WHERE id = ?
			`
		} else {
			query = `
			UPDATE invoice_maintenances 
			SET total_adjustment = COALESCE(total_adjustment, 0) + ?,
			    status = CASE 
			        WHEN (grand_total - (COALESCE(total_adjustment, 0) + ?)) <= 0 THEN 'PAID' 
			        ELSE status 
			    END
			WHERE id = ?
			`
			result := tx.Exec(query, adjustmentAmount, adjustmentAmount, refID)
			if result.Error != nil {
				utils.LogErrors(childSpan, result.Error)
				return tx, result.Error
			}
			return tx, nil
		}

		result := tx.Exec(query, refID)
		if result.Error != nil {
			utils.LogErrors(childSpan, result.Error)
			return tx, result.Error
		}
	} else {
		return tx, fmt.Errorf("invalid reference type: %s", refType)
	}

	return tx, nil
}

func (r *InvoiceAdjustmentRepository) RestoreReferenceInvoiceAdjustmentAmount(tx *gorm.DB, refID uint, refType string, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentRepository-RestoreReferenceInvoiceAdjustmentAmount", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if refType == "sales_invoice" {
		query := `
		UPDATE sales_invoices 
		SET total_adjustment = history_total_adjustment,
		    status = history_status,
		    history_total_adjustment = NULL,
		    history_status = NULL
		WHERE id = ?
		`
		result := tx.Exec(query, refID)
		if result.Error != nil {
			utils.LogErrors(childSpan, result.Error)
			return tx, result.Error
		}
	} else if refType == "invoice_dp" {
		query := `
		UPDATE invoice_dps 
		SET total_adjustment = history_total_adjustment,
		    status = history_status,
		    history_total_adjustment = NULL,
		    history_status = NULL
		WHERE id = ?
		`
		result := tx.Exec(query, refID)
		if result.Error != nil {
			utils.LogErrors(childSpan, result.Error)
			return tx, result.Error
		}
	} else if refType == "invoice_maintenance" {
		query := `
		UPDATE invoice_maintenances 
		SET total_adjustment = history_total_adjustment,
		    status = history_status,
		    history_total_adjustment = NULL,
		    history_status = NULL
		WHERE id = ?
		`
		result := tx.Exec(query, refID)
		if result.Error != nil {
			utils.LogErrors(childSpan, result.Error)
			return tx, result.Error
		}
	} else {
		return tx, fmt.Errorf("invalid reference type: %s", refType)
	}

	return tx, nil
}

func (r *InvoiceAdjustmentRepository) Commit(tx *gorm.DB) error {
	return tx.Commit().Error
}
