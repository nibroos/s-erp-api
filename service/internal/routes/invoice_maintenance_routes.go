package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/config"
	"github.com/nibroos/s-erp-api/service/internal/controller/rest"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

func SetupInvoiceMaintenanceRoutes(invoiceMaintenances fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, rabbitmq *config.RabbitMQ, tracer opentracing.Tracer) {
	invoiceMaintenanceRepo := repository.NewInvoiceMaintenanceRepository(gormDB, sqlDB, utilRepo, rabbitmq, tracer)
	invoiceMaintenanceService := service.NewInvoiceMaintenanceService(invoiceMaintenanceRepo, utilRepo, rabbitmq, tracer)
	invoiceMaintenanceController := rest.NewInvoiceMaintenanceController(invoiceMaintenanceService, invoiceMaintenanceRepo, tracer)

	invoiceMaintenances.Post("/index-invoice-maintenance", invoiceMaintenanceController.GetInvoiceMaintenances)
	invoiceMaintenances.Post("/index-detail-invoice-maintenance", invoiceMaintenanceController.GetInvoiceMaintenancesDetails)
	invoiceMaintenances.Post("/show-invoice-maintenance", invoiceMaintenanceController.GetInvoiceMaintenanceByID)
	invoiceMaintenances.Post("/create-invoice-maintenance", invoiceMaintenanceController.CreateInvoiceMaintenance)
	invoiceMaintenances.Post("/update-invoice-maintenance", invoiceMaintenanceController.UpdateInvoiceMaintenance)
	invoiceMaintenances.Post("/delete-invoice-maintenance", invoiceMaintenanceController.DeleteInvoiceMaintenance)
	invoiceMaintenances.Post("/restore-invoice-maintenance", invoiceMaintenanceController.RestoreInvoiceMaintenance)
	invoiceMaintenances.Post("/excel-invoice-maintenance", invoiceMaintenanceController.ExcelGetInvoiceMaintenances)
	invoiceMaintenances.Post("/csv-invoice-maintenance", invoiceMaintenanceController.CsvGetInvoiceMaintenances)
	invoiceMaintenances.Post("/pdf-invoice-maintenance", invoiceMaintenanceController.Pdf)
	invoiceMaintenances.Post("/index-ref-so-dt", invoiceMaintenanceController.GetRefSalesOrderForInvoiceMaintenance)
	invoiceMaintenances.Post("/approve-invoice-maintenance", invoiceMaintenanceController.ApproveInvoiceMaintenances)
	invoiceMaintenances.Post("/cancel-approve-invoice-maintenance", invoiceMaintenanceController.CancelApproveInvoiceMaintenances)
	invoiceMaintenances.Post("/widget-invoice-maintenance", invoiceMaintenanceController.GetWidgetInvoiceMaintenances)
	invoiceMaintenances.Post("/repeat-invoice-maintenance", invoiceMaintenanceController.RepeatInvoiceMaintenances)
	invoiceMaintenances.Post("/emails-invoice-maintenance", invoiceMaintenanceController.EmailsApproveInvoiceMaintenances)
}
