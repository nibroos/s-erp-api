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

func SetupBranchRoutes(branch fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, tracer opentracing.Tracer) {
	branchRepo := repository.NewBranchRepository(gormDB, sqlDB, tracer)
	branchService := service.NewBranchService(branchRepo, tracer)
	branchController := rest.NewBranchController(branchService, branchRepo, tracer)

	// prefix /branch

	// branch.Post("/index-branch", middleware.PermissionMiddleware("index-branch"), branchController.GetBranches)
	branch.Post("/index-branch", branchController.GetBranches)
	branch.Post("/show-branch", branchController.GetBranchByID)
	branch.Post("/create-branch", branchController.CreateBranch)
	branch.Post("/update-branch", branchController.UpdateBranch)
	branch.Post("/delete-branch", branchController.DeleteBranch)
	branch.Post("/restore-branch", branchController.RestoreBranch)
}
