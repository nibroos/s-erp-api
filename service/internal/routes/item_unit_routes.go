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

func SetupItemUnitRoutes(itemUnits fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	itemUnitRepo := repository.NewItemUnitRepository(gormDB, sqlDB, tracer)
	itemUnitService := service.NewItemUnitService(itemUnitRepo, utilRepo, tracer)
	itemUnitController := rest.NewItemUnitController(itemUnitService, itemUnitRepo, tracer)

	// itemUnits.Post("/index-item-unit", middleware.PermissionMiddleware("read_masters"), itemUnitController.GetItemUnits)
	itemUnits.Post("/index-item-unit", itemUnitController.GetItemUnits)
	itemUnits.Post("/show-item-unit", itemUnitController.GetItemUnitByID)
	itemUnits.Post("/create-item-unit", itemUnitController.CreateItemUnit)
	itemUnits.Post("/update-item-unit", itemUnitController.UpdateItemUnit)
	itemUnits.Post("/delete-item-unit", itemUnitController.DeleteItemUnit)
	itemUnits.Post("/restore-item-unit", itemUnitController.RestoreItemUnit)
	itemUnits.Post("/excel-item-unit", itemUnitController.ExcelGetItemUnits)
	itemUnits.Post("/csv-item-unit", itemUnitController.CsvGetItemUnits)
}
