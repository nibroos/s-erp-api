package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type AccountSettingService struct {
	repo     *repository.AccountSettingRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewAccountSettingService(repo *repository.AccountSettingRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *AccountSettingService {
	return &AccountSettingService{repo: repo, utilRepo: utilRepo, tracer: tracer}
}

func (s *AccountSettingService) GetAccountSettingUser(ctx *fiber.Ctx, params *dtos.GetUserByIDParams, span opentracing.Span) (*dtos.AccountSettingDetailDTO, error) {
	childSpan := opentracing.StartSpan("AccountSettingService-GetAccountSettingUser", opentracing.ChildOf(span.Context()))

	user, err := s.repo.GetAccountSettingUser(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return user, nil
}

func (s *AccountSettingService) UpdateAccountSetting(ctx *fiber.Ctx, tx *gorm.DB, user *models.User, fieldsToUpdate []string, span opentracing.Span) (*models.User, error) {
	childSpan := opentracing.StartSpan("AccountSettingService-UpdateAccountSetting", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if utils.Contains(fieldsToUpdate, "password") && user.Password != "" {
		hashedPassword, err := utils.HashPassword(user.Password, childSpan)
		if err != nil {
			return nil, err
		}
		user.Password = hashedPassword
	}

	if err := s.repo.UpdateAccountSetting(tx, user, fieldsToUpdate, childSpan); err != nil {
		return nil, err
	}

	return user, nil
}
