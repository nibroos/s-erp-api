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

type CompanyProfileRepository struct {
	db     *gorm.DB
	sqlDB  *sqlx.DB
	tracer opentracing.Tracer
}

func NewCompanyProfileRepository(db *gorm.DB, sqlDB *sqlx.DB, tracer opentracing.Tracer) *CompanyProfileRepository {
	return &CompanyProfileRepository{
		db:     db,
		sqlDB:  sqlDB,
		tracer: tracer,
	}
}

func (r *CompanyProfileRepository) GetCompanyProfiles(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.CompanyProfileListDTO, int, error) {
	childSpan := opentracing.StartSpan("CompanyProfileRepository-GetCompanyProfiles")

	companyProfiles := []dtos.CompanyProfileListDTO{}
	var total int

	query := `SELECT *
    FROM ( 
        SELECT 
					cp.id, 
					cp.is_primary,
					cp.parent_id,
					cp.vat_id,
					cp.pph23_id,
					cp.company_owner_name,
					cp.company_sign_name,
					cp.company_name,
					cp.company_address,
					cp.company_phone,
					cp.company_email,
					cp.company_website,
					cp.company_logo,
					cp.company_sign,
					cp.company_description,
					cp.company_remark,
					cp.company_status,
					cp.created_at,
					cp.updated_at,
					cp.deleted_at,
					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM company_profiles cp
        LEFT JOIN users cu ON cp.created_by_id = cu.id
        LEFT JOIN users uu ON cp.updated_by_id = uu.id
				WHERE cp.deleted_at IS NULL
				ORDER BY cp.is_primary DESC
    ) AS alias WHERE 1=1`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT 
					cp.id, 
					cp.is_primary,
					cp.parent_id,
					cp.company_owner_name,
					cp.company_sign_name,
					cp.company_name,
					cp.company_address,
					cp.company_phone,
					cp.company_email,
					cp.company_website,
					cp.company_logo,
					cp.company_sign,
					cp.company_description,
					cp.company_remark,
					cp.company_status,
					cp.created_at,
					cp.updated_at,
					cp.deleted_at,
					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM company_profiles cp
        LEFT JOIN users cu ON cp.created_by_id = cu.id
        LEFT JOIN users uu ON cp.updated_by_id = uu.id
				WHERE cp.deleted_at IS NULL
    ) AS alias WHERE 1=1`

	var args []interface{}

	i := 1
	for key, value := range filters {
		switch key {
		case "company_name", "company_description", "company_remark":
			if value != "" {
				query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				countQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

	if filters["ids"] != "" {
		query += fmt.Sprintf(" AND id IN (%s)", filters["ids"])
	}
	if value, ok := filters["global"]; ok && value != "" {
		query += fmt.Sprintf(" AND (company_name ILIKE $%d OR company_description ILIKE $%d OR company_remark ILIKE $%d)", i, i+1, i+2)
		countQuery += fmt.Sprintf(" AND (company_name ILIKE $%d OR company_description ILIKE $%d OR company_remark ILIKE $%d)", i, i+1, i+2)
		args = append(args, "%"+value+"%", "%"+value+"%", "%"+value+"%")
		i += 3
	}

	countArgs := append([]interface{}{}, args...)

	var wg sync.WaitGroup
	var countErr, selectErr error

	// Goroutine for count query
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "company_name")
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
		selectSpan := opentracing.StartSpan("SelectQuery", opentracing.ChildOf(childSpan.Context()))

		err := r.sqlDB.SelectContext(ctx.Context(), &companyProfiles, query, args...)
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

	if len(companyProfiles) > 0 {
		for i := range companyProfiles {
			if companyProfiles[i].CompanyLogo != nil {
				logo := utils.AddHostURLToImageURL(*companyProfiles[i].CompanyLogo)
				companyProfiles[i].CompanyLogo = &logo
			}
			if companyProfiles[i].CompanySign != nil {
				sign := utils.AddHostURLToImageURL(*companyProfiles[i].CompanySign)
				companyProfiles[i].CompanySign = &sign
			}
		}
	}

	return companyProfiles, total, nil
}

func (r *CompanyProfileRepository) GetCompanyProfileByID(ctx *fiber.Ctx, params *dtos.GetCompanyProfileParams, span opentracing.Span) (*dtos.CompanyProfileDetailDTO, error) {
	childSpan := r.tracer.StartSpan("CompanyProfileRepository-GetCompanyProfileByID", opentracing.ChildOf(span.Context()))
	var CompanyProfile dtos.CompanyProfileDetailDTO

	query := `SELECT 
					cp.id, 
					cp.is_primary,
					cp.parent_id,
					cp.vat_id,
					cp.pph23_id,
					cp.company_owner_name,
					cp.company_sign_name,
					cp.company_name,
					cp.company_address,
					cp.company_phone,
					cp.company_email,
					cp.company_website,
					cp.company_logo,
					cp.company_sign,
					cp.company_description,
					cp.company_remark,
					cp.company_status,
					cp.created_at,
					cp.updated_at,
					cp.deleted_at,
	cu.name as created_by_name,
	uu.name as updated_by_name

	FROM company_profiles cp
	LEFT JOIN users cu ON cp.created_by_id = cu.id
	LEFT JOIN users uu ON cp.updated_by_id = uu.id
	WHERE 1=1`

	var args []interface{}

	if params.IsPrimary == nil {
		i := 1
		query += " AND cp.id = $1"
		args = append(args, params.ID)
		i++
	}

	isDeletedQuery := ` AND cp.deleted_at IS NULL`
	if params.IsDeleted != nil && *params.IsDeleted == 1 {
		isDeletedQuery = " AND cp.deleted_at IS NOT NULL"
	}

	if params.IsPrimary != nil && *params.IsPrimary == 1 {
		query += " AND cp.is_primary = 1"
	}

	query += isDeletedQuery

	if err := r.sqlDB.Get(&CompanyProfile, query, args...); err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	if CompanyProfile.CompanyLogo != nil {
		logo := utils.AddHostURLToImageURL(*CompanyProfile.CompanyLogo)
		CompanyProfile.CompanyLogo = &logo
	}
	if CompanyProfile.CompanySign != nil {
		sign := utils.AddHostURLToImageURL(*CompanyProfile.CompanySign)
		CompanyProfile.CompanySign = &sign
	}

	return &CompanyProfile, nil
}

// BeginTransaction starts a new transaction
func (r *CompanyProfileRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *CompanyProfileRepository) CreateCompanyProfile(tx *gorm.DB, CompanyProfile *models.CompanyProfile, span opentracing.Span) error {
	childSpan := r.tracer.StartSpan("CompanyProfileRepository-CreateCompanyProfile", opentracing.ChildOf(span.Context()))
	if err := tx.Create(CompanyProfile).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *CompanyProfileRepository) UpdateCompanyProfile(tx *gorm.DB, CompanyProfile *models.CompanyProfile, span opentracing.Span) error {
	childSpan := r.tracer.StartSpan("CompanyProfileRepository-UpdateCompanyProfile", opentracing.ChildOf(span.Context()))

	if err := tx.Model(&models.CompanyProfile{}).Where("id = ?", CompanyProfile.ID).Select("*").Omit("created_at", "created_by_id").Updates(CompanyProfile).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil

}

func (r *CompanyProfileRepository) DeleteCompanyProfile(tx *gorm.DB, params *dtos.GetCompanyProfileParams, span opentracing.Span) error {
	childSpan := r.tracer.StartSpan("CompanyProfileRepository-DeleteCompanyProfile", opentracing.ChildOf(span.Context()))

	// if err := tx.Unscoped().Delete(&models.CompanyProfile{}, id).Error; err != nil {
	if err := tx.Delete(&models.CompanyProfile{}, params.ID).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil

}

func (s *CompanyProfileRepository) RestoreCompanyProfile(tx *gorm.DB, params *dtos.GetCompanyProfileParams, span opentracing.Span) error {
	childSpan := s.tracer.StartSpan("CompanyProfileRepository-RestoreCompanyProfile", opentracing.ChildOf(span.Context()))

	var CompanyProfile models.CompanyProfile
	if err := tx.Unscoped().Model(&CompanyProfile).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil

}
