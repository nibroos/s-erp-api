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

func SetupIOTypeRoutes(ioTypes fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	ioTypeRepo := repository.NewIOTypeRepository(gormDB, sqlDB, tracer)
	ioTypeService := service.NewIOTypeService(ioTypeRepo, utilRepo, tracer)
	ioTypeController := rest.NewIOTypeController(ioTypeService, ioTypeRepo, tracer)

	ioTypes.Post("/index-io-type", ioTypeController.GetIOTypes)
	ioTypes.Post("/show-io-type", ioTypeController.GetIOTypeByID)
	ioTypes.Post("/create-io-type", ioTypeController.CreateIOType)
	ioTypes.Post("/update-io-type", ioTypeController.UpdateIOType)
	ioTypes.Post("/delete-io-type", ioTypeController.DeleteIOType)
	ioTypes.Post("/restore-io-type", ioTypeController.RestoreIOType)
	ioTypes.Post("/excel-io-type", ioTypeController.ExcelGetIOTypes)
	ioTypes.Post("/csv-io-type", ioTypeController.CsvGetIOTypes)
}
