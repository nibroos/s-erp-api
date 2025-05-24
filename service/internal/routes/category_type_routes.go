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

func SetupCategoryTypeRoutes(categoryTypes fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	categoryTypeRepo := repository.NewCategoryTypeRepository(gormDB, sqlDB, tracer)
	categoryTypeService := service.NewCategoryTypeService(categoryTypeRepo, utilRepo, tracer)
	categoryTypeController := rest.NewCategoryTypeController(categoryTypeService, categoryTypeRepo, tracer)

	// categoryTypes.Post("/index-category-type", middleware.PermissionMiddleware("index-category-type"), categoryTypeController.GetCategoryTypes)
	categoryTypes.Post("/index-category-type", categoryTypeController.GetCategoryTypes)
	categoryTypes.Post("/show-category-type", categoryTypeController.GetCategoryTypeByID)
	categoryTypes.Post("/create-category-type", categoryTypeController.CreateCategoryType)
	categoryTypes.Post("/update-category-type", categoryTypeController.UpdateCategoryType)
	categoryTypes.Post("/delete-category-type", categoryTypeController.DeleteCategoryType)
	categoryTypes.Post("/restore-category-type", categoryTypeController.RestoreCategoryType)
	categoryTypes.Post("/excel-category-type", categoryTypeController.ExcelGetCategoryTypes)
	categoryTypes.Post("/csv-category-type", categoryTypeController.CsvGetCategoryTypes)
}
