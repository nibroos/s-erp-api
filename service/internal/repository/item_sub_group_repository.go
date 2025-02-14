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

type ItemSubGroupRepository struct {
	db    *gorm.DB
	sqlDB *sqlx.DB
}

func NewItemSubGroupRepository(db *gorm.DB, sqlDB *sqlx.DB) *ItemSubGroupRepository {
	return &ItemSubGroupRepository{
		db:    db,
		sqlDB: sqlDB,
	}
}

func (r *ItemSubGroupRepository) GetItemSubGroups(ctx context.Context, filters map[string]string) ([]dtos.ItemSubGroupListDTO, int, error) {
	itemSubGroups := []dtos.ItemSubGroupListDTO{}
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
				WHERE g.name = 'item_sub_groups'
    ) AS alias WHERE 1=1 AND deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM (
        SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
        cu.name as created_by_name,
        uu.name as updated_by_name

        FROM mix_values m
        LEFT JOIN users cu ON m.created_by_id = cu.id
        LEFT JOIN users uu ON m.updated_by_id = uu.id
				LEFT JOIN groups g ON m.group_id = g.id
				WHERE g.name = 'item_sub_groups'
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
		err := r.sqlDB.SelectContext(ctx, &itemSubGroups, query, args...)
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

	return itemSubGroups, total, nil
}

func (r *ItemSubGroupRepository) GetItemSubGroupByID(ctx context.Context, params *dtos.GetItemSubGroupParams) (*dtos.ItemSubGroupDetailDTO, error) {
	var itemSubGroup dtos.ItemSubGroupDetailDTO

	query := `SELECT m.id, m.name, m.description, m.remark, m.status, m.created_at, m.updated_at, m.deleted_at,
	cu.name as created_by_name,
	uu.name as updated_by_name

	FROM mix_values m
	LEFT JOIN users cu ON m.created_by_id = cu.id
	LEFT JOIN users uu ON m.updated_by_id = uu.id
	LEFT JOIN groups g ON m.group_id = g.id
	WHERE g.name = 'item_sub_groups'`

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

	if err := r.sqlDB.Get(&itemSubGroup, query, args...); err != nil {
		return nil, err
	}

	return &itemSubGroup, nil
}

// BeginTransaction starts a new transaction
func (r *ItemSubGroupRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *ItemSubGroupRepository) CreateItemSubGroup(tx *gorm.DB, itemSubGroup *models.MixValue) error {
	if err := tx.Create(itemSubGroup).Error; err != nil {
		return err
	}
	return nil
}

func (r *ItemSubGroupRepository) UpdateItemSubGroup(tx *gorm.DB, itemSubGroup *models.MixValue) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("*").Omit("created_at", "created_by_id").Updates(itemSubGroup).Error; err != nil {
			return err
		}
		return nil
	})

}

func (r *ItemSubGroupRepository) DeleteItemSubGroup(tx *gorm.DB, id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.MixValue{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *ItemSubGroupRepository) RestoreItemSubGroup(tx *gorm.DB, id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var itemSubGroup models.MixValue
		if err := tx.Unscoped().Model(&itemSubGroup).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		return nil
	})
}
