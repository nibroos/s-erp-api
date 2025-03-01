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

func SetupProductRoutes(products fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	productRepo := repository.NewProductRepository(gormDB, sqlDB, utilRepo, tracer)
	productService := service.NewProductService(productRepo, utilRepo, tracer)
	productController := rest.NewProductController(productService, productRepo, tracer)

	// products.Post("/index-product", middleware.PermissionMiddleware("read_masters"), productController.GetProducts)
	products.Post("/index-product", productController.GetProducts)
	products.Post("/show-product", productController.GetProductByID)
	products.Post("/create-product", productController.CreateProduct)
	products.Post("/update-product", productController.UpdateProduct)
	products.Post("/delete-product", productController.DeleteProduct)
	products.Post("/restore-product", productController.RestoreProduct)
	products.Post("/excel-product", productController.ExcelGetProducts)
	products.Post("/csv-product", productController.CsvGetProducts)
}
