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

type UnitRepository struct {
	db    *gorm.DB
	sqlDB *sqlx.DB
}

func NewUnitRepository(db *gorm.DB, sqlDB *sqlx.DB) *UnitRepository {
	return &UnitRepository{
		db:    db,
		sqlDB: sqlDB,
	}
}

func (r *UnitRepository) GetUnits(ctx context.Context, filters map[string]string) ([]dtos.UnitListDTO, int, error) {
	units := []dtos.UnitListDTO{}
	var total int

	query := `SELECT *
    FROM ( 
        SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM mix_values m
				LEFT JOIN groups g ON m.group_id = g.id
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
				WHERE g.name = 'units'
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM mix_values m
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
				LEFT JOIN groups g ON m.group_id = g.id
				WHERE g.name = 'units'
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

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
		err := r.sqlDB.SelectContext(ctx, &units, query, args...)
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

	return units, total, nil
}

func (r *UnitRepository) GetUnitByID(ctx context.Context, params *dtos.GetUnitParams) (*dtos.UnitDetailDTO, error) {
	var unit dtos.UnitDetailDTO

	query := `SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
	cu.name as created_by_name,
	uu.name as updated_by_name

	FROM mix_values m
	LEFT JOIN users cu ON m.created_by_id = cu.id
	LEFT JOIN users uu ON m.updated_by_id = uu.id
	LEFT JOIN groups g ON m.group_id = g.id
	WHERE g.name = 'units'`

	var args []interface{}

	i := 1
	query += " AND m.id = $1"
	args = append(args, params.ID)
	i++

	isDeletedQuery := ` AND m.deleted_at IS NULL`
	if params.IsDeleted != nil && *params.IsDeleted == 1 {
		isDeletedQuery = " AND m.deleted_at IS NOT NULL"
	}

	query += isDeletedQuery

	if err := r.sqlDB.Get(&unit, query, args...); err != nil {
		return nil, err
	}

	return &unit, nil
}

// BeginTransaction starts a new transaction
func (r *UnitRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *UnitRepository) CreateUnit(tx *gorm.DB, unit *models.MixValue) error {
	if err := tx.Create(unit).Error; err != nil {
		return err
	}
	return nil
}

func (r *UnitRepository) UpdateUnit(tx *gorm.DB, unit *models.MixValue) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Updates(unit).Error; err != nil {
			return err
		}
		return nil
	})

}

func (r *UnitRepository) DeleteUnit(tx *gorm.DB, params *dtos.GetUnitParams) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.MixValue{}, params.ID).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *UnitRepository) RestoreUnit(tx *gorm.DB, params *dtos.GetUnitParams) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var unit models.MixValue
		if err := tx.Unscoped().Model(&unit).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		return nil
	})
}
