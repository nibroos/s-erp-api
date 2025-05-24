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

func SetupPaymentTypeRoutes(paymentTypes fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	paymentTypeRepo := repository.NewPaymentTypeRepository(gormDB, sqlDB, tracer)
	paymentTypeService := service.NewPaymentTypeService(paymentTypeRepo, utilRepo, tracer)
	paymentTypeController := rest.NewPaymentTypeController(paymentTypeService, paymentTypeRepo, tracer)

	// paymentTypes.Post("/index-payment-type", middleware.PermissionMiddleware("index-payment-type"), paymentTypeController.GetPaymentTypes)
	paymentTypes.Post("/index-payment-type", paymentTypeController.GetPaymentTypes)
	paymentTypes.Post("/show-payment-type", paymentTypeController.GetPaymentTypeByID)
	paymentTypes.Post("/create-payment-type", paymentTypeController.CreatePaymentType)
	paymentTypes.Post("/update-payment-type", paymentTypeController.UpdatePaymentType)
	paymentTypes.Post("/delete-payment-type", paymentTypeController.DeletePaymentType)
	paymentTypes.Post("/restore-payment-type", paymentTypeController.RestorePaymentType)
	paymentTypes.Post("/excel-payment-type", paymentTypeController.ExcelGetPaymentTypes)
	paymentTypes.Post("/csv-payment-type", paymentTypeController.CsvGetPaymentTypes)
}
