package service

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type TaskService struct {
	repo     *repository.TaskRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewTaskService(repo *repository.TaskRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *TaskService {
	return &TaskService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *TaskService) GetTasks(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.TaskListDTO, int, error) {
	childSpan := opentracing.StartSpan("TaskService-GetTasks", opentracing.ChildOf(span.Context()))

	tasks, total, err := s.repo.GetTasks(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return tasks, total, nil
}

func (s *TaskService) CreateTask(ctx *fiber.Ctx, task *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("TaskService-CreateTask", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateTask(tx, task, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return task, nil
}

func (s *TaskService) GetTaskByID(ctx *fiber.Ctx, params *dtos.GetTaskParams, span opentracing.Span) (*dtos.TaskDetailDTO, error) {
	childSpan := opentracing.StartSpan("TaskService-GetTaskByID", opentracing.ChildOf(span.Context()))

	task, err := s.repo.GetTaskByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return task, nil
}

func (s *TaskService) UpdateTask(ctx *fiber.Ctx, task *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("TaskService-UpdateTask", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateTask(tx, task, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return task, nil
}

func (s *TaskService) DeleteTask(ctx *fiber.Ctx, params *dtos.GetTaskParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TaskService-DeleteTask", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteTask(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *TaskService) RestoreTask(ctx *fiber.Ctx, params *dtos.GetTaskParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TaskService-RestoreTask", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreTask(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *TaskService) ExcelGetTasks(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("TaskService-ExcelGetTasks", opentracing.ChildOf(span.Context()))

	tasks, _, err := s.GetTasks(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "tasks"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("Tasks", "A1", &[]string{"ID", "Name", "Description", "Created At", "Updated At"})

	for i, task := range tasks {
		row := []interface{}{
			task.ID,
			task.Name,
			task.Description,
			task.CreatedAt,
			task.UpdatedAt,
		}
		file.SetSheetRow("Tasks", fmt.Sprintf("A%d", i+2), &row)
	}

	// Set active sheet of the workbook
	file.SetActiveSheet(index)

	// Save the file
	if err := file.SaveAs("output.xlsx"); err != nil {
		fmt.Println("Error saving file:", err)
		return nil, err
	}

	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// github.com/xuri/excelize/v2
func (s *TaskService) CsvGetTasks(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("TaskService-CsvGetTasks", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	tasks, _, err := s.GetTasks(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	// get company profile
	companyProfileParams := dtos.GetCompanyProfileParams{ID: 1}
	companyProfile, err := s.utilRepo.GetCompanyProfileByID(ctx, &companyProfileParams)
	appName := "App"
	if err != nil {
		defer childSpan.Finish()
	} else {
		appName = *companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "Tasks\n"
	csv += "\n"

	csv += "ID,Name,Description,Remark,Created At,Updated At\n"
	// Build CSV rows
	for _, task := range tasks {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s\n",
			task.ID,
			task.Name,
			utils.GetPtrVal(task.Description),
			utils.GetPtrVal(task.Remark),
			utils.GetPtrVal(task.CreatedAt),
			utils.GetPtrVal(task.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
