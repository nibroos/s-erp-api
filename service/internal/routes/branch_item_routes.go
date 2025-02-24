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

func SetupBranchItemRoutes(branchItems fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	branchItemRepo := repository.NewBranchItemRepository(gormDB, sqlDB, tracer)
	branchItemService := service.NewBranchItemService(branchItemRepo, utilRepo, tracer)
	branchItemController := rest.NewBranchItemController(branchItemService, branchItemRepo, tracer)

	// branchItems.Post("/index-branch-item", middleware.PermissionMiddleware("read_masters"), branchItemController.GetBranchItems)
	branchItems.Post("/index-branch-item", branchItemController.GetBranchItems)
	branchItems.Post("/show-branch-item", branchItemController.GetBranchItemByID)
	branchItems.Post("/create-branch-item", branchItemController.CreateBranchItem)
	branchItems.Post("/update-branch-item", branchItemController.UpdateBranchItem)
	branchItems.Post("/delete-branch-item", branchItemController.DeleteBranchItem)
	branchItems.Post("/restore-branch-item", branchItemController.RestoreBranchItem)
	branchItems.Post("/excel-branch-item", branchItemController.ExcelGetBranchItems)
	branchItems.Post("/csv-branch-item", branchItemController.CsvGetBranchItems)
}
