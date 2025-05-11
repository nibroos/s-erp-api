package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/config"
	"github.com/nibroos/s-erp-api/service/internal/controller/rest"
	"github.com/nibroos/s-erp-api/service/internal/middleware"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

// SetupRoutes sets up the REST routes for the user service.
func SetupRoutes(app *fiber.App, gormDB *gorm.DB, sqlDB *sqlx.DB, rabbitmq *config.RabbitMQ, tracer opentracing.Tracer) {
	app.Use(middleware.SafetyMiddleware())

	// Test route to trigger panic
	app.Get("/api/v1/test-panic", func(c *fiber.Ctx) error {
		// Simulate a nil pointer dereference
		var ptr *string
		_ = *ptr // This will trigger a panic
		return c.SendString("This won't be reached")
	})

	// Test route to trigger panic with custom message
	app.Get("/api/v1/test-panic-message", func(c *fiber.Ctx) error {
		panic("This is a test panic message")
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

	// util service
	utilRepo := repository.NewUtilRepository(gormDB, sqlDB)
	userRepo := repository.NewUserRepository(gormDB, sqlDB, utilRepo, tracer)

	// Setup auth routes
	auth.Post("/login", rest.NewUserController(service.NewUserService(repository.NewUserRepository(gormDB, sqlDB, utilRepo, tracer), utilRepo, tracer), userRepo, tracer).Login)
	auth.Post("/register", rest.NewUserController(service.NewUserService(repository.NewUserRepository(gormDB, sqlDB, utilRepo, tracer), utilRepo, tracer), userRepo, tracer).Register)

	// Protected routes
	app.Use(middleware.JWTMiddleware())
	// app.Use(middleware.ConvertToClientTimezone())
	app.Use(middleware.JaegerTracingMiddleware(tracer))
	// app.Use(middleware.JaegerMiddleware(tracer))

	// Grouped routes
	users := version.Group("/users")
	SetupUserRoutes(users, gormDB, sqlDB, utilRepo, tracer)

	identifiers := version.Group("/identifiers")
	SetupIdentifierRoutes(identifiers, gormDB, sqlDB)

	contacts := version.Group("/contacts")
	SetupContactRoutes(contacts, gormDB, sqlDB)

	addresses := version.Group("/addresses")
	SetupAddressRoutes(addresses, gormDB, sqlDB)

	roles := version.Group("/roles")
	SetupRoleRoutes(roles, gormDB, sqlDB, utilRepo, tracer)

	itemGroups := version.Group("/item-groups")
	SetupItemGroupRoutes(itemGroups, gormDB, sqlDB, utilRepo, tracer)

	itemSubGroups := version.Group("/item-sub-groups")
	SetupItemSubGroupRoutes(itemSubGroups, gormDB, sqlDB, utilRepo, tracer)

	companyProfiles := version.Group("/company-profiles")
	SetupCompanyProfileRoutes(companyProfiles, gormDB, sqlDB, tracer)

	branches := version.Group("/branches")
	SetupBranchRoutes(branches, gormDB, sqlDB, tracer)

	customerTypes := version.Group("/customer-types")
	SetupCustomerTypeRoutes(customerTypes, gormDB, sqlDB, utilRepo, tracer)

	orderTypes := version.Group("/order-types")
	SetupOrderTypeRoutes(orderTypes, gormDB, sqlDB, utilRepo, tracer)

	currencies := version.Group("/currencies")
	SetupCurrencyRoutes(currencies, gormDB, sqlDB, utilRepo, tracer)

	units := version.Group("/units")
	SetupUnitRoutes(units, gormDB, sqlDB, utilRepo, tracer)

	tasks := version.Group("/tasks")
	SetupTaskRoutes(tasks, gormDB, sqlDB, utilRepo, tracer)

	vats := version.Group("/vats")
	SetupVatRoutes(vats, gormDB, sqlDB, utilRepo, tracer)

	pph23s := version.Group("/pph23s")
	SetupPph23Routes(pph23s, gormDB, sqlDB, utilRepo, tracer)

	customers := version.Group("/customers")
	SetupCustomerRoutes(customers, gormDB, sqlDB, utilRepo, tracer)

	msItems := version.Group("/ms-items")
	SetupMsItemRoutes(msItems, gormDB, sqlDB, utilRepo, tracer)

	itemUnits := version.Group("/item-units")
	SetupItemUnitRoutes(itemUnits, gormDB, sqlDB, utilRepo, tracer)

	branchItems := version.Group("/branch-items")
	SetupBranchItemRoutes(branchItems, gormDB, sqlDB, utilRepo, tracer)

	products := version.Group("/products")
	SetupProductRoutes(products, gormDB, sqlDB, utilRepo, tracer)

	quotations := version.Group("/quotations")
	SetupQuotationRoutes(quotations, gormDB, sqlDB, utilRepo, tracer)

	salesOrders := version.Group("/sales-orders")
	SetupSalesOrderRoutes(salesOrders, gormDB, sqlDB, utilRepo, tracer)

	inventories := version.Group("/inventories")
	SetupInventoryRoutes(inventories, gormDB, sqlDB, utilRepo, rabbitmq, tracer)

	shippingTerms := version.Group("/shipping-terms")
	SetupShippingTermRoutes(shippingTerms, gormDB, sqlDB, utilRepo, tracer)

	paymentTerms := version.Group("/payment-terms")
	SetupPaymentTermRoutes(paymentTerms, gormDB, sqlDB, utilRepo, tracer)

	purchaseTypes := version.Group("/purchase-types")
	SetupPurchaseTypeRoutes(purchaseTypes, gormDB, sqlDB, utilRepo, tracer)

	ioTypes := version.Group("/io-types")
	SetupIOTypeRoutes(ioTypes, gormDB, sqlDB, utilRepo, tracer)

	warehouses := version.Group("/warehouses")
	SetupWarehouseRoutes(warehouses, gormDB, sqlDB, utilRepo, tracer)

	accountSettings := version.Group("/account-setting")
	SetupAccountSettingRoutes(accountSettings, gormDB, sqlDB, utilRepo, tracer)

	purchaseOrders := version.Group("/purchase-orders")
	SetupPurchaseOrderRoutes(purchaseOrders, gormDB, sqlDB, utilRepo, tracer)

	invoiceDps := version.Group("/invoice-dps")
	SetupInvoiceDpRoutes(invoiceDps, gormDB, sqlDB, utilRepo, tracer)

	salesInvoices := version.Group("/sales-invoices")
	SetupSalesInvoiceRoutes(salesInvoices, gormDB, sqlDB, utilRepo, tracer)

	invoiceMaintenances := version.Group("/invoice-maintenances")
	SetupInvoiceMaintenanceRoutes(invoiceMaintenances, gormDB, sqlDB, utilRepo, tracer)

	invoiceAdjustments := version.Group("/invoice-adjustments")
	SetupInvoiceAdjustmentRoutes(invoiceAdjustments, gormDB, sqlDB, utilRepo, tracer)

	// Scheduler route
	// cron := cron.New()
	// schedulerController := rest.NewSchedulerController(cron, gormDB, sqlDB)
	// version.Post("/scheduler/schedule", schedulerController.Schedule)

	// version.Post("/scheduler/list", schedulerController.ListSchedules)

	// Start the cron scheduler
	// cron.Start()
}
