package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/controller/rest"
	"github.com/nibroos/s-erp-api/service/internal/middleware"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

// SetupRoutes sets up the REST routes for the user service.
func SetupRoutes(app *fiber.App, gormDB *gorm.DB, sqlDB *sqlx.DB, tracer opentracing.Tracer) {
	// Public routes
	app.Get("/api/v1/users/test", func(c *fiber.Ctx) error {
		return c.SendString("REST Users Service!")
	})

	version := app.Group("/api/v1")

	// Seeder route
	version.Post("/seeders/run", rest.NewSeederController(sqlDB.DB).RunSeeders)

	auth := version.Group("/auth")

	version.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Service is running",
		})
	})

	// Setup auth routes
	auth.Post("/login", rest.NewUserController(service.NewUserService(repository.NewUserRepository(gormDB, sqlDB))).Login)
	auth.Post("/register", rest.NewUserController(service.NewUserService(repository.NewUserRepository(gormDB, sqlDB))).Register)

	// Protected routes
	app.Use(middleware.JWTMiddleware())
	// app.Use(middleware.ConvertToClientTimezone())
	app.Use(middleware.JaegerTracingMiddleware(tracer))
	// app.Use(middleware.JaegerMiddleware(tracer))

	// util service
	utilRepo := repository.NewUtilRepository(gormDB, sqlDB)

	// Grouped routes
	users := version.Group("/users")
	SetupUserRoutes(users, gormDB, sqlDB)

	identifiers := version.Group("/identifiers")
	SetupIdentifierRoutes(identifiers, gormDB, sqlDB)

	contacts := version.Group("/contacts")
	SetupContactRoutes(contacts, gormDB, sqlDB)

	addresses := version.Group("/addresses")
	SetupAddressRoutes(addresses, gormDB, sqlDB)

	itemGroup := version.Group("/item-groups")
	SetupItemGroupRoutes(itemGroup, gormDB, sqlDB, utilRepo, tracer)

	itemSubGroup := version.Group("/item-sub-groups")
	SetupItemSubGroupRoutes(itemSubGroup, gormDB, sqlDB, utilRepo)

	companyProfile := version.Group("/company-profiles")
	SetupCompanyProfileRoutes(companyProfile, gormDB, sqlDB, tracer)

	customerType := version.Group("/customer-types")
	SetupCustomerTypeRoutes(customerType, gormDB, sqlDB, utilRepo, tracer)

	currency := version.Group("/currencies")
	SetupCurrencyRoutes(currency, gormDB, sqlDB, utilRepo, tracer)

	unit := version.Group("/units")
	SetupUnitRoutes(unit, gormDB, sqlDB, utilRepo, tracer)

	vat := version.Group("/vats")
	SetupVatRoutes(vat, gormDB, sqlDB, utilRepo, tracer)

	pph23 := version.Group("/pph23s")
	SetupPph23Routes(pph23, gormDB, sqlDB, utilRepo, tracer)

	// Scheduler route
	// cron := cron.New()
	// schedulerController := rest.NewSchedulerController(cron, gormDB, sqlDB)
	// version.Post("/scheduler/schedule", schedulerController.Schedule)

	// version.Post("/scheduler/list", schedulerController.ListSchedules)

	// Start the cron scheduler
	// cron.Start()
}
