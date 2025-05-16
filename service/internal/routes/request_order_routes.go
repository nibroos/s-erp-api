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

func SetupRequestOrderRoutes(requestOrders fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	requestOrderRepo := repository.NewRequestOrderRepository(gormDB, sqlDB, utilRepo, tracer)
	requestOrderService := service.NewRequestOrderService(requestOrderRepo, utilRepo, tracer)
	requestOrderController := rest.NewRequestOrderController(requestOrderService, requestOrderRepo, tracer)

	requestOrders.Post("/index-request-order", requestOrderController.GetRequestOrders)
	requestOrders.Post("/show-request-order", requestOrderController.GetRequestOrderByID)
	requestOrders.Post("/create-request-order", requestOrderController.CreateRequestOrder)
	requestOrders.Post("/update-request-order", requestOrderController.UpdateRequestOrder)
	requestOrders.Post("/delete-request-order", requestOrderController.DeleteRequestOrder)
	requestOrders.Post("/restore-request-order", requestOrderController.RestoreRequestOrder)
	requestOrders.Post("/index-ref-so-dt", requestOrderController.GetRefSalesOrderDts)
	requestOrders.Post("/index-ref-product", requestOrderController.GetRefProductForRequestOrder)
	requestOrders.Post("/widget-request-order", requestOrderController.GetWidgetRequestOrders)
}
