package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type BranchRepository struct {
	db     *gorm.DB
	sqlDB  *sqlx.DB
	tracer opentracing.Tracer
}

func NewBranchRepository(db *gorm.DB, sqlDB *sqlx.DB, tracer opentracing.Tracer) *BranchRepository {
	return &BranchRepository{
		db:     db,
		sqlDB:  sqlDB,
		tracer: tracer,
	}
}

func (r *BranchRepository) GetBranches(ctx context.Context, filters map[string]string, span opentracing.Span) ([]dtos.BranchListDTO, int, error) {
	childSpan := opentracing.StartSpan("BranchRepository-GetBranches")

	branches := []dtos.BranchListDTO{}
	var total int

	query := `SELECT *
    FROM ( 
        SELECT 
					b.id, 
					b.company_profile_id,
					b.parent_id,
					b.owner_name,
					b.sign_name,
					b.name,
					b.address,
					b.phone,
					b.email,
					b.website,
					b.logo,
					b.sign,
					b.description,
					b.remark,
					b.status,
					b.created_at,
					b.updated_at,
					b.deleted_at,

					c.company_name,

					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM branches b
				LEFT JOIN company_profiles c ON b.company_profile_id = c.id
        LEFT JOIN users cu ON b.created_by_id = cu.id
        LEFT JOIN users uu ON b.updated_by_id = uu.id
				WHERE b.deleted_at IS NULL
				ORDER BY b.name DESC
    ) AS alias WHERE 1=1`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT 
					b.id, 
					b.company_profile_id,
					b.parent_id,
					b.owner_name,
					b.sign_name,
					b.name,
					b.address,
					b.phone,
					b.email,
					b.website,
					b.logo,
					b.sign,
					b.description,
					b.remark,
					b.status,

					c.company_name,

					b.created_at,
					b.updated_at,
					b.deleted_at,
					cu.name as created_by_name,
					uu.name as updated_by_name

        FROM branches b
				LEFT JOIN company_profiles c ON b.company_profile_id = c.id
        LEFT JOIN users cu ON b.created_by_id = cu.id
        LEFT JOIN users uu ON b.updated_by_id = uu.id
				WHERE b.deleted_at IS NULL
    ) AS alias WHERE 1=1`

	var args []interface{}

	i := 1
	for key, value := range filters {
		switch key {
		case "name", "description", "remark":
			if value != "" {
				query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				countQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

	if value, ok := filters["global"]; ok && value != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d OR remark ILIKE $%d)", i, i+1, i+2)
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d OR remark ILIKE $%d)", i, i+1, i+2)
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

			err := r.sqlDB.GetContext(ctx, &total, countQuery, countArgs...)
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
		selectSpan := opentracing.StartSpan("SelectQuery", opentracing.ChildOf(childSpan.Context()))

		err := r.sqlDB.SelectContext(ctx, &branches, query, args...)
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

	if len(branches) > 0 {
		for i := range branches {
			if branches[i].Logo != nil {
				logo := utils.AddHostURLToImageURL(*branches[i].Logo)
				branches[i].Logo = &logo
			}
			if branches[i].Sign != nil {
				sign := utils.AddHostURLToImageURL(*branches[i].Sign)
				branches[i].Sign = &sign
			}
		}
	}

	return branches, total, nil
}

func (r *BranchRepository) GetBranchByID(ctx context.Context, params *dtos.GetBranchParams, span opentracing.Span) (*dtos.BranchDetailDTO, error) {
	childSpan := r.tracer.StartSpan("BranchRepository-GetBranchByID", opentracing.ChildOf(span.Context()))
	var Branch dtos.BranchDetailDTO

	query := `SELECT 
					b.id, 
					b.company_profile_id,
					b.parent_id,
					b.owner_name,
					b.sign_name,
					b.name,
					b.address,
					b.phone,
					b.email,
					b.website,
					b.logo,
					b.sign,
					b.description,
					b.remark,
					b.status,
					b.created_at,
					b.updated_at,
					b.deleted_at,
					c.company_name,
	cu.name as created_by_name,
	uu.name as updated_by_name

	FROM branches b
	LEFT JOIN company_profiles c ON b.company_profile_id = c.id
	LEFT JOIN users cu ON b.created_by_id = cu.id
	LEFT JOIN users uu ON b.updated_by_id = uu.id
	WHERE 1=1`

	var args []interface{}

	i := 1
	query += " AND b.id = $1"
	args = append(args, params.ID)
	i++

	isDeletedQuery := ` AND b.deleted_at IS NULL`
	if params.IsDeleted != nil && *params.IsDeleted == 1 {
		isDeletedQuery = " AND b.deleted_at IS NOT NULL"
	}

	query += isDeletedQuery

	if err := r.sqlDB.Get(&Branch, query, args...); err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	if Branch.Logo != nil {
		logo := utils.AddHostURLToImageURL(*Branch.Logo)
		Branch.Logo = &logo
	}
	if Branch.Sign != nil {
		sign := utils.AddHostURLToImageURL(*Branch.Sign)
		Branch.Sign = &sign
	}

	return &Branch, nil
}

// BeginTransaction starts a new transaction
func (r *BranchRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *BranchRepository) CreateBranch(tx *gorm.DB, Branch *models.Branch, span opentracing.Span) error {
	childSpan := r.tracer.StartSpan("BranchRepository-CreateBranch", opentracing.ChildOf(span.Context()))
	if err := tx.Create(Branch).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *BranchRepository) UpdateBranch(tx *gorm.DB, Branch *models.Branch, span opentracing.Span) error {
	childSpan := r.tracer.StartSpan("BranchRepository-UpdateBranch", opentracing.ChildOf(span.Context()))
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Branch{}).Where("id = ?", Branch.ID).Select("*").Omit("created_at", "created_by_id").Updates(Branch).Error; err != nil {
			defer childSpan.Finish()
			utils.LogErrors(childSpan, err)
			return err
		}
		return nil
	})
}

func (r *BranchRepository) DeleteBranch(tx *gorm.DB, params *dtos.GetBranchParams, span opentracing.Span) error {
	childSpan := r.tracer.StartSpan("BranchRepository-DeleteBranch", opentracing.ChildOf(span.Context()))
	return r.db.Transaction(func(tx *gorm.DB) error {
		// if err := tx.Unscoped().Delete(&models.Branch{}, id).Error; err != nil {
		if err := tx.Delete(&models.Branch{}, params.ID).Error; err != nil {
			defer childSpan.Finish()
			utils.LogErrors(childSpan, err)
			return err
		}
		return nil
	})
}

func (s *BranchRepository) RestoreBranch(tx *gorm.DB, params *dtos.GetBranchParams, span opentracing.Span) error {
	childSpan := s.tracer.StartSpan("BranchRepository-RestoreBranch", opentracing.ChildOf(span.Context()))
	return s.db.Transaction(func(tx *gorm.DB) error {
		var Branch models.Branch
		if err := tx.Unscoped().Model(&Branch).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
			defer childSpan.Finish()
			utils.LogErrors(childSpan, err)
			return err
		}
		return nil
	})
}
