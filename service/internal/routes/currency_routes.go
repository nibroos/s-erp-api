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

func SetupCurrencyRoutes(currencies fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	currencyRepo := repository.NewCurrencyRepository(gormDB, sqlDB, tracer)
	currencyService := service.NewCurrencyService(currencyRepo, utilRepo, tracer)
	currencyController := rest.NewCurrencyController(currencyService, currencyRepo, tracer)

	// currencies.Post("/index-currency", middleware.PermissionMiddleware("index-currency"), currencyController.GetCurrencies)
	currencies.Post("/index-currency", currencyController.GetCurrencies)
	currencies.Post("/show-currency", currencyController.GetCurrencyByID)
	currencies.Post("/create-currency", currencyController.CreateCurrency)
	currencies.Post("/update-currency", currencyController.UpdateCurrency)
	currencies.Post("/delete-currency", currencyController.DeleteCurrency)
	currencies.Post("/restore-currency", currencyController.RestoreCurrency)
	currencies.Post("/excel-currency", currencyController.ExcelGetCurrencies)
	currencies.Post("/csv-currency", currencyController.CsvGetCurrencies)
}
