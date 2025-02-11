package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/controller/rest"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"gorm.io/gorm"
)

func SetupVatRoutes(vats fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository) {
	vatRepo := repository.NewVatRepository(gormDB, sqlDB)
	vatService := service.NewVatService(vatRepo, utilRepo)
	vatController := rest.NewVatController(vatService, vatRepo)

	// prefix /vats

	// vats.Post("/index-vat", middleware.PermissionMiddleware("index-vat"), vatController.GetVats)
	vats.Post("/index-vat", vatController.GetVats)
	vats.Post("/show-vat", vatController.GetVatByID)
	vats.Post("/create-vat", vatController.CreateVat)
	vats.Post("/update-vat", vatController.UpdateVat)
	vats.Post("/delete-vat", vatController.DeleteVat)
	vats.Post("/restore-vat", vatController.RestoreVat)
	vats.Post("/excel-vat", vatController.ExcelGetVats)
	vats.Post("/csv-vat", vatController.CsvGetVats)
}
