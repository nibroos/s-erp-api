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

func SetupCompanyProfileRoutes(companyProfile fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, tracer opentracing.Tracer) {
	companyProfileRepo := repository.NewCompanyProfileRepository(gormDB, sqlDB, tracer)
	companyProfileService := service.NewCompanyProfileService(companyProfileRepo, tracer)
	companyProfileController := rest.NewCompanyProfileController(companyProfileService, companyProfileRepo, tracer)

	// prefix /companyProfile

	// companyProfile.Post("/index-company-profile", middleware.PermissionMiddleware("index-company-profile"), companyProfileController.GetCompanyProfiles)
	companyProfile.Post("/index-company-profile", companyProfileController.GetCompanyProfiles)
	companyProfile.Post("/show-company-profile", companyProfileController.GetCompanyProfileByID)
	// companyProfile.Post("/show-primary-company-profile", companyProfileController.GetPrimaryCompanyProfileByID)
	companyProfile.Post("/create-company-profile", companyProfileController.CreateCompanyProfile)
	companyProfile.Post("/update-company-profile", companyProfileController.UpdateCompanyProfile)
	companyProfile.Post("/delete-company-profile", companyProfileController.DeleteCompanyProfile)
	companyProfile.Post("/restore-company-profile", companyProfileController.RestoreCompanyProfile)
	companyProfile.Post("/index-bank-information", companyProfileController.GetBankInformationsWithCompany)
}
