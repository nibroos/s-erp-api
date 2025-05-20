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

func SetupInvoiceAdjustmentRoutes(invoiceAdjustments fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	invoiceAdjustmentRepo := repository.NewInvoiceAdjustmentRepository(gormDB, sqlDB, utilRepo, tracer)
	invoiceAdjustmentService := service.NewInvoiceAdjustmentService(invoiceAdjustmentRepo, utilRepo, tracer)
	invoiceAdjustmentController := rest.NewInvoiceAdjustmentController(invoiceAdjustmentService, invoiceAdjustmentRepo, tracer)

	invoiceAdjustments.Post("/index-invoice-adjustment", invoiceAdjustmentController.GetInvoiceAdjustments)
	invoiceAdjustments.Post("/show-invoice-adjustment", invoiceAdjustmentController.GetInvoiceAdjustmentByID)
	invoiceAdjustments.Post("/create-invoice-adjustment", invoiceAdjustmentController.CreateInvoiceAdjustment)
	invoiceAdjustments.Post("/update-invoice-adjustment", invoiceAdjustmentController.UpdateInvoiceAdjustment)
	invoiceAdjustments.Post("/delete-invoice-adjustment", invoiceAdjustmentController.DeleteInvoiceAdjustment)
	invoiceAdjustments.Post("/restore-invoice-adjustment", invoiceAdjustmentController.RestoreInvoiceAdjustment)
	invoiceAdjustments.Post("/index-reference-invoices", invoiceAdjustmentController.GetReferenceInvoices)
	invoiceAdjustments.Post("/excel-invoice-adjustment", invoiceAdjustmentController.ExcelGetInvoiceAdjustments)
	invoiceAdjustments.Post("/csv-invoice-adjustment", invoiceAdjustmentController.CsvGetInvoiceAdjustments)
}
