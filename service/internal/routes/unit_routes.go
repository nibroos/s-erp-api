package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/controller/rest"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"gorm.io/gorm"
)

func SetupUnitRoutes(units fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository) {
	unitRepo := repository.NewUnitRepository(gormDB, sqlDB)
	unitService := service.NewUnitService(unitRepo, utilRepo)
	unitController := rest.NewUnitController(unitService, unitRepo)

	// prefix /units

	// units.Post("/index-unit", middleware.PermissionMiddleware("index-unit"), unitController.GetUnits)
	units.Post("/index-unit", unitController.GetUnits)
	units.Post("/show-unit", unitController.GetUnitByID)
	units.Post("/create-unit", unitController.CreateUnit)
	units.Post("/update-unit", unitController.UpdateUnit)
	units.Post("/delete-unit", unitController.DeleteUnit)
	units.Post("/restore-unit", unitController.RestoreUnit)
	units.Post("/excel-unit", unitController.ExcelGetUnits)
	units.Post("/csv-unit", unitController.CsvGetUnits)
}
