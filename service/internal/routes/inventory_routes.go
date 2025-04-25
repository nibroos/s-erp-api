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

func SetupInventoryRoutes(inventories fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	inventoryRepo := repository.NewInventoryRepository(gormDB, sqlDB, utilRepo, tracer)
	inventoryService := service.NewInventoryService(inventoryRepo, utilRepo, tracer)
	inventoryController := rest.NewInventoryController(inventoryService, inventoryRepo, tracer)

	inventories.Post("/index-inventory", inventoryController.GetInventories)
	inventories.Post("/show-inventory", inventoryController.GetInventoryByID)
	inventories.Post("/create-inventory", inventoryController.CreateInventory)
	inventories.Post("/update-inventory", inventoryController.UpdateInventory)
	inventories.Post("/delete-inventory", inventoryController.DeleteInventory)
	inventories.Post("/restore-inventory", inventoryController.RestoreInventory)
	inventories.Post("/excel-inventory", inventoryController.ExcelGetInventories)
	inventories.Post("/csv-inventory", inventoryController.CsvGetInventories)
	inventories.Post("/index-ref-so-dt", inventoryController.GetRefIndexSoDts)
}
