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

func SetupInvoiceDpRoutes(invoiceDps fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	invoiceDpRepo := repository.NewInvoiceDpRepository(gormDB, sqlDB, utilRepo, tracer)
	invoiceDpService := service.NewInvoiceDpService(invoiceDpRepo, utilRepo, tracer)
	invoiceDpController := rest.NewInvoiceDpController(invoiceDpService, invoiceDpRepo, tracer)

	invoiceDps.Post("/index-invoice-dp", invoiceDpController.GetInvoiceDps)
	invoiceDps.Post("/show-invoice-dp", invoiceDpController.GetInvoiceDpByID)
	invoiceDps.Post("/create-invoice-dp", invoiceDpController.CreateInvoiceDp)
	invoiceDps.Post("/update-invoice-dp", invoiceDpController.UpdateInvoiceDp)
	invoiceDps.Post("/delete-invoice-dp", invoiceDpController.DeleteInvoiceDp)
	invoiceDps.Post("/restore-invoice-dp", invoiceDpController.RestoreInvoiceDp)
	// invoiceDps.Post("/excel-invoice-dp", invoiceDpController.ExcelGetInvoiceDps)
	// invoiceDps.Post("/csv-invoice-dp", invoiceDpController.CsvGetInvoiceDps)
	invoiceDps.Post("/index-ref-so-dt", invoiceDpController.GetRefSalesOrderDts)
	invoiceDps.Post("/widget-invoice-dp", invoiceDpController.GetWidgetInvoiceDps)
}
