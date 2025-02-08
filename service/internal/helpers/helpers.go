package helpers

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/repository"
)

// type CompanyProfileRepository struct {
// 	db    *gorm.DB
// 	sqlDB *sqlx.DB
// }

// func NewCompanyProfileRepository(db *gorm.DB, sqlDB *sqlx.DB) *CompanyProfileRepository {
// 	return &CompanyProfileRepository{
// 		db:    db,
// 		sqlDB: sqlDB,
// 	}
// }

// func (r *CompanyProfileRepository) GetCompanyProfileByID(ctx context.Context, params *dtos.GetCompanyProfileParams) (*dtos.CompanyProfileDetailDTO, error) {
// 	var CompanyProfile dtos.CompanyProfileDetailDTO

// 	query := `SELECT cp.id, cp.company_name, cp.company_address, cp.company_phone, cp.company_email, cp.company_website, cp.company_logo, cp.company_description, cp.company_remark, cp.company_status, cp.created_at, cp.updated_at, cp.deleted_at,
// 	cu.name as created_by_name,
// 	uu.name as updated_by_name

// 	FROM company_profiles cp
// 	LEFT JOIN users cu ON cp.created_by_id = cu.id
// 	LEFT JOIN users uu ON cp.updated_by_id = uu.id
// 	WHERE 1=1`

// 	var args []interface{}

// 	i := 1
// 	query += " AND cp.id = $1"
// 	args = append(args, params.ID)
// 	i++

// 	isDeletedQuery := ` AND cp.deleted_at IS NULL`
// 	if params.IsDeleted != nil && *params.IsDeleted == 1 {
// 		isDeletedQuery = " AND cp.deleted_at IS NOT NULL"
// 	}

// 	query += isDeletedQuery

// 	err := r.sqlDB.Get(&CompanyProfile, query, args...)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &CompanyProfile, nil
// }

// GetCompanyProfileByID retrieves the company profile by ID
func GetCompanyProfileByID(ctx context.Context, repo *repository.CompanyProfileRepository, params *dtos.GetCompanyProfileParams) (*dtos.CompanyProfileDetailDTO, error) {
	return repo.GetCompanyProfileByID(ctx, params)
}
