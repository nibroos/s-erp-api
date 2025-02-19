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

func SetupOrderTypeRoutes(orderTypes fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	orderTypeRepo := repository.NewOrderTypeRepository(gormDB, sqlDB, tracer)
	orderTypeService := service.NewOrderTypeService(orderTypeRepo, utilRepo, tracer)
	orderTypeController := rest.NewOrderTypeController(orderTypeService, orderTypeRepo, tracer)

	// orderTypes.Post("/index-order-type", middleware.PermissionMiddleware("index-order-type"), orderTypeController.GetOrderTypes)
	orderTypes.Post("/index-order-type", orderTypeController.GetOrderTypes)
	orderTypes.Post("/show-order-type", orderTypeController.GetOrderTypeByID)
	orderTypes.Post("/create-order-type", orderTypeController.CreateOrderType)
	orderTypes.Post("/update-order-type", orderTypeController.UpdateOrderType)
	orderTypes.Post("/delete-order-type", orderTypeController.DeleteOrderType)
	orderTypes.Post("/restore-order-type", orderTypeController.RestoreOrderType)
	orderTypes.Post("/excel-order-type", orderTypeController.ExcelGetOrderTypes)
	orderTypes.Post("/csv-order-type", orderTypeController.CsvGetOrderTypes)
}
