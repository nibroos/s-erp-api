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

// BulkUpdate updates multiple rows in a single query.
//
// idColumn: The name of the primary key column (e.g., "id").
//
// data: A slice of maps, where each map contains the primary key value and the new values for the columns to update.
//
// Example data:
//
//	[]map[string]interface{}{
//	    {"id": 1, "price_buy": 400, "quantity": 10},
//	    {"id": 2, "price_buy": 500, "quantity": 20},
//	}
func (r *UtilRepository) BulkUpdate(tx *gorm.DB, tableName string, idColumn string, data []map[string]interface{}, span opentracing.Span) error {
	if len(data) == 0 {
		return nil // No data to update
	}

	// Start a child span for tracing
	childSpan := opentracing.StartSpan("BulkUpdate", opentracing.ChildOf(span.Context()))

	// Step 1: Extract column names (excluding the primary key)
	columns := make([]string, 0)
	for key := range data[0] {
		if key != idColumn {
			columns = append(columns, key)
		}
	}

	// Step 2: Build the SET clause with CASE statements for each column
	var setBuilder strings.Builder
	var args []interface{}
	for _, column := range columns {
		setBuilder.WriteString(fmt.Sprintf("%s = CASE %s ", column, idColumn))

		// Add WHEN-THEN clauses for each row
		for _, item := range data {
			setBuilder.WriteString("WHEN ? THEN ? ")
			args = append(args, item[idColumn], item[column]) // Append the ID and column value
		}
		setBuilder.WriteString("END, ")
	}

	// Remove the trailing comma and space
	setClause := strings.TrimSuffix(setBuilder.String(), ", ")

	// Step 3: Build the WHERE clause
	var whereBuilder strings.Builder
	whereBuilder.WriteString(fmt.Sprintf("%s IN (", idColumn))
	placeholders := strings.Repeat("?, ", len(data)-1) + "?"
	whereBuilder.WriteString(placeholders)
	whereBuilder.WriteString(")")

	// Step 4: Combine the full query
	query := fmt.Sprintf(`
        UPDATE %s
        SET %s
        WHERE %s
    `, tableName, setClause, whereBuilder.String())

	// Step 5: Append IDs to the args for the WHERE clause
	for _, item := range data {
		args = append(args, item[idColumn])
	}

	// Step 6: Execute the query
	result := tx.Exec(query, args...)
	if result.Error != nil {
		defer childSpan.Finish()
		childSpan.LogKV("rows_affected", result.RowsAffected)
		return result.Error
	}

	return nil
}
