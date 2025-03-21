package repository

import (
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type AccountSettingRepository struct {
	db       *gorm.DB
	sqlDB    *sqlx.DB
	utilRepo *UtilRepository
	tracer   opentracing.Tracer
}

func NewAccountSettingRepository(db *gorm.DB, sqlDB *sqlx.DB, utilRepo *UtilRepository, tracer opentracing.Tracer) *AccountSettingRepository {
	return &AccountSettingRepository{
		db:       db,
		sqlDB:    sqlDB,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (r *AccountSettingRepository) GetAccountSettingUser(ctx *fiber.Ctx, params *dtos.GetUserByIDParams, span opentracing.Span) (*dtos.AccountSettingDetailDTO, error) {
	childSpan := opentracing.StartSpan("AccountSettingRepository-GetUserByID", opentracing.ChildOf(span.Context()))
	var user dtos.AccountSettingDetailDTO

	query := `SELECT 
		u.id, u.username, u.name, u.email, u.address, u.status, u.phone_number,
		u.profile_image_url, u.password, u.created_at, u.updated_at
	FROM users u 
	WHERE u.id = $1 AND u.deleted_at IS NULL`

	var args []interface{}
	args = append(args, params.ID)

	var wg sync.WaitGroup
	var userErr, roleErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		userSpan := opentracing.StartSpan("UserQuery", opentracing.ChildOf(childSpan.Context()))
		err := r.sqlDB.GetContext(ctx.Context(), &user, query, args...)
		if err != nil {
			userSpan.LogKV("query", query)
			utils.LogErrors(userSpan, err)
			userErr = err
		}
	}()

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

	wg.Wait()

	if userErr != nil {
		defer childSpan.Finish()
		return nil, userErr
	}

	if roleErr != nil {
		defer childSpan.Finish()
		return nil, roleErr
	}

	return &user, nil
}

func (r *AccountSettingRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *AccountSettingRepository) UpdateAccountSetting(tx *gorm.DB, user *models.User, fieldsToUpdate []string, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("AccountSettingRepository-UpdateAccountSetting", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if err := tx.Select(fieldsToUpdate).Updates(user).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return err
	}
	return nil
}
