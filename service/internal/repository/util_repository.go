package repository

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	var companyProfile dtos.CompanyProfileDetailDTO

	query := `SELECT cp.id, cp.company_name, cp.company_address, cp.company_phone, cp.company_email, cp.company_website, cp.company_logo, cp.company_description, cp.company_remark, cp.company_status, cp.created_at, cp.updated_at, cp.deleted_at,
	cp.company_owner_name,
	cp.company_email_password,
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

	if err := r.sqlDB.Get(&companyProfile, query, args...); err != nil {
		return nil, err
	}

	if companyProfile.CompanyLogo != nil {
		logoUrl := utils.AddHostURLToImageURL(*companyProfile.CompanyLogo)
		companyProfile.CompanyLogoUrl = &logoUrl

	}
	if companyProfile.CompanySign != nil {
		signUrl := utils.AddHostURLToImageURL(*companyProfile.CompanySign)
		companyProfile.CompanySignUrl = &signUrl
	}

	return &companyProfile, nil
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
	childSpan := opentracing.StartSpan("utils-BulkUpdate", opentracing.ChildOf(span.Context()))

	if len(data) == 0 {
		return nil // No data to upsert
	}

	// Separate data into inserts and updates
	var insertsData []map[string]interface{}
	var updatesData []map[string]interface{}

	for _, row := range data {
		if id, ok := row[idColumn]; ok {
			// Check if ID is 0 (or 0 in different numeric types)
			isZero := false
			switch v := id.(type) {
			case int:
				isZero = v == 0
			case uint:
				isZero = v == 0
			case int64:
				isZero = v == 0
			case float64:
				isZero = v == 0
			}

			if isZero {
				// Remove ID for insert operations
				delete(row, idColumn)
				insertsData = append(insertsData, row)
			} else {
				updatesData = append(updatesData, row)
			}
		}
	}

	// Handle inserts
	if len(insertsData) > 0 {
		if err := r.handleInserts(tx, table, insertsData); err != nil {
			defer childSpan.Finish()
			childSpan.LogKV("error", err.Error())
			return err
		}
	}

	// Handle updates
	if len(updatesData) > 0 {
		if err := r.handleUpdates(tx, table, idColumn, updatesData); err != nil {
			defer childSpan.Finish()
			childSpan.LogKV("error", err.Error())
			return err
		}
	}

	return nil
}

func (r *UtilRepository) handleInserts(tx *gorm.DB, table string, data []map[string]interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// Get columns from the first row
	columns := make([]string, 0, len(data[0]))
	for column := range data[0] {
		columns = append(columns, column)
	}

	// Prepare VALUES clause and arguments
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

	// Construct insert query
	query := fmt.Sprintf(`
        INSERT INTO %s (%s)
        VALUES %s
    `,
		table,
		strings.Join(columns, ", "),
		strings.Join(valueStrings, ", "),
	)

	result := tx.Exec(query, valueArgs...)
	return result.Error
}

func (r *UtilRepository) handleUpdates(tx *gorm.DB, table string, idColumn string, data []map[string]interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// Get columns from the first row
	columns := make([]string, 0, len(data[0]))
	for column := range data[0] {
		columns = append(columns, column)
	}

	// Prepare VALUES and UPDATE clauses
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

	// Prepare UPDATE clause
	var updates []string
	for _, column := range columns {
		if column != idColumn {
			updates = append(updates, fmt.Sprintf("%s = EXCLUDED.%s", column, column))
		}
	}

	// Construct upsert query
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

	result := tx.Exec(query, valueArgs...)
	return result.Error
}

// s.utilRepo.LockProducts(ctx, tx, productIDs, childSpan); err != nil {
func (r *UtilRepository) LockRowTable(ctx *fiber.Ctx, tx *gorm.DB, ids []*uint, tableName string, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("LockRowTable", opentracing.ChildOf(span.Context()))
	if len(ids) == 0 {
		return tx, nil
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE id IN (?) FOR UPDATE", tableName)
	result := tx.Exec(query, ids)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		childSpan.LogKV("error", result.Error.Error())
		return tx, result.Error
	}

	return tx, nil
}

// UpdateAttachmentsDesc
func (r *UtilRepository) CreateSentEmails(ctx *fiber.Ctx, tx *gorm.DB, sentEmails []models.SentEmail, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("UtilRepository-CreateSentEmails", opentracing.ChildOf(span.Context()))

	// if err := r.Upsert(tx, "sent_emails", "id", sentEmails, childSpan); err != nil {
	// 	utils.LogErrors(childSpan, err)
	// 	return nil, err
	// }
	if err := tx.Create(&sentEmails).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *UtilRepository) UpdateSentEmail(tx *gorm.DB, sentEmail *models.SentEmail, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("UtilRepository-UpdateSentEmail", opentracing.ChildOf(span.Context()))
	if err := tx.Model(&models.SentEmail{}).Where("id = ?", sentEmail.ID).Select("*").Omit(
		"created_at", "created_by_id",
	).Updates(&sentEmail).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

func (r *UtilRepository) DeleteSentEmailsByRefIDs(ctx *fiber.Ctx, tx *gorm.DB, refIDs []uint, refType string, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("UtilRepository-DeleteSentEmailsByRefIDs", opentracing.ChildOf(span.Context()))
	if err := tx.Unscoped().Where("ref_id IN (?)", refIDs).Where("ref_type = ?", refType).Delete(&models.SentEmail{}).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}

// UpsertModel performs an upsert operation for any model using GORM
func (r *UtilRepository) UpsertModel(tx *gorm.DB, model interface{}, data interface{}, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("UtilRepository-UpsertModel", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	// Try to update first
	result := tx.Model(model).Updates(data)
	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return result.Error
	}

	// If no rows were affected, create new record
	if result.RowsAffected == 0 {
		result = tx.Create(data)
		if result.Error != nil {
			utils.LogErrors(childSpan, result.Error)
			return result.Error
		}
	}

	return nil
}

// BatchUpsertModels performs batch upsert operations for any model slice
func (r *UtilRepository) BatchUpsertModels(tx *gorm.DB, models interface{}, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("UtilRepository-BatchUpsertModels", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	result := tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(models)

	if result.Error != nil {
		utils.LogErrors(childSpan, result.Error)
		return result.Error
	}

	return nil
}
