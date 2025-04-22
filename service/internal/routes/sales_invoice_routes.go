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

func SetupSalesInvoiceRoutes(salesInvoices fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	salesInvoiceRepo := repository.NewSalesInvoiceRepository(gormDB, sqlDB, utilRepo, tracer)
	salesInvoiceService := service.NewSalesInvoiceService(salesInvoiceRepo, utilRepo, tracer)
	salesInvoiceController := rest.NewSalesInvoiceController(salesInvoiceService, salesInvoiceRepo, tracer)

	salesInvoices.Post("/index-sales-invoice", salesInvoiceController.GetSalesInvoices)
	salesInvoices.Post("/show-sales-invoice", salesInvoiceController.GetSalesInvoiceByID)
	salesInvoices.Post("/create-sales-invoice", salesInvoiceController.CreateSalesInvoice)
	salesInvoices.Post("/update-sales-invoice", salesInvoiceController.UpdateSalesInvoice)
	salesInvoices.Post("/delete-sales-invoice", salesInvoiceController.DeleteSalesInvoice)
	salesInvoices.Post("/restore-sales-invoice", salesInvoiceController.RestoreSalesInvoice)
	// salesInvoices.Post("/excel-sales-invoice", salesInvoiceController.ExcelGetSalesInvoices)
	// salesInvoices.Post("/csv-sales-invoice", salesInvoiceController.CsvGetSalesInvoices)
	salesInvoices.Post("/index-ref-so-dt", salesInvoiceController.GetRefSalesOrderDts)
}
