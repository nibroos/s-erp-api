package repository

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type CustomerRepository struct {
	db     *gorm.DB
	sqlDB  *sqlx.DB
	tracer opentracing.Tracer
}

func NewCustomerRepository(db *gorm.DB, sqlDB *sqlx.DB, tracer opentracing.Tracer) *CustomerRepository {
	return &CustomerRepository{
		db:     db,
		sqlDB:  sqlDB,
		tracer: tracer,
	}
}

func (r *CustomerRepository) GetCustomers(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.CustomerListDTO, int, error) {
	// Create a child span for the controller
	childSpan := opentracing.StartSpan("CustomerRepository-GetCustomers", opentracing.ChildOf(span.Context()))

	customers := []dtos.CustomerListDTO{}
	var total int

	query := `SELECT *
    FROM ( 
        SELECT DISTINCT ON (c.id)
					c.id, c.customer_type_id, c.agent_id, c.name, c.code, c.address, c.phone, c.email, c.pic, c.status, c.created_at, c.updated_at, c.deleted_at,
				ct.name as customer_type_name,
				ag.name as agent_name,

        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM customers c
				LEFT JOIN mix_values ct ON c.customer_type_id = ct.id
				LEFT JOIN customers ag ON ag.agent_id = ag.id
        LEFT JOIN users cu ON c.created_by_id = cu.id
        LEFT JOIN users uu ON c.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT DISTINCT ON (c.id) 
					c.id, c.customer_type_id, c.agent_id, c.name, c.code, c.address, c.phone, c.email, c.pic, c.status, c.created_at, c.updated_at, c.deleted_at,
				ct.name as customer_type_name,
				ag.name as agent_name,

        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM customers c
				LEFT JOIN mix_values ct ON c.customer_type_id = ct.id
				LEFT JOIN customers ag ON ag.agent_id = ag.id
        LEFT JOIN users cu ON c.created_by_id = cu.id
        LEFT JOIN users uu ON c.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	var args []interface{}

	i := 1
	for key, value := range filters {
		switch key {
		case "name", "code", "address", "phone", "email", "pic":
			if value != "" {
				query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				countQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

	filterKey := []string{
		"customer_type_id",
		"agent_id",
		"status",
	}

	for key := range filterKey {
		if filters[filterKey[key]] != "" {
			query += fmt.Sprintf(" AND %s = $%d", filterKey[key], i)
			countQuery += fmt.Sprintf(" AND %s = $%d", filterKey[key], i)
			args = append(args, filters[filterKey[key]])
			i++
		}
	}

	if value, ok := filters["global"]; ok && value != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR code ILIKE $%d OR address ILIKE $%d OR phone ILIKE $%d OR email ILIKE $%d OR pic ILIKE $%d)", i, i+1, i+2, i+3, i+4, i+5)
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR code ILIKE $%d OR address ILIKE $%d OR phone ILIKE $%d OR email ILIKE $%d OR pic ILIKE $%d)", i, i+1, i+2, i+3, i+4, i+5)
		args = append(args, "%"+value+"%", "%"+value+"%", "%"+value+"%", "%"+value+"%", "%"+value+"%", "%"+value+"%")
		i += 6
	}

	countArgs := append([]interface{}{}, args...)

	var wg sync.WaitGroup
	var countErr, selectErr error

	// Goroutine for count query
	wg.Add(1)
	go func() {
		defer wg.Done()
		if filters["is_csv"] != "1" {
			// Create a span for the count query
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "name")
	orderDirection := utils.GetStringOrDefault(filters["order_direction"], "asc")
	query += fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)

	perPage := utils.GetIntOrDefault(filters["per_page"], 10)
	currentPage := utils.GetIntOrDefault(filters["page"], 1)

	// if is_csv
	if filters["is_csv"] != "1" {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
		args = append(args, perPage, (currentPage-1)*perPage)
	}

	// Goroutine for select query
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Create a span for the select query
		selectSpan := opentracing.StartSpan("SelectQuery", opentracing.ChildOf(childSpan.Context()))

		err := r.sqlDB.SelectContext(ctx.Context(), &customers, query, args...)
		if err != nil {
			selectSpan.LogKV("query", query)
			utils.LogErrors(selectSpan, err)
			selectErr = err
		}
	}()

	// Wait for both goroutines to finish
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

	return customers, total, nil
}

func (r *CustomerRepository) GetCustomerByID(ctx *fiber.Ctx, params *dtos.GetCustomerParams, span opentracing.Span) (*dtos.CustomerDetailDTO, error) {
	childSpan := opentracing.StartSpan("CustomerRepository-GetCustomerByID", opentracing.ChildOf(span.Context()))
	var customer dtos.CustomerDetailDTO

	query := `
	SELECT 
		c.id, c.customer_type_id, c.agent_id, c.name, c.code, c.address, c.phone, c.email, c.pic, c.status, c.created_at, c.updated_at, c.deleted_at,
    ct.name as customer_type_name,
    ag.name as agent_name,
    cu.name as created_by_name,
    uu.name as updated_by_name

	FROM customers c
	LEFT JOIN mix_values ct on c.customer_type_id = ct.id
	LEFT JOIN customers ag on c.agent_id = ag.id
	LEFT JOIN users cu on c.created_by_id = cu.id
	LEFT JOIN users uu on c.updated_by_id = uu.id
	WHERE 1=1`

	var args []interface{}

	i := 1
	query += " AND c.id = $1"
	args = append(args, params.ID)
	i++

	isDeletedQuery := ` AND c.deleted_at IS NULL`
	if params.IsDeleted != nil && *params.IsDeleted == 1 {
		isDeletedQuery = " AND c.deleted_at IS NOT NULL"
	}

	query += isDeletedQuery

	if err := r.sqlDB.Get(&customer, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return &customer, nil
}

// BeginTransaction starts a new transaction
func (r *CustomerRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *CustomerRepository) CreateCustomer(tx *gorm.DB, customer *models.Customer, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CustomerRepository-CreateCustomer", opentracing.ChildOf(span.Context()))
	if err := tx.Create(customer).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *CustomerRepository) UpdateCustomer(tx *gorm.DB, customer *models.Customer, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CustomerRepository-UpdateCustomer", opentracing.ChildOf(span.Context()))

	if err := tx.Select("*").Omit("created_at", "created_by_id").Updates(customer).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *CustomerRepository) DeleteCustomer(tx *gorm.DB, params *dtos.GetCustomerParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CustomerRepository-DeleteCustomer", opentracing.ChildOf(span.Context()))

	if err := tx.Delete(&models.Customer{}, params.ID).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (s *CustomerRepository) RestoreCustomer(tx *gorm.DB, params *dtos.GetCustomerParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CustomerRepository-RestoreCustomer", opentracing.ChildOf(span.Context()))

	var customer models.Customer
	if err := tx.Unscoped().Model(&customer).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}
