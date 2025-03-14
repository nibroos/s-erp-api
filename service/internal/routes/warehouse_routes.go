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

func SetupWarehouseRoutes(warehouses fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	warehouseRepo := repository.NewWarehouseRepository(gormDB, sqlDB, tracer)
	warehouseService := service.NewWarehouseService(warehouseRepo, utilRepo, tracer)
	warehouseController := rest.NewWarehouseController(warehouseService, warehouseRepo, tracer)

	warehouses.Post("/index-warehouse", warehouseController.GetWarehouses)
	warehouses.Post("/show-warehouse", warehouseController.GetWarehouseByID)
	warehouses.Post("/create-warehouse", warehouseController.CreateWarehouse)
	warehouses.Post("/update-warehouse", warehouseController.UpdateWarehouse)
	warehouses.Post("/delete-warehouse", warehouseController.DeleteWarehouse)
	warehouses.Post("/restore-warehouse", warehouseController.RestoreWarehouse)
	warehouses.Post("/excel-warehouse", warehouseController.ExcelGetWarehouses)
	warehouses.Post("/csv-warehouse", warehouseController.CsvGetWarehouses)
}
