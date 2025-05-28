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

type UserRepository struct {
	db       *gorm.DB
	sqlDB    *sqlx.DB
	utilRepo *UtilRepository
	tracer   opentracing.Tracer
}

func NewUserRepository(db *gorm.DB, sqlDB *sqlx.DB, utilRepo *UtilRepository, tracer opentracing.Tracer) *UserRepository {
	return &UserRepository{
		db:       db,
		sqlDB:    sqlDB,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (r *UserRepository) GetUsers(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.UserListDTO, int, error) {
	childSpan := opentracing.StartSpan("UserRepository-GetUsers", opentracing.ChildOf(span.Context()))
	users := []dtos.UserListDTO{}
	var total int

	query := `SELECT *
		FROM (
			SELECT
				u.id, u.username, u.name, u.email, u.branch_id, u.address, u.status,
				b.name as branch_name,
				p.mv2_id as role_id,
				mv.name as role_name
			FROM users u
			LEFT JOIN branches b ON u.branch_id = b.id
			LEFT JOIN pools p ON u.id = p.mv1_id
			LEFT JOIN mix_values mv ON p.mv2_id = mv.id
			LEFT JOIN groups g1 ON p.group1_id = g1.id
			LEFT JOIN groups g2 ON p.group2_id = g2.id
			WHERE g1.name = 'users' AND g2.name = 'roles'
		) AS alias WHERE 1=1`

	countQuery := `SELECT COUNT(*) FROM (
		SELECT
			u.id, u.username, u.name, u.email, u.branch_id, u.address, u.status,
			b.name as branch_name,
			p.mv2_id as role_id,
			mv.name as role_name
		FROM users u
		LEFT JOIN branches b ON u.branch_id = b.id
		LEFT JOIN pools p ON u.id = p.mv1_id
		LEFT JOIN mix_values mv ON p.mv2_id = mv.id
		LEFT JOIN groups g1 ON p.group1_id = g1.id
		LEFT JOIN groups g2 ON p.group2_id = g2.id
		WHERE g1.name = 'users' AND g2.name = 'roles'
	) AS alias WHERE 1=1`

	var args []interface{}

	i := 1
	for key, value := range filters {
		switch key {
		case "username", "name", "email":
			if value != "" {
				query += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				countQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, i)
				args = append(args, "%"+value+"%")
				i++
			}
		}
	}

	filterIDsKey := map[string]string{
		"role_ids": "role_id",
	}

	for key, valueID := range filterIDsKey {
		if value, ok := filters[key]; ok && value != "" {
			query += fmt.Sprintf(" AND %s IN (%s)", valueID, value)
			countQuery += fmt.Sprintf(" AND %s IN (%s)", valueID, value)
		}
	}

	if filters["ids"] != "" {
		query += fmt.Sprintf(" AND id IN (%s)", filters["ids"])
	}
	if value, ok := filters["global"]; ok && value != "" {
		query += fmt.Sprintf(" AND (username ILIKE $%d OR name ILIKE $%d OR email ILIKE $%d)", i, i+1, i+2)
		countQuery += fmt.Sprintf(" AND (username ILIKE $%d OR name ILIKE $%d OR email ILIKE $%d)", i, i+1, i+2)
		args = append(args, "%"+value+"%", "%"+value+"%", "%"+value+"%")
		i += 3
	}

	orderColumn := utils.GetStringOrDefault(filters["order_column"], "id")
	orderDirection := utils.GetStringOrDefault(filters["order_direction"], "asc")
	query += fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)

	perPage := utils.GetIntOrDefault(filters["per_page"], 10)
	currentPage := utils.GetIntOrDefault(filters["page"], 1)

	countArgs := append([]interface{}{}, args...)

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)

	args = append(args, perPage, (currentPage-1)*perPage)

	var wg sync.WaitGroup
	var selectErr error

	wg.Add(2)

	// Goroutine for count query
	go func() {
		defer wg.Done()
		selectSpan := opentracing.StartSpan("CountQuery", opentracing.ChildOf(childSpan.Context()))
		err := r.sqlDB.GetContext(ctx.Context(), &total, countQuery, countArgs...)
		if err != nil {
			selectSpan.LogKV("query", query)
			utils.LogErrors(selectSpan, err)
			selectErr = err
		}
	}()

	// Goroutine for select query
	go func() {
		defer wg.Done()
		selectSpan := opentracing.StartSpan("SelectQuery", opentracing.ChildOf(childSpan.Context()))
		err := r.sqlDB.SelectContext(ctx.Context(), &users, query, args...)
		if err != nil {
			selectSpan.LogKV("query", query)
			utils.LogErrors(selectSpan, err)
			selectErr = err
		}
	}()

	// Wait for both goroutines to finish
	wg.Wait()

	defer childSpan.Finish()

	if selectErr != nil {
		return nil, 0, selectErr
	}

	return users, total, nil
}

func (r *UserRepository) GetUserByID(ctx *fiber.Ctx, params *dtos.GetUserByIDParams, span opentracing.Span) (*dtos.UserDetailDTO, error) {
	childSpan := opentracing.StartSpan("UserRepository-GetUserByID", opentracing.ChildOf(span.Context()))
	var user dtos.UserDetailDTO

	query := `SELECT 
		u.branch_id, 
			u.id, u.username, u.name, u.email, u.address, u.status, u.password,
			b.name as branch_name
	FROM users u 
	LEFT JOIN branches b ON u.branch_id = b.id
	WHERE u.id = $1`

	var args []interface{}
	args = append(args, params.ID)

	isDeletedQuery := ` AND u.deleted_at IS NULL`
	if params.IsDeleted != nil && *params.IsDeleted == 1 {
		isDeletedQuery = " AND u.deleted_at IS NOT NULL"
	}

	query += isDeletedQuery

	var wg sync.WaitGroup
	var userErr, roleErr, permissionErr error

	// Goroutine for user query
	wg.Add(1)
	go func() {
		defer wg.Done()

		userSpan := opentracing.StartSpan("UserQuery", opentracing.ChildOf(childSpan.Context()))

		err := r.sqlDB.GetContext(ctx.Context(), &user, query, args...)
		// userErr <- err
		if err != nil {
			userSpan.LogKV("query", query)
			utils.LogErrors(userSpan, err)
			userErr = err
		}

	}()

	// Goroutine for role query
	wg.Add(1)
	go func() {
		defer wg.Done()
		roleSpan := opentracing.StartSpan("RoleQuery", opentracing.ChildOf(childSpan.Context()))
		var roleNames []string
		roleQuery := `
            SELECT mv.name 
            FROM pools p
            JOIN mix_values mv ON p.mv2_id = mv.id
            JOIN groups g1 ON p.group1_id = g1.id
            JOIN groups g2 ON p.group2_id = g2.id
            WHERE p.deleted_at IS NULL AND
						g1.name = 'users' AND g2.name = 'roles' 
            AND p.deleted_at IS NULL
            AND p.mv1_id = $1
        `
		err := r.sqlDB.SelectContext(ctx.Context(), &roleNames, roleQuery, params.ID)
		if err != nil {
			roleSpan.LogKV("query", roleQuery)
			utils.LogErrors(roleSpan, err)
			roleErr = err
		}
		user.Roles = roleNames
	}()

	// Goroutine for permission query
	wg.Add(1)
	go func() {
		defer wg.Done()
		permissionSpan := opentracing.StartSpan("PermissionQuery", opentracing.ChildOf(childSpan.Context()))
		var permissionNames []string
		permissionQuery := `
            SELECT mv.name 
            FROM pools p
            JOIN mix_values mv ON p.mv2_id = mv.id
            JOIN groups g1 ON p.group1_id = g1.id
            JOIN groups g2 ON p.group2_id = g2.id
            WHERE p.deleted_at IS NULL AND
						g1.name = 'roles' AND g2.name = 'permissions' AND p.mv1_id IN (
                SELECT mv.id 
                FROM pools p
                JOIN mix_values mv ON p.mv2_id = mv.id
                JOIN groups g1 ON p.group1_id = g1.id
                JOIN groups g2 ON p.group2_id = g2.id
                WHERE g1.name = 'users' AND g2.name = 'roles' AND p.mv1_id = $1
            )
        `
		err := r.sqlDB.SelectContext(ctx.Context(), &permissionNames, permissionQuery, params.ID)
		if err != nil {
			permissionSpan.LogKV("query", permissionQuery)
			utils.LogErrors(permissionSpan, err)
			permissionErr = err
		}
		user.Permissions = permissionNames
	}()

	// Wait for both goroutines to finish
	wg.Wait()

	if userErr != nil {
		defer childSpan.Finish()
		return nil, userErr
	}

	if roleErr != nil {
		defer childSpan.Finish()
		return nil, roleErr
	}

	if permissionErr != nil {
		defer childSpan.Finish()
		return nil, permissionErr
	}

	return &user, nil
}
func (r *UserRepository) GetUserByEmail(ctx *fiber.Ctx, email string) (*dtos.UserDetailDTO, error) {
	var user dtos.UserDetailDTO

	query := `SELECT 
		u.branch_id,
		u.id, u.username, u.name, u.email, u.password, u.address,
		wh.id as warehouse_id,
		b.name as branch_name
	FROM users u
	LEFT JOIN branches b ON u.branch_id = b.id
	LEFT JOIN mix_values wh ON u.branch_id = wh.branch_id AND wh.is_main = 1
	WHERE u.deleted_at IS NULL AND (u.email = $1 OR u.username = $1)`
	if err := r.sqlDB.GetContext(ctx.Context(), &user, query, email); err != nil {
		return nil, err
	}

	id := user.ID

	// Channels for concurrent execution
	roleChan := make(chan error)
	permissionChan := make(chan error)

	// Goroutine for role query
	go func() {
		var roleNames []string
		roleQuery := `
            SELECT mv.name
            FROM pools p
            JOIN mix_values mv ON p.mv2_id = mv.id
            JOIN groups g1 ON p.group1_id = g1.id
            JOIN groups g2 ON p.group2_id = g2.id
            WHERE p.deleted_at IS NULL AND
						g1.name = 'users' AND g2.name = 'roles' AND p.mv1_id = $1
        `
		err := r.sqlDB.SelectContext(ctx.Context(), &roleNames, roleQuery, id)
		if err == nil {
			user.Roles = roleNames
		}
		roleChan <- err
	}()

	// Goroutine for permission query
	go func() {
		var permissionNames []string
		permissionQuery := `
            SELECT mv.name
            FROM pools p
            JOIN mix_values mv ON p.mv2_id = mv.id
            JOIN groups g1 ON p.group1_id = g1.id
            JOIN groups g2 ON p.group2_id = g2.id
            WHERE p.deleted_at IS NULL AND
						g1.name = 'roles' AND g2.name = 'permissions' AND p.mv1_id IN (
                SELECT mv.id
                FROM pools p
                JOIN mix_values mv ON p.mv2_id = mv.id
                JOIN groups g1 ON p.group1_id = g1.id
                JOIN groups g2 ON p.group2_id = g2.id
                WHERE g1.name = 'users' AND g2.name = 'roles' AND p.mv1_id = $1
            )
        `
		err := r.sqlDB.SelectContext(ctx.Context(), &permissionNames, permissionQuery, id)
		if err == nil {
			user.Permissions = permissionNames
		}
		permissionChan <- err
	}()

	// Wait for both goroutines to finish
	roleErr := <-roleChan
	permissionErr := <-permissionChan

	if roleErr != nil {
		return nil, roleErr
	}

	if permissionErr != nil {
		return nil, permissionErr
	}

	return &user, nil
}

// BeginTransaction starts a new transaction
func (r *UserRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *UserRepository) AttachRoles(tx *gorm.DB, user *models.User, roleIDs []uint32, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("UserRepository-AttachRoles", opentracing.ChildOf(span.Context()))

	// Prepare batch insert for new role_user relationships
	var pools []models.Pool
	for _, roleID := range roleIDs {
		pool := models.Pool{
			Group1ID: utils.GroupIDUsers, // users
			Group2ID: utils.GroupIDRoles, // roles
			Mv1ID:    uint32(user.ID),
			Mv2ID:    roleID,
		}
		pools = append(pools, pool)
	}

	params := &dtos.GetUserParams{ID: user.ID}
	// delete existing roles
	if err := r.DeleteRolesByUserID(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		return err
	}

	// Insert all role_user relationships in a single query
	if len(pools) > 0 {
		if err := tx.Create(&pools).Error; err != nil {
			defer childSpan.Finish()
			return err
		}
	}

	return nil
}

func (r *UserRepository) CreateUser(tx *gorm.DB, user *models.User, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("UserRepository-CreateUser", opentracing.ChildOf(span.Context()))
	if err := tx.Create(user).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *UserRepository) UpdateUser(tx *gorm.DB, user *models.User, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("UserRepository-UpdateUser")

	if err := tx.Select("*").Omit("created_at", "created_by_id").Updates(user).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (r *UserRepository) DeleteUser(tx *gorm.DB, params *dtos.GetUserParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("UserRepository-DeleteUser", opentracing.ChildOf(span.Context()))

	// if err := tx.Unscoped().Delete(&models.User{}, id).Error; err != nil {
	if err := tx.Delete(&models.User{}, params).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

// DeleteRolesByUserID
func (r *UserRepository) DeleteRolesByUserID(tx *gorm.DB, params *dtos.GetUserParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("UserRepository-DeleteRolesByUserID", opentracing.ChildOf(span.Context()))
	if err := tx.Exec(`
			UPDATE pools SET deleted_at = NOW() 
			WHERE group1_id = ? AND mv1_id = ?
			AND group2_id = ?
		`, utils.GroupIDUsers, params.ID, utils.GroupIDRoles).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

func (s *UserRepository) RestoreUser(tx *gorm.DB, params *dtos.GetUserParams, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("UserRepository-RestoreUser", opentracing.ChildOf(span.Context()))
	var user models.User
	if err := tx.Unscoped().Model(&user).Where("id = ?", params.ID).Update("deleted_at", nil).Error; err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}

// commit or rollback
func (r *UserRepository) Commit(tx *gorm.DB) error {
	return tx.Commit().Error
}
