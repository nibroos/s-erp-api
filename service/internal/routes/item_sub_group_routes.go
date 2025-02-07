package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/controller/rest"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"gorm.io/gorm"
)

func SetupItemSubGroupRoutes(itemSubGroup fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB) {
	itemGroupRepo := repository.NewItemSubGroupRepository(gormDB, sqlDB)
	itemGroupService := service.NewItemSubGroupService(itemGroupRepo)
	itemGroupController := rest.NewItemSubGroupController(itemGroupService)

	// prefix /itemSubGroup

	// itemSubGroup.Post("/index-item-sub-group", middleware.PermissionMiddleware("index-item-sub-group"), itemGroupController.GetItemSubGroups)
	itemSubGroup.Post("/index-item-sub-group", itemGroupController.GetItemSubGroups)
	itemSubGroup.Post("/show-item-sub-group", itemGroupController.GetItemSubGroupByID)
	itemSubGroup.Post("/create-item-sub-group", itemGroupController.CreateItemSubGroup)
	itemSubGroup.Post("/update-item-sub-group", itemGroupController.UpdateItemSubGroup)
	itemSubGroup.Post("/delete-item-sub-group", itemGroupController.DeleteItemSubGroup)
	itemSubGroup.Post("/restore-item-sub-group", itemGroupController.RestoreItemSubGroup)
}
