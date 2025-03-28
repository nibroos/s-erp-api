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

type InvoiceDpRepository struct {
	db       *gorm.DB
	sqlDB    *sqlx.DB
	utilRepo *UtilRepository
	tracer   opentracing.Tracer
}

func NewInvoiceDpRepository(db *gorm.DB, sqlDB *sqlx.DB, utilRepo *UtilRepository, tracer opentracing.Tracer) *InvoiceDpRepository {
	return &InvoiceDpRepository{
		db:       db,
		sqlDB:    sqlDB,
		tracer:   tracer,
		utilRepo: utilRepo,
	}
}

func (r *InvoiceDpRepository) GetInvoiceDps(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.InvoiceDpListDTO, int, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-GetInvoiceDps", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	invoiceDps := []dtos.InvoiceDpListDTO{}

	var total int

	filterDBColumnKey := []string{
		"idp.invoice_no", "idp.remark",
		"c.name",
		"idt.remark",
		"idtb.remark",
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
		"customer_id":     "idp.customer_id",
		"currency_id":     "idp.currency_id",
		"payment_term_id": "idp.payment_term_id",
		"vat_id":          "idp.vat_id",
		"pph23_id":        "idp.pph23_id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":     "idp.customer_id",
		"currency_ids":     "idp.currency_id",
		"payment_term_ids": "idp.payment_term_id",
		"pph23_ids":        "idp.pph23_id",
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
		"vat_ids": []string{"idp.vat_id", "idt.vat_id"},
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
        SELECT DISTINCT ON (idp.id)
					idp.id, idp.customer_id, idp.currency_id, idp.payment_term_id, idp.vat_id, idp.pph23_id, idp.branch_id,
					idp.invoice_no, idp.remark, 
					idp.exchange_rate, idp.pph23_percentage, idp.vat_percentage, idp.dp_percentage, idp.total_qty, idp.subtotal, idp.total_discount, idp.total_pph23, idp.total_vat, idp.grand_total, idp.created_by_id, idp.updated_by_id, idp.deleted_by_id, idp.created_at, idp.updated_at, idp.deleted_at,
					TO_CHAR(idp.invoice_date, 'YYYY-MM-DD') as invoice_date,
					idp.discount_amount, idp.discount_percentage, idp.discount_percentage_amount, idp.discount_final, idp.discount_type, idp.total_amount_products,

					c.name as customer_name,
					cur.name as currency_name,
					pt.name as payment_term_name,
					vat.name as vat_name,
					pph.name as pph23_name,
					b.name as branch_name,

					idt.remark as invoice_dp_dt_remark,
					idtb.remark as invoice_dp_dt_bom_remark,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM invoice_dps idp
				LEFT JOIN invoice_dp_dts idt ON idt.invoice_dp_id = idp.id
				LEFT JOIN invoice_dp_dt_boms idtb ON idtb.invoice_dp_dt_id = idt.id
				LEFT JOIN customers c ON idp.customer_id = c.id
				LEFT JOIN mix_values cur ON idp.currency_id = cur.id
				LEFT JOIN mix_values pt ON idp.payment_term_id = pt.id
				LEFT JOIN mix_values vat ON idp.vat_id = vat.id
				LEFT JOIN mix_values pph ON idp.pph23_id = pph.id
				LEFT JOIN branches b ON idp.branch_id = b.id

        LEFT JOIN users cu ON idp.created_by_id = cu.id
        LEFT JOIN users uu ON idp.updated_by_id = uu.id
				WHERE 1=1` + condition + queryGlobal + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery

	for key, value := range filters {
		switch key {
		case "invoice_no", "remark":
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

		err := r.sqlDB.SelectContext(ctx.Context(), &invoiceDps, query, args...)
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

	return invoiceDps, total, nil
}

func (r *InvoiceDpRepository) GetInvoiceDpByID(ctx *fiber.Ctx, params *dtos.GetInvoiceDpParams, tx *gorm.DB, span opentracing.Span) (*dtos.InvoiceDpDetailDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-GetInvoiceDpByID", opentracing.ChildOf(span.Context()))
	var invoiceDp dtos.InvoiceDpDetailDTO

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	baseQuery := `
    FROM ( 
			SELECT DISTINCT ON (idp.id)
				idp.id, idp.customer_id, idp.currency_id, idp.payment_term_id, idp.vat_id, idp.pph23_id, idp.branch_id,
				idp.invoice_no, idp.remark, 
				idp.exchange_rate, idp.pph23_percentage, idp.vat_percentage, idp.dp_percentage, idp.total_qty, idp.subtotal, idp.total_discount, idp.total_pph23, idp.total_vat, idp.grand_total, idp.created_by_id, idp.updated_by_id, idp.deleted_by_id, idp.created_at, idp.updated_at, idp.deleted_at,
				TO_CHAR(idp.invoice_date, 'YYYY-MM-DD') as invoice_date,
				idp.discount_amount, idp.discount_percentage, idp.discount_percentage_amount, idp.discount_final, idp.discount_type, idp.total_amount_products,

				cu.name as created_by_name,
				uu.name as updated_by_name

			FROM invoice_dps idp
			LEFT JOIN invoice_dp_dts idt ON idt.invoice_dp_id = idp.id
			LEFT JOIN invoice_dp_dt_boms idtb ON idtb.invoice_dp_dt_id = idt.id

			LEFT JOIN users cu ON idp.created_by_id = cu.id
			LEFT JOIN users uu ON idp.updated_by_id = uu.id
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

	if err := r.sqlDB.Get(&invoiceDp, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		childSpan.LogKV("query", query)
		return nil, err
	}

	return &invoiceDp, nil
}

func (r *InvoiceDpRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *InvoiceDpRepository) Rollback() *gorm.DB {
	return r.db.Rollback()
}

func (r *InvoiceDpRepository) CreateInvoiceDp(tx *gorm.DB, invoiceDp *models.InvoiceDp, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-CreateInvoiceDp", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Create(&invoiceDp)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceDpRepository) CreateInvoiceDpDts(tx *gorm.DB, invoiceDpDts []models.InvoiceDpDt, span opentracing.Span) (*gorm.DB, []models.InvoiceDpDt, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-CreateInvoiceDpDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Create(&invoiceDpDts)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, nil, result.Error
	}

	return tx, invoiceDpDts, nil
}

func (r *InvoiceDpRepository) BulkCreateInvoiceDpDtBoms(tx *gorm.DB, invoiceDpDtBoms []map[string]interface{}, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-BulkCreateInvoiceDpDtBoms", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(invoiceDpDtBoms) == 0 {
		return tx, nil
	}

	result := tx.Table("invoice_dp_dt_boms").Create(invoiceDpDtBoms)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceDpRepository) GetInvoiceDpDts(ctx *fiber.Ctx, invoiceDpID uint, span opentracing.Span) ([]dtos.InvoiceDpDtListDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-GetInvoiceDpDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceDpDts := []dtos.InvoiceDpDtListDTO{}

	query := `
	SELECT 
		idt.id, idt.product_uuid, idt.invoice_dp_id, idt.item_unit_id, idt.vat_id, idt.pph23_id, 
		idt.ref_id, idt.product_id, idt.ref_type, idt.product_type, idt.remark, 
		idt.dp_percentage, idt.is_vat, idt.is_pph23, idt.qty, idt.price, idt.subtotal, 
		idt.discount_amount, idt.discount_percentage, idt.discount_percentage_num, 
		idt.discount_percentage_amount, idt.discount_final, idt.discount_type, 
		idt.total_amount, idt.total_dp, idt.created_by_id, idt.updated_by_id, idt.deleted_by_id, 
		idt.created_at, idt.updated_at, idt.deleted_at,
		
		p.name as product_name, p.code as product_code,
		iu.name as unit_name,
		v.name as vat_name,
		pph.name as pph23_name,
		
		cu.name as created_by_name,
		uu.name as updated_by_name
	FROM invoice_dp_dts idt
	LEFT JOIN products p ON idt.product_id = p.id
	LEFT JOIN item_units iu ON idt.item_unit_id = iu.id
	LEFT JOIN mix_values v ON idt.vat_id = v.id
	LEFT JOIN mix_values pph ON idt.pph23_id = pph.id
	LEFT JOIN users cu ON idt.created_by_id = cu.id
	LEFT JOIN users uu ON idt.updated_by_id = uu.id
	WHERE idt.invoice_dp_id = $1 AND idt.deleted_at IS NULL
	ORDER BY idt.id ASC
	`

	err := r.sqlDB.SelectContext(ctx.Context(), &invoiceDpDts, query, invoiceDpID)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return invoiceDpDts, nil
}

func (r *InvoiceDpRepository) GetInvoiceDpDtBoms(ctx *fiber.Ctx, invoiceDpID uint, span opentracing.Span) ([]dtos.InvoiceDpDtBomListDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-GetInvoiceDpDtBoms", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceDpDtBoms := []dtos.InvoiceDpDtBomListDTO{}

	query := `
	SELECT 
		idtb.id, idtb.product_uuid, idtb.invoice_dp_id, idtb.invoice_dp_dt_id, 
		idtb.bom_id, idtb.product_id, idtb.item_unit_id, idtb.remark, 
		idtb.qty, idtb.price, idtb.subtotal, idtb.created_by_id, idtb.updated_by_id, 
		idtb.deleted_by_id, idtb.created_at, idtb.updated_at, idtb.deleted_at,
		
		p.name as product_name, p.code as product_code,
		iu.name as unit_name,
		
		cu.name as created_by_name,
		uu.name as updated_by_name
	FROM invoice_dp_dt_boms idtb
	LEFT JOIN products p ON idtb.product_id = p.id
	LEFT JOIN item_units iu ON idtb.item_unit_id = iu.id
	LEFT JOIN users cu ON idtb.created_by_id = cu.id
	LEFT JOIN users uu ON idtb.updated_by_id = uu.id
	WHERE idtb.invoice_dp_id = $1 AND idtb.deleted_at IS NULL
	ORDER BY idtb.id ASC
	`

	err := r.sqlDB.SelectContext(ctx.Context(), &invoiceDpDtBoms, query, invoiceDpID)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return invoiceDpDtBoms, nil
}

func (r *InvoiceDpRepository) UpdateInvoiceDp(tx *gorm.DB, invoiceDp *models.InvoiceDp, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-UpdateInvoiceDp", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Model(&models.InvoiceDp{}).Where("id = ?", invoiceDp.ID).Updates(invoiceDp)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceDpRepository) BulkCreateInvoiceDpDts(tx *gorm.DB, invoiceDpDts []models.InvoiceDpDt, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-BulkCreateInvoiceDpDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(invoiceDpDts) == 0 {
		return tx, nil
	}

	result := tx.Create(&invoiceDpDts)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceDpRepository) BulkUpdateInvoiceDpDts(tx *gorm.DB, invoiceDpDts []models.InvoiceDpDt, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-BulkUpdateInvoiceDpDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(invoiceDpDts) == 0 {
		return tx, nil
	}

	for _, invoiceDpDt := range invoiceDpDts {
		result := tx.Model(&models.InvoiceDpDt{}).Where("id = ?", invoiceDpDt.ID).Updates(invoiceDpDt)
		if result.Error != nil {
			utils.LogErrors(childSpan, result.Error)
			return tx, result.Error
		}
	}

	return tx, nil
}

func (r *InvoiceDpRepository) BulkUpdateInvoiceDpDtBoms(tx *gorm.DB, invoiceDpDtBoms []map[string]interface{}, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-BulkUpdateInvoiceDpDtBoms", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(invoiceDpDtBoms) == 0 {
		return tx, nil
	}

	for _, invoiceDpDtBom := range invoiceDpDtBoms {
		result := tx.Model(&models.InvoiceDpDtBom{}).Where("id = ?", invoiceDpDtBom["id"]).Updates(invoiceDpDtBom)
		if result.Error != nil {
			utils.LogErrors(childSpan, result.Error)
			return tx, result.Error
		}
	}

	return tx, nil
}

func (r *InvoiceDpRepository) DeleteInvoiceDpDtBomsByIDs(tx *gorm.DB, invoiceDpDtBomIDs []uint, userID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-DeleteInvoiceDpDtBomsByIDs", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(invoiceDpDtBomIDs) == 0 {
		return tx, nil
	}

	result := tx.Model(&models.InvoiceDpDtBom{}).Where("id IN ?", invoiceDpDtBomIDs).Updates(map[string]interface{}{
		"deleted_by_id": userID,
		"deleted_at":    time.Now(),
	})
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceDpRepository) DeleteInvoiceDpDtsByIDs(tx *gorm.DB, invoiceDpDtIDs []uint, userID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-DeleteInvoiceDpDtsByIDs", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(invoiceDpDtIDs) == 0 {
		return tx, nil
	}

	result := tx.Model(&models.InvoiceDpDt{}).Where("id IN ?", invoiceDpDtIDs).Updates(map[string]interface{}{
		"deleted_by_id": userID,
		"deleted_at":    time.Now(),
	})
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceDpRepository) DeleteInvoiceDp(tx *gorm.DB, invoiceDpID uint, userID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-DeleteInvoiceDp", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Model(&models.InvoiceDp{}).Where("id = ?", invoiceDpID).Updates(map[string]interface{}{
		"deleted_by_id": userID,
		"deleted_at":    time.Now(),
	})
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceDpRepository) GetRefSalesOrderDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefSalesOrderDtListDTO, int, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-GetRefSalesOrderDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	soDts := []dtos.RefSalesOrderDtListDTO{}

	var total int

	filterDBColumnKey := []string{
		"so.sales_order_no", "so.po_buyer_no", "so.remark",
		"c.name",
		"sodt.remark",
		"sodtb.remark",
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
					sodt.disc_final, sodt.disc_type, sodt.total_am, sodt.created_by_id, 
					sodt.updated_by_id, sodt.deleted_by_id, sodt.created_at, sodt.updated_at, sodt.deleted_at,

					so.customer_id, so.order_type_id, so.currency_id, so.vat_id as head_vat_id, 
					so.pph23_id as head_pph23_id, so.vat_perc as head_vat_perc, 
					so.pph23_perc as head_pph23_perc, so.disc_am as head_disc_am, 
					so.disc_perc as head_disc_perc, so.markup_perc as head_markup_perc, 
					so.remark as head_remark, so.exchange_rate, so.sales_order_no, 
					TO_CHAR(so.due_at, 'YYYY-MM-DD') as due_at,

					c.name as customer_name,
					p.name as product_name, p.code as product_code, p.sku as item_sku,
					iu.name as unit_name,
					v.name as vat_name,
					pph.name as pph23_name,
					
					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM so_dts sodt
				LEFT JOIN sales_orders so ON sodt.sales_order_id = so.id
				LEFT JOIN so_dt_boms sodtb ON sodtb.so_dt_id = sodt.id
				LEFT JOIN customers c ON so.customer_id = c.id
				LEFT JOIN products p ON sodt.item_id = p.id
				LEFT JOIN item_units iu ON sodt.item_unit_id = iu.id
				LEFT JOIN mix_values v ON sodt.vat_id = v.id
				LEFT JOIN mix_values pph ON sodt.pph23_id = pph.id

        LEFT JOIN users cu ON sodt.created_by_id = cu.id
        LEFT JOIN users uu ON sodt.updated_by_id = uu.id
				WHERE 1=1 AND so.status = 'approved'` + condition + queryGlobal + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery

	for key, value := range filters {
		switch key {
		case "sales_order_no", "po_buyer_no", "remark":
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

func (r *InvoiceDpRepository) GetSoDtBoms(ctx *fiber.Ctx, soDtIDs []uint, span opentracing.Span) ([]dtos.SalesOrderSoDtBomListDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-GetSoDtBoms", opentracing.ChildOf(span.Context()))
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
		
		p.name as product_name, p.code as product_code, p.sku as item_sku,
		p.barcode as item_barcode, p.factory_code as item_factory_code,
		p.specification as item_specification,
		iu.name as unit_name,
		
		cu.name as created_by_name,
		uu.name as updated_by_name
	FROM so_dt_boms sodtb
	LEFT JOIN products p ON sodtb.item_id = p.id
	LEFT JOIN item_units iu ON sodtb.item_unit_id = iu.id
	LEFT JOIN users cu ON sodtb.created_by_id = cu.id
	LEFT JOIN users uu ON sodtb.updated_by_id = uu.id
	WHERE sodtb.so_dt_id IN (?) AND sodtb.deleted_at IS NULL
	ORDER BY sodtb.id ASC
	`

	query = r.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Raw(query, soDtIDs)
	})

	err := r.sqlDB.SelectContext(ctx.Context(), &soDtBoms, query, soDtIDs)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return soDtBoms, nil
}

func (r *InvoiceDpRepository) GetSoDtQtyUpdateForInvoice(ctx *fiber.Ctx, soDtIDs []uint, span opentracing.Span) ([]dtos.GetSoDtQtyUpdateForInvoiceDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-GetSoDtQtyUpdateForInvoice", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(soDtIDs) == 0 {
		return []dtos.GetSoDtQtyUpdateForInvoiceDTO{}, nil
	}

	soDtsQtyUpdate := []dtos.GetSoDtQtyUpdateForInvoiceDTO{}

	query := `
	SELECT 
		sodt.id, sodt.id as so_dt_id, sodt.sales_order_id,
		COALESCE(sodt.qty_invoiced, 0) as qty_invoiced
	FROM so_dts sodt
	WHERE sodt.id IN (?) AND sodt.deleted_at IS NULL
	`

	query = r.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Raw(query, soDtIDs)
	})

	err := r.sqlDB.SelectContext(ctx.Context(), &soDtsQtyUpdate, query, soDtIDs)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return soDtsQtyUpdate, nil
}

func (r *InvoiceDpRepository) BulkUpdateSoDtsQty(tx *gorm.DB, soDtsQtyUpdate []map[string]interface{}, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-BulkUpdateSoDtsQty", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(soDtsQtyUpdate) == 0 {
		return tx, nil
	}

	for _, soDt := range soDtsQtyUpdate {
		result := tx.Model(&models.SoDt{}).Where("id = ?", soDt["id"]).Updates(map[string]interface{}{
			"qty_invoiced": soDt["qty_invoiced"],
		})
		if result.Error != nil {
			utils.LogErrors(childSpan, result.Error)
			return tx, result.Error
		}
	}

	return tx, nil
}

func (r *InvoiceDpRepository) UpdateSalesOrderStatus(tx *gorm.DB, salesOrderID uint, status string, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-UpdateSalesOrderStatus", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Model(&models.SalesOrder{}).Where("id = ?", salesOrderID).Updates(map[string]interface{}{
		"status": status,
	})
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceDpRepository) Commit(tx *gorm.DB) error {
	return tx.Commit().Error
}
