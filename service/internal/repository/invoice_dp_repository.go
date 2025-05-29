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
		"idp.invoice_no", "idp.remark", "idp.status", "idp.title",
		"c.name",
		"idt.remark",
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
		condition += fmt.Sprintf(" AND idp.id IN (%s)", filters["ids"])
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

	if value, ok := filters["status"]; ok && value != "" {
		condition += fmt.Sprintf(" AND idp.status = $%d", i)
		args = append(args, value)
		i++
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
		"vat_ids": {"idp.vat_id", "idt.vat_id"},
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

		filterDateTypeKey := map[string]string{
			"invoice_date": "idp.invoice_date",
			"due_date":     "idp.due_date",
		}

		dateTypeColumn := "idp.invoice_date"
		for key := range filterDateTypeKey {
			if key == filters["date_type"] {
				dateTypeColumn = filterDateTypeKey[filters["date_type"]]
			}
		}

		condition += fmt.Sprintf(" AND (%s BETWEEN $%d AND $%d)", dateTypeColumn, i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
	}

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (idp.id)
					idp.id, idp.customer_id, idp.currency_id, idp.payment_term_id, idp.vat_id, idp.pph23_id, idp.branch_id, idp.bank_id,
					idp.invoice_no, idp.remark, idp.status, idp.title, bk.name as bank_name, bk.account_number, bk.account_name, bk.account_name,
					idp.exchange_rate, idp.pph23_percentage, idp.vat_percentage, idp.dp_percentage, idp.total_qty, idp.subtotal, idp.total_discount, idp.total_pph23, idp.total_vat, idp.grand_total, idp.created_by_id, idp.updated_by_id, idp.deleted_by_id, idp.created_at, idp.updated_at, idp.deleted_at,
					TO_CHAR(idp.invoice_date, 'YYYY-MM-DD') as invoice_date, TO_CHAR(idp.due_date, 'YYYY-MM-DD') as due_date,
					idp.discount_amount, idp.discount_percentage, idp.discount_percentage_amount, idp.discount_final, idp.discount_type, idp.total_amount_products, idp.total_dp_products,

					c.name as customer_name,
					cur.name as currency_name,
					pt.name as payment_term_name,
					vat.name as vat_name,
					pph.name as pph23_name,
					b.name as branch_name,

					idt.remark as invoice_dp_dt_remark,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM invoice_dps idp
				LEFT JOIN invoice_dp_dts idt ON idt.invoice_dp_id = idp.id
				LEFT JOIN customers c ON idp.customer_id = c.id
				LEFT JOIN mix_values cur ON idp.currency_id = cur.id
				LEFT JOIN mix_values pt ON idp.payment_term_id = pt.id
				LEFT JOIN mix_values vat ON idp.vat_id = vat.id
				LEFT JOIN mix_values pph ON idp.pph23_id = pph.id
				LEFT JOIN branches b ON idp.branch_id = b.id
				LEFT JOIN bank_informations bk ON idp.bank_id = bk.id

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
		case "invoice_no", "remark", "status", "title":
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
            idp.id, idp.customer_id, idp.currency_id, idp.payment_term_id, idp.vat_id, idp.pph23_id, idp.branch_id, idp.bank_id,
            idp.invoice_no, idp.remark, idp.status, idp.title,
            idp.exchange_rate, idp.pph23_percentage, idp.vat_percentage, idp.dp_percentage, idp.total_qty, idp.subtotal, idp.total_discount, idp.total_pph23, idp.total_vat, idp.grand_total, idp.created_by_id, idp.updated_by_id, idp.deleted_by_id, idp.created_at, idp.updated_at, idp.deleted_at,
            TO_CHAR(idp.invoice_date, 'YYYY-MM-DD') as invoice_date, TO_CHAR(idp.due_date, 'YYYY-MM-DD') as due_date,
            idp.discount_amount, idp.discount_percentage, idp.discount_percentage_amount, idp.discount_final, idp.discount_type, idp.total_amount_products, idp.total_dp_products, idp.rev_no,

            br.company_profile_id,

			b.name as bank_name,
            b.account_name,
            b.account_number,

            cu.name as created_by_name,
            uu.name as updated_by_name,

			c.name as customer_name,
            c.code as customer_code,
            c.phone,
            c.pic,
            c.address

        FROM invoice_dps idp
        LEFT JOIN invoice_dp_dts idt ON idt.invoice_dp_id = idp.id
		LEFT JOIN customers c ON idp.customer_id = c.id
        LEFT JOIN users cu ON idp.created_by_id = cu.id
        LEFT JOIN users uu ON idp.updated_by_id = uu.id
        LEFT JOIN branches br ON idp.branch_id = br.id
		LEFT JOIN bank_informations b ON idp.bank_id = b.id
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

	if err := r.sqlDB.Get(&invoiceDp, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		childSpan.LogKV("query", query)
		return nil, err
	}

	return &invoiceDp, nil
}

func (r *InvoiceDpRepository) GetUpdatedInvoiceDpDts(ctx *fiber.Ctx, invoiceDpID uint, span opentracing.Span) ([]dtos.InvoiceDpDtListUpdateDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-GetUpdatedInvoiceDpDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceDpDts := []dtos.InvoiceDpDtListUpdateDTO{}

	query := `
	SELECT 
		idt.id, idt.product_uuid, idt.invoice_dp_id, idt.item_unit_id, idt.vat_id, idt.pph23_id, 
		idt.ref_id, idt.ref_dt_id, idt.product_id, idt.ref_type, idt.product_type, idt.remark, 
		idt.dp_percentage, idt.is_vat, idt.is_pph23, idt.qty, idt.price, idt.subtotal,
		idt.discount, idt.total_amount, idt.total_dp, idt.created_by_id, idt.updated_by_id, idt.deleted_by_id, 
		idt.created_at, idt.updated_at, idt.deleted_at,
		
		idt.id as invoice_dp_dt_id,
		isg.id as item_sub_group_id,
		ig.id as item_group_id,
		isg.name as item_sub_group_name,
		ig.name as item_group_name,
		u.name as unit_name,
		p.name as item_name,
		p.code as item_code,
		
		cu.name as created_by_name,
		uu.name as updated_by_name
	FROM invoice_dp_dts idt
	LEFT JOIN products p ON idt.product_id = p.id
	LEFT JOIN item_units iu ON idt.item_unit_id = iu.id
	LEFT JOIN mix_values u ON iu.unit_id = u.id
	LEFT JOIN mix_values isg ON p.item_sub_group_id = isg.id
	LEFT JOIN mix_values ig ON isg.parent_id = ig.id
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

func (r *InvoiceDpRepository) GetInvoiceDpDts(ctx *fiber.Ctx, invoiceDpID uint, span opentracing.Span) ([]dtos.InvoiceDpDtListDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-GetInvoiceDpDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceDpDts := []dtos.InvoiceDpDtListDTO{}

	query := `
	SELECT 
		idt.id, idt.product_uuid, idt.invoice_dp_id, idt.item_unit_id, idt.vat_id, idt.pph23_id, 
		idt.ref_id, idt.ref_dt_id, idt.product_id, idt.product_id as item_id, idt.ref_type, idt.product_type, idt.remark, 
		idt.dp_percentage, idt.is_vat, idt.is_pph23, idt.qty, idt.price, idt.subtotal,
		idt.discount, idt.total_amount, idt.total_dp, idt.created_by_id, idt.updated_by_id, idt.deleted_by_id, 
		idt.created_at, idt.updated_at, idt.deleted_at,
		
		p.name as item_name, p.code as item_code,
		u.name as unit_name,
		v.name as vat_name,
		pph.name as pph23_name,
		
		cu.name as created_by_name,
		uu.name as updated_by_name,

		CASE WHEN idt.ref_type = 'so' THEN so.sales_order_no ELSE NULL END as ref_num
	FROM invoice_dp_dts idt
	LEFT JOIN products p ON idt.product_id = p.id
	LEFT JOIN item_units iu ON idt.item_unit_id = iu.id
	LEFT JOIN mix_values u ON iu.unit_id = u.id
	LEFT JOIN mix_values v ON idt.vat_id = v.id
	LEFT JOIN mix_values pph ON idt.pph23_id = pph.id
	LEFT JOIN sales_orders so ON idt.ref_id = so.id AND idt.ref_type = 'so'
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

	if err := tx.Save(&invoiceDpDts).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return tx, err
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

func (r *InvoiceDpRepository) RestoreInvoiceDp(ctx *fiber.Ctx, params *dtos.GetInvoiceDpParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-RestoreInvoiceDp", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var invoiceDp models.InvoiceDp
	if err := tx.Unscoped().Model(&invoiceDp).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
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
		"p.name", "p.code",
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

	var refDtIDs []uint
	if invoiceID, ok := filters["invoice_id"]; ok && invoiceID != "" {
		var invoiceDpDts []struct {
			RefDtID *uint `db:"ref_dt_id"`
		}

		refDtQuery := `
			SELECT ref_dt_id 
			FROM invoice_dp_dts 
			WHERE invoice_dp_id = $1 
			AND ref_type = 'so' 
			AND deleted_at IS NULL
		`

		if err := r.sqlDB.SelectContext(ctx.Context(), &invoiceDpDts, refDtQuery, invoiceID); err != nil {
			utils.LogErrors(childSpan, err)
			return nil, 0, err
		}

		for _, dt := range invoiceDpDts {
			if dt.RefDtID != nil {
				refDtIDs = append(refDtIDs, *dt.RefDtID)
			}
		}

		if len(refDtIDs) > 0 {
			refDtIDsStr := make([]string, len(refDtIDs))
			for i, id := range refDtIDs {
				refDtIDsStr[i] = fmt.Sprintf("%d", id)
			}
			condition += fmt.Sprintf(" AND (sodt.id IN (%s) OR (so.status NOT IN ('INVOICE', 'CANCELLED', 'FINISH')))", strings.Join(refDtIDsStr, ","))
		} else {
			condition += " AND so.status NOT IN ('INVOICE', 'CANCELLED', 'FINISH') AND sodt.total_dp IS NULL"
		}
	} else if filters["specific_ids"] != "" {
		condition += fmt.Sprintf(" AND (sodt.id IN (%s) OR (so.status NOT IN ('INVOICE', 'CANCELLED', 'FINISH')))", filters["specific_ids"])
	} else {
		condition += " AND so.status NOT IN ('INVOICE', 'CANCELLED', 'FINISH') AND sodt.total_dp IS NULL"
	}

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
		"vat_ids": {"so.vat_id", "sodt.vat_id"},
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

	// if date_type, start_date, end_date filled
	if filters["date_type"] != "" && filters["start_date"] != "" && filters["end_date"] != "" {
		filterDateTypeKey := map[string]string{
			"shipping_at": "so.shipping_at",
			"order_at":    "so.order_at",
			"due_at":      "so.due_at",
			"agree_at":    "so.agree_at",
		}

		dateTypeColumn := filterDateTypeKey[filters["date_type"]]
		condition += fmt.Sprintf(" AND (%s BETWEEN $%d AND $%d)", dateTypeColumn, i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
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
								LEFT JOIN invoice_dp_dts idt ON idt.ref_dt_id = sodt.id AND idt.ref_type = 'so' AND idt.deleted_at IS NULL

        LEFT JOIN users cu ON sodt.created_by_id = cu.id
        LEFT JOIN users uu ON sodt.updated_by_id = uu.id
                WHERE idt.id IS NULL AND sodt.deleted_at IS NULL
								` + condition + queryGlobal + `
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

// func (r *InvoiceDpRepository) UpdateSalesOrderStatus(tx *gorm.DB, salesOrderID uint, status string, span opentracing.Span) (*gorm.DB, error) {
// 	childSpan := opentracing.StartSpan("InvoiceDpRepository-UpdateSalesOrderStatus", opentracing.ChildOf(span.Context()))
// 	defer childSpan.Finish()

// 	result := tx.Model(&models.SalesOrder{}).Where("id = ?", salesOrderID).Updates(map[string]interface{}{
// 		"status": status,
// 	})
// 	if result.Error != nil {
// 		utils.LogErrors(childSpan, result.Error)
// 		return tx, result.Error
// 	}

// 	return tx, nil
// }

func (r *InvoiceDpRepository) UpdateSoDtsTotalDp(tx *gorm.DB, invoiceDpDts []models.InvoiceDpDt, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-UpdateSoDtsTotalDp", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var updates []struct {
		RefDtID uint    `gorm:"column:ref_dt_id"`
		TotalDp float64 `gorm:"column:total_dp"`
	}

	var soDtIDs []uint
	for _, invoiceDpDt := range invoiceDpDts {
		if invoiceDpDt.RefType != nil && *invoiceDpDt.RefType == "so" &&
			invoiceDpDt.RefDtID != nil && invoiceDpDt.TotalDp != nil {
			updates = append(updates, struct {
				RefDtID uint    `gorm:"column:ref_dt_id"`
				TotalDp float64 `gorm:"column:total_dp"`
			}{
				RefDtID: *invoiceDpDt.RefDtID,
				TotalDp: *invoiceDpDt.TotalDp,
			})
			soDtIDs = append(soDtIDs, *invoiceDpDt.RefDtID)
		}
	}

	if len(updates) > 0 {
		if err := tx.Exec("SELECT id FROM so_dts WHERE id IN ? FOR UPDATE", soDtIDs).Error; err != nil {
			utils.LogErrors(childSpan, err)
			return tx, err
		}

		query := `
        UPDATE so_dts AS s
        SET total_dp = t.total_dp
        FROM (
            SELECT unnest($1::integer[]) AS ref_dt_id, 
                   unnest($2::numeric[]) AS total_dp
        ) t
        WHERE s.id = t.ref_dt_id
        `

		refDtIDs := make([]uint, len(updates))
		totalDps := make([]float64, len(updates))

		for i, update := range updates {
			refDtIDs[i] = update.RefDtID
			totalDps[i] = update.TotalDp
		}

		if err := tx.Exec(query, pq.Array(refDtIDs), pq.Array(totalDps)).Error; err != nil {
			utils.LogErrors(childSpan, err)
			return tx, err
		}
	}

	return tx, nil
}

func (r *InvoiceDpRepository) ResetSoDtsTotalDp(tx *gorm.DB, deletedInvoiceDpDtIDs []uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-ResetSoDtsTotalDp", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(deletedInvoiceDpDtIDs) == 0 {
		return tx, nil
	}

	var deletedDts []struct {
		ID      uint     `gorm:"column:id"`
		RefDtID *uint    `gorm:"column:ref_dt_id"`
		RefType *string  `gorm:"column:ref_type"`
		TotalDp *float64 `gorm:"column:total_dp"`
	}

	if err := tx.Table("invoice_dp_dts").
		Select("id, ref_dt_id, ref_type, total_dp").
		Where("id IN ?", deletedInvoiceDpDtIDs).
		Find(&deletedDts).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return tx, err
	}

	var soDtIDs []uint
	for _, dt := range deletedDts {
		if dt.RefDtID != nil && dt.RefType != nil && *dt.RefType == "so" {
			soDtIDs = append(soDtIDs, *dt.RefDtID)
		}
	}

	if len(soDtIDs) > 0 {
		query := `
        UPDATE so_dts
        SET history_total_dp = total_dp,
            total_dp = NULL
        WHERE id IN (?)
        `

		if err := tx.Exec("SELECT id FROM so_dts WHERE id IN ? FOR UPDATE", soDtIDs).Error; err != nil {
			utils.LogErrors(childSpan, err)
			return tx, err
		}

		if err := tx.Exec(query, soDtIDs).Error; err != nil {
			utils.LogErrors(childSpan, err)
			return tx, err
		}
	}

	return tx, nil
}

func (r *InvoiceDpRepository) GetInvoiceDpCreatedThisMonth(ctx *fiber.Ctx, tx *gorm.DB, customerID uint, span opentracing.Span) (int, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-GetCustomerInvoiceDpCreatedThisMonth", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var count int
	query := `
    SELECT COUNT(*) 
    FROM invoice_dps 
    WHERE EXTRACT(MONTH FROM created_at) = EXTRACT(MONTH FROM CURRENT_DATE) 
    AND EXTRACT(YEAR FROM created_at) = EXTRACT(YEAR FROM CURRENT_DATE)
    `

	err := r.sqlDB.GetContext(ctx.Context(), &count, query)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return 0, err
	}

	return count, nil
}

func (r *InvoiceDpRepository) GetInvoiceDpWidgetData(ctx *fiber.Ctx, span opentracing.Span) (map[string]interface{}, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-GetInvoiceDpWidgetData", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	result := make(map[string]interface{})
	var wg sync.WaitGroup
	var errChan = make(chan error, 5)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var totalInvoiceDp int
		query := `SELECT COUNT(*) FROM invoice_dps WHERE deleted_at IS NULL`

		args := []interface{}{}
		i := 1

		if !isAdmin && branchID != nil {
			query += fmt.Sprintf(" AND branch_id = $%d", i)
			args = append(args, branchID)
			i++
		}

		err := r.sqlDB.GetContext(ctx.Context(), &totalInvoiceDp, query, args...)
		if err != nil {
			utils.LogErrors(childSpan, err)
			errChan <- err
			return
		}
		result["total_invoice_dp"] = totalInvoiceDp
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var totalInvoiceDpThisMonth int
		query := `
			SELECT COUNT(*) 
			FROM invoice_dps 
			WHERE deleted_at IS NULL
			AND EXTRACT(MONTH FROM created_at) = EXTRACT(MONTH FROM CURRENT_DATE) 
			AND EXTRACT(YEAR FROM created_at) = EXTRACT(YEAR FROM CURRENT_DATE)
		`

		args := []interface{}{}
		i := 1

		if !isAdmin && branchID != nil {
			query += fmt.Sprintf(" AND branch_id = $%d", i)
			args = append(args, branchID)
			i++
		}

		err := r.sqlDB.GetContext(ctx.Context(), &totalInvoiceDpThisMonth, query, args...)
		if err != nil {
			utils.LogErrors(childSpan, err)
			errChan <- err
			return
		}
		result["total_invoice_dp_this_month"] = totalInvoiceDpThisMonth
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var totalInvoiceDpAmount float64
		query := `
			SELECT COALESCE(SUM(grand_total), 0) 
			FROM invoice_dps 
			WHERE deleted_at IS NULL
		`

		args := []interface{}{}
		i := 1

		if !isAdmin && branchID != nil {
			query += fmt.Sprintf(" AND branch_id = $%d", i)
			args = append(args, branchID)
			i++
		}

		err := r.sqlDB.GetContext(ctx.Context(), &totalInvoiceDpAmount, query, args...)
		if err != nil {
			utils.LogErrors(childSpan, err)
			errChan <- err
			return
		}
		result["total_invoice_dp_amount"] = totalInvoiceDpAmount
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var totalInvoiceDpAmountThisMonth float64
		query := `
			SELECT COALESCE(SUM(grand_total), 0) 
			FROM invoice_dps 
			WHERE deleted_at IS NULL
			AND EXTRACT(MONTH FROM created_at) = EXTRACT(MONTH FROM CURRENT_DATE) 
			AND EXTRACT(YEAR FROM created_at) = EXTRACT(YEAR FROM CURRENT_DATE)
		`

		args := []interface{}{}
		i := 1

		if !isAdmin && branchID != nil {
			query += fmt.Sprintf(" AND branch_id = $%d", i)
			args = append(args, branchID)
			i++
		}

		err := r.sqlDB.GetContext(ctx.Context(), &totalInvoiceDpAmountThisMonth, query, args...)
		if err != nil {
			utils.LogErrors(childSpan, err)
			errChan <- err
			return
		}
		result["total_invoice_dp_amount_this_month"] = totalInvoiceDpAmountThisMonth
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		type StatusCount struct {
			Status string `db:"status"`
			Count  int    `db:"count"`
		}

		var statusCounts []StatusCount
		query := `
			SELECT status, COUNT(*) as count
			FROM invoice_dps
			WHERE deleted_at IS NULL
		`

		args := []interface{}{}
		i := 1

		if !isAdmin && branchID != nil {
			query += fmt.Sprintf(" AND branch_id = $%d", i)
			args = append(args, branchID)
			i++
		}

		query += " GROUP BY status"

		err := r.sqlDB.SelectContext(ctx.Context(), &statusCounts, query, args...)
		if err != nil {
			utils.LogErrors(childSpan, err)
			errChan <- err
			return
		}

		statusMap := make(map[string]int)
		for _, sc := range statusCounts {
			statusMap[sc.Status] = sc.Count
		}
		result["invoice_dp_status_counts"] = statusMap
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (r *InvoiceDpRepository) GetWidgetInvoiceDps(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.InvoiceDpStatusWidget, int, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-GetWidgetInvoiceDps", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	var widgets []dtos.InvoiceDpStatusWidget
	var total int

	filterDBColumnKey := []string{
		"idp.invoice_no", "idp.remark", "idp.status", "idp.title",
		"c.name",
		"idt.remark",
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
		condition += fmt.Sprintf(" AND idp.id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"status":          "idp.status",
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

	if value, ok := filters["status"]; ok && value != "" {
		condition += fmt.Sprintf(" AND idp.status = $%d", i)
		args = append(args, value)
		i++
	}

	for key, value := range filters {
		switch key {
		case "invoice_no", "remark", "idp.title":
			if value != "" {
				condition += fmt.Sprintf(" AND idp.%s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

	if !isAdmin && branchID != nil {
		condition += fmt.Sprintf(" AND idp.branch_id = $%d", i)
		args = append(args, branchID)
		i++
	}

	if isAdmin && filters["branch_id"] != "" {
		condition += fmt.Sprintf(" AND idp.branch_id = $%d", i)
		args = append(args, filters["branch_id"])
		i++
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
		"vat_ids": {"idp.vat_id", "idt.vat_id"},
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

	// if date_type, start_date, end_date filled
	if filters["date_type"] != "" && filters["start_date"] != "" && filters["end_date"] != "" {

		filterDateTypeKey := map[string]string{
			"invoice_date": "idp.invoice_date",
			"due_date":     "idp.due_date",
		}

		dateTypeColumn := "idp.invoice_date"
		for key := range filterDateTypeKey {
			if key == filters["date_type"] {
				dateTypeColumn = filterDateTypeKey[filters["date_type"]]
			}
		}

		condition += fmt.Sprintf(" AND (%s BETWEEN $%d AND $%d)", dateTypeColumn, i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
	}

	query := `
    WITH status_values AS (
        SELECT status, row_number() over () as status_order
        FROM (VALUES
            ('TOTAL', 0),
            ('PAID', 1),
            ('UNPAID', 2),
            ('CANCELLED', 3)
        ) AS s(status)
    ),
    filtered_invoices AS (
        SELECT DISTINCT
            idp.id,
            idp.status as idp_status,
            idp.total_qty,
            idp.grand_total
        FROM invoice_dps idp
        LEFT JOIN invoice_dp_dts idt ON idt.invoice_dp_id = idp.id
        LEFT JOIN customers c ON idp.customer_id = c.id
        WHERE idp.deleted_at IS NULL
        ` + condition + queryGlobal + `
    ),
    invoice_dp_stats AS (
        SELECT
            sv.status,
            sv.status_order,
            COUNT(DISTINCT fi.id) as order_count,
            COALESCE(SUM(fi.total_qty), 0) as total_qty,
            COALESCE(SUM(fi.grand_total), 0) as grand_total
        FROM status_values sv
        LEFT JOIN filtered_invoices fi ON
            (sv.status = fi.idp_status) OR
            (sv.status = 'TOTAL') OR
            (sv.status = 'UNPAID' AND fi.idp_status NOT IN ('PAID', 'CANCELLED'))
        GROUP BY sv.status, sv.status_order
    )
    SELECT
        sv.status,
        order_count,
        total_qty,
        grand_total
    FROM invoice_dp_stats sv
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

func (r *InvoiceDpRepository) ResetSoDtsTotalDpForCancelled(tx *gorm.DB, invoiceDpID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceDpRepository-ResetSoDtsTotalDpForCancelled", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var invoiceDpDts []struct {
		ID      uint     `gorm:"column:id"`
		RefDtID *uint    `gorm:"column:ref_dt_id"`
		RefType *string  `gorm:"column:ref_type"`
		TotalDp *float64 `gorm:"column:total_dp"`
	}

	if err := tx.Table("invoice_dp_dts").
		Select("id, ref_dt_id, ref_type, total_dp").
		Where("invoice_dp_id = ? AND deleted_at IS NULL", invoiceDpID).
		Find(&invoiceDpDts).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return tx, err
	}

	var soDtIDs []uint
	for _, dt := range invoiceDpDts {
		if dt.RefDtID != nil && dt.RefType != nil && *dt.RefType == "so" {
			soDtIDs = append(soDtIDs, *dt.RefDtID)
		}
	}

	if len(soDtIDs) > 0 {
		query := `
        UPDATE so_dts
        SET history_total_dp = total_dp,
            total_dp = NULL
        WHERE id IN (?)
        `

		if err := tx.Exec("SELECT id FROM so_dts WHERE id IN ? FOR UPDATE", soDtIDs).Error; err != nil {
			utils.LogErrors(childSpan, err)
			return tx, err
		}

		if err := tx.Exec(query, soDtIDs).Error; err != nil {
			utils.LogErrors(childSpan, err)
			return tx, err
		}
	}

	return tx, nil
}

func (r *InvoiceDpRepository) Commit(tx *gorm.DB) error {
	return tx.Commit().Error
}
