package repository

import (
	"fmt"
	"os"
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
	"gorm.io/gorm/clause"
)

type TicketRepository struct {
	db       *gorm.DB
	sqlDB    *sqlx.DB
	utilRepo *UtilRepository
	rabbitmq *config.RabbitMQ
	tracer   opentracing.Tracer
}

func NewTicketRepository(db *gorm.DB, sqlDB *sqlx.DB, utilRepo *UtilRepository, rabbitmq *config.RabbitMQ, tracer opentracing.Tracer) *TicketRepository {
	return &TicketRepository{
		db:       db,
		sqlDB:    sqlDB,
		rabbitmq: rabbitmq,
		tracer:   tracer,
		utilRepo: utilRepo,
	}
}

func (r *TicketRepository) GetTickets(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.TicketListDTO, int, error) {
	childSpan := opentracing.StartSpan("TicketRepository-GetTickets", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	products := []dtos.TicketListDTO{}

	var total int

	filterDBColumnKey := []string{
		"so.title", "so.ticket_no", "so.remark", "so.priority_type", "so.status", "so.issue_desc", "so.issue_solution",
		"pi.name",
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
		"product_id":    "so.product_id",
		"priority_type": "so.priority_type",
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
		"customer_ids":   "so.customer_id",
		"product_ids":    "so.product_id",
		"branch_ids":     "so.branch_id",
		"priority_types": "so.priority_type",
		"statuses":       "so.status",
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

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (so.id)
					so.id, so.customer_id, so.branch_id, so.product_id, 
					so.title, so.ticket_no, so.remark, so.priority_type,   
					so.status, so.created_by_id, so.updated_by_id, so.deleted_by_id, so.created_at, so.updated_at, so.deleted_at,
					TO_CHAR(so.reported_at, 'YYYY-MM-DD') as reported_at,

					pi.name as product_name,
					c.name as customer_name,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM tickets so
				LEFT JOIN products pi ON so.product_id = pi.id
				LEFT JOIN customers c ON so.customer_id = c.id

        LEFT JOIN users cu ON so.created_by_id = cu.id
        LEFT JOIN users uu ON so.updated_by_id = uu.id
				WHERE 1=1` + condition + queryGlobal + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery

	for key, value := range filters {
		switch key {
		case "title", "ticket_no", "remark":
			if value != "" {
				query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				countQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

	// if date_type, start_date, end_date filled
	if filters["start_date"] != "" && filters["end_date"] != "" {
		query += fmt.Sprintf(" AND (reported_at BETWEEN $%d AND $%d)", i, i+1)
		countQuery += fmt.Sprintf(" AND (reported_at BETWEEN $%d AND $%d)", i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "reported_at")
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

func (r *TicketRepository) GetTicketByID(ctx *fiber.Ctx, params *dtos.GetTicketParams, tx *gorm.DB, span opentracing.Span) (*dtos.TicketDetailDTO, error) {
	childSpan := opentracing.StartSpan("TicketRepository-GetTicketByID", opentracing.ChildOf(span.Context()))
	var ticket dtos.TicketDetailDTO

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	baseQuery := `
    FROM ( 
			SELECT DISTINCT ON (so.id)
				so.id, so.customer_id, so.branch_id, so.product_id, 
				so.title, so.ticket_no, so.remark, so.issue_desc, so.issue_solution, so.priority_type,   
				so.status, so.created_by_id, so.updated_by_id, so.deleted_by_id, so.created_at, so.updated_at, so.deleted_at,
				TO_CHAR(so.reported_at, 'YYYY-MM-DD') as reported_at,

				so.ticket_no_ori,

				cu.name as created_by_name,
				uu.name as updated_by_name

			FROM tickets so
			LEFT JOIN products pi ON so.product_id = pi.id
			LEFT JOIN customers c ON so.customer_id = c.id

			LEFT JOIN users cu ON so.created_by_id = cu.id
			LEFT JOIN users uu ON so.updated_by_id = uu.id
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

	if err := r.sqlDB.Get(&ticket, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		childSpan.LogKV("query", query)
		return nil, err
	}

	return &ticket, nil
}

func (r *TicketRepository) GetScheduleByTicketID(ctx *fiber.Ctx, params *dtos.GetTicketParams, tx *gorm.DB, span opentracing.Span) (*dtos.ScheduleDetailDTO, error) {
	childSpan := opentracing.StartSpan("TicketRepository-GetScheduleByTicketID", opentracing.ChildOf(span.Context()))
	var schedule dtos.ScheduleDetailDTO

	baseQuery := `
    FROM ( 
			SELECT DISTINCT ON (s.id)
				s.id, s.assignee_id, COALESCE(s.customer_id, so.customer_id) as customer_id, s.sales_order_id, s.uuid, s.steps_id, s.title, s.module_type, s.remark, s.status, s.color, s.created_by_id, s.updated_by_id, s.deleted_by_id, s.deleted_at,

				TO_CHAR(s.start_at, 'YYYY-MM-DD') as start_at,
				TO_CHAR(s.end_at, 'YYYY-MM-DD') as end_at,

				ass.name as assignee_name,
				cu.name as created_by_name,
				uu.name as updated_by_name

			FROM schedules s
			LEFT JOIN tickets so ON s.sales_order_id = so.id

			LEFT JOIN users ass ON s.assignee_id = ass.id
			LEFT JOIN users cu ON s.created_by_id = cu.id
			LEFT JOIN users uu ON s.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	var args []interface{}

	i := 1
	query += " AND sales_order_id = $1 AND module_type ='tickets'"
	args = append(args, params.ID)
	i++

	if err := r.sqlDB.Get(&schedule, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		childSpan.LogKV("query", query)
		return nil, err
	}

	return &schedule, nil
}

// BeginTransaction starts a new transaction
func (r *TicketRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

// Rollback all changes in the transaction
func (r *TicketRepository) Rollback() *gorm.DB {
	return r.db.Rollback()
}

func (r *TicketRepository) CreateTicket(tx *gorm.DB, ticket *models.Ticket, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("TicketRepository-CreateTicket", opentracing.ChildOf(span.Context()))
	if err := tx.Create(ticket).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *TicketRepository) UpdateTicket(tx *gorm.DB, ticket *models.Ticket, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TicketRepository-UpdateTicket", opentracing.ChildOf(span.Context()))

	if err := tx.Where("id = ?", ticket.ID).Select("*").Omit(
		"created_at", "created_by_id", "branch_id",
	).Updates(ticket).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *TicketRepository) DeleteTicket(tx *gorm.DB, params *dtos.GetTicketParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TicketRepository-DeleteTicket", opentracing.ChildOf(span.Context()))

	if err := tx.Delete(&models.Ticket{}, params.ID).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil

}

func (r *TicketRepository) DeleteScheduleTasksByScheduleID(tx *gorm.DB, params *dtos.GetTicketParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TicketRepository-DeleteScheduleTasksByScheduleID", opentracing.ChildOf(span.Context()))

	if err := tx.Model(&models.ScheduleTask{}).Where("schedule_id = ?", params.ID).Delete(&models.ScheduleTask{}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil

}

func (r *TicketRepository) DeleteScheduleByID(tx *gorm.DB, params *dtos.GetTicketParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TicketRepository-DeleteScheduleByID", opentracing.ChildOf(span.Context()))

	if err := tx.Model(&models.Schedule{}).Where("id = ?", params.ID).Delete(&models.Schedule{}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil

}

func (s *TicketRepository) RestoreTicket(tx *gorm.DB, params *dtos.GetTicketParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TicketRepository-RestoreTicket", opentracing.ChildOf(span.Context()))

	var ticket models.Ticket
	if err := tx.Unscoped().Model(&ticket).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

// Lock Ticket Header
func (r *TicketRepository) LockTicketHeader(ctx *fiber.Ctx, tx *gorm.DB, req dtos.FormTicketRequest, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TicketRepository-LockTicketHeader", opentracing.ChildOf(span.Context()))

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", []uint{*req.ID}).Find(&models.Ticket{}).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)

		return err
	}

	return nil
}

// GetCustomerTicketCreatedThisMonth
func (r *TicketRepository) GetCustomerTicketCreatedThisMonth(ctx *fiber.Ctx, tx *gorm.DB, customerID uint, span opentracing.Span) (int, error) {
	childSpan := opentracing.StartSpan("TicketRepository-GetCustomerTicketCreatedThisMonth", opentracing.ChildOf(span.Context()))

	var total int

	baseQuery := `
		FROM (
			SELECT COUNT(*) as total
			FROM tickets so
			WHERE so.customer_id = $1 AND so.created_at >= date_trunc('month', CURRENT_DATE)
			AND so.deleted_at IS NULL
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

// GetCustomerTicketCreatedThisMonth
func (r *TicketRepository) GetGlobalTicketCreatedThisMonth(ctx *fiber.Ctx, tx *gorm.DB, customerID uint, span opentracing.Span) (int, error) {
	childSpan := opentracing.StartSpan("TicketRepository-GetGlobalTicketCreatedThisMonth", opentracing.ChildOf(span.Context()))

	var total int

	baseQuery := `
		FROM (
			SELECT COUNT(*) as total
			FROM tickets so
			WHERE so.created_at >= date_trunc('month', CURRENT_DATE)
			AND so.deleted_at IS NULL
		) AS alias WHERE 1=1`

	query := `SELECT *
		` + baseQuery

	err := tx.Raw(query).Scan(&total).Error
	if err != nil {
		utils.LogErrors(childSpan, err)
		return 0, err
	}

	return total, nil
}

// CreateSchedule
func (r *TicketRepository) CreateSchedule(ctx *fiber.Ctx, tx *gorm.DB, schedule *models.Schedule, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("TicketRepository-CreateSchedule", opentracing.ChildOf(span.Context()))

	if err := tx.Create(schedule).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)

		return nil, err
	}

	return tx, nil
}

// CreateScheduleSteps
func (r *TicketRepository) CreateScheduleSteps(ctx *fiber.Ctx, tx *gorm.DB, scheduleSteps []*models.ScheduleTask, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("TicketRepository-CreateScheduleSteps", opentracing.ChildOf(span.Context()))

	if err := tx.Create(&scheduleSteps).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)

		return nil, err
	}

	return tx, nil
}

// CreateScheduleTasks
func (r *TicketRepository) CreateScheduleTasks(ctx *fiber.Ctx, tx *gorm.DB, scheduleTasks []*models.ScheduleTask, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("TicketRepository-CreateScheduleTasks", opentracing.ChildOf(span.Context()))

	if err := tx.Create(&scheduleTasks).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)

		return nil, err
	}

	return tx, nil
}

// GetScheduleTasksByScheduleID
func (r *TicketRepository) GetScheduleTasksByScheduleID(ctx *fiber.Ctx, filters map[string]string, scheduleIDs []uint, span opentracing.Span) ([]dtos.ScheduleTaskListDTO, error) {
	childSpan := opentracing.StartSpan("TicketRepository-GetScheduleTasksByScheduleID", opentracing.ChildOf(span.Context()))

	scheduleTasks := []dtos.ScheduleTaskListDTO{}

	filterDBColumnKey := []string{
		"st.title", "st.color", "st.remark",
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
		condition += fmt.Sprintf(" AND st.id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"schedule_id": "st.schedule_id",
		"entity_id":   "st.entity_id",
		"assignee_id": "st.assignee_id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		// "customer_ids":       "so.customer_id",
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
		// "vat_ids": []string{"so.vat_id", "sd.vat_id"},
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

	if len(scheduleIDs) > 0 {
		condition += fmt.Sprintf(" AND st.schedule_id = ANY($%d)", i)
		args = append(args, pq.Array(scheduleIDs))
		i++
	}

	filterKeyLike := map[string]string{
		"title":  "st.title",
		"color":  "st.color",
		"remark": "st.remark",
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
        SELECT DISTINCT ON (st.id)
					st.id, st.schedule_id, st.assignee_id, st.parent_id, st.entity_id, st.entity_type, st.uuid, st.parent_uuid, st.title, st.remark, st.order_item, st.color, st.is_checked,
					
					TO_CHAR(st.start_at, 'YYYY-MM-DD') as start_at,
					TO_CHAR(st.end_at, 'YYYY-MM-DD') as end_at,

					st.created_by_id, st.updated_by_id, st.deleted_by_id, st.created_at, st.updated_at, st.deleted_at,

					cu.name as created_by_name,
					uu.name as updated_by_name

				FROM schedule_tasks st

        LEFT JOIN users cu ON st.created_by_id = cu.id
        LEFT JOIN users uu ON st.updated_by_id = uu.id
				WHERE 1=1` + condition + queryGlobal + `
			
    ) AS alias WHERE 1=1 AND deleted_at IS NULL 
		ORDER BY order_item ASC`

	query := `SELECT *
		` + baseQuery

	selectSpan := opentracing.StartSpan("SelectQuery", opentracing.ChildOf(childSpan.Context()))

	err := r.sqlDB.SelectContext(ctx.Context(), &scheduleTasks, query, args...)
	if err != nil {
		selectSpan.LogKV("query", query)
		utils.LogErrors(selectSpan, err)
		return nil, err
	}

	return scheduleTasks, nil
}

func (r *TicketRepository) UpdateTicketSchedule(tx *gorm.DB, ticketSchedule *models.Schedule, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TicketRepository-UpdateTicketSchedule", opentracing.ChildOf(span.Context()))

	if err := tx.Where("id = ?", ticketSchedule.ID).Select("*").Omit(
		"created_at", "created_by_id",
	).Updates(ticketSchedule).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *TicketRepository) DeleteScheduleStepsWhereNotIn(ctx *fiber.Ctx, tx *gorm.DB, scheduleID uint, scheduleStepIDs []uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SoDtRepository-DeleteScheduleStepsWhereNotIn", opentracing.ChildOf(span.Context()))

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	now := time.Now()
	deletedAt := gorm.DeletedAt{Time: now, Valid: true}

	query := tx.Model(&models.ScheduleTask{}).Where("schedule_id = ? AND entity_type = 'steps' AND deleted_at IS NULL", scheduleID)

	if len(scheduleStepIDs) > 0 {
		query = query.Where("id NOT IN (?)", scheduleStepIDs)
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

// bulk/batch update schedule steps
func (r *TicketRepository) UpdateScheduleSteps(ctx *fiber.Ctx, tx *gorm.DB, scheduleSteps []*models.ScheduleTask, userID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SoDtRepository-UpdateScheduleSteps", opentracing.ChildOf(span.Context()))

	data := make([]map[string]interface{}, 0)
	for _, scheduleStep := range scheduleSteps {
		data = append(data, map[string]interface{}{
			"id":            scheduleStep.ID,
			"schedule_id":   scheduleStep.ScheduleID,
			"assignee_id":   scheduleStep.AssigneeID,
			"parent_id":     scheduleStep.ParentID,
			"entity_id":     scheduleStep.EntityID,
			"entity_type":   scheduleStep.EntityType,
			"uuid":          scheduleStep.UUID,
			"parent_uuid":   scheduleStep.ParentUUID,
			"title":         scheduleStep.Title,
			"remark":        scheduleStep.Remark,
			"order_item":    scheduleStep.OrderItem,
			"color":         scheduleStep.Color,
			"is_checked":    scheduleStep.IsChecked,
			"start_at":      scheduleStep.StartAt,
			"end_at":        scheduleStep.EndAt,
			"updated_by_id": userID,
			"updated_at":    time.Now(),
		})
	}

	// if err := r.utilRepo.BulkUpdate(tx, "scheduleSteps", "id", data, childSpan); err != nil {
	if err := r.utilRepo.Upsert(tx, "schedule_tasks", "id", data, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *TicketRepository) GetUpdatedScheduleStepsByTicketScheduleIDs(ctx *fiber.Ctx, tx *gorm.DB, ticketScheduleIDs []uint, span opentracing.Span) ([]dtos.UpdatedScheduleStepListDTO, error) {
	childSpan := opentracing.StartSpan("TicketRepository-GetSoDtsByTicketIDs", opentracing.ChildOf(span.Context()))

	steps := []dtos.UpdatedScheduleStepListDTO{}

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (st.id)
					st.id, st.schedule_id, st.assignee_id, st.parent_id, st.entity_id, st.entity_type, st.uuid, st.parent_uuid, st.title, st.remark, st.order_item, st.color, st.is_checked,
					
					TO_CHAR(st.start_at, 'YYYY-MM-DD') as start_at,
					TO_CHAR(st.end_at, 'YYYY-MM-DD') as end_at,

					st.created_by_id, st.updated_by_id, st.deleted_by_id, st.created_at, st.updated_at, st.deleted_at,

					cu.name as created_by_name,
					uu.name as updated_by_name

				FROM schedule_tasks st

        LEFT JOIN users cu ON st.created_by_id = cu.id
        LEFT JOIN users uu ON st.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	var args []interface{}
	i := 1

	if len(ticketScheduleIDs) > 0 {
		query += " AND schedule_id = ANY($1)"
		args = append(args, pq.Array(ticketScheduleIDs))
		i++

	}

	// GORM Raw
	if err := tx.Raw(query, args...).Scan(&steps).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return steps, nil
}

func (r *TicketRepository) DeleteScheduleTasksWhereNotIn(ctx *fiber.Ctx, tx *gorm.DB, ticketID uint, scheduleTaskIDs []uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SoDtRepository-DeleteScheduleTasksWhereNotIn", opentracing.ChildOf(span.Context()))

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	now := time.Now()
	deletedAt := gorm.DeletedAt{Time: now, Valid: true}

	query := tx.Model(&models.ScheduleTask{}).Where("schedule_id = ? AND entity_type = 'tasks' AND deleted_at IS NULL", ticketID)

	if len(scheduleTaskIDs) > 0 {
		query = query.Where("id NOT IN (?)", scheduleTaskIDs)
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

func (r *TicketRepository) UpdateScheduleTasks(tx *gorm.DB, tasks []map[string]interface{}, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SoDtRepository-UpdateScheduleTasks", opentracing.ChildOf(span.Context()))

	if err := r.utilRepo.Upsert(tx, "schedule_tasks", "id", tasks, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

// CreateTicketFiles
func (r *TicketRepository) CreateTicketFiles(ctx *fiber.Ctx, tx *gorm.DB, letters []*models.Letter, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("TicketRepository-CreateTicketFiles", opentracing.ChildOf(span.Context()))

	if err := tx.Create(&letters).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)

		return nil, err
	}

	return tx, nil
}

// GetAttachmentsByTicketID
func (r *TicketRepository) GetAttachmentsByTicketID(ctx *fiber.Ctx, tx *gorm.DB, ticketID uint, attachmentType string, span opentracing.Span) ([]dtos.SalesOrderAttachmentsDTO, error) {
	childSpan := opentracing.StartSpan("TicketRepository-GetAttachmentsByTicketID", opentracing.ChildOf(span.Context()))

	attachments := []dtos.SalesOrderAttachmentsDTO{}

	var args []interface{}

	i := 1

	condition := ""

	condition += fmt.Sprintf(" AND ltr.ref_id = $%d AND ltr.ref_type = 'tickets'", i)
	args = append(args, ticketID)
	i++

	if attachmentType != "" {
		condition += fmt.Sprintf(" AND ltr.file_prop->>'attachment_type' = $%d", i)
		args = append(args, attachmentType)
		i++
	}

	baseQuery := `
    FROM ( 
			SELECT DISTINCT ON (ltr.id)
				ltr.id, ltr.ref_id, ltr.ref_type, ltr.file_type, ltr.file_url, ltr.file_name,  ltr.remark, ltr.created_at, ltr.deleted_at,
				-- json file_size
				ltr.file_prop->>'file_size' as file_size,
				ltr.file_prop->>'device_type' as device_type,
				ltr.file_prop->>'attachment_type' as attachment_type,

				TO_CHAR(ltr.created_at, 'YYYY-MM-DD') as created_at,
				1 as is_checked,

				cu.name as created_by_name,
				uu.name as updated_by_name

			FROM letters ltr

			LEFT JOIN users cu ON ltr.created_by_id = cu.id
			LEFT JOIN users uu ON ltr.updated_by_id = uu.id
			WHERE 1=1 ` + condition + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	// query += " AND ref_id = $1"
	// args = append(args, ticketID)
	// i++

	if err := r.sqlDB.SelectContext(ctx.Context(), &attachments, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return attachments, nil
}

// GetAttachmentsByTicketID
func (r *TicketRepository) GetAttachmentsByScheduleID(ctx *fiber.Ctx, tx *gorm.DB, scheduleID uint, span opentracing.Span) ([]dtos.ScheduleAttachmentsDTO, error) {
	childSpan := opentracing.StartSpan("TicketRepository-GetAttachmentsByScheduleID", opentracing.ChildOf(span.Context()))

	attachments := []dtos.ScheduleAttachmentsDTO{}

	baseQuery := `
    FROM ( 
			SELECT DISTINCT ON (ltr.id)
				ltr.id, ltr.ref_id, ltr.ref_type, ltr.file_type, ltr.file_url, ltr.file_name,  ltr.remark, ltr.created_at, ltr.deleted_at,
				
				ltr.file_prop->>'attachment_type' as attachment_type,
				ltr.file_prop->>'file_size' as file_size,
				ltr.file_prop->>'device_type' as device_type,

				TO_CHAR(ltr.created_at, 'YYYY-MM-DD') as created_at,

				cu.name as created_by_name,
				uu.name as updated_by_name

			FROM letters ltr

			LEFT JOIN users cu ON ltr.created_by_id = cu.id
			LEFT JOIN users uu ON ltr.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT *
		` + baseQuery

	var args []interface{}

	i := 1
	query += " AND ref_id = $1 AND ref_type = 'schedules'"
	args = append(args, scheduleID)
	i++

	if err := r.sqlDB.SelectContext(ctx.Context(), &attachments, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return attachments, nil
}

// DeleteTicketFilesByIDs: Delete database & file record
func (r *TicketRepository) DeleteTicketFilesByIDs(ctx *fiber.Ctx, tx *gorm.DB, letterIDs []uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TicketRepository-DeleteTicketFilesByIDs", opentracing.ChildOf(span.Context()))

	// First get the file paths before marking records as deleted
	var letters []models.Letter
	if err := tx.Where("id IN (?) AND deleted_at IS NULL", letterIDs).Find(&letters).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	now := time.Now()
	deletedAt := gorm.DeletedAt{Time: now, Valid: true}

	// Mark records as deleted in database
	query := tx.Model(&models.Letter{}).Where("id IN (?) AND deleted_at IS NULL", letterIDs)
	if err := query.Updates(map[string]interface{}{
		"deleted_by_id": userID,
		"deleted_at":    deletedAt,
	}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	// Create error channel for collecting deletion errors
	errChan := make(chan error, len(letters))
	var wg sync.WaitGroup

	// Delete physical files concurrently
	for _, letter := range letters {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			// Check if file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				errChan <- fmt.Errorf("file %s does not exist", filePath)
				return
			}

			// Delete the file
			if err := os.Remove(filePath); err != nil {
				errChan <- fmt.Errorf("failed to delete file %s: %v", filePath, err)
				return
			}
		}(letter.FileUrl)
	}

	// Wait for all deletions to complete
	wg.Wait()
	close(errChan)

	// Collect any errors that occurred during file deletion
	var errors []error
	for err := range errChan {
		if err != nil {
			errors = append(errors, err)
			utils.LogErrors(childSpan, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some files failed to delete: %v", errors)
	}

	return nil
}

// DeleteTicketScheduleByTicketID
func (r *TicketRepository) DeleteTicketScheduleByTicketID(ctx *fiber.Ctx, tx *gorm.DB, ticketID uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TicketRepository-DeleteTicketScheduleByTicketID", opentracing.ChildOf(span.Context()))

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	now := time.Now()
	deletedAt := gorm.DeletedAt{Time: now, Valid: true}

	query := tx.Model(&models.Schedule{}).Where("sales_order_id = ? AND module_type = 'tickets' AND deleted_at IS NULL", ticketID)

	if err := query.Updates(map[string]interface{}{
		"deleted_by_id": userID,
		"deleted_at":    deletedAt,
	}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

// UpdateAttachmentsDesc
func (r *TicketRepository) UpdateAttachmentsDesc(ctx *fiber.Ctx, tx *gorm.DB, attachments []map[string]interface{}, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("TicketRepository-UpdateAttachmentsDesc", opentracing.ChildOf(span.Context()))

	if err := r.utilRepo.Upsert(tx, "letters", "id", attachments, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *TicketRepository) GetCalendars(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.CalendarListDTO, int, error) {
	childSpan := opentracing.StartSpan("TicketRepository-GetCalendars", opentracing.ChildOf(span.Context()))

	claims, _ := auth.GetAuthUser(ctx)
	branchID := claims["bid"]

	isAdmin := utils.IsAdmin(ctx)

	calendars := []dtos.CalendarListDTO{}

	var total int

	filterDBColumnKey := []string{
		"so.title", "so.ticket_no", "so.remark", "so.ship_dest",
		"pi.name",
		"it.name",
		"sd.remark",
		"sd.gen_code",
		"sdb.remark",
		"sdb.gen_code",
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
		"due_at":        "so.due_at",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	// // date_type, start_at, end_at
	// if filters["start_at"] != "" && filters["end_at"] != "" {
	// 	condition += fmt.Sprintf(" AND (s.start_at BETWEEN $%d AND $%d) OR (s.end_at BETWEEN $%d AND $%d)", i, i+1, i, i+1)
	// 	args = append(args, filters["start_at"], filters["end_at"])
	// 	i += 2
	// }

	filterIDsKey := map[string]string{
		"customer_ids":   "so.customer_id",
		"order_type_ids": "so.order_type_id",
		"currency_ids":   "so.currency_id",
		"payment_ids":    "so.payment_id",
		"pph23_ids":      "so.pph23_id",
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
		"vat_ids": []string{"so.vat_id", "sd.vat_id"},
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

	joinCondition := ""
	selectJoinCondition := ""

	filterKeyJoin := map[string]string{
		// "is_task_exists": "JOIN schedules s ON s.sales_order_id = so.id AND s.deleted_at IS NULL LEFT JOIN schedule_tasks stp ON stp.schedule_id = s.id AND stp.deleted_at IS NULL AND stp.entity_type = 'steps' LEFT JOIN schedule_tasks st ON st.parent_id = stp.id AND st.deleted_at IS NULL AND st.entity_type = 'tasks'",
		"is_task_exists": "JOIN schedule_tasks st ON st.schedule_id = s.id AND st.deleted_at IS NULL AND st.entity_type = 'tasks' AND st.order_item = 3",
	}
	filterKeySelectJoin := map[string]string{
		// "is_task_exists": "JOIN schedules s ON s.sales_order_id = so.id AND s.deleted_at IS NULL LEFT JOIN schedule_tasks stp ON stp.schedule_id = s.id AND stp.deleted_at IS NULL AND stp.entity_type = 'steps' LEFT JOIN schedule_tasks st ON st.parent_id = stp.id AND st.deleted_at IS NULL AND st.entity_type = 'tasks'",
		"is_task_exists": "st.title as task_title,",
	}
	for key, join := range filterKeyJoin {
		if filters[key] == "1" {
			joinCondition += fmt.Sprintf(" %s", join)
			selectJoinCondition += fmt.Sprintf(" %s", filterKeySelectJoin[key])
		}
	}

	customCondition := ""
	filterKeyCustom := map[string]string{
		// "is_task_exists": " AND st.is_checked = 1",
	}
	for _, join := range filterKeyCustom {
		customCondition += fmt.Sprintf("%s", join)
	}

	baseQuery := `
    FROM ( 
        SELECT DISTINCT ON (s.id)
					s.id,
					s.sales_order_id,
					-- s.status,
					s.title,
					s.color,
					s.total_task_step_4_done,
					s.total_all_tasks_done,
					s.total_tasks,
					COALESCE(s.total_tasks_4, 0) as total_tasks_4,
					TO_CHAR(so.reported_at, 'YYYY-MM-DD') as reported_at,
					
					-- IF NULL THEN TODAY
					TO_CHAR(COALESCE(s.start_at, CURRENT_DATE), 'YYYY-MM-DD') as start,
					TO_CHAR(COALESCE(s.end_at, CURRENT_DATE), 'YYYY-MM-DD') as end,
					` + selectJoinCondition + `

					ot.name as order_type_name,
					c.name as customer_name

				FROM schedules s
				LEFT JOIN tickets so ON s.sales_order_id = so.id AND so.deleted_at IS NULL
				LEFT JOIN mix_values ot ON so.order_type_id = ot.id
				LEFT JOIN customers c ON so.customer_id = c.id
				` + joinCondition + `
				WHERE 1=1` + condition + queryGlobal + customCondition + `
				AND s.deleted_at IS NULL
    ) AS alias WHERE 1=1`

	query := `SELECT *
		` + baseQuery

	countQuery := `SELECT COUNT(*) as total
		` + baseQuery

	for key, value := range filters {
		switch key {
		case "title", "ticket_no", "ship_dest", "remark":
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "start")
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

		err := r.sqlDB.SelectContext(ctx.Context(), &calendars, query, args...)
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

	return calendars, total, nil
}

func (r *TicketRepository) GetScheduleByID(ctx *fiber.Ctx, params *dtos.GetTicketParams, tx *gorm.DB, span opentracing.Span) (*dtos.ScheduleSingleDetailDTO, error) {
	childSpan := opentracing.StartSpan("TicketRepository-GetScheduleByID", opentracing.ChildOf(span.Context()))
	var schedule dtos.ScheduleSingleDetailDTO

	baseQuery := `
    FROM ( 
			SELECT DISTINCT ON (s.id)
				s.id, s.assignee_id, s.customer_id, s.sales_order_id, s.title, s.module_type, s.color, s.remark,
					s.total_task_step_4_done,
					s.total_all_tasks_done,
					s.total_tasks,
					COALESCE(s.total_tasks_4, 0) as total_tasks_4,

				TO_CHAR(s.start_at, 'YYYY-MM-DD') as start_at,
				TO_CHAR(s.end_at, 'YYYY-MM-DD') as end_at,

				c.name as customer_name,
				ass.name as assignee_name,
				cu.name as created_by_name,
				uu.name as updated_by_name

			FROM schedules s
			LEFT JOIN customers c ON s.customer_id = c.id
			LEFT JOIN users ass ON s.assignee_id = ass.id
			LEFT JOIN users cu ON s.created_by_id = cu.id
			LEFT JOIN users uu ON s.updated_by_id = uu.id
			WHERE s.deleted_at IS NULL
    ) AS alias WHERE 1=1`

	query := `SELECT *
		` + baseQuery

	var args []interface{}

	i := 1
	query += " AND id = $1"
	args = append(args, params.ID)
	i++

	if err := r.sqlDB.Get(&schedule, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		childSpan.LogKV("query", query)
		return nil, err
	}

	return &schedule, nil
}

func (r *TicketRepository) GetScheduleTaskTotalDoneByID(ctx *fiber.Ctx, tx *gorm.DB, scheduleID uint, span opentracing.Span) ([]dtos.ListScheduleTaskByScheduleID, error) {
	childSpan := opentracing.StartSpan("TicketRepository-GetScheduleTaskTotalDoneByID", opentracing.ChildOf(span.Context()))
	var schedule []dtos.ListScheduleTaskByScheduleID

	baseQuery := `
    FROM ( 
			SELECT DISTINCT ON (st.id)
				st.id,
				st.schedule_id,
				ststep.order_item as step_order_item,
				st.is_checked

			FROM schedule_tasks st
			LEFT JOIN schedules s ON st.schedule_id = s.id AND s.deleted_at IS NULL
			LEFT JOIN schedule_tasks ststep ON st.parent_id = ststep.id AND ststep.deleted_at IS NULL AND ststep.entity_type = 'steps'
			WHERE s.deleted_at IS NULL AND st.deleted_at IS NULL AND st.entity_type = 'tasks'
    ) AS alias WHERE 1=1`

	query := `SELECT *
		` + baseQuery

	var args []interface{}

	i := 1
	query += " AND schedule_id = $1"
	args = append(args, scheduleID)
	i++

	if err := tx.Raw(query, args...).Scan(&schedule).Error; err != nil {
		utils.LogErrors(childSpan, err)
		childSpan.LogKV("query", query)
		return schedule, err
	}

	return schedule, nil
}

// bulk/batch update soDts
func (r *TicketRepository) UpdateTicketScheduleApp(tx *gorm.DB, schedule []map[string]interface{}, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SoDtRepository-UpdateSoDts", opentracing.ChildOf(span.Context()))

	if err := r.utilRepo.Upsert(tx, "schedules", "id", schedule, childSpan); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *TicketRepository) GetWidgetTickets(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.TicketStatusWidget, int, error) {
	childSpan := opentracing.StartSpan("TicketRepository-GetWidgetTickets", opentracing.ChildOf(span.Context()))

	widgets := []dtos.TicketStatusWidget{}

	var total int

	filterDBColumnKey := []string{
		"so.title", "so.ticket_no", "so.remark", "so.priority_type", "so.status", "so.issue_desc", "so.issue_solution",
		"pi.name",
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
		"product_id":    "so.product_id",
		"priority_type": "so.priority_type",
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
		"customer_ids":   "so.customer_id",
		"product_ids":    "so.product_id",
		"branch_ids":     "so.branch_id",
		"priority_types": "so.priority_type",
		"statuses":       "so.status",
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
	// 'OPEN' | 'IN PROGRESS' | 'RESOLVED' | 'CLOSED'
	query := `
				WITH status_values AS (
        SELECT status, row_number() over () as status_order
        FROM (VALUES 
            ('TOTAL', 0),      -- Set TOTAL with order 0 to appear first
            ('OPEN', 1),
						('IN PROGRESS', 2),
						('RESOLVED', 3),
						('CLOSED', 4)
        ) AS s(status)
    ),
    filtered_orders AS (
        SELECT DISTINCT ON (so.id)
            so.id,             
            so.status as so_status
        FROM tickets so
        LEFT JOIN products pi ON so.product_id = pi.id
        WHERE so.deleted_at IS NULL
        ` + condition + queryGlobal + `
    ),
    ticket_stats AS (
        SELECT 
            sv.status,
            sv.status_order,
            COUNT(DISTINCT fo.id) as ticket_count
        FROM status_values sv
        LEFT JOIN filtered_orders fo ON sv.status = fo.so_status OR sv.status = 'TOTAL'
        GROUP BY sv.status, sv.status_order
    )
    SELECT 
        sv.status,
        ticket_count,
				'tickets' as widget_type
    FROM ticket_stats sv
    ORDER BY status_order`

	// // Group by status
	// query += " GROUP BY sv.status ORDER BY sv.status"

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
