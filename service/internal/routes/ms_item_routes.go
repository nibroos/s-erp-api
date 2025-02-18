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

func SetupMsItemRoutes(msItems fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	msItemRepo := repository.NewMsItemRepository(gormDB, sqlDB, tracer)
	msItemService := service.NewMsItemService(msItemRepo, utilRepo, tracer)
	msItemController := rest.NewMsItemController(msItemService, msItemRepo, tracer)

	// msItems.Post("/index-ms-item", middleware.PermissionMiddleware("read_masters"), msItemController.GetMsItems)
	msItems.Post("/index-ms-item", msItemController.GetMsItems)
	msItems.Post("/show-ms-item", msItemController.GetMsItemByID)
	msItems.Post("/create-ms-item", msItemController.CreateMsItem)
	msItems.Post("/update-ms-item", msItemController.UpdateMsItem)
	msItems.Post("/delete-ms-item", msItemController.DeleteMsItem)
	msItems.Post("/restore-ms-item", msItemController.RestoreMsItem)
	msItems.Post("/excel-ms-item", msItemController.ExcelGetMsItems)
	msItems.Post("/csv-ms-item", msItemController.CsvGetMsItems)
}
