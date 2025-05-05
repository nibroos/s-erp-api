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

func SetupInvoiceMaintenanceRoutes(invoiceMaintenances fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	invoiceMaintenanceRepo := repository.NewInvoiceMaintenanceRepository(gormDB, sqlDB, utilRepo, tracer)
	invoiceMaintenanceService := service.NewInvoiceMaintenanceService(invoiceMaintenanceRepo, utilRepo, tracer)
	invoiceMaintenanceController := rest.NewInvoiceMaintenanceController(invoiceMaintenanceService, invoiceMaintenanceRepo, tracer)

	invoiceMaintenances.Post("/index-invoice-maintenance", invoiceMaintenanceController.GetInvoiceMaintenances)
	invoiceMaintenances.Post("/show-invoice-maintenance", invoiceMaintenanceController.GetInvoiceMaintenanceByID)
	invoiceMaintenances.Post("/create-invoice-maintenance", invoiceMaintenanceController.CreateInvoiceMaintenance)
	invoiceMaintenances.Post("/update-invoice-maintenance", invoiceMaintenanceController.UpdateInvoiceMaintenance)
	invoiceMaintenances.Post("/delete-invoice-maintenance", invoiceMaintenanceController.DeleteInvoiceMaintenance)
	invoiceMaintenances.Post("/restore-invoice-maintenance", invoiceMaintenanceController.RestoreInvoiceMaintenance)
	// invoiceMaintenances.Post("/excel-invoice-maintenance", invoiceMaintenanceController.ExcelGetInvoiceMaintenances)
	// invoiceMaintenances.Post("/csv-invoice-maintenance", invoiceMaintenanceController.CsvGetInvoiceMaintenances)
	invoiceMaintenances.Post("/index-ref-so-dt", invoiceMaintenanceController.GetRefSalesOrderForInvoiceMaintenance)
	invoiceMaintenances.Post("/approve-invoice-maintenance", invoiceMaintenanceController.ApproveInvoiceMaintenances)
	invoiceMaintenances.Post("/cancel-approve-invoice-maintenance", invoiceMaintenanceController.CancelApproveInvoiceMaintenances)
}
