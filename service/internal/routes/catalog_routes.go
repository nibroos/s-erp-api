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

func SetupCatalogRoutes(catalogs fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	catalogRepo := repository.NewCatalogRepository(gormDB, sqlDB, tracer)
	catalogService := service.NewCatalogService(catalogRepo, utilRepo, tracer)
	catalogController := rest.NewCatalogController(catalogService, catalogRepo, tracer)

	// catalogs.Post("/index-catalog", middleware.PermissionMiddleware("read_masters"), catalogController.GetCatalogs)
	catalogs.Post("/index-catalog", catalogController.GetCatalogs)
	catalogs.Post("/show-catalog", catalogController.GetCatalogByID)
	catalogs.Post("/create-catalog", catalogController.CreateCatalog)
	catalogs.Post("/update-catalog", catalogController.UpdateCatalog)
	catalogs.Post("/delete-catalog", catalogController.DeleteCatalog)
	catalogs.Post("/restore-catalog", catalogController.RestoreCatalog)
	catalogs.Post("/excel-catalog", catalogController.ExcelGetCatalogs)
	catalogs.Post("/csv-catalog", catalogController.CsvGetCatalogs)
}
