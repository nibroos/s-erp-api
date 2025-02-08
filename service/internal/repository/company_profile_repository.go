package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"gorm.io/gorm"
)

type CompanyProfileRepository struct {
	db    *gorm.DB
	sqlDB *sqlx.DB
}

func NewCompanyProfileRepository(db *gorm.DB, sqlDB *sqlx.DB) *CompanyProfileRepository {
	return &CompanyProfileRepository{
		db:    db,
		sqlDB: sqlDB,
	}
}

func (r *CompanyProfileRepository) GetCompanyProfiles(ctx context.Context, filters map[string]string) ([]dtos.CompanyProfileListDTO, int, error) {
	companyProfiles := []dtos.CompanyProfileListDTO{}
	var total int

	query := `SELECT *
    FROM ( 
        SELECT cp.id, cp.company_name, cp.company_address, cp.company_phone, cp.company_email, cp.company_website, cp.company_logo, cp.company_description, cp.company_remark, cp.company_status, cp.created_at, cp.updated_at, cp.deleted_at,
        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM company_profiles cp
        LEFT JOIN users cu ON cp.created_by_id = cu.id
        LEFT JOIN users uu ON cp.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT 
        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM company_profiles cp
        LEFT JOIN users cu ON cp.created_by_id = cu.id
        LEFT JOIN users uu ON cp.updated_by_id = uu.id
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

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

	if value, ok := filters["global"]; ok && value != "" {
		query += fmt.Sprintf(" AND (company_name ILIKE $%d OR company_description ILIKE $%d OR company_remark ILIKE $%d)", i, i+1, i+2)
		countQuery += fmt.Sprintf(" AND (company_name ILIKE $%d OR company_description ILIKE $%d OR company_remark ILIKE $%d)", i, i+1, i+2)
		args = append(args, "%"+value+"%", "%"+value+"%", "%"+value+"%")
		i += 3
	}

	countArgs := append([]interface{}{}, args...)

	// Channels for concurrent execution
	countChan := make(chan error)
	selectChan := make(chan error)

	// Goroutine for count query
	go func() {
		if filters["is_csv"] != "1" {
			err := r.sqlDB.GetContext(ctx, &total, countQuery, countArgs...)
			countChan <- err
		} else {
			countChan <- nil
		}
	}()

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
	go func() {
		err := r.sqlDB.SelectContext(ctx, &companyProfiles, query, args...)
		selectChan <- err
	}()

	// Wait for both goroutines to finish
	countErr := <-countChan
	selectErr := <-selectChan

	if countErr != nil {
		return nil, 0, countErr
	}

	if selectErr != nil {
		return nil, 0, selectErr
	}

	return companyProfiles, total, nil
}

func (r *CompanyProfileRepository) GetCompanyProfileByID(ctx context.Context, params *dtos.GetCompanyProfileParams) (*dtos.CompanyProfileDetailDTO, error) {
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

// BeginTransaction starts a new transaction
func (r *CompanyProfileRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *CompanyProfileRepository) CreateCompanyProfile(tx *gorm.DB, CompanyProfile *models.CompanyProfile) error {
	if err := tx.Create(CompanyProfile).Error; err != nil {
		return err
	}
	return nil
}

func (r *CompanyProfileRepository) UpdateCompanyProfile(tx *gorm.DB, CompanyProfile *models.CompanyProfile) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(CompanyProfile).Error; err != nil {
			return err
		}
		return nil
	})

}

func (r *CompanyProfileRepository) DeleteCompanyProfile(tx *gorm.DB, params *dtos.GetCompanyProfileParams) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// if err := tx.Unscoped().Delete(&models.CompanyProfile{}, id).Error; err != nil {
		if err := tx.Delete(&models.CompanyProfile{}, params.ID).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *CompanyProfileRepository) RestoreCompanyProfile(tx *gorm.DB, params *dtos.GetCompanyProfileParams) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var CompanyProfile models.CompanyProfile
		if err := tx.Unscoped().First(&CompanyProfile, params.ID).Error; err != nil {
			return err
		}
		return tx.Model(&CompanyProfile).Update("deleted_at", nil).Error
	})
}
