package repository

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type UtilRepository struct {
	db    *gorm.DB
	sqlDB *sqlx.DB
}

func NewUtilRepository(db *gorm.DB, sqlDB *sqlx.DB) *UtilRepository {
	return &UtilRepository{
		db:    db,
		sqlDB: sqlDB,
	}
}

func (r *UtilRepository) GetCompanyProfileByID(ctx *fiber.Ctx, params *dtos.GetCompanyProfileParams) (*dtos.CompanyProfileDetailDTO, error) {
	var CompanyProfile dtos.CompanyProfileDetailDTO

	query := `SELECT cp.id, cp.company_name, cp.company_address, cp.company_phone, cp.company_email, cp.company_website, cp.company_logo, cp.company_description, cp.company_remark, cp.company_status, cp.created_at, cp.updated_at, cp.deleted_at,
	cu.name as created_by_name,
	uu.name as updated_by_name

	FROM company_profiles cp
	LEFT JOIN users cu ON cp.created_by_id = cu.id
	LEFT JOIN users uu ON cp.updated_by_id = uu.id
	WHERE 1=1`

	var args []interface{}

	i := 1
	query += " AND cp.id = $1"
	args = append(args, params.ID)
	i++

	isDeletedQuery := ` AND cp.deleted_at IS NULL`
	if params.IsDeleted != nil && *params.IsDeleted == 1 {
		isDeletedQuery = " AND cp.deleted_at IS NOT NULL"
	}

	query += isDeletedQuery

	if err := r.sqlDB.Get(&CompanyProfile, query, args...); err != nil {
		return nil, err
	}

	return &CompanyProfile, nil
}

// Upsert performs an INSERT ... ON CONFLICT operation for any table.
//
// table: The table name.
//
// idColumn: The column with a unique constraint.
//
// data: A map of column-value pairs to insert/update.
//
//	[]map[string]interface{}{
//	    {"id": 1, "price_buy": 400, "quantity": 10},
//	    {"id": 2, "price_buy": 500, "quantity": 20},
//	}
func (r *UtilRepository) Upsert(tx *gorm.DB, table string, idColumn string, data []map[string]interface{}, span opentracing.Span) error {
	// Start a child span for tracing
	childSpan := opentracing.StartSpan("BulkUpdate", opentracing.ChildOf(span.Context()))

	if len(data) == 0 {
		return nil // No data to upsert
	}

	// Get the columns from the first row
	columns := make([]string, 0, len(data[0]))
	for column := range data[0] {
		columns = append(columns, column)
	}

	// Prepare the VALUES clause and arguments
	var valueStrings []string
	var valueArgs []interface{}
	for _, row := range data {
		var placeholders []string
		for _, column := range columns {
			placeholders = append(placeholders, fmt.Sprintf("$%d", len(valueArgs)+1))
			valueArgs = append(valueArgs, row[column])
		}
		valueStrings = append(valueStrings, fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")))
	}

	// Prepare the UPDATE clause
	var updates []string
	for _, column := range columns {
		if column != idColumn { // Skip the conflict column in the UPDATE clause
			updates = append(updates, fmt.Sprintf("%s = EXCLUDED.%s", column, column))
		}
	}

	// Construct the query
	query := fmt.Sprintf(`
        INSERT INTO %s (%s)
        VALUES %s
        ON CONFLICT (%s)
        DO UPDATE SET %s
    `,
		table,
		strings.Join(columns, ", "),
		strings.Join(valueStrings, ", "),
		idColumn,
		strings.Join(updates, ", "),
	)

	// Execute the query
	result := tx.Exec(query, valueArgs...)
	if result.Error != nil {
		defer childSpan.Finish()
		childSpan.LogKV("rows_affected", result.RowsAffected)
		return result.Error
	}

	return nil
}
