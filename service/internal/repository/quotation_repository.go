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

type QuotationRepository struct {
	db       *gorm.DB
	sqlDB    *sqlx.DB
	utilRepo *UtilRepository
	tracer   opentracing.Tracer
}

func NewQuotationRepository(db *gorm.DB, sqlDB *sqlx.DB, utilRepo *UtilRepository, tracer opentracing.Tracer) *QuotationRepository {
	return &QuotationRepository{
		db:       db,
		sqlDB:    sqlDB,
		tracer:   tracer,
		utilRepo: utilRepo,
	}
}

func (r *QuotationRepository) GetQuotations(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.QuotationListDTO, int, error) {
	childSpan := opentracing.StartSpan("QuotationRepository-GetQuotations", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	products := []dtos.QuotationListDTO{}

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
		"expired_at":    "q.expired_at",
		"due_at":        "q.due_at",
		"is_approve":    "q.is_approve",
	}

	for key, _ := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", value, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":   "q.customer_id",
		"order_type_ids": "q.order_type_id",
		"currency_ids":   "q.currency_id",
		"payment_ids":    "q.payment_id",
		"pph23_ids":      "q.pph23_id",
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

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (q.id)
					q.id, q.customer_id, q.order_type_id, q.currency_id, q.vat_id, q.payment_id, q.pph23_id, q.branch_id, q.quo_no, q.title, q.remark, q.status, q.is_approved, q.exchange_rate, q.pph23_perc, q.total_qty, q.subtotal, q.total_discount, q.total_pph23, q.total_vat, q.grand_total, q.created_by_id, q.updated_by_id, q.deleted_by_id, q.created_at, q.updated_at, q.deleted_at,
					TO_CHAR(q.due_at, 'YYYY-MM-DD') as due_at,
					TO_CHAR(q.expired_at, 'YYYY-MM-DD') as expired_at,
					q.vat_perc, q.disc_am, q.disc_perc, q.disc_perc_am, q.disc_final, q.disc_type,

					pi.id as product_id,
					it.id as item_id,
					qd.vat_id as quo_dt_vat_id,

					pi.name as product_name,
					it.name as item_name,
					cur.name as currency_name,
					vat.name as vat_name,
					pph.name as pph23_name,

					ot.name as order_type_name,
					c.name as customer_name,

					qd.remark as quo_dt_remark,
					qd.gen_code as quo_dt_gen_code,
					qdb.remark as quo_dt_bom_remark,
					qdb.gen_code as quo_dt_bom_gen_code,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM quotations q
				LEFT JOIN quo_dts qd ON qd.quotation_id = q.id
				LEFT JOIN products pi ON qd.item_id = pi.id
				LEFT JOIN item_units iu ON qd.item_unit_id = iu.id
				LEFT JOIN quo_dt_boms qdb ON qdb.quo_dt_id = qd.id
				LEFT JOIN products it ON qdb.item_id = it.id

				LEFT JOIN mix_values cur ON q.currency_id = cur.id
				LEFT JOIN mix_values vat ON q.vat_id = vat.id
				LEFT JOIN mix_values pph ON q.pph23_id = pph.id
				LEFT JOIN mix_values ot ON q.order_type_id = ot.id
				LEFT JOIN customers c ON q.customer_id = c.id

        LEFT JOIN users cu ON q.created_by_id = cu.id
        LEFT JOIN users uu ON q.updated_by_id = uu.id
				WHERE 1=1` + condition + queryGlobal + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery

	for key, value := range filters {
		switch key {
		case "quo_no", "title", "remark":
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

func (r *QuotationRepository) GetQuotationByID(ctx *fiber.Ctx, params *dtos.GetQuotationParams, tx *gorm.DB, span opentracing.Span) (*dtos.QuotationDetailDTO, error) {
	childSpan := opentracing.StartSpan("QuotationRepository-GetQuotationByID", opentracing.ChildOf(span.Context()))
	var quotation dtos.QuotationDetailDTO

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (q.id)
					q.id, q.customer_id, q.order_type_id, q.currency_id, q.vat_id, q.payment_id, q.pph23_id, q.branch_id, q.quo_no, q.title, q.remark, q.status, q.is_approved, q.exchange_rate, q.pph23_perc, q.total_qty, q.subtotal, q.total_discount, q.total_pph23, q.total_vat, q.grand_total, q.created_by_id, q.updated_by_id, q.deleted_by_id, q.created_at, q.updated_at, q.deleted_at,
					TO_CHAR(q.due_at, 'YYYY-MM-DD') as due_at,
					TO_CHAR(q.expired_at, 'YYYY-MM-DD') as expired_at,
					q.vat_perc, q.disc_am, q.disc_perc, q.disc_perc_am, q.disc_final, q.disc_type,
					-- q.quotation_id,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM quotations q
				LEFT JOIN quo_dts qd ON qd.quotation_id = q.id
				LEFT JOIN products pi ON qd.item_id = pi.id
				LEFT JOIN item_units iu ON qd.item_unit_id = iu.id
				LEFT JOIN quo_dt_boms qdb ON qdb.quo_dt_id = qd.id
				LEFT JOIN products it ON qdb.item_id = it.id

				LEFT JOIN mix_values cur ON q.currency_id = cur.id
				LEFT JOIN mix_values vat ON q.vat_id = vat.id
				LEFT JOIN mix_values pph ON q.pph23_id = pph.id

        LEFT JOIN users cu ON q.created_by_id = cu.id
        LEFT JOIN users uu ON q.updated_by_id = uu.id
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

	if err := r.sqlDB.Get(&quotation, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		childSpan.LogKV("query", query)
		return nil, err
	}

	return &quotation, nil
}

// BeginTransaction starts a new transaction
func (r *QuotationRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

// Rollback all changes in the transaction
func (r *QuotationRepository) Rollback() *gorm.DB {
	return r.db.Rollback()
}

func (r *QuotationRepository) CreateQuotation(tx *gorm.DB, quotation *models.Quotation, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("QuotationRepository-CreateQuotation", opentracing.ChildOf(span.Context()))
	if err := tx.Create(quotation).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *QuotationRepository) UpdateQuotation(tx *gorm.DB, quotation *models.Quotation, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuotationRepository-UpdateQuotation", opentracing.ChildOf(span.Context()))

	if err := tx.Where("id = ?", quotation.ID).Select("*").Omit(
		"created_at", "created_by_id", "branch_id",
	).Updates(quotation).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *QuotationRepository) DeleteQuotation(tx *gorm.DB, params *dtos.GetQuotationParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuotationRepository-DeleteQuotation", opentracing.ChildOf(span.Context()))

	if err := tx.Delete(&models.Quotation{}, params.ID).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil

}

func (s *QuotationRepository) RestoreQuotation(tx *gorm.DB, params *dtos.GetQuotationParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuotationRepository-RestoreQuotation", opentracing.ChildOf(span.Context()))

	var quotation models.Quotation
	if err := tx.Unscoped().Model(&quotation).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *QuotationRepository) CreateQuoDts(tx *gorm.DB, quoDts []models.QuoDt, quotationID uint, span opentracing.Span) (*gorm.DB, []models.QuoDt, error) {
	childSpan := opentracing.StartSpan("QuoDtRepository-CreateQuoDts", opentracing.ChildOf(span.Context()))

	tx = tx.Create(&quoDts)
	if tx.Error != nil {
		utils.LogErrors(childSpan, tx.Error)
		tx.Rollback()
		log.Fatalf("Failed to insert quoDts: %v", tx.Error)
	}

	return tx, quoDts, nil
}

// bulk/batch update quoDts
func (r *QuotationRepository) UpdateQuoDts(tx *gorm.DB, quoDts []models.QuoDt, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("QuoDtRepository-UpdateQuoDts", opentracing.ChildOf(span.Context()))

	data := make([]map[string]interface{}, 0)
	for _, quoDt := range quoDts {
		data = append(data, map[string]interface{}{
			"id":           quoDt.ID,
			"product_uuid": quoDt.ProductUuid,
			"quotation_id": quoDt.QuotationID,
			"ref_id":       quoDt.RefID,
			"vat_id":       quoDt.VatID,
			"pph23_id":     quoDt.Pph23ID,
			"item_unit_id": quoDt.ItemUnitID,
			"item_id":      quoDt.ItemID,
			"ref_type":     quoDt.RefType,
			"item_type":    quoDt.ItemType,
			// "ref_json":      quoDt.RefJSON,
			// "item_json":     quoDt.ItemJSON,
			"gen_code":       quoDt.GenCode,
			"remark":         quoDt.Remark,
			"is_lock_vat":    quoDt.IsLockVat,
			"vat_perc":       quoDt.VatPerc,
			"vat_perc_am":    quoDt.VatPercAm,
			"is_lock_pph23":  quoDt.IsLockPph23,
			"pph23_perc":     quoDt.Pph23Perc,
			"pph23_perc_am":  quoDt.Pph23PercAm,
			"qty_so":         quoDt.QtySO,
			"qty":            quoDt.Qty,
			"price_sell":     quoDt.PriceSell,
			"price_buy":      quoDt.PriceBuy,
			"subtotal_sell":  quoDt.SubtotalSell,
			"subtotal_buy":   quoDt.SubtotalBuy,
			"disc_am":        quoDt.DiscAm,
			"disc_perc":      quoDt.DiscPerc,
			"disc_perc_num":  quoDt.DiscPercNum,
			"disc_perc_am":   quoDt.DiscPercAm,
			"disc_final":     quoDt.DiscFinal,
			"disc_type":      quoDt.DiscType,
			"total_am":       quoDt.TotalAm,
			"head_disc_am":   quoDt.HeadDiscAm,
			"head_disc_perc": quoDt.HeadDiscPerc,
			"updated_by_id":  quoDt.UpdatedByID,
			"updated_at":     time.Now(),
		})
	}

	// if err := r.utilRepo.BulkUpdate(tx, "quoDts", "id", data, childSpan); err != nil {
	if err := r.utilRepo.Upsert(tx, "quo_dts", "id", data, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *QuotationRepository) DeleteQuoDtsWhereNotIn(ctx *fiber.Ctx, tx *gorm.DB, quotationID uint, quoDtIDs []uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("QuoDtRepository-DeleteQuoDtsWhereNotIn", opentracing.ChildOf(span.Context()))

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	now := time.Now()
	deletedAt := gorm.DeletedAt{Time: now, Valid: true}

	query := tx.Model(&models.QuoDt{}).Where("quotation_id = ? AND deleted_at IS NULL", quotationID)

	if len(quoDtIDs) > 0 {
		query = query.Where("id NOT IN (?)", quoDtIDs)
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

// func (r *QuotationRepository) GetQuoDtsByQuotationIDs(ctx *fiber.Ctx, quotationIDs uint, span opentracing.Span) ([]dtos.QuotationQuoDtListDTO, error) {
func (r *QuotationRepository) GetQuoDtsByQuotationIDs(ctx *fiber.Ctx, tx *gorm.DB, quotationIDs []uint, span opentracing.Span) ([]dtos.QuotationQuoDtListDTO, error) {
	childSpan := opentracing.StartSpan("QuotationRepository-GetQuoDtsByQuotationIDs", opentracing.ChildOf(span.Context()))

	quoDts := []dtos.QuotationQuoDtListDTO{}

	query := `SELECT qd.id, qd.quotation_id, qd.product_uuid,
		qd.item_unit_id, qd.vat_id, qd.ref_id, qd.item_id, qd.ref_type, qd.item_type, qd.gen_code, qd.remark, qd.vat_perc, qd.qty_so, qd.qty, qd.price_sell, qd.price_buy, qd.subtotal_sell, qd.subtotal_buy, qd.vat_perc, qd.vat_perc_am, qd.disc_am, qd.disc_perc, qd.disc_perc_num, qd.disc_perc_am, qd.disc_final, qd.disc_type, qd.total_am, qd.created_by_id, qd.updated_by_id, qd.deleted_by_id, qd.created_at, qd.updated_at, qd.deleted_at,
		qd.created_at, qd.updated_at, qd.deleted_at,

		qd.id as quo_dt_id,
		isg.id as item_sub_group_id,
		ig.id as item_group_id,
		isg.name as item_sub_group_name,
		ig.name as item_group_name,
		u.name as unit_name,
		pi.name as item_name,
		pi.code as item_code,

		cu.name as created_by_name,
		uu.name as updated_by_name

	FROM quo_dts qd
	LEFT JOIN quotations p ON qd.quotation_id = p.id
	LEFT JOIN products pi ON qd.item_id = pi.id
	LEFT JOIN item_units iu ON qd.item_unit_id = iu.id
	LEFT JOIN mix_values u ON iu.unit_id = u.id
	LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
	LEFT JOIN mix_values ig ON isg.parent_id = ig.id
	LEFT JOIN users cu ON qd.created_by_id = cu.id
	LEFT JOIN users uu ON qd.updated_by_id = uu.id
	WHERE qd.deleted_at IS NULL`

	var args []interface{}
	i := 1

	if len(quotationIDs) > 0 {
		query += " AND qd.quotation_id = ANY($1)"
		args = append(args, pq.Array(quotationIDs))
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
	// if err := sqlxTx.SelectContext(ctx.Context(), &quoDts, query, args...); err != nil {
	// 	utils.LogErrors(childSpan, err)
	// 	return nil, err
	// }

	if err := r.sqlDB.SelectContext(ctx.Context(), &quoDts, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return quoDts, nil
}

// func (r *QuotationRepository) GetQuoDtsByQuotationIDs(ctx *fiber.Ctx, quotationIDs uint, span opentracing.Span) ([]dtos.QuotationQuoDtListDTO, error) {
func (r *QuotationRepository) GetUpdatedQuoDtsByQuotationIDs(ctx *fiber.Ctx, tx *gorm.DB, quotationIDs []uint, span opentracing.Span) ([]dtos.QuotationQuoDtListUpdateDTO, error) {
	childSpan := opentracing.StartSpan("QuotationRepository-GetQuoDtsByQuotationIDs", opentracing.ChildOf(span.Context()))

	quoDts := []dtos.QuotationQuoDtListUpdateDTO{}

	query := `SELECT qd.id, qd.quotation_id, qd.product_uuid,
		qd.item_unit_id, qd.vat_id, qd.ref_id, qd.item_id, qd.ref_type, qd.item_type, qd.gen_code, qd.remark, qd.vat_perc, qd.qty_so, qd.qty, qd.price_sell, qd.price_buy, qd.subtotal_sell, qd.subtotal_buy, qd.vat_perc, qd.vat_perc_am, qd.disc_am, qd.disc_perc, qd.disc_perc_num, qd.disc_perc_am, qd.disc_final, qd.disc_type, qd.total_am, qd.created_by_id, qd.updated_by_id, qd.deleted_by_id, qd.created_at, qd.updated_at, qd.deleted_at,
		qd.created_at, qd.updated_at, qd.deleted_at,

		qd.id as quo_dt_id,
		isg.id as item_sub_group_id,
		ig.id as item_group_id,
		isg.name as item_sub_group_name,
		ig.name as item_group_name,
		u.name as unit_name,
		pi.name as item_name,
		pi.code as item_code,

		cu.name as created_by_name,
		uu.name as updated_by_name

	FROM quo_dts qd
	LEFT JOIN quotations p ON qd.quotation_id = p.id
	LEFT JOIN products pi ON qd.item_id = pi.id
	LEFT JOIN item_units iu ON qd.item_unit_id = iu.id
	LEFT JOIN mix_values u ON iu.unit_id = u.id
	LEFT JOIN mix_values isg ON pi.item_sub_group_id = isg.id
	LEFT JOIN mix_values ig ON isg.parent_id = ig.id
	LEFT JOIN users cu ON qd.created_by_id = cu.id
	LEFT JOIN users uu ON qd.updated_by_id = uu.id
	WHERE qd.deleted_at IS NULL`

	var args []interface{}
	i := 1

	if len(quotationIDs) > 0 {
		query += " AND qd.quotation_id = ANY($1)"
		args = append(args, pq.Array(quotationIDs))
		i++

	}

	// GORM Raw
	if err := tx.Raw(query, args...).Scan(&quoDts).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return quoDts, nil
}

func (r *QuotationRepository) CreateQuoDtBoms(tx *gorm.DB, quoDtBoms []map[string]interface{}, quotationID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("QuoDtRepository-CreateQuoDtBoms", opentracing.ChildOf(span.Context()))

	err := r.utilRepo.Upsert(tx, "quo_dt_boms", "id", quoDtBoms, span)
	// err := tx.Create(&quoDtBoms).Error
	if err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *QuotationRepository) DeleteQuoDtBomsWhereNotIn(ctx *fiber.Ctx, tx *gorm.DB, quotationID uint, quoDtBomIDs []uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuoDtRepository-DeleteQuoDtBomsWhereNotIn", opentracing.ChildOf(span.Context()))

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	now := time.Now()
	deletedAt := gorm.DeletedAt{Time: now, Valid: true}

	query := tx.Model(&models.QuoDtBom{}).Where("quotation_id = ? AND deleted_at IS NULL", quotationID)

	if len(quoDtBomIDs) > 0 {
		query = query.Where("id NOT IN (?)", quoDtBomIDs)
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

// bulk/batch update quoDts
func (r *QuotationRepository) UpdateQuoDtBoms(tx *gorm.DB, quoDtBoms []map[string]interface{}, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuoDtRepository-UpdateQuoDtBoms", opentracing.ChildOf(span.Context()))

	// data := make([]map[string]interface{}, 0)
	// for _, quoDtBom := range quoDtBoms {
	// 	itemJson := "{}"
	// 	genCode := ""

	// 	data = append(data, map[string]interface{}{
	// 		"id":            quoDtBom["id"],
	// 		"quotation_id":  quoDtBom["quotation_id"],
	// 		"quo_dt_id":     quoDtBom["quo_dt_id"],
	// 		"product_id":    quoDtBom["product_id"],
	// 		"item_id":       quoDtBom["item_id"],
	// 		"item_unit_id":  quoDtBom["item_unit_id"],
	// 		"item_json":     itemJson,
	// 		"gen_code":      genCode,
	// 		"remark":        quoDtBom["remark"],
	// 		"qty":           quoDtBom["qty"],
	// 		"price_sell":    quoDtBom["price_sell"],
	// 		"price_buy":     quoDtBom["price_buy"],
	// 		"subtotal_sell": quoDtBom["subtotal_sell"],
	// 		"subtotal_buy":  quoDtBom["subtotal_buy"],
	// 		"updated_by_id": quoDtBom["updated_by_id"],
	// 		"updated_at":    time.Now(),
	// 	})
	// }

	// log.Println("quoDtBoms", quoDtBoms)

	// if err := r.utilRepo.BulkUpdate(tx, "quoDtBoms", "id", data, childSpan); err != nil {
	if err := r.utilRepo.Upsert(tx, "quo_dt_boms", "id", quoDtBoms, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *QuotationRepository) DeleteQuoDtBomsByQuotationID(tx *gorm.DB, params *dtos.GetQuotationParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuoDtRepository-DeleteQuoDtBomsByQuotationID", opentracing.ChildOf(span.Context()))

	if err := tx.Where("quotation_id = ?", params.ID).Delete(&models.QuoDtBom{}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *QuotationRepository) DeleteQuoDtsByQuotationID(tx *gorm.DB, params *dtos.GetQuotationParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuoDtRepository-DeleteQuoDtsByQuotationID", opentracing.ChildOf(span.Context()))

	if err := tx.Where("quotation_id = ?", params.ID).Delete(&models.QuoDt{}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *QuotationRepository) GetQuoDtsBomByQuotations(ctx *fiber.Ctx, filters map[string]string, quotationIDs []uint, span opentracing.Span) ([]dtos.QuotationQuoDtBomListDTO, error) {
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
		"is_approve":    "q.is_approve",
	}

	for key, _ := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", value, i)
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

// Lock Quotation Header
func (r *QuotationRepository) LockQuotationHeader(ctx *fiber.Ctx, tx *gorm.DB, req dtos.UpdateQuotationRequest, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuotationRepository-LockQuotationHeader", opentracing.ChildOf(span.Context()))

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", []uint{req.ID}).Find(&models.Quotation{}).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)

		return err
	}

	return nil
}

func (r *QuotationRepository) LockQuoDts(ctx *fiber.Ctx, tx *gorm.DB, quoDtIDs []*uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuotationRepository-LockQuoDts", opentracing.ChildOf(span.Context()))

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", quoDtIDs).Find(&models.QuoDt{}).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)

		return err
	}

	return nil
}

func (r *QuotationRepository) LockQuoDtBoms(ctx *fiber.Ctx, tx *gorm.DB, quoDtBomIDs []*uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuotationRepository-LockQuoDtBoms", opentracing.ChildOf(span.Context()))

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", quoDtBomIDs).Find(&models.QuoDtBom{}).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)

		return err
	}

	return nil
}
