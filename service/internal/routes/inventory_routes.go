package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/config"
	"github.com/nibroos/s-erp-api/service/internal/controller/rest"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

func SetupInventoryRoutes(inventories fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, rabbitmq *config.RabbitMQ, tracer opentracing.Tracer) {
	inventoryRepo := repository.NewInventoryRepository(gormDB, sqlDB, utilRepo, rabbitmq, tracer)
	inventoryService := service.NewInventoryService(inventoryRepo, utilRepo, rabbitmq, tracer)
	inventoryController := rest.NewInventoryController(inventoryService, inventoryRepo, rabbitmq, tracer)

	inventories.Post("/index-inventory", inventoryController.GetInventories)
	inventories.Post("/index-inventory-status", inventoryController.GetInventoriesStatus)
	inventories.Post("/show-inventory", inventoryController.GetInventoryByID)
	inventories.Post("/create-inventory", inventoryController.CreateInventory)
	inventories.Post("/update-inventory", inventoryController.UpdateInventory)
	inventories.Post("/delete-inventory", inventoryController.DeleteInventory)
	inventories.Post("/restore-inventory", inventoryController.RestoreInventory)
	inventories.Post("/excel-inventory", inventoryController.ExcelGetInventories)
	inventories.Post("/csv-inventory", inventoryController.CsvGetInventories)
	inventories.Post("/pdf-inventory", inventoryController.Pdf)
	inventories.Post("/index-ref-so-dt", inventoryController.GetRefIndexSoDts)
	inventories.Post("/index-ref-ro-dt", inventoryController.GetRefIndexRoDts)
	inventories.Post("/index-ref-po-dt", inventoryController.GetRefIndexPoDts)
	inventories.Post("/index-ref-inv-dt", inventoryController.GetRefIndexInvDts)

	inventories.Post("/stocks/index-stock", inventoryController.GetStocks)
	inventories.Post("/stocks/index-stock-closings", inventoryController.GetStockClosings)
	inventories.Post("/stocks/create-stock-closing", inventoryController.CreateStockClosings)
	inventories.Post("/stocks/create-update-adjustment", inventoryController.GetStocks)
}
