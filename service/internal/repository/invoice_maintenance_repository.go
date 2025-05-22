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
	"github.com/nibroos/s-erp-api/service/internal/config"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type InvoiceMaintenanceRepository struct {
	db       *gorm.DB
	sqlDB    *sqlx.DB
	utilRepo *UtilRepository
	rabbitmq *config.RabbitMQ
	tracer   opentracing.Tracer
}

func NewInvoiceMaintenanceRepository(db *gorm.DB, sqlDB *sqlx.DB, utilRepo *UtilRepository, rabbitmq *config.RabbitMQ, tracer opentracing.Tracer) *InvoiceMaintenanceRepository {
	return &InvoiceMaintenanceRepository{
		db:       db,
		sqlDB:    sqlDB,
		tracer:   tracer,
		rabbitmq: rabbitmq,
		utilRepo: utilRepo,
	}
}

func (r *InvoiceMaintenanceRepository) GetInvoiceMaintenances(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.InvoiceMaintenanceListDTO, int, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-GetInvoiceMaintenances", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	invoiceMaintenances := []dtos.InvoiceMaintenanceListDTO{}

	var total int

	filterDBColumnKey := []string{
		"im.invoice_no", "im.remark", "im.status", "im.title",
		"c.name",
		"imdt.remark",
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
		condition += fmt.Sprintf(" AND im.id IN (%s)", filters["ids"])
	}

	if filters["invoice_maintenance_ids"] != "" {
		condition += fmt.Sprintf(" AND im.id IN (%s)", filters["invoice_maintenance_ids"])
	}

	filterKey := map[string]string{
		"customer_id":     "im.customer_id",
		"currency_id":     "im.currency_id",
		"payment_term_id": "im.payment_term_id",
		"vat_id":          "im.vat_id",
		"pph23_id":        "im.pph23_id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":     "im.customer_id",
		"currency_ids":     "im.currency_id",
		"payment_term_ids": "im.payment_term_id",
		"pph23_ids":        "im.pph23_id",
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
		"vat_ids": []string{"im.vat_id", "imdt.vat_id"},
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
	log.Println("condition", condition)

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (im.id)
					im.id, im.customer_id, im.currency_id, im.payment_term_id, im.vat_id, im.pph23_id, im.branch_id, im.bank_id,
					im.invoice_no, im.remark, im.status, im.approved_status, im.approved_by_id, im.rev_no, im.title, bk.name as bank_name, bk.account_number as account_number, bk.account_name as account_name,
					im.exchange_rate, im.pph23_percentage, im.vat_percentage, im.total_qty, im.subtotal, im.total_discount, im.total_pph23, im.total_vat, im.grand_total, im.created_by_id, im.updated_by_id, im.deleted_by_id, im.created_at, im.updated_at, im.deleted_at,
					TO_CHAR(im.invoice_date, 'YYYY-MM-DD') as invoice_date, TO_CHAR(im.due_date, 'YYYY-MM-DD') as due_date,
					im.discount_amount, im.discount_percentage, im.discount_percentage_amount, im.discount_final, im.discount_type, im.total_amount_products, im.total_dp_products, im.total_balance_products, im.total_adjustment,

					(im.due_date - CURRENT_DATE) as days_remaining,
					CASE
							WHEN im.due_date < CURRENT_DATE THEN 'expired'
							ELSE 'expiring soon'
					END AS status_expired,

					c.name as customer_name,
					c.address as customer_address,
					c.email as customer_email,
					cur.name as currency_name,
					pt.name as payment_term_name,
					vat.name as vat_name,
					pph.name as pph23_name,
					b.name as branch_name,

					imdt.remark as invoice_maintenance_dt_remark,

					cu.name as created_by_name,
					uu.name as updated_by_name,
					au.name as approved_by_name

        FROM invoice_maintenances im
				LEFT JOIN invoice_maintenance_dts imdt ON imdt.invoice_maintenance_id = im.id
				LEFT JOIN customers c ON im.customer_id = c.id
				LEFT JOIN mix_values cur ON im.currency_id = cur.id
				LEFT JOIN mix_values pt ON im.payment_term_id = pt.id
				LEFT JOIN mix_values vat ON im.vat_id = vat.id
				LEFT JOIN mix_values pph ON im.pph23_id = pph.id
				LEFT JOIN branches b ON im.branch_id = b.id
				LEFT JOIN bank_informations bk ON im.bank_id = bk.id

        LEFT JOIN users cu ON im.created_by_id = cu.id
        LEFT JOIN users uu ON im.updated_by_id = uu.id
		LEFT JOIN users au ON im.approved_by_id = au.id
				WHERE 1=1` + condition + queryGlobal + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery

	log.Println("query", query)

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

	type tempInvoiceMaintenanceDTO struct {
		dtos.InvoiceMaintenanceListDTO
		// InvoiceMaintenanceDts interface{}
	}
	var tempResults []tempInvoiceMaintenanceDTO

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

		// err := r.sqlDB.SelectContext(ctx.Context(), &invoiceMaintenances, query, args...)
		// err := r.sqlDB.SelectContext(ctx.Context(), &tempResults, query, args...)
		err := r.db.Raw(query, args...).Scan(&tempResults).Error
		if err != nil {
			selectSpan.LogKV("query", query)
			utils.LogErrors(selectSpan, err)
			selectErr = err
		}
	}()

	wg.Wait()

	// Convert to final type
	invoiceMaintenances = make([]dtos.InvoiceMaintenanceListDTO, len(tempResults))
	for i, temp := range tempResults {
		invoiceMaintenances[i] = temp.InvoiceMaintenanceListDTO
		// invoiceMaintenances[i].InvoiceMaintenanceDts = []*dtos.InvoiceMaintenanceDtListDTO{}
	}

	if countErr != nil || selectErr != nil {
		defer childSpan.Finish()
	}

	if countErr != nil {
		return nil, 0, countErr
	}

	if selectErr != nil {
		return nil, 0, selectErr
	}

	return invoiceMaintenances, total, nil
}

func (r *InvoiceMaintenanceRepository) GetInvoiceMaintenanceDtsRawByIDs(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.InvoiceMaintenanceDtListNoBomDTO, int, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-GetInvoiceMaintenanceDtsRawByIDs", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	invoiceMaintenances := []dtos.InvoiceMaintenanceDtListNoBomDTO{}

	var total int

	filterDBColumnKey := []string{
		"im.invoice_no", "im.remark", "im.status", "im.title",
		"c.name",
		"imdt.remark",
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
		condition += fmt.Sprintf(" AND im.id IN (%s)", filters["ids"])
	}

	if filters["invoice_maintenance_ids"] != "" {
		condition += fmt.Sprintf(" AND imdt.invoice_maintenance_id IN (%s)", filters["invoice_maintenance_ids"])
	}

	filterKey := map[string]string{
		"customer_id":     "im.customer_id",
		"currency_id":     "im.currency_id",
		"payment_term_id": "im.payment_term_id",
		"vat_id":          "im.vat_id",
		"pph23_id":        "im.pph23_id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":     "im.customer_id",
		"currency_ids":     "im.currency_id",
		"payment_term_ids": "im.payment_term_id",
		"pph23_ids":        "im.pph23_id",
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
		"vat_ids": []string{"im.vat_id", "imdt.vat_id"},
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
        SELECT DISTINCT ON (imdt.id)
					imdt.id, imdt.product_uuid, imdt.invoice_maintenance_id, imdt.item_unit_id, imdt.vat_id, imdt.pph23_id, imdt.ref_id, imdt.ref_dt_id,
					imdt.product_id, imdt.ref_type, imdt.ref_json, imdt.product_type, imdt.product_json, imdt.remark, imdt.is_vat, imdt.is_pph23,
					imdt.qty, imdt.price, imdt.subtotal, imdt.discount, imdt.total_amount, imdt.total_dp, imdt.total_balance,

					p.name as item_name, 
					u.name as unit_name,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM invoice_maintenance_dts imdt
				LEFT JOIN invoice_maintenances im ON imdt.invoice_maintenance_id = im.id
				LEFT JOIN customers c ON im.customer_id = c.id
				LEFT JOIN mix_values cur ON im.currency_id = cur.id
				LEFT JOIN mix_values pt ON im.payment_term_id = pt.id
				LEFT JOIN mix_values vat ON im.vat_id = vat.id
				LEFT JOIN mix_values pph ON im.pph23_id = pph.id
				LEFT JOIN item_units iu ON imdt.item_unit_id = iu.id

				LEFT JOIN products p ON imdt.product_id = p.id
				LEFT JOIN mix_values u ON iu.unit_id = u.id
				LEFT JOIN branches b ON im.branch_id = b.id
				LEFT JOIN bank_informations bk ON im.bank_id = bk.id

        LEFT JOIN users cu ON im.created_by_id = cu.id
        LEFT JOIN users uu ON im.updated_by_id = uu.id
		LEFT JOIN users au ON im.approved_by_id = au.id
				WHERE 1=1 AND imdt.deleted_at IS NULL ` + condition + queryGlobal + `
    ) AS alias WHERE 1=1 `

	query := `SELECT *
		` + baseQuery

	for key, value := range filters {
		switch key {
		case "invoice_no", "remark", "status", "title":
			if value != "" {
				query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

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

	var selectErr error

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "id")
	orderDirection := utils.GetStringOrDefault(filters["order_direction"], "desc")
	query += fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)

	perPage := utils.GetIntOrDefault(filters["per_page"], 10)
	currentPage := utils.GetIntOrDefault(filters["page"], 1)

	if filters["is_csv"] != "1" {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
		args = append(args, perPage, (currentPage-1)*perPage)
	}

	log.Println("query", query)

	// err := r.sqlDB.SelectContext(ctx.Context(), &invoiceMaintenances, query, args...)
	err := r.db.Raw(query, args...).Scan(&invoiceMaintenances).Error
	if err != nil {
		childSpan.LogKV("query", query)
		utils.LogErrors(childSpan, err)
		selectErr = err
	}

	if selectErr != nil {
		defer childSpan.Finish()
		return nil, 0, selectErr
	}

	return invoiceMaintenances, total, nil
}

func (r *InvoiceMaintenanceRepository) GetInvoiceMaintenanceByID(ctx *fiber.Ctx, params *dtos.GetInvoiceMaintenanceParams, tx *gorm.DB, span opentracing.Span) (*dtos.InvoiceMaintenanceDetailDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-GetInvoiceMaintenanceByID", opentracing.ChildOf(span.Context()))
	var invoiceMaintenance dtos.InvoiceMaintenanceDetailDTO

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (im.id)
            im.id, im.customer_id, im.currency_id, im.payment_term_id, im.vat_id, im.pph23_id, im.branch_id, im.bank_id,
            im.invoice_no, im.remark, im.status, im.approved_status, im.rev_no, im.title,
            im.exchange_rate, im.pph23_percentage, im.vat_percentage, im.total_qty, im.subtotal, im.total_discount, im.total_pph23, im.total_vat, im.grand_total, im.created_by_id, im.updated_by_id, im.deleted_by_id, im.created_at, im.updated_at, im.deleted_at,
            TO_CHAR(im.invoice_date, 'YYYY-MM-DD') as invoice_date, TO_CHAR(im.due_date, 'YYYY-MM-DD') as due_date,
            im.discount_amount, im.discount_percentage, im.discount_percentage_amount, im.discount_final, im.discount_type, im.total_amount_products, im.total_dp_products, im.total_balance_products,

						br.company_profile_id,
						c.name as customer_name,
						c.code as customer_code,
						c.phone as phone,
						c.address as address,
						cur.name as currency_name,
						vat.name as vat_name,
						pph.name as pph23_name,
						py.account_name,
						py.name as bank_name,
						
            cu.name as created_by_name,
            uu.name as updated_by_name

        FROM invoice_maintenances im
        LEFT JOIN invoice_maintenance_dts imdt ON imdt.invoice_maintenance_id = im.id
				LEFT JOIN bank_informations py ON im.bank_id = py.id
				LEFT JOIN mix_values cur ON im.currency_id = cur.id
				LEFT JOIN mix_values pt ON im.payment_term_id = pt.id
				LEFT JOIN mix_values vat ON im.vat_id = vat.id
				LEFT JOIN mix_values pph ON im.pph23_id = pph.id
				LEFT JOIN customers c ON im.customer_id = c.id
				LEFT JOIN branches br ON im.branch_id = br.id
        LEFT JOIN users cu ON im.created_by_id = cu.id
        LEFT JOIN users uu ON im.updated_by_id = uu.id
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

	if err := r.sqlDB.Get(&invoiceMaintenance, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		childSpan.LogKV("query", query)
		return nil, err
	}

	return &invoiceMaintenance, nil
}

func (r *InvoiceMaintenanceRepository) GetInvoiceMaintenanceByNoBomID(ctx *fiber.Ctx, params *dtos.GetInvoiceMaintenanceParams, tx *gorm.DB, span opentracing.Span) (*dtos.InvoiceMaintenanceDetailNoBomDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-GetInvoiceMaintenanceByNoBomID", opentracing.ChildOf(span.Context()))
	var invoiceMaintenance dtos.InvoiceMaintenanceDetailNoBomDTO

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (im.id)
            im.id, im.customer_id, im.currency_id, im.payment_term_id, im.vat_id, im.pph23_id, im.branch_id, im.bank_id,
            im.invoice_no, im.remark, im.status, im.approved_status, im.rev_no, im.title,
            im.exchange_rate, im.pph23_percentage, im.vat_percentage, im.total_qty, im.subtotal, im.total_discount, im.total_pph23, im.total_vat, im.grand_total, im.created_by_id, im.updated_by_id, im.deleted_by_id, im.created_at, im.updated_at, im.deleted_at,
            TO_CHAR(im.invoice_date, 'YYYY-MM-DD') as invoice_date, TO_CHAR(im.due_date, 'YYYY-MM-DD') as due_date,
            im.discount_amount, im.discount_percentage, im.discount_percentage_amount, im.discount_final, im.discount_type, im.total_amount_products, im.total_dp_products, im.total_balance_products,

						br.company_profile_id,
						c.name as customer_name,
						c.code as customer_code,
						c.phone as phone,
						c.address as address,
						cur.name as currency_name,
						vat.name as vat_name,
						pph.name as pph23_name,
						py.account_name,
						py.name as bank_name,
						
            cu.name as created_by_name,
            uu.name as updated_by_name

        FROM invoice_maintenances im
        LEFT JOIN invoice_maintenance_dts imdt ON imdt.invoice_maintenance_id = im.id
				LEFT JOIN bank_informations py ON im.bank_id = py.id
				LEFT JOIN mix_values cur ON im.currency_id = cur.id
				LEFT JOIN mix_values pt ON im.payment_term_id = pt.id
				LEFT JOIN mix_values vat ON im.vat_id = vat.id
				LEFT JOIN mix_values pph ON im.pph23_id = pph.id
				LEFT JOIN customers c ON im.customer_id = c.id
				LEFT JOIN branches br ON im.branch_id = br.id
        LEFT JOIN users cu ON im.created_by_id = cu.id
        LEFT JOIN users uu ON im.updated_by_id = uu.id
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

	if err := r.sqlDB.Get(&invoiceMaintenance, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		childSpan.LogKV("query", query)
		return nil, err
	}

	return &invoiceMaintenance, nil
}

func (r *InvoiceMaintenanceRepository) GetUpdatedInvoiceMaintenanceDts(ctx *fiber.Ctx, invoiceMaintenanceID uint, span opentracing.Span) ([]dtos.InvoiceMaintenanceDtListUpdateDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-GetUpdatedInvoiceMaintenanceDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceMaintenanceDts := []dtos.InvoiceMaintenanceDtListUpdateDTO{}

	query := `
	SELECT InvoiceMaintenanceDetailDTO
		imdt.id, imdt.product_uuid, imdt.invoice_maintenance_id, imdt.item_unit_id, imdt.vat_id, imdt.pph23_id, 
				imdt.ref_id, imdt.ref_dt_id, imdt.product_id, imdt.ref_type, imdt.product_type, imdt.remark, 
		imdt.is_vat, imdt.is_pph23, imdt.qty, imdt.price, imdt.subtotal,
		imdt.discount, imdt.total_amount, imdt.total_dp, imdt.total_balance, imdt.created_by_id, imdt.updated_by_id, imdt.deleted_by_id, 
		imdt.created_at, imdt.updated_at, imdt.deleted_at,
		
		imdt.id as invoice_maintenance_dt_id,
		isg.id as item_sub_group_id,
		ig.id as item_group_id,
		isg.name as item_sub_group_name,
		ig.name as item_group_name,
		u.name as unit_name,
		p.name as item_name,
		p.code as item_code,
		
		cu.name as created_by_name,
		uu.name as updated_by_name
	FROM invoice_maintenance_dts imdt
	LEFT JOIN products p ON imdt.product_id = p.id
	LEFT JOIN item_units iu ON imdt.item_unit_id = iu.id
	LEFT JOIN mix_values u ON iu.unit_id = u.id
	LEFT JOIN mix_values isg ON p.item_sub_group_id = isg.id
	LEFT JOIN mix_values ig ON isg.parent_id = ig.id
	LEFT JOIN users cu ON imdt.created_by_id = cu.id
	LEFT JOIN users uu ON imdt.updated_by_id = uu.id
	WHERE imdt.invoice_maintenance_id = $1 AND imdt.deleted_at IS NULL
	ORDER BY imdt.id ASC
	`

	err := r.sqlDB.SelectContext(ctx.Context(), &invoiceMaintenanceDts, query, invoiceMaintenanceID)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return invoiceMaintenanceDts, nil
}

func (r *InvoiceMaintenanceRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *InvoiceMaintenanceRepository) Rollback() *gorm.DB {
	return r.db.Rollback()
}

func (r *InvoiceMaintenanceRepository) CreateInvoiceMaintenance(tx *gorm.DB, invoiceMaintenance *models.InvoiceMaintenance, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-CreateInvoiceMaintenance", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Create(&invoiceMaintenance)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceMaintenanceRepository) CreateInvoiceMaintenanceDts(tx *gorm.DB, invoiceMaintenanceDts []models.InvoiceMaintenanceDt, span opentracing.Span) (*gorm.DB, []models.InvoiceMaintenanceDt, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-CreateInvoiceMaintenanceDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Create(&invoiceMaintenanceDts)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, nil, result.Error
	}

	return tx, invoiceMaintenanceDts, nil
}

func (r *InvoiceMaintenanceRepository) GetInvoiceMaintenanceDts(ctx *fiber.Ctx, invoiceMaintenanceID uint, isDeleted *int, span opentracing.Span) ([]dtos.InvoiceMaintenanceDtListDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-GetInvoiceMaintenanceDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceMaintenanceDts := []dtos.InvoiceMaintenanceDtListDTO{}

	query := `
	SELECT 
		imdt.id, imdt.product_uuid, imdt.invoice_maintenance_id, imdt.item_unit_id, imdt.vat_id, imdt.pph23_id, 
		imdt.ref_id, imdt.ref_dt_id, imdt.product_id, imdt.ref_type, imdt.product_type, imdt.remark, 
		imdt.is_vat, imdt.is_pph23, imdt.qty, imdt.price, imdt.subtotal,
		imdt.discount, imdt.total_amount, imdt.total_dp, imdt.total_balance, imdt.created_by_id, imdt.updated_by_id, imdt.deleted_by_id, 
		imdt.created_at, imdt.updated_at, imdt.deleted_at,
		
		p.name as item_name, p.code as item_code,
		u.name as unit_name,
		v.name as vat_name,
		pph.name as pph23_name,
		
		cu.name as created_by_name,
		uu.name as updated_by_name,

		CASE WHEN imdt.ref_type = 'so' THEN so.sales_order_no ELSE NULL END as ref_num
	FROM invoice_maintenance_dts imdt
	LEFT JOIN products p ON imdt.product_id = p.id
	LEFT JOIN item_units iu ON imdt.item_unit_id = iu.id
	LEFT JOIN mix_values u ON iu.unit_id = u.id
	LEFT JOIN mix_values v ON imdt.vat_id = v.id
	LEFT JOIN mix_values pph ON imdt.pph23_id = pph.id
	LEFT JOIN sales_orders so ON imdt.ref_id = so.id AND imdt.ref_type = 'so'
	LEFT JOIN users cu ON imdt.created_by_id = cu.id
	LEFT JOIN users uu ON imdt.updated_by_id = uu.id
	WHERE imdt.invoice_maintenance_id = $1
	`

	if isDeleted != nil && *isDeleted == 1 {
		query += " AND imdt.deleted_at IS NOT NULL"
	} else {
		query += " AND imdt.deleted_at IS NULL"
	}

	query += " ORDER BY imdt.id ASC"

	err := r.sqlDB.SelectContext(ctx.Context(), &invoiceMaintenanceDts, query, invoiceMaintenanceID)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return invoiceMaintenanceDts, nil
}

func (r *InvoiceMaintenanceRepository) UpdateInvoiceMaintenance(tx *gorm.DB, invoiceMaintenance *models.InvoiceMaintenance, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-UpdateInvoiceMaintenance", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Model(&models.InvoiceMaintenance{}).Where("id = ?", invoiceMaintenance.ID).Updates(invoiceMaintenance)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceMaintenanceRepository) BulkCreateInvoiceMaintenanceDts(tx *gorm.DB, invoiceMaintenanceDts []models.InvoiceMaintenanceDt, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-BulkCreateInvoiceMaintenanceDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(invoiceMaintenanceDts) == 0 {
		return tx, nil
	}

	result := tx.Create(&invoiceMaintenanceDts)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceMaintenanceRepository) BulkUpdateInvoiceMaintenanceDts(tx *gorm.DB, invoiceMaintenanceDts []models.InvoiceMaintenanceDt, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-BulkUpdateInvoiceMaintenanceDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(invoiceMaintenanceDts) == 0 {
		return tx, nil
	}

	if err := tx.Save(&invoiceMaintenanceDts).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return tx, err
	}

	return tx, nil
}

func (r *InvoiceMaintenanceRepository) DeleteInvoiceMaintenanceDtsByIDs(tx *gorm.DB, invoiceMaintenanceDtIDs []uint, userID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-DeleteInvoiceMaintenanceDtsByIDs", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(invoiceMaintenanceDtIDs) == 0 {
		return tx, nil
	}

	result := tx.Model(&models.InvoiceMaintenanceDt{}).Where("id IN ?", invoiceMaintenanceDtIDs).Updates(map[string]interface{}{
		"deleted_by_id": userID,
		"deleted_at":    time.Now(),
	})
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceMaintenanceRepository) DeleteInvoiceMaintenance(tx *gorm.DB, invoiceMaintenanceID uint, userID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-DeleteInvoiceMaintenance", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Model(&models.InvoiceMaintenance{}).Where("id = ?", invoiceMaintenanceID).Updates(map[string]interface{}{
		"deleted_by_id": userID,
		"deleted_at":    time.Now(),
	})
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceMaintenanceRepository) RestoreInvoiceMaintenance(ctx *fiber.Ctx, params *dtos.GetInvoiceMaintenanceParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-RestoreInvoiceMaintenance", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var invoiceMaintenance models.InvoiceMaintenance
	if err := tx.Unscoped().Model(&invoiceMaintenance).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *InvoiceMaintenanceRepository) GetRefSalesOrderForInvoiceMaintenance(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefSalesOrderForInvoiceMaintenanceListDTO, int, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-GetRefSalesOrderForInvoiceMaintenance", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	soDts := []dtos.RefSalesOrderForInvoiceMaintenanceListDTO{}

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

	var refDtIDs []uint
	if invoiceID, ok := filters["invoice_id"]; ok && invoiceID != "" {
		var invoiceMaintenanceDts []struct {
			RefDtID *uint `db:"ref_dt_id"`
		}

		refDtQuery := `
			SELECT ref_dt_id 
			FROM invoice_maintenance_dts 
			WHERE invoice_maintenance_id = $1 
			AND ref_type = 'so' 
			AND deleted_at IS NULL
		`

		if err := r.sqlDB.SelectContext(ctx.Context(), &invoiceMaintenanceDts, refDtQuery, invoiceID); err != nil {
			utils.LogErrors(childSpan, err)
			return nil, 0, err
		}

		for _, dt := range invoiceMaintenanceDts {
			if dt.RefDtID != nil {
				refDtIDs = append(refDtIDs, *dt.RefDtID)
			}
		}

		if len(refDtIDs) > 0 {
			refDtIDsStr := make([]string, len(refDtIDs))
			for i, id := range refDtIDs {
				refDtIDsStr[i] = fmt.Sprintf("%d", id)
			}
			condition += fmt.Sprintf(" AND (sodt.id IN (%s) OR (so.status NOT IN ('CANCELED', 'FINISH') AND CURRENT_DATE BETWEEN so.agree_at AND so.due_at))", strings.Join(refDtIDsStr, ","))
		} else {
			condition += " AND so.status NOT IN ('CANCELED', 'FINISH') AND CURRENT_DATE BETWEEN so.agree_at AND so.due_at"
		}
	} else if filters["specific_ids"] != "" {
		condition += fmt.Sprintf(" AND (sodt.id IN (%s) OR (so.status NOT IN ('CANCELED', 'FINISH') AND CURRENT_DATE BETWEEN so.agree_at AND so.due_at))", filters["specific_ids"])
	} else {
		condition += " AND so.status NOT IN ('CANCELED', 'FINISH') AND CURRENT_DATE BETWEEN so.agree_at AND so.due_at"
	}

	if filters["ids"] != "" {
		condition += fmt.Sprintf(" AND id IN (%s)", filters["ids"])
	}

	if value, ok := filters["sales_order_no"]; ok && value != "" {
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

				so.customer_id, so.order_type_id, so.currency_id, so.vat_id as head_vat_id, so.payment_id,
				so.pph23_id as head_pph23_id, so.vat_perc as head_vat_perc, 
				so.pph23_perc as head_pph23_perc, so.disc_am as head_disc_am, 
				so.disc_perc as head_disc_perc, so.markup_perc as head_markup_perc, 
				so.remark as head_remark, so.exchange_rate, so.sales_order_no, so.po_buyer_no, 
				so.order_at as order_date, so.shipping_at as shipping_date,
				TO_CHAR(so.agree_at, 'YYYY-MM-DD') as agree_at,
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
				WHERE 1=1 AND so.order_type_id = 130
                AND ot.name = 'Maintenance'` + condition + queryGlobal + `
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

func (r *InvoiceMaintenanceRepository) GetSoDtBoms(ctx *fiber.Ctx, soDtIDs []uint, span opentracing.Span) ([]dtos.SalesOrderSoDtBomListDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-GetSoDtBoms", opentracing.ChildOf(span.Context()))
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

func (r *InvoiceMaintenanceRepository) BulkUpdateSalesOrdersStatus(tx *gorm.DB, salesOrderIDs []uint, status string, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-BulkUpdateSalesOrdersStatus", opentracing.ChildOf(span.Context()))
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

func (r *InvoiceMaintenanceRepository) RestoreSalesOrdersStatus(tx *gorm.DB, salesOrderIDs []uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-RestoreSalesOrdersStatus", opentracing.ChildOf(span.Context()))
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

func (r *InvoiceMaintenanceRepository) LockSalesOrders(tx *gorm.DB, salesOrderIDs []uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-LockSalesOrders", opentracing.ChildOf(span.Context()))
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

func (r *InvoiceMaintenanceRepository) LockInvoiceMaintenance(tx *gorm.DB, invoiceMaintenanceID uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-LockInvoiceMaintenance", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	lockQuery := "SELECT id FROM invoice_maintenances WHERE id = ? FOR UPDATE"
	if err := tx.Exec(lockQuery, invoiceMaintenanceID).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *InvoiceMaintenanceRepository) GetInvoiceMaintenanceForUpdate(tx *gorm.DB, invoiceMaintenanceID uint, span opentracing.Span) (*models.InvoiceMaintenance, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-GetInvoiceMaintenanceForUpdate", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var invoiceMaintenance models.InvoiceMaintenance
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", invoiceMaintenanceID).First(&invoiceMaintenance).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return &invoiceMaintenance, nil
}

func (r *InvoiceMaintenanceRepository) GetInvoiceMaintenanceCreatedThisMonth(ctx *fiber.Ctx, tx *gorm.DB, customerID uint, span opentracing.Span) (int, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-GetCustomerInvoiceMaintenanceCreatedThisMonth", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var count int
	query := `
    SELECT COUNT(*) 
    FROM invoice_maintenances 
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

func (r *InvoiceMaintenanceRepository) GetSalesOrdersByIDs(ctx *fiber.Ctx, salesOrderIDs []uint, span opentracing.Span) ([]models.SalesOrder, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-GetSalesOrdersByIDs", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(salesOrderIDs) == 0 {
		return []models.SalesOrder{}, nil
	}

	var salesOrders []models.SalesOrder
	if err := r.db.Where("id IN ?", salesOrderIDs).Find(&salesOrders).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return salesOrders, nil
}

func (r *InvoiceMaintenanceRepository) GetInvoiceMaintenanceDtsByIDs(ctx *fiber.Ctx, invoiceMaintenanceDtIDs []uint, span opentracing.Span) ([]models.InvoiceMaintenanceDt, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-GetInvoiceMaintenanceDtsByIDs", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(invoiceMaintenanceDtIDs) == 0 {
		return []models.InvoiceMaintenanceDt{}, nil
	}

	var invoiceMaintenanceDts []models.InvoiceMaintenanceDt
	if err := r.db.Where("id IN ?", invoiceMaintenanceDtIDs).Find(&invoiceMaintenanceDts).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return invoiceMaintenanceDts, nil
}

func (r *InvoiceMaintenanceRepository) GetSalesOrderIDsFromInvoiceMaintenanceDts(ctx *fiber.Ctx, invoiceMaintenanceID uint, span opentracing.Span) ([]uint, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-GetSalesOrderIDsFromInvoiceMaintenanceDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var salesOrderIDs []uint
	query := `
    SELECT DISTINCT ref_id 
    FROM invoice_maintenance_dts 
    WHERE invoice_maintenance_id = ? 
    AND ref_type = 'so' 
    AND deleted_at IS NULL
    `

	if err := r.db.Raw(query, invoiceMaintenanceID).Scan(&salesOrderIDs).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return salesOrderIDs, nil
}

func (r *InvoiceMaintenanceRepository) GetSalesOrderIDsFromInvoiceMaintenanceDtsByIDs(ctx *fiber.Ctx, invoiceMaintenanceDtIDs []uint, span opentracing.Span) ([]uint, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-GetSalesOrderIDsFromInvoiceMaintenanceDtsByIDs", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(invoiceMaintenanceDtIDs) == 0 {
		return []uint{}, nil
	}

	var salesOrderIDs []uint
	query := `
    SELECT DISTINCT ref_id 
    FROM invoice_maintenance_dts 
    WHERE id IN ? 
    AND ref_type = 'so' 
    AND deleted_at IS NULL
    `

	if err := r.db.Raw(query, invoiceMaintenanceDtIDs).Scan(&salesOrderIDs).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return salesOrderIDs, nil
}

func (r *InvoiceMaintenanceRepository) UpdateSalesOrdersStatusToInvoice(tx *gorm.DB, salesOrderIDs []uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-UpdateSalesOrdersStatusToInvoice", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(salesOrderIDs) == 0 {
		return tx, nil
	}

	return r.BulkUpdateSalesOrdersStatus(tx, salesOrderIDs, "INVOICE", childSpan)
}

func (r *InvoiceMaintenanceRepository) RestoreSalesOrdersStatusFromInvoice(tx *gorm.DB, salesOrderIDs []uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-RestoreSalesOrdersStatusFromInvoice", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(salesOrderIDs) == 0 {
		return tx, nil
	}

	return r.RestoreSalesOrdersStatus(tx, salesOrderIDs, childSpan)
}

func (r *InvoiceMaintenanceRepository) UpdateSalesOrdersStatusForInvoiceMaintenance(tx *gorm.DB, req dtos.UpdateSalesOrderStatusForInvoiceMaintenanceRequest, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-UpdateSalesOrdersStatusForInvoiceMaintenance", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	query := `
        UPDATE sales_orders 
        SET 
            status = ?,
            history_status = status
        WHERE id = ?
    `

	result := tx.Exec(query, req.Status, req.ID)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceMaintenanceRepository) BulkApproveInvoiceMaintenances(tx *gorm.DB, ids []uint, userID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-BulkApproveInvoiceMaintenances", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(ids) == 0 {
		return tx, nil
	}

	lockQuery := "SELECT id FROM invoice_maintenances WHERE id IN ? FOR UPDATE"
	if err := tx.Exec(lockQuery, ids).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return tx, err
	}

	result := tx.Model(&models.InvoiceMaintenance{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"approved_status": "APPROVED",
			"approved_by_id":  userID,
			"updated_by_id":   userID,
			"updated_at":      time.Now(),
		})

	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceMaintenanceRepository) BulkCancelApproveInvoiceMaintenances(tx *gorm.DB, ids []uint, userID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-BulkCancelApproveInvoiceMaintenances", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(ids) == 0 {
		return tx, nil
	}

	lockQuery := "SELECT id FROM invoice_maintenances WHERE id IN ? FOR UPDATE"
	if err := tx.Exec(lockQuery, ids).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return tx, err
	}

	result := tx.Model(&models.InvoiceMaintenance{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"approved_status": "PENDING",
			"approved_by_id":  nil,
			"updated_by_id":   userID,
			"updated_at":      time.Now(),
		})

	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceMaintenanceRepository) UpdateSoDtInvoiceStatus(tx *gorm.DB, soDtID uint, status interface{}, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-UpdateSoDtInvoiceStatus", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var result *gorm.DB
	if status == nil {
		result = tx.Exec("UPDATE so_dts SET invoice_status = NULL WHERE id = ?", soDtID)
	} else {
		result = tx.Exec("UPDATE so_dts SET invoice_status = ? WHERE id = ?", status, soDtID)
	}

	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceMaintenanceRepository) BulkUpdateSoDtInvoiceStatus(tx *gorm.DB, soDtIDs []uint, status interface{}, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-BulkUpdateSoDtInvoiceStatus", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(soDtIDs) == 0 {
		return tx, nil
	}

	var result *gorm.DB
	if status == nil {
		result = tx.Exec("UPDATE so_dts SET invoice_status = NULL WHERE id IN ?", soDtIDs)
	} else {
		result = tx.Exec("UPDATE so_dts SET invoice_status = ? WHERE id IN ?", status, soDtIDs)
	}

	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return tx, result.Error
	}

	return tx, nil
}

func (r *InvoiceMaintenanceRepository) CheckAndUpdateSalesOrderStatus(tx *gorm.DB, salesOrderID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-CheckAndUpdateSalesOrderStatus", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var totalItems int64
	var invoicedItems int64

	if err := tx.Model(&models.SoDt{}).Where("sales_order_id = ? AND deleted_at IS NULL", salesOrderID).Count(&totalItems).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return tx, err
	}

	if err := tx.Model(&models.SoDt{}).Where("sales_order_id = ? AND invoice_status = 'INVOICE' AND deleted_at IS NULL", salesOrderID).Count(&invoicedItems).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return tx, err
	}

	type SalesOrderStatus struct {
		Status        string `gorm:"column:status"`
		HistoryStatus string `gorm:"column:history_status"`
	}

	var soStatus SalesOrderStatus
	if err := tx.Model(&models.SalesOrder{}).Where("id = ?", salesOrderID).Select("status, history_status").Scan(&soStatus).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return tx, err
	}

	if totalItems > 0 && invoicedItems == totalItems {
		result := tx.Exec("UPDATE sales_orders SET status = 'INVOICE', history_status = CASE WHEN status != 'INVOICE' THEN status ELSE history_status END WHERE id = ?", salesOrderID)
		if result.Error != nil {
			utils.LogErrors(childSpan, result.Error)
			return tx, result.Error
		}
	} else if invoicedItems < totalItems && soStatus.Status == "INVOICE" {
		result := tx.Exec("UPDATE sales_orders SET status = history_status, history_status = status WHERE id = ? AND status = 'INVOICE'", salesOrderID)
		if result.Error != nil {
			utils.LogErrors(childSpan, result.Error)
			return tx, result.Error
		}
	}

	return tx, nil
}

func (r *InvoiceMaintenanceRepository) GetSoDtIDsFromInvoiceMaintenanceDts(ctx *fiber.Ctx, invoiceMaintenanceID uint, span opentracing.Span) ([]uint, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-GetSoDtIDsFromInvoiceMaintenanceDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var soDtIDs []uint
	query := `
	SELECT ref_dt_id 
	FROM invoice_maintenance_dts 
	WHERE invoice_maintenance_id = ? 
	AND ref_type = 'so' 
	AND ref_dt_id IS NOT NULL
	AND deleted_at IS NULL
	`

	if err := r.db.Raw(query, invoiceMaintenanceID).Scan(&soDtIDs).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return soDtIDs, nil
}

func (r *InvoiceMaintenanceRepository) GetSoDtInvoiceStatus(ctx *fiber.Ctx, soDtIDs []uint, span opentracing.Span) (map[uint]string, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-GetSoDtInvoiceStatus", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(soDtIDs) == 0 {
		return make(map[uint]string), nil
	}

	type SoDtStatus struct {
		ID            uint   `db:"id"`
		InvoiceStatus string `db:"invoice_status"`
	}

	var statuses []SoDtStatus
	query := `
    SELECT id, invoice_status 
    FROM so_dts 
    WHERE id IN (?) AND deleted_at IS NULL
    `

	query = r.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Raw(query, soDtIDs)
	})

	err := r.sqlDB.SelectContext(ctx.Context(), &statuses, query)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	statusMap := make(map[uint]string)
	for _, status := range statuses {
		if status.InvoiceStatus != "" {
			statusMap[status.ID] = status.InvoiceStatus
		}
	}

	return statusMap, nil
}

func (r *InvoiceMaintenanceRepository) GetWidgetInvoiceMaintenances(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.InvoiceMaintenanceStatusWidget, int, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-GetWidgetInvoiceMaintenances", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	var widgets []dtos.InvoiceMaintenanceStatusWidget
	var total int

	filterDBColumnKey := []string{
		"im.invoice_no", "im.remark", "im.status", "im.title",
		"c.name",
		"imdt.remark",
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
		condition += fmt.Sprintf(" AND im.id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"status":          "im.status",
		"customer_id":     "im.customer_id",
		"currency_id":     "im.currency_id",
		"payment_term_id": "im.payment_term_id",
		"vat_id":          "im.vat_id",
		"pph23_id":        "im.pph23_id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	for key, value := range filters {
		switch key {
		case "invoice_no", "remark", "title":
			if value != "" {
				condition += fmt.Sprintf(" AND im.%s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

	if !isAdmin && branchID != nil {
		condition += fmt.Sprintf(" AND im.branch_id = $%d", i)
		args = append(args, branchID)
		i++
	}

	if isAdmin && filters["branch_id"] != "" {
		condition += fmt.Sprintf(" AND im.branch_id = $%d", i)
		args = append(args, filters["branch_id"])
		i++
	}

	filterIDsKey := map[string]string{
		"customer_ids":     "im.customer_id",
		"currency_ids":     "im.currency_id",
		"payment_term_ids": "im.payment_term_id",
		"pph23_ids":        "im.pph23_id",
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
		"vat_ids": []string{"im.vat_id", "imdt.vat_id"},
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

	if filters["start_date"] != "" && filters["end_date"] != "" {
		condition += fmt.Sprintf(" AND (im.invoice_date BETWEEN $%d AND $%d)", i, i+1)
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
            ('CANCELED', 3)
        ) AS s(status)
    ),
    filtered_invoices AS (
        SELECT DISTINCT
            im.id,
            im.status as im_status,
            im.total_qty,
            im.grand_total
        FROM invoice_maintenances im
        LEFT JOIN invoice_maintenance_dts imdt ON imdt.invoice_maintenance_id = im.id
        LEFT JOIN customers c ON im.customer_id = c.id
        WHERE im.deleted_at IS NULL
        ` + condition + queryGlobal + `
    ),
    invoice_maintenance_stats AS (
        SELECT
            sv.status,
            sv.status_order,
            COUNT(DISTINCT fi.id) as order_count,
            COALESCE(SUM(fi.total_qty), 0) as total_qty,
            COALESCE(SUM(fi.grand_total), 0) as grand_total
        FROM status_values sv
        LEFT JOIN filtered_invoices fi ON
            (sv.status = fi.im_status) OR
            (sv.status = 'TOTAL') OR
            (sv.status = 'UNPAID' AND fi.im_status NOT IN ('PAID', 'CANCELED'))
        GROUP BY sv.status, sv.status_order
    )
    SELECT
        sv.status,
        order_count,
        total_qty,
        grand_total
    FROM invoice_maintenance_stats sv
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

func (r *InvoiceMaintenanceRepository) GetInvoiceMaintenanceModelByID(ctx *fiber.Ctx, id uint, span opentracing.Span) (*models.InvoiceMaintenance, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-GetInvoiceMaintenanceModelByID", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var invoiceMaintenance models.InvoiceMaintenance
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&invoiceMaintenance).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return &invoiceMaintenance, nil
}

func (r *InvoiceMaintenanceRepository) GetInvoiceMaintenanceDtModels(ctx *fiber.Ctx, invoiceMaintenanceID uint, span opentracing.Span) ([]models.InvoiceMaintenanceDt, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-GetInvoiceMaintenanceDtModels", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var invoiceMaintenanceDts []models.InvoiceMaintenanceDt
	if err := r.db.Where("invoice_maintenance_id = ? AND deleted_at IS NULL", invoiceMaintenanceID).Find(&invoiceMaintenanceDts).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return invoiceMaintenanceDts, nil
}

func (r *InvoiceMaintenanceRepository) LockInvoiceNumberingProcess(tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-LockInvoiceNumberingProcess", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	lockQuery := "SELECT pg_advisory_xact_lock(1001)"
	if err := tx.Exec(lockQuery).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *InvoiceMaintenanceRepository) RepeatInvoiceMaintenances(ctx *fiber.Ctx, req dtos.RepeatInvoiceMaintenanceRequest, userID uint, span opentracing.Span) ([]dtos.RepeatInvoiceMaintenanceResult, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceRepository-RepeatInvoiceMaintenances", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	claims, _ := auth.GetAuthUser(ctx)

	var branchID uint
	if bid, ok := claims["bid"]; ok {
		switch v := bid.(type) {
		case uint:
			branchID = v
		case float64:
			branchID = uint(v)
		case int:
			branchID = uint(v)
		case int64:
			branchID = uint(v)
		default:
			utils.LogErrors(childSpan, fmt.Errorf("invalid branch ID type: %T", bid))
			return nil, fmt.Errorf("invalid branch ID type")
		}
	} else {
		utils.LogErrors(childSpan, fmt.Errorf("branch ID not found in claims"))
		return nil, fmt.Errorf("branch ID not found in claims")
	}

	results := make([]dtos.RepeatInvoiceMaintenanceResult, 0, len(req.Invoices))

	for _, item := range req.Invoices {
		tx := r.db.Begin()
		if tx.Error != nil {
			utils.LogErrors(childSpan, tx.Error)
			return nil, tx.Error
		}

		if err := r.LockInvoiceNumberingProcess(tx, childSpan); err != nil {
			tx.Rollback()
			utils.LogErrors(childSpan, err)
			return nil, err
		}

		originalInvoice, err := r.GetInvoiceMaintenanceModelByID(ctx, item.ID, childSpan)
		if err != nil {
			tx.Rollback()
			utils.LogErrors(childSpan, err)
			return nil, err
		}

		count, err := r.GetInvoiceMaintenanceCreatedThisMonth(ctx, tx, *originalInvoice.CustomerID, childSpan)
		if err != nil {
			tx.Rollback()
			utils.LogErrors(childSpan, err)
			return nil, err
		}

		originalDts, err := r.GetInvoiceMaintenanceDtModels(ctx, item.ID, childSpan)
		if err != nil {
			tx.Rollback()
			utils.LogErrors(childSpan, err)
			return nil, err
		}

		newInvoice, err := utils.MapRepeatInvoiceMaintenance(ctx, originalInvoice, item, userID, branchID, count+1, childSpan)
		if err != nil {
			tx.Rollback()
			utils.LogErrors(childSpan, err)
			return nil, err
		}

		tx, err = r.CreateInvoiceMaintenance(tx, &newInvoice, childSpan)
		if err != nil {
			tx.Rollback()
			utils.LogErrors(childSpan, err)
			return nil, err
		}

		newDts, err := utils.MapRepeatInvoiceMaintenanceDts(ctx, originalDts, newInvoice.ID, userID, childSpan)
		if err != nil {
			tx.Rollback()
			utils.LogErrors(childSpan, err)
			return nil, err
		}

		if _, _, err := r.CreateInvoiceMaintenanceDts(tx, newDts, childSpan); err != nil {
			tx.Rollback()
			utils.LogErrors(childSpan, err)
			return nil, err
		}

		if err := tx.Commit().Error; err != nil {
			utils.LogErrors(childSpan, err)
			return nil, err
		}

		results = append(results, dtos.RepeatInvoiceMaintenanceResult{
			OriginalID:  item.ID,
			DuplicateID: newInvoice.ID,
		})
	}

	return results, nil
}

func (r *InvoiceMaintenanceRepository) Commit(tx *gorm.DB) error {
	return tx.Commit().Error
}
