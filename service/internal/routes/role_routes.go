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

func SetupRoleRoutes(roles fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	roleRepo := repository.NewRoleRepository(gormDB, sqlDB, tracer)
	roleService := service.NewRoleService(roleRepo, utilRepo, tracer)
	roleController := rest.NewRoleController(roleService, roleRepo, tracer)

	// roles.Post("/index-role", middleware.PermissionMiddleware("index-role"), roleController.GetRoles)
	roles.Post("/index-role", roleController.GetRoles)
	roles.Post("/show-role", roleController.GetRoleByID)
	roles.Post("/create-role", roleController.CreateRole)
	roles.Post("/update-role", roleController.UpdateRole)
	roles.Post("/delete-role", roleController.DeleteRole)
	roles.Post("/restore-role", roleController.RestoreRole)
	roles.Post("/excel-role", roleController.ExcelGetRoles)
	roles.Post("/csv-role", roleController.CsvGetRoles)
}
