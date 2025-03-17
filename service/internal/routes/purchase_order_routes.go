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

func SetupPurchaseOrderRoutes(purchaseOrders fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	purchaseOrderRepo := repository.NewPurchaseOrderRepository(gormDB, sqlDB, utilRepo, tracer)
	purchaseOrderService := service.NewPurchaseOrderService(purchaseOrderRepo, utilRepo, tracer)
	purchaseOrderController := rest.NewPurchaseOrderController(purchaseOrderService, purchaseOrderRepo, tracer)

	purchaseOrders.Post("/create-purchase-order", purchaseOrderController.CreatePurchaseOrder)
}
