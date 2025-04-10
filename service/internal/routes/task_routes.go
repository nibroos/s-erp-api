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

func SetupTaskRoutes(tasks fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	taskRepo := repository.NewTaskRepository(gormDB, sqlDB, tracer)
	taskService := service.NewTaskService(taskRepo, utilRepo, tracer)
	taskController := rest.NewTaskController(taskService, taskRepo, tracer)

	// tasks.Post("/index-task", middleware.PermissionMiddleware("index-task"), taskController.GetTasks)
	tasks.Post("/index-task", taskController.GetTasks)
	tasks.Post("/show-task", taskController.GetTaskByID)
	tasks.Post("/create-task", taskController.CreateTask)
	tasks.Post("/update-task", taskController.UpdateTask)
	tasks.Post("/delete-task", taskController.DeleteTask)
	tasks.Post("/restore-task", taskController.RestoreTask)
	tasks.Post("/excel-task", taskController.ExcelGetTasks)
	tasks.Post("/csv-task", taskController.CsvGetTasks)
}
