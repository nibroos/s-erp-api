package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/controller/rest"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"gorm.io/gorm"
)

func SetupItemGroupRoutes(itemGroups fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB) {
	itemGroupRepo := repository.NewItemGroupRepository(gormDB, sqlDB)
	itemGroupService := service.NewItemGroupService(itemGroupRepo)
	itemGroupController := rest.NewItemGroupController(itemGroupService)

	// prefix /itemGroups

	// itemGroups.Post("/index-item-group", middleware.PermissionMiddleware("index-item-group"), itemGroupController.GetItemGroups)
	itemGroups.Post("/index-item-group", itemGroupController.GetItemGroups)
	itemGroups.Post("/show-item-group", itemGroupController.GetItemGroupByID)
	itemGroups.Post("/create-item-group", itemGroupController.CreateItemGroup)
	itemGroups.Post("/update-item-group", itemGroupController.UpdateItemGroup)
	itemGroups.Post("/delete-item-group", itemGroupController.DeleteItemGroup)
	itemGroups.Post("/restore-item-group", itemGroupController.RestoreItemGroup)
}
