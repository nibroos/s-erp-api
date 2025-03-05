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

func SetupPaymentTermRoutes(paymentTerms fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	paymentTermRepo := repository.NewPaymentTermRepository(gormDB, sqlDB, tracer)
	paymentTermService := service.NewPaymentTermService(paymentTermRepo, utilRepo, tracer)
	paymentTermController := rest.NewPaymentTermController(paymentTermService, paymentTermRepo, tracer)

	paymentTerms.Post("/index-payment-term", paymentTermController.GetPaymentTerms)
	paymentTerms.Post("/show-payment-term", paymentTermController.GetPaymentTermByID)
	paymentTerms.Post("/create-payment-term", paymentTermController.CreatePaymentTerm)
	paymentTerms.Post("/update-payment-term", paymentTermController.UpdatePaymentTerm)
	paymentTerms.Post("/delete-payment-term", paymentTermController.DeletePaymentTerm)
	paymentTerms.Post("/restore-payment-term", paymentTermController.RestorePaymentTerm)
	paymentTerms.Post("/excel-payment-term", paymentTermController.ExcelGetPaymentTerms)
	paymentTerms.Post("/csv-payment-term", paymentTermController.CsvGetPaymentTerms)
}
