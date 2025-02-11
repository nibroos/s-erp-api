package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/controller/rest"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

func SetupCustomerTypeRoutes(customerTypes fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer, ctx *fiber.Ctx) {

	// prefix /customerTypesparentSpan := opentracing.StartSpan(ctx.Path(), opentracing.ChildOf(utils.JaegerMiddleware(ctx, c.tracer)))
	parentSpan := opentracing.StartSpan(ctx.Path(), opentracing.ChildOf(utils.JaegerMiddleware(ctx, tracer)))
	requestBody := ctx.Body()
	responseBody := ctx.Response().Body()
	parentSpan.LogKV("request_body", string(requestBody))
	parentSpan.LogKV("response_body", string(responseBody))
	defer parentSpan.Finish()

	customerTypeRepo := repository.NewCustomerTypeRepository(gormDB, sqlDB, parentSpan)
	customerTypeService := service.NewCustomerTypeService(customerTypeRepo, utilRepo, parentSpan)
	customerTypeController := rest.NewCustomerTypeController(customerTypeService, customerTypeRepo, parentSpan)

	// customerTypes.Post("/index-customer-type", middleware.PermissionMiddleware("index-customer-type"), customerTypeController.GetCustomerTypes)
	customerTypes.Post("/index-customer-type", customerTypeController.GetCustomerTypes)
	customerTypes.Post("/show-customer-type", customerTypeController.GetCustomerTypeByID)
	customerTypes.Post("/create-customer-type", customerTypeController.CreateCustomerType)
	customerTypes.Post("/update-customer-type", customerTypeController.UpdateCustomerType)
	customerTypes.Post("/delete-customer-type", customerTypeController.DeleteCustomerType)
	customerTypes.Post("/restore-customer-type", customerTypeController.RestoreCustomerType)
	customerTypes.Post("/excel-customer-type", customerTypeController.ExcelGetCustomerTypes)
	customerTypes.Post("/csv-customer-type", customerTypeController.CsvGetCustomerTypes)
}
