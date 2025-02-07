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

type ItemGroupRepository struct {
	db    *gorm.DB
	sqlDB *sqlx.DB
}

func NewItemGroupRepository(db *gorm.DB, sqlDB *sqlx.DB) *ItemGroupRepository {
	return &ItemGroupRepository{
		db:    db,
		sqlDB: sqlDB,
	}
}

func (r *ItemGroupRepository) GetItemGroups(ctx context.Context, filters map[string]string) ([]dtos.ItemGroupListDTO, int, error) {
	modules := []dtos.ItemGroupListDTO{}
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
				WHERE g.name = 'item_groups'
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM mix_values m
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
				LEFT JOIN groups g ON m.group_id = g.id
				WHERE g.name = 'item_groups'
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
		err := r.sqlDB.GetContext(ctx, &total, countQuery, countArgs...)
		countChan <- err
	}()

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "name")
	orderDirection := utils.GetStringOrDefault(filters["order_direction"], "asc")
	query += fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)

	perPage := utils.GetIntOrDefault(filters["per_page"], 10)
	currentPage := utils.GetIntOrDefault(filters["page"], 1)

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, perPage, (currentPage-1)*perPage)

	// Goroutine for select query
	go func() {
		err := r.sqlDB.SelectContext(ctx, &modules, query, args...)
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

	return modules, total, nil
}

func (r *ItemGroupRepository) GetItemGroupByID(ctx context.Context, id uint) (*dtos.ItemGroupDetailDTO, error) {
	var itemGroup dtos.ItemGroupDetailDTO

	query := `SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
	cu.name as created_by_name,
	uu.name as updated_by_name

	FROM mix_values m
	LEFT JOIN users cu ON m.created_by_id = cu.id
	LEFT JOIN users uu ON m.updated_by_id = uu.id
	LEFT JOIN groups g ON m.group_id = g.id
	WHERE m.id = $1 AND g.name = 'item_groups' AND m.deleted_at IS NULL`

	if err := r.sqlDB.Get(&itemGroup, query, id); err != nil {
		return nil, err
	}

	return &itemGroup, nil
}

// BeginTransaction starts a new transaction
func (r *ItemGroupRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *ItemGroupRepository) CreateItemGroup(tx *gorm.DB, itemGroup *models.MixValue) error {
	if err := tx.Create(itemGroup).Error; err != nil {
		return err
	}
	return nil
}

func (r *ItemGroupRepository) UpdateItemGroup(tx *gorm.DB, itemGroup *models.MixValue) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(itemGroup).Error; err != nil {
			return err
		}
		return nil
	})

}

func (r *ItemGroupRepository) DeleteItemGroup(tx *gorm.DB, id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// if err := tx.Unscoped().Delete(&models.MixValue{}, id).Error; err != nil {
		if err := tx.Delete(&models.MixValue{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *ItemGroupRepository) RestoreItemGroup(tx *gorm.DB, id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var itemGroup models.MixValue
		if err := tx.Unscoped().First(&itemGroup, id).Error; err != nil {
			return err
		}
		return tx.Model(&itemGroup).Update("deleted_at", nil).Error
	})
}
