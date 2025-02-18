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

func SetupCustomerRoutes(customers fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	customerRepo := repository.NewCustomerRepository(gormDB, sqlDB, tracer)
	customerService := service.NewCustomerService(customerRepo, utilRepo, tracer)
	customerController := rest.NewCustomerController(customerService, customerRepo, tracer)

	// customers.Post("/index-customer", middleware.PermissionMiddleware("read_masters"), customerController.GetCustomers)
	customers.Post("/index-customer", customerController.GetCustomers)
	customers.Post("/show-customer", customerController.GetCustomerByID)
	customers.Post("/create-customer", customerController.CreateCustomer)
	customers.Post("/update-customer", customerController.UpdateCustomer)
	customers.Post("/delete-customer", customerController.DeleteCustomer)
	customers.Post("/restore-customer", customerController.RestoreCustomer)
	customers.Post("/excel-customer", customerController.ExcelGetCustomers)
	customers.Post("/csv-customer", customerController.CsvGetCustomers)
}
