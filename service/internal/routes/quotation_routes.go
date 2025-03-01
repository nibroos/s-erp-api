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

func SetupQuotationRoutes(quotations fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	quotationRepo := repository.NewQuotationRepository(gormDB, sqlDB, utilRepo, tracer)
	quotationService := service.NewQuotationService(quotationRepo, utilRepo, tracer)
	quotationController := rest.NewQuotationController(quotationService, quotationRepo, tracer)

	// quotations.Post("/index-quotation", middleware.PermissionMiddleware("read_masters"), quotationController.GetQuotations)
	quotations.Post("/index-quotation", quotationController.GetQuotations)
	quotations.Post("/show-quotation", quotationController.GetQuotationByID)
	quotations.Post("/create-quotation", quotationController.CreateQuotation)
	quotations.Post("/update-quotation", quotationController.UpdateQuotation)
	quotations.Post("/delete-quotation", quotationController.DeleteQuotation)
	quotations.Post("/restore-quotation", quotationController.RestoreQuotation)
	quotations.Post("/excel-quotation", quotationController.ExcelGetQuotations)
	quotations.Post("/csv-quotation", quotationController.CsvGetQuotations)
}
