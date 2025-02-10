package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/controller/rest"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"gorm.io/gorm"
)

func SetupItemSubGroupRoutes(itemSubGroup fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository) {
	itemSubGroupRepo := repository.NewItemSubGroupRepository(gormDB, sqlDB)
	itemSubGroupService := service.NewItemSubGroupService(itemSubGroupRepo, utilRepo)
	itemSubGroupController := rest.NewItemSubGroupController(itemSubGroupService, itemSubGroupRepo)

	// prefix /itemSubGroup

	// itemSubGroup.Post("/index-item-sub-group", middleware.PermissionMiddleware("index-item-sub-group"), itemSubGroupController.GetItemSubGroups)
	itemSubGroup.Post("/index-item-sub-group", itemSubGroupController.GetItemSubGroups)
	itemSubGroup.Post("/show-item-sub-group", itemSubGroupController.GetItemSubGroupByID)
	itemSubGroup.Post("/create-item-sub-group", itemSubGroupController.CreateItemSubGroup)
	itemSubGroup.Post("/update-item-sub-group", itemSubGroupController.UpdateItemSubGroup)
	itemSubGroup.Post("/delete-item-sub-group", itemSubGroupController.DeleteItemSubGroup)
	itemSubGroup.Post("/restore-item-sub-group", itemSubGroupController.RestoreItemSubGroup)
	itemSubGroup.Post("/excel-item-sub-group", itemSubGroupController.ExcelGetItemSubGroups)
	itemSubGroup.Post("/csv-item-sub-group", itemSubGroupController.CsvGetItemSubGroups)
}
