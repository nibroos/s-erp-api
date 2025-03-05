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

func SetupPurchaseTypeRoutes(purchaseTypes fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	purchaseTypeRepo := repository.NewPurchaseTypeRepository(gormDB, sqlDB, tracer)
	purchaseTypeService := service.NewPurchaseTypeService(purchaseTypeRepo, utilRepo, tracer)
	purchaseTypeController := rest.NewPurchaseTypeController(purchaseTypeService, purchaseTypeRepo, tracer)

	purchaseTypes.Post("/index-purchase-type", purchaseTypeController.GetPurchaseTypes)
	purchaseTypes.Post("/show-purchase-type", purchaseTypeController.GetPurchaseTypeByID)
	purchaseTypes.Post("/create-purchase-type", purchaseTypeController.CreatePurchaseType)
	purchaseTypes.Post("/update-purchase-type", purchaseTypeController.UpdatePurchaseType)
	purchaseTypes.Post("/delete-purchase-type", purchaseTypeController.DeletePurchaseType)
	purchaseTypes.Post("/restore-purchase-type", purchaseTypeController.RestorePurchaseType)
	purchaseTypes.Post("/excel-purchase-type", purchaseTypeController.ExcelGetPurchaseTypes)
	purchaseTypes.Post("/csv-purchase-type", purchaseTypeController.CsvGetPurchaseTypes)
}
