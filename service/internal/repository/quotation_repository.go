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
					q.id, q.customer_id, q.order_type_id, q.currency_id, q.vat_id, q.payment_id, q.pph23_id, q.branch_id, q.quo_no, q.title, q.remark, q.status, q.is_approved, q.exchange_rate, q.vat_perc, q.pph23_perc, q.total_qty, q.subtotal, q.total_discount, q.total_pph23, q.total_vat, q.grand_total, q.due_at, q.expired_at, q.created_by_id, q.updated_by_id, q.deleted_by_id, q.created_at, q.updated_at, q.deleted_at,

					pi.id as product_id,
					it.id as item_id,
					qd.vat_id as quo_dt_vat_id,

					pi.name as product_name,
					it.name as item_name,
					cur.name as currency_name,
					vat.name as vat_name,
					pph.name as pph23_name,

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
					q.id, q.customer_id, q.order_type_id, q.currency_id, q.vat_id, q.payment_id, q.pph23_id, q.branch_id, q.quo_no, q.title, q.remark, q.status, q.is_approved, q.exchange_rate, q.vat_perc, q.pph23_perc, q.total_qty, q.subtotal, q.total_discount, q.total_pph23, q.total_vat, q.grand_total, q.due_at, q.expired_at, q.created_by_id, q.updated_by_id, q.deleted_by_id, q.created_at, q.updated_at, q.deleted_at,
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

	// sqlConn, err := tx.DB()
	// if err != nil {
	// 	utils.LogErrors(childSpan, err)
	// 	return nil, err
	// }

	// // Use sqlx with the extracted SQL connection
	// sqlxDB := sqlx.NewDb(sqlConn, "postgres")

	// if err := sqlxDB.GetContext(ctx.Context(), &quotation, query, args...); err != nil {
	// 	utils.LogErrors(childSpan, err)
	// 	return nil, err
	// }

	if err := r.sqlDB.Get(&quotation, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
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

	if err := tx.Select("*").Omit(
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

func (r *QuotationRepository) CreateQuoDts(tx *gorm.DB, quoDts []*models.QuoDt, quotationID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("QuoDtRepository-CreateQuoDts", opentracing.ChildOf(span.Context()))

	result := tx.CreateInBatches(quoDts, len(quoDts))

	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		// return result.Error
		return nil, result.Error
	}

	return result, nil
}

// bulk/batch update quoDts
func (r *QuotationRepository) UpdateQuoDts(tx *gorm.DB, quoDts []*models.QuoDt, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("QuoDtRepository-UpdateQuoDts", opentracing.ChildOf(span.Context()))

	data := make([]map[string]interface{}, 0)
	for _, quoDt := range quoDts {
		data = append(data, map[string]interface{}{
			"id":           quoDt.ID,
			"quotation_id": quoDt.QuotationID,
			"ref_id":       quoDt.RefID,
			"vat_id":       quoDt.VatID,
			"item_unit_id": quoDt.ItemUnitID,
			"item_id":      quoDt.ItemID,
			"ref_type":     quoDt.RefType,
			"item_type":    quoDt.ItemType,
			// "ref_json":      quoDt.RefJSON,
			"gen_code":      quoDt.GenCode,
			"remark":        quoDt.Remark,
			"vat_perc":      quoDt.VatPerc,
			"vat_perc_am":   quoDt.VatPercAm,
			"qty_so":        quoDt.QtySO,
			"qty":           quoDt.Qty,
			"price_sell":    quoDt.PriceSell,
			"price_buy":     quoDt.PriceBuy,
			"subtotal_sell": quoDt.SubtotalSell,
			"subtotal_buy":  quoDt.SubtotalBuy,
			"disc_am":       quoDt.DiscAm,
			"disc_perc":     quoDt.DiscPerc,
			"disc_perc_num": quoDt.DiscPercNum,
			"disc_perc_am":  quoDt.DiscPercAm,
			"disc_final":    quoDt.DiscFinal,
			"disc_type":     quoDt.DiscType,
			"total_am":      quoDt.TotalAm,
			"updated_by_id": quoDt.UpdatedByID,
			"updated_at":    time.Now(),
		})
	}

	// if err := r.utilRepo.BulkUpdate(tx, "quoDts", "id", data, childSpan); err != nil {
	if err := r.utilRepo.Upsert(tx, "quo_dts", "id", data, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *QuotationRepository) DeleteQuoDtsWhereNotIn(tx *gorm.DB, quotationID uint, quoDtIDs []uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("QuoDtRepository-DeleteQuoDtsWhereNotIn", opentracing.ChildOf(span.Context()))

	if err := tx.Where("quotation_id = ? AND id NOT IN ?", quotationID, quoDtIDs).Delete(&models.QuoDt{}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

// func (r *QuotationRepository) GetQuoDtsByQuotationIDs(ctx *fiber.Ctx, quotationIDs uint, span opentracing.Span) ([]dtos.QuotationQuoDtListDTO, error) {
func (r *QuotationRepository) GetQuoDtsByQuotationIDs(ctx *fiber.Ctx, tx *gorm.DB, quotationIDs []uint, span opentracing.Span) ([]dtos.QuotationQuoDtListDTO, error) {
	childSpan := opentracing.StartSpan("QuotationRepository-GetQuoDtsByQuotationID", opentracing.ChildOf(span.Context()))

	quoDts := []dtos.QuotationQuoDtListDTO{}

	query := `SELECT qd.id, qd.quotation_id, 
		qd.item_unit_id, qd.vat_id, qd.ref_id, qd.item_id, qd.ref_type, qd.item_type, qd.gen_code, qd.remark, qd.vat_perc, qd.qty_so, qd.qty, qd.price_sell, qd.price_buy, qd.subtotal_sell, qd.subtotal_buy, qd.vat_perc, qd.vat_perc_am, qd.disc_am, qd.disc_perc, qd.disc_perc_num, qd.disc_perc_am, qd.disc_final, qd.disc_type, qd.total_am, qd.created_by_id, qd.updated_by_id, qd.deleted_by_id, qd.created_at, qd.updated_at, qd.deleted_at,
		qd.created_at, qd.updated_at, qd.deleted_at,

		isg.id as item_sub_group_id,
		ig.id as item_group_id,
		isg.name as item_sub_group_name,
		ig.name as item_group_name,
		u.name as unit_name,
		pi.name as item_name,

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
	// i := 1

	if len(quotationIDs) > 0 {
		query += " AND qd.quotation_id IN (" + utils.JoinUintsToString(quotationIDs, ",") + ")"
	}

	// sqlConn, err := tx.DB()
	// if err != nil {
	// 	utils.LogErrors(childSpan, err)
	// 	return nil, err
	// }

	// // Use sqlx with the extracted SQL connection
	// sqlxDB := sqlx.NewDb(sqlConn, "postgres")

	// if err := sqlxDB.SelectContext(ctx.Context(), &quoDts, query, args...); err != nil {
	// 	utils.LogErrors(childSpan, err)
	// 	return nil, err
	// }

	// log.Println("GetQuoDtsByQuotationIDs-result", quoDts)

	// return quoDts, nil

	// if len(quotationIDs) > 0 {
	// 	query += " AND qd.quotation_id IN (" + utils.JoinUintsToString(quotationIDs, ",") + ")"
	// }

	// if err := tx.Raw(query).Find(&quoDts).Error; err != nil {
	// 	utils.LogErrors(childSpan, err)
	// 	return nil, err
	// }

	// log.Println("GetQuoDtsByQuotationIDs-result", quoDts)

	// return quoDts, nil

	// Extract the underlying SQL connection from the GORM transaction
	sqlConn, err := tx.DB()
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	// Use sqlx with the extracted SQL connection
	sqlxDB := sqlx.NewDb(sqlConn, "postgres")

	if err := sqlxDB.SelectContext(ctx.Context(), &quoDts, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	log.Println("GetQuoDtsByQuotationIDs-result", quoDts)

	return quoDts, nil
}

func (r *QuotationRepository) CreateQuoDtBoms(tx *gorm.DB, quoDtBoms []*models.QuoDtBom, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuoDtBomRepository-CreateQuoDtBoms", opentracing.ChildOf(span.Context()))

	result := tx.CreateInBatches(quoDtBoms, len(quoDtBoms))

	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return result.Error
	}

	return nil
}

func (r *QuotationRepository) DeleteQuoDtBomsWhereNotIn(tx *gorm.DB, quotationID uint, quoDtIDs []uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuoDtRepository-DeleteQuoDtsWhereNotIn", opentracing.ChildOf(span.Context()))

	if err := tx.Where("quotation_id = ? AND id NOT IN ?", quotationID, quoDtIDs).Delete(&models.QuoDt{}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

// bulk/batch update quoDts
func (r *QuotationRepository) UpdateQuoDtBoms(tx *gorm.DB, quoDtBoms []*models.QuoDtBom, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuoDtRepository-UpdateQuoDtBoms", opentracing.ChildOf(span.Context()))

	data := make([]map[string]interface{}, 0)
	for _, quoDtBom := range quoDtBoms {
		data = append(data, map[string]interface{}{
			"id":           quoDtBom.ID,
			"quotation_id": quoDtBom.QuotationID,
			"quo_dt_id":    quoDtBom.QuoDtID,
			"product_id":   quoDtBom.ProductID,
			"item_id":      quoDtBom.ItemID,
			"item_unit_id": quoDtBom.ItemUnitID,
			// "ref_json":      quoDtBom.RefJSON,
			"gen_code":      quoDtBom.GenCode,
			"remark":        quoDtBom.Remark,
			"qty":           quoDtBom.Qty,
			"price_sell":    quoDtBom.PriceSell,
			"price_buy":     quoDtBom.PriceBuy,
			"subtotal_sell": quoDtBom.SubtotalSell,
			"subtotal_buy":  quoDtBom.SubtotalBuy,
			"updated_by_id": quoDtBom.UpdatedByID,
			"updated_at":    time.Now(),
		})
	}

	// if err := r.utilRepo.BulkUpdate(tx, "quoDtBoms", "id", data, childSpan); err != nil {
	if err := r.utilRepo.Upsert(tx, "quo_dt_boms", "id", data, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *QuotationRepository) DeleteQuoDtBomsByQuotationID(tx *gorm.DB, params *dtos.GetQuotationParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuoDtRepository-DeleteQuoDtsWhereNotIn", opentracing.ChildOf(span.Context()))

	if err := tx.Where("quotation_id = ?", params.ID).Delete(&models.QuoDtBom{}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *QuotationRepository) DeleteQuoDtsByQuotationID(tx *gorm.DB, params *dtos.GetQuotationParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuoDtRepository-DeleteQuoDtsWhereNotIn", opentracing.ChildOf(span.Context()))

	if err := tx.Where("quotation_id = ?", params.ID).Delete(&models.QuoDt{}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}
