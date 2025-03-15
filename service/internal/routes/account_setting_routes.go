package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/controller/rest"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

func SetupAccountSettingRoutes(accountSettings fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	accountSettingRepo := repository.NewAccountSettingRepository(gormDB, sqlDB, utilRepo, tracer)
	accountSettingService := service.NewAccountSettingService(accountSettingRepo, utilRepo, tracer)
	accountSettingController := rest.NewAccountSettingController(accountSettingService, accountSettingRepo, tracer)

	accountSettings.Post("/show-account-setting", accountSettingController.GetAccountSettingUser)
	accountSettings.Post("/update-account-setting", accountSettingController.UpdateAccountSetting)
}
