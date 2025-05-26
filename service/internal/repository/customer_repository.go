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

	condition := ""
	var args []interface{}
	i := 1

	filterEqual := map[string]string{
		"status":           "c.status",
		"is_active":        "c.status",
		"is_crm":           "c.is_crm",
		"customer_type_id": "c.customer_type_id",
		"agent_id":         "c.agent_id",
		"category_type_id": "c.category_type_id",
		"currency_id":      "c.currency_id",
		"payment_type_id":  "cc.payment_type_id",
	}
	for key, colDB := range filterEqual {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", colDB, i)
			args = append(args, value)
			i++
		}
	}

	baseQuery := `FROM ( 
        SELECT DISTINCT ON (c.id)
					c.id, c.customer_type_id, c.agent_id, c.currency_id, c.shortname, c.name, c.code, c.address, c.phone, c.email, c.pic, c.status, c.created_at, c.updated_at, c.deleted_at,
				ct.name as customer_type_name,
				ag.name as agent_name,

				cgt.name as category_type_name,
				pymt.name as payment_type_name,
				TO_CHAR(c.contract_date, 'YYYY-MM-DD') as contract_date,
				COALESCE(c.is_contract, 0) as is_contract,
				cc.price as contract_price,
				TO_CHAR(cc.due_at, 'YYYY-MM-DD') as due_at,
				COALESCE(c.is_crm, 0) as is_crm,

        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM customers c
				LEFT JOIN mix_values ct ON c.customer_type_id = ct.id
				LEFT JOIN customers ag ON ag.agent_id = ag.id

				LEFT JOIN mix_values cgt ON c.category_type_id = cgt.id
				
				-- get active today to early date, if today 2024-01-10, there's 2 contracts 2023-01-01 and 2024-01-20, get 2024-01-20
				-- LEFT JOIN customer_contracts cc ON c.id = cc.customer_id AND cc.deleted_at IS NULL AND (cc.due_at >= NOW()::DATE OR cc.due_at <= NOW()::DATE) ORDER BY cc.due_at ASC
				LEFT JOIN (
						SELECT DISTINCT ON (customer_id) *
						FROM customer_contracts 
						WHERE deleted_at IS NULL 
						AND (due_at >= NOW()::DATE)
						ORDER BY customer_id, due_at ASC
				) cc ON c.id = cc.customer_id
				LEFT JOIN mix_values pymt ON cc.payment_type_id = pymt.id

        LEFT JOIN users cu ON c.created_by_id = cu.id
        LEFT JOIN users uu ON c.updated_by_id = uu.id
				WHERE 1=1 ` + condition + `
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	query := `SELECT * ` + baseQuery

	countQuery := `SELECT COUNT(*) ` + baseQuery
	// countQuery := `SELECT COUNT(*) FROM (
	//       SELECT DISTINCT ON (c.id)
	// 				c.id, c.customer_type_id, c.agent_id, c.currency_id, c.shortname, c.name, c.code, c.address, c.phone, c.email, c.pic, c.status, c.created_at, c.updated_at, c.deleted_at,
	// 			ct.name as customer_type_name,
	// 			ag.name as agent_name,

	//       cu.name as created_by_name,
	//       uu.name as updated_by_name

	//       FROM customers c
	// 			LEFT JOIN mix_values ct ON c.customer_type_id = ct.id
	// 			LEFT JOIN customers ag ON ag.agent_id = ag.id
	//       LEFT JOIN users cu ON c.created_by_id = cu.id
	//       LEFT JOIN users uu ON c.updated_by_id = uu.id
	// 			WHERE 1=1 ` + condition + `
	//   ) AS alias WHERE 1=1 AND deleted_at IS NULL`

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

	// filterKey := []string{
	// 	"customer_type_id",
	// 	"agent_id",
	// 	"status",
	// }

	// for key := range filterKey {
	// 	if filters[filterKey[key]] != "" {
	// 		query += fmt.Sprintf(" AND %s = $%d", filterKey[key], i)
	// 		countQuery += fmt.Sprintf(" AND %s = $%d", filterKey[key], i)
	// 		args = append(args, filters[filterKey[key]])
	// 		i++
	// 	}
	// }

	if filters["ids"] != "" {
		query += fmt.Sprintf(" AND id IN (%s)", filters["ids"])
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
		c.id, c.customer_type_id, c.agent_id, c.currency_id, c.shortname, c.name, c.code, c.address, c.phone, c.email, c.pic, c.status, 
		c.remark, c.owner_name, c.owner_phone, c.owner_email, c.category_type_id, c.is_contract,
		c.pic_name, c.pic_phone,
		c.created_at, c.updated_at, c.deleted_at,
		COALESCE(c.is_crm, 0) as is_crm,

		TO_CHAR(c.contract_date, 'YYYY-MM-DD') as contract_date,
		
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

func (r *CustomerRepository) GetCustomerPicEmails(ctx *fiber.Ctx, params *dtos.GetCustomerParams, span opentracing.Span) ([]dtos.FormCustomerPICEmailsRequest, error) {
	childSpan := opentracing.StartSpan("CustomerRepository-GetCustomerPicEmails", opentracing.ChildOf(span.Context()))
	var customer []dtos.FormCustomerPICEmailsRequest

	query := `
	SELECT 
		c.id, c.name, c.is_main,
		c.created_by_id, c.updated_by_id, c.deleted_by_id

	FROM pic_emails c
	WHERE 1=1`

	var args []interface{}

	i := 1
	query += " AND c.customer_id = $1"
	args = append(args, params.ID)
	i++

	isDeletedQuery := ` AND c.deleted_at IS NULL`
	if params.IsDeleted != nil && *params.IsDeleted == 1 {
		isDeletedQuery = " AND c.deleted_at IS NOT NULL"
	}

	query += isDeletedQuery

	if err := r.sqlDB.SelectContext(ctx.Context(), &customer, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return customer, nil
}

func (r *CustomerRepository) GetCustomerContracts(ctx *fiber.Ctx, params *dtos.GetCustomerParams, span opentracing.Span) ([]dtos.FormCustomerContractsRequest, error) {
	childSpan := opentracing.StartSpan("CustomerRepository-GetCustomerContracts", opentracing.ChildOf(span.Context()))
	var customer []dtos.FormCustomerContractsRequest

	query := `
	SELECT 
		c.id, c.product_id, c.price, c.payment_type_id, c.qty, c.remark,
		c.created_by_id, c.updated_by_id, c.deleted_by_id,

		TO_CHAR(c.agree_at, 'YYYY-MM-DD') as agree_at,
		TO_CHAR(c.due_at, 'YYYY-MM-DD') as due_at,
		TO_CHAR(c.installation_at, 'YYYY-MM-DD') as installation_at,
		TO_CHAR(c.warranty_at, 'YYYY-MM-DD') as warranty_at,
		c.customer_id,

		p.name as product_name,
		pt.name as payment_type_name,
		isg.name as item_sub_group_name,
		ig.name as item_group_name

	FROM customer_contracts c
	JOIN products p ON c.product_id = p.id
	LEFT JOIN mix_values isg ON p.item_sub_group_id = isg.id
	LEFT JOIN mix_values ig ON isg.parent_id = ig.id
	LEFT JOIN mix_values pt ON c.payment_type_id = pt.id
	LEFT JOIN customers cu ON c.customer_id = cu.id
	WHERE 1=1 AND c.deleted_at IS NULL AND c.product_id IS NOT NULL`

	var args []interface{}

	i := 1
	query += " AND c.customer_id = $1"
	args = append(args, params.ID)
	i++

	isDeletedQuery := ` AND c.deleted_at IS NULL`
	if params.IsDeleted != nil && *params.IsDeleted == 1 {
		isDeletedQuery = " AND c.deleted_at IS NOT NULL"
	}

	query += isDeletedQuery

	if err := r.sqlDB.SelectContext(ctx.Context(), &customer, query, args...); err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return customer, nil
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

// // delete customer_contracts where not in customer_id & id
// func (r *CustomerRepository) DeleteCustomerContracts(tx *gorm.DB, customerContractIDs []uint, customerID uint, span opentracing.Span) error {
// 	childSpan := opentracing.StartSpan("CustomerRepository-DeleteCustomerContracts", opentracing.ChildOf(span.Context()))

// 	if err := tx.Where("customer_id = ? AND id NOT IN (?)", customerID, customerContractIDs).Delete(&models.CustomerContract{}).Error; err != nil {
// 		defer childSpan.Finish()
// 		utils.LogErrors(childSpan, err)
// 		return err
// 	}
// 	return nil
// }

func (r *CustomerRepository) DeleteCustomerContracts(tx *gorm.DB, customerContractIDs []uint, customerID uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CustomerRepository-DeleteCustomerContracts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	query := tx.Where("customer_id = ?", customerID)

	// Only add the NOT IN clause if the array is not empty
	if len(customerContractIDs) > 0 {
		query = query.Where("id NOT IN (?)", customerContractIDs)
	}

	if err := query.Delete(&models.CustomerContract{}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

// create/update customer_contracts
func (r *CustomerRepository) CreateUpdateCustomerContracts(tx *gorm.DB, customerContracts []models.CustomerContract, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CustomerRepository-CreateUpdateCustomerContracts", opentracing.ChildOf(span.Context()))

	if err := tx.Model(&models.CustomerContract{}).Updates(&customerContracts).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

// delete customer_contracts where not in customer_id & id
func (r *CustomerRepository) DeletePicEmails(tx *gorm.DB, customerEmailIDs []uint, customerID uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CustomerRepository-DeletePicEmails", opentracing.ChildOf(span.Context()))

	if err := tx.Where("customer_id = ? AND id NOT IN (?)", customerID, customerEmailIDs).Delete(&models.PicEmail{}).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

// create/update customer_contracts
func (r *CustomerRepository) CreateUpdatePicEmails(tx *gorm.DB, customerEmails []models.PicEmail, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CustomerRepository-CreateUpdatePicEmails", opentracing.ChildOf(span.Context()))

	if err := tx.Model(&models.PicEmail{}).Updates(&customerEmails).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}
