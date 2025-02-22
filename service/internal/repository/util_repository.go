package repository

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
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
