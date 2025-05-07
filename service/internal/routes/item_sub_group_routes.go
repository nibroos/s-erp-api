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

func SetupItemSubGroupRoutes(itemSubGroups fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	itemSubGroupRepo := repository.NewItemSubGroupRepository(gormDB, sqlDB, tracer)
	itemSubGroupService := service.NewItemSubGroupService(itemSubGroupRepo, utilRepo, tracer)
	itemSubGroupController := rest.NewItemSubGroupController(itemSubGroupService, itemSubGroupRepo, tracer)

	// itemSubGroups.Post("/index-item-sub-group", middleware.PermissionMiddleware("index-item-sub-group"), itemSubGroupController.GetItemSubGroups)
	itemSubGroups.Post("/index-item-sub-group", itemSubGroupController.GetItemSubGroups)
	itemSubGroups.Post("/show-item-sub-group", itemSubGroupController.GetItemSubGroupByID)
	itemSubGroups.Post("/create-item-sub-group", itemSubGroupController.CreateItemSubGroup)
	itemSubGroups.Post("/update-item-sub-group", itemSubGroupController.UpdateItemSubGroup)
	itemSubGroups.Post("/delete-item-sub-group", itemSubGroupController.DeleteItemSubGroup)
	itemSubGroups.Post("/restore-item-sub-group", itemSubGroupController.RestoreItemSubGroup)
	itemSubGroups.Post("/excel-item-sub-group", itemSubGroupController.ExcelGetItemSubGroups)
	itemSubGroups.Post("/csv-item-sub-group", itemSubGroupController.CsvGetItemSubGroups)

	itemSubGroups.Post("/is-item-sub-group", itemSubGroupController.IsItemSubGroupByID)
}
