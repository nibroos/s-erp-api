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

func SetupPph23Routes(pph23s fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	pph23Repo := repository.NewPph23Repository(gormDB, sqlDB, tracer)
	pph23Service := service.NewPph23Service(pph23Repo, utilRepo, tracer)
	pph23Controller := rest.NewPph23Controller(pph23Service, pph23Repo, tracer)

	// pph23s.Post("/index-pph23", middleware.PermissionMiddleware("index-pph23"), pph23Controller.GetPph23s)
	pph23s.Post("/index-pph23", pph23Controller.GetPph23s)
	pph23s.Post("/show-pph23", pph23Controller.GetPph23ByID)
	pph23s.Post("/create-pph23", pph23Controller.CreatePph23)
	pph23s.Post("/update-pph23", pph23Controller.UpdatePph23)
	pph23s.Post("/delete-pph23", pph23Controller.DeletePph23)
	pph23s.Post("/restore-pph23", pph23Controller.RestorePph23)
	pph23s.Post("/excel-pph23", pph23Controller.ExcelGetPph23s)
	pph23s.Post("/csv-pph23", pph23Controller.CsvGetPph23s)
}
