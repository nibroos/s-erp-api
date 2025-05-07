package repository

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

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
					cp.company_city,
					cp.company_province,
					cp.company_district,
					cp.company_postal_code,
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
					cp.company_city,
					cp.company_province,
					cp.company_district,
					cp.company_postal_code,
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

func (r *CompanyProfileRepository) GetCompanyProfileByID(ctx *fiber.Ctx, params *dtos.GetCompanyProfileParams, span opentracing.Span) (*dtos.CompanyProfileWithBanksDTO, error) {
	childSpan := r.tracer.StartSpan("CompanyProfileRepository-GetCompanyProfileByID", opentracing.ChildOf(span.Context()))
	var companyProfile dtos.CompanyProfileDetailDTO

	query := `SELECT 
					cp.id, 
					cp.is_primary,
					cp.parent_id,
					cp.vat_id,
					cp.pph23_id,
					cp.company_owner_name,
					cp.company_sign_name,
					cp.company_name,
					cp.company_city,
					cp.company_province,
					cp.company_district,
					cp.company_postal_code,
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

	if err := r.sqlDB.Get(&companyProfile, query, args...); err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	if companyProfile.CompanyLogo != nil {
		logo := utils.AddHostURLToImageURL(*companyProfile.CompanyLogo)
		companyProfile.CompanyLogo = &logo
	}
	if companyProfile.CompanySign != nil {
		sign := utils.AddHostURLToImageURL(*companyProfile.CompanySign)
		companyProfile.CompanySign = &sign
	}

	bankQuery := `SELECT 
					bi.id, 
					bi.commpany_profile_id,
					bi.name,
					bi.account_number,
					bi.account_name,
					bi.description,
					bi.created_at,
					bi.updated_at,
					bi.deleted_at,
					cp.company_name as company_name,
					cu.name as created_by_name,
					uu.name as updated_by_name
				FROM bank_informations bi
				LEFT JOIN company_profiles cp ON bi.commpany_profile_id = cp.id
				LEFT JOIN users cu ON bi.created_by_id = cu.id
				LEFT JOIN users uu ON bi.updated_by_id = uu.id
				WHERE bi.commpany_profile_id = $1 AND bi.deleted_at IS NULL`

	var bankInformations []dtos.BankInformationListDTO
	if err := r.sqlDB.Select(&bankInformations, bankQuery, companyProfile.ID); err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	result := &dtos.CompanyProfileWithBanksDTO{
		CompanyProfileDetailDTO: companyProfile,
		BankInformations:        bankInformations,
	}

	return result, nil
}

func (r *CompanyProfileRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *CompanyProfileRepository) CreateCompanyProfile(tx *gorm.DB, companyProfile *models.CompanyProfile, bankInformations []*models.BankInformation, span opentracing.Span) error {
	childSpan := r.tracer.StartSpan("CompanyProfileRepository-CreateCompanyProfile", opentracing.ChildOf(span.Context()))

	if err := tx.Create(companyProfile).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}

	if len(bankInformations) > 0 {
		for _, bank := range bankInformations {
			bank.CommpanyProfileID = &companyProfile.ID
			bank.CreatedByID = &companyProfile.CreatedByID

			if err := tx.Create(bank).Error; err != nil {
				defer childSpan.Finish()
				utils.LogErrors(childSpan, err)
				return err
			}
		}
	}

	return nil
}

func (r *CompanyProfileRepository) UpdateCompanyProfile(tx *gorm.DB, companyProfile *models.CompanyProfile, bankInformations []*models.BankInformation, span opentracing.Span) error {
	childSpan := r.tracer.StartSpan("CompanyProfileRepository-UpdateCompanyProfile", opentracing.ChildOf(span.Context()))

	if err := tx.Model(&models.CompanyProfile{}).Where("id = ?", companyProfile.ID).Select("*").Omit("created_at", "created_by_id").Updates(companyProfile).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}

	var existingBankIDs []uint
	if err := tx.Model(&models.BankInformation{}).Where("commpany_profile_id = ? AND deleted_at IS NULL", companyProfile.ID).Pluck("id", &existingBankIDs).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}

	updatedBankIDs := make(map[uint]bool)

	if len(bankInformations) > 0 {
		for _, bank := range bankInformations {
			if bank.ID > 0 {
				updatedBankIDs[bank.ID] = true
				if err := tx.Model(&models.BankInformation{}).Where("id = ?", bank.ID).Select("*").Omit("created_at", "created_by_id").Updates(bank).Error; err != nil {
					defer childSpan.Finish()
					utils.LogErrors(childSpan, err)
					return err
				}
			} else {
				bank.CommpanyProfileID = &companyProfile.ID
				bank.CreatedByID = &companyProfile.UpdatedByID
				bank.UpdatedByID = &companyProfile.UpdatedByID

				if err := tx.Create(bank).Error; err != nil {
					defer childSpan.Finish()
					utils.LogErrors(childSpan, err)
					return err
				}
				updatedBankIDs[bank.ID] = true
			}
		}
	}

	for _, existingID := range existingBankIDs {
		if !updatedBankIDs[existingID] {
			if err := tx.Delete(&models.BankInformation{}, existingID).Error; err != nil {
				defer childSpan.Finish()
				utils.LogErrors(childSpan, err)
				return err
			}
		}
	}

	return nil
}

func (r *CompanyProfileRepository) DeleteCompanyProfile(tx *gorm.DB, params *dtos.GetCompanyProfileParams, span opentracing.Span) error {
	childSpan := r.tracer.StartSpan("CompanyProfileRepository-DeleteCompanyProfile", opentracing.ChildOf(span.Context()))

	if err := tx.Model(&models.BankInformation{}).Where("commpany_profile_id = ?", params.ID).Update("deleted_at", time.Now()).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}

	if err := tx.Delete(&models.CompanyProfile{}, params.ID).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (s *CompanyProfileRepository) RestoreCompanyProfile(tx *gorm.DB, params *dtos.GetCompanyProfileParams, span opentracing.Span) error {
	childSpan := s.tracer.StartSpan("CompanyProfileRepository-RestoreCompanyProfile", opentracing.ChildOf(span.Context()))

	var companyProfile models.CompanyProfile
	if err := tx.Unscoped().Model(&companyProfile).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}

	if err := tx.Unscoped().Model(&models.BankInformation{}).Where("commpany_profile_id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}

	return nil
}

func (r *CompanyProfileRepository) GetBankInformations(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.BankInformationListDTO, int, error) {
	childSpan := opentracing.StartSpan("CompanyProfileRepository-GetBankInformations")

	bankInformations := []dtos.BankInformationListDTO{}
	var total int

	// condition := ""
	var args []interface{}
	i := 1

	query := `SELECT *
    FROM ( 
        SELECT 
			bi.id, 
			bi.commpany_profile_id,
			bi.name,
			bi.account_number,
			bi.account_name,
			bi.description,
			bi.created_at,
			bi.updated_at,
			bi.deleted_at,
			cp.company_name as company_name,
			cu.name as created_by_name,
			uu.name as updated_by_name

        FROM bank_informations bi
        LEFT JOIN company_profiles cp ON bi.commpany_profile_id = cp.id
        LEFT JOIN users cu ON bi.created_by_id = cu.id
        LEFT JOIN users uu ON bi.updated_by_id = uu.id
		WHERE bi.deleted_at IS NULL
    ) AS alias WHERE 1=1`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT 
			bi.id, 
			bi.commpany_profile_id,
			bi.name,
			bi.account_number,
			bi.account_name,
			bi.description,
			bi.created_at,
			bi.updated_at,
			bi.deleted_at,
			cp.company_name as company_name,
			cu.name as created_by_name,
			uu.name as updated_by_name

        FROM bank_informations bi
        LEFT JOIN company_profiles cp ON bi.commpany_profile_id = cp.id
        LEFT JOIN users cu ON bi.created_by_id = cu.id
        LEFT JOIN users uu ON bi.updated_by_id = uu.id
		WHERE bi.deleted_at IS NULL
    ) AS alias WHERE 1=1`

	for key, value := range filters {
		switch key {
		case "name", "description":
			if value != "" {
				query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				countQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		case "commpany_profile_id":
			if value != "" {
				query += fmt.Sprintf(" AND commpany_profile_id = $%d", i)
				countQuery += fmt.Sprintf(" AND commpany_profile_id = $%d", i)
				args = append(args, value)
				i++
			}
		}
	}

	if value, ok := filters["global"]; ok && value != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR account_number ILIKE $%d OR account_name ILIKE $%d)", i, i+1, i+2)
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR account_number ILIKE $%d OR account_name ILIKE $%d)", i, i+1, i+2)
		args = append(args, "%"+value+"%", "%"+value+"%", "%"+value+"%")
		i += 3
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

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "name")
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

		err := r.sqlDB.SelectContext(ctx.Context(), &bankInformations, query, args...)
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

	return bankInformations, total, nil
}

func (r *CompanyProfileRepository) GetBankInformationByID(ctx *fiber.Ctx, params *dtos.GetBankInformationParams, span opentracing.Span) (*dtos.BankInformationDetailDTO, error) {
	childSpan := r.tracer.StartSpan("CompanyProfileRepository-GetBankInformationByID", opentracing.ChildOf(span.Context()))
	var bankInformation dtos.BankInformationDetailDTO

	query := `SELECT 
			bi.id, 
			bi.commpany_profile_id,
			bi.name,
			bi.account_number,
			bi.account_name,
			bi.description,
			bi.created_at,
			bi.updated_at,
			bi.deleted_at,
			cp.company_name as company_name,
			cu.name as created_by_name,
			uu.name as updated_by_name

		FROM bank_informations bi
		LEFT JOIN company_profiles cp ON bi.commpany_profile_id = cp.id
		LEFT JOIN users cu ON bi.created_by_id = cu.id
		LEFT JOIN users uu ON bi.updated_by_id = uu.id
		WHERE bi.id = $1`

	isDeletedQuery := ` AND bi.deleted_at IS NULL`
	if params.IsDeleted != nil && *params.IsDeleted == 1 {
		isDeletedQuery = " AND bi.deleted_at IS NOT NULL"
	}

	query += isDeletedQuery

	if err := r.sqlDB.Get(&bankInformation, query, params.ID); err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return &bankInformation, nil
}

func (r *CompanyProfileRepository) CreateBankInformation(tx *gorm.DB, bankInformation *models.BankInformation, span opentracing.Span) error {
	childSpan := r.tracer.StartSpan("CompanyProfileRepository-CreateBankInformation", opentracing.ChildOf(span.Context()))
	if err := tx.Create(bankInformation).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *CompanyProfileRepository) UpdateBankInformation(tx *gorm.DB, bankInformation *models.BankInformation, span opentracing.Span) error {
	childSpan := r.tracer.StartSpan("CompanyProfileRepository-UpdateBankInformation", opentracing.ChildOf(span.Context()))

	if err := tx.Model(&models.BankInformation{}).Where("id = ?", bankInformation.ID).Select("*").Omit("created_at", "created_by_id").Updates(bankInformation).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *CompanyProfileRepository) DeleteBankInformation(tx *gorm.DB, id uint, span opentracing.Span) error {
	childSpan := r.tracer.StartSpan("CompanyProfileRepository-DeleteBankInformation", opentracing.ChildOf(span.Context()))

	if err := tx.Delete(&models.BankInformation{}, id).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *CompanyProfileRepository) RestoreBankInformation(tx *gorm.DB, id uint, span opentracing.Span) error {
	childSpan := r.tracer.StartSpan("CompanyProfileRepository-RestoreBankInformation", opentracing.ChildOf(span.Context()))

	var bankInformation models.BankInformation
	if err := tx.Unscoped().Model(&bankInformation).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *CompanyProfileRepository) GetBankInformationsWithCompany(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.BankInformationWithCompanyDTO, int, error) {
	childSpan := r.tracer.StartSpan("CompanyProfileRepository-GetBankInformationsWithCompany", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var bankInformations []dtos.BankInformationWithCompanyDTO
	var total int

	query := `
        SELECT 
            bi.id, 
            bi.commpany_profile_id,
            cp.id as company_id,
            cp.company_name, 
            cp.company_email,
            cp.company_phone,
            cp.company_address,
            bi.name, 
            bi.account_number, 
            bi.account_name, 
            bi.description,
            cb.name as created_by_name,
            ub.name as updated_by_name,
            to_char(bi.created_at, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"') as created_at,
            to_char(bi.updated_at, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"') as updated_at,
            to_char(bi.deleted_at, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"') as deleted_at
        FROM 
            bank_informations bi
        LEFT JOIN 
            company_profiles cp ON bi.commpany_profile_id = cp.id
        LEFT JOIN 
            users cb ON bi.created_by_id = cb.id
        LEFT JOIN 
            users ub ON bi.updated_by_id = ub.id
        WHERE 
            bi.deleted_at IS NULL AND cp.id = 1
            AND (bi.name != '' OR bi.account_number != '' OR bi.account_name != '')
    `

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filters["global"] != "" {
		conditions = append(conditions, fmt.Sprintf("(bi.name ILIKE $%d OR bi.account_number ILIKE $%d OR bi.account_name ILIKE $%d OR cp.company_name ILIKE $%d)", argIndex, argIndex, argIndex, argIndex))
		args = append(args, "%"+filters["global"]+"%")
		argIndex++
	}

	if filters["name"] != "" {
		conditions = append(conditions, fmt.Sprintf("bi.name ILIKE $%d", argIndex))
		args = append(args, "%"+filters["name"]+"%")
		argIndex++
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) as count_query", query)
	err := r.sqlDB.Get(&total, countQuery, args...)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, 0, err
	}

	perPage, err := strconv.Atoi(filters["per_page"])
	if err != nil {
		perPage = 10
	}

	page, err := strconv.Atoi(filters["page"])
	if err != nil {
		page = 1
	}

	offset := (page - 1) * perPage

	orderColumn := filters["order_column"]
	if orderColumn == "" {
		orderColumn = "id"
	}

	orderDirection := filters["order_direction"]
	if orderDirection == "" {
		orderDirection = "asc"
	}

	query += fmt.Sprintf(" ORDER BY bi.%s %s LIMIT $%d OFFSET $%d", orderColumn, orderDirection, argIndex, argIndex+1)
	args = append(args, perPage, offset)

	err = r.sqlDB.Select(&bankInformations, query, args...)
	if err != nil {
		utils.LogErrors(childSpan, err)
		return nil, 0, err
	}

	return bankInformations, total, nil
}
