package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/controller/rest"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"gorm.io/gorm"
)

func SetupCompanyProfileRoutes(companyProfile fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB) {
	itemGroupRepo := repository.NewCompanyProfileRepository(gormDB, sqlDB)
	itemGroupService := service.NewCompanyProfileService(itemGroupRepo)
	itemGroupController := rest.NewCompanyProfileController(itemGroupService)

	// prefix /companyProfile

	// companyProfile.Post("/index-company-profile", middleware.PermissionMiddleware("index-company-profile"), itemGroupController.GetCompanyProfiles)
	companyProfile.Post("/index-company-profile", itemGroupController.GetCompanyProfiles)
	companyProfile.Post("/show-company-profile", itemGroupController.GetCompanyProfileByID)
	companyProfile.Post("/create-company-profile", itemGroupController.CreateCompanyProfile)
	companyProfile.Post("/update-company-profile", itemGroupController.UpdateCompanyProfile)
	companyProfile.Post("/delete-company-profile", itemGroupController.DeleteCompanyProfile)
	companyProfile.Post("/restore-company-profile", itemGroupController.RestoreCompanyProfile)
}
