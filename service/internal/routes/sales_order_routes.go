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

func SetupSalesOrderRoutes(salesOrders fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	salesOrderRepo := repository.NewSalesOrderRepository(gormDB, sqlDB, utilRepo, tracer)
	salesOrderService := service.NewSalesOrderService(salesOrderRepo, utilRepo, tracer)
	salesOrderController := rest.NewSalesOrderController(salesOrderService, salesOrderRepo, tracer)

	// salesOrders.Post("/index-sales-order", middleware.PermissionMiddleware("read_masters"), salesOrderController.GetSalesOrders)
	// salesOrders.Post("/index-project-app", salesOrderController.GetProjectsApp)
	salesOrders.Post("/show-project-app", salesOrderController.GetSalesOrderByID)
	salesOrders.Post("/index-sales-order", salesOrderController.GetSalesOrders)
	salesOrders.Post("/show-sales-order", salesOrderController.GetSalesOrderByID)
	salesOrders.Post("/create-sales-order", salesOrderController.CreateSalesOrder)
	salesOrders.Post("/update-sales-order", salesOrderController.UpdateSalesOrder)
	salesOrders.Post("/delete-sales-order", salesOrderController.DeleteSalesOrder)
	salesOrders.Post("/restore-sales-order", salesOrderController.RestoreSalesOrder)
	salesOrders.Post("/excel-sales-order", salesOrderController.ExcelGetSalesOrders)
	salesOrders.Post("/csv-sales-order", salesOrderController.CsvGetSalesOrders)
	salesOrders.Post("/index-ref-quo-dt", salesOrderController.GetRefIndexQuoDts)

	salesOrders.Post("/index-calendar", salesOrderController.GetCalendars)
	salesOrders.Post("/update-sales-order-schedule", salesOrderController.UpdateScheduleSalesOrder)
	salesOrders.Post("/update-sales-order-schedule-app", salesOrderController.UpdateScheduleSalesOrderApp)
	salesOrders.Post("/update-sales-order-schedule-app-upload", salesOrderController.UpdateScheduleSalesOrderAppUpload)
}
