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

func SetupShippingTermRoutes(shippingTerms fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	shippingTermRepo := repository.NewShippingTermRepository(gormDB, sqlDB, tracer)
	shippingTermService := service.NewShippingTermService(shippingTermRepo, utilRepo, tracer)
	shippingTermController := rest.NewShippingTermController(shippingTermService, shippingTermRepo, tracer)

	shippingTerms.Post("/index-shipping-term", shippingTermController.GetShippingTerms)
	shippingTerms.Post("/show-shipping-term", shippingTermController.GetShippingTermByID)
	shippingTerms.Post("/create-shipping-term", shippingTermController.CreateShippingTerm)
	shippingTerms.Post("/update-shipping-term", shippingTermController.UpdateShippingTerm)
	shippingTerms.Post("/delete-shipping-term", shippingTermController.DeleteShippingTerm)
	shippingTerms.Post("/restore-shipping-term", shippingTermController.RestoreShippingTerm)
	shippingTerms.Post("/excel-shipping-term", shippingTermController.ExcelGetShippingTerms)
	shippingTerms.Post("/csv-shipping-term", shippingTermController.CsvGetShippingTerms)
}
