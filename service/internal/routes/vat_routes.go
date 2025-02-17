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

func SetupVatRoutes(vats fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	vatRepo := repository.NewVatRepository(gormDB, sqlDB, tracer)
	vatService := service.NewVatService(vatRepo, utilRepo, tracer)
	vatController := rest.NewVatController(vatService, vatRepo, tracer)

	// vats.Post("/index-vat", middleware.PermissionMiddleware("index-vat"), vatController.GetVats)
	vats.Post("/index-vat", vatController.GetVats)
	vats.Post("/show-vat", vatController.GetVatByID)
	vats.Post("/create-vat", vatController.CreateVat)
	vats.Post("/update-vat", vatController.UpdateVat)
	vats.Post("/delete-vat", vatController.DeleteVat)
	vats.Post("/restore-vat", vatController.RestoreVat)
	vats.Post("/excel-vat", vatController.ExcelGetVats)
	vats.Post("/csv-vat", vatController.CsvGetVats)

	vats.Post("/index-vat-history", vatController.GetVatHistories)
	vats.Post("/show-vat-history", vatController.GetVatHistoryByID)
	vats.Post("/update-vat-history", vatController.UpdateVatHistory)
	vats.Post("/delete-vat-history", vatController.DeleteVatHistory)
	vats.Post("/restore-vat-history", vatController.RestoreVatHistory)
	vats.Post("/excel-vat-history", vatController.ExcelGetVatsHistory)
	vats.Post("/csv-vat-history", vatController.CsvGetVatsHistory)
}
