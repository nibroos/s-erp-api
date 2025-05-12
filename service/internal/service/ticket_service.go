package service

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type TicketService struct {
	repo     *repository.TicketRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewTicketService(repo *repository.TicketRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *TicketService {
	return &TicketService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *TicketService) GetTickets(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.TicketListDTO, int, error) {
	childSpan := opentracing.StartSpan("TicketService-GetTickets", opentracing.ChildOf(span.Context()))

	tickets, total, err := s.repo.GetTickets(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return tickets, total, nil
}

func (s *TicketService) CreateTicket(ctx *fiber.Ctx, req dtos.FormTicketRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.Ticket, *gorm.DB, error) {
	childSpan := opentracing.StartSpan("TicketService-CreateTicket", opentracing.ChildOf(span.Context()))

	customerSoCreatedThisMonthNumber, err := s.repo.GetCustomerTicketCreatedThisMonth(ctx, tx, *req.CustomerID, childSpan)
	globalSoCreatedThisMonthNumber, err := s.repo.GetGlobalTicketCreatedThisMonth(ctx, tx, *req.CustomerID, childSpan)

	ticket, err := utils.MapCreateUpdateTicket(ctx, req, nil, userID, branchID, customerSoCreatedThisMonthNumber, globalSoCreatedThisMonthNumber, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, tx, err
	}

	if tx, err = s.repo.CreateTicket(tx, &ticket, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	// files, err := ctx.FormFile("files")
	issueFiles := form.File["issue_files"]
	if len(issueFiles) > 0 {
		propJson := map[string]interface{}{}
		propJson["attachment_type"] = "issue"
		propJson["ref_type"] = "tickets"

		// handle new issueFiles upload
		newFiles, err := utils.MapNewSalesOrderFiles(ctx, issueFiles, *ticket.ID, userID, propJson, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, tx, err
		}

		// create new letters
		if tx, err = s.repo.CreateTicketFiles(ctx, tx, newFiles, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, tx, err
		}
	}

	solutionFiles := form.File["solution_files"]
	if len(solutionFiles) > 0 {
		propJson := map[string]interface{}{}
		propJson["attachment_type"] = "solution"
		propJson["ref_type"] = "tickets"

		// handle new solutionFiles upload
		newFiles, err := utils.MapNewSalesOrderFiles(ctx, issueFiles, *ticket.ID, userID, propJson, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, tx, err
		}

		// create new letters
		if tx, err = s.repo.CreateTicketFiles(ctx, tx, newFiles, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, tx, err
		}
	}
	// go routine to create schedule entity related to sales order
	// go func() {

	if req.Schedule != nil {
		req.Schedule.SalesOrderID = ticket.ID
		req.Schedule.CustomerID = ticket.CustomerID
		req.Schedule.ModuleType = "tickets"
		tx, err = s.CreateSchedule(ctx, *req.Schedule, &ticket, userID, tx, childSpan)

		if err != nil {
			childSpan.Finish()
			tx.Rollback()
			return nil, tx, err
		}
	}
	// }()

	return &ticket, tx, nil
}

// CreateSchedule
func (s *TicketService) CreateSchedule(ctx *fiber.Ctx, req dtos.UpdateScheduleRequest, ticket *models.Ticket, userID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("TicketService-CreateSchedule", opentracing.ChildOf(span.Context()))

	schedule, err := utils.MapCreateUpdateSchedule(ctx, req, userID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return tx, err
	}

	if tx, err = s.repo.CreateSchedule(ctx, tx, &schedule, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return tx, err
	}

	// create steps
	if len(req.Steps) > 0 {
		steps, err := utils.MapCreateScheduleSteps(ctx, req.Steps, *schedule.ID, userID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return tx, err
		}

		if tx, err = s.repo.CreateScheduleSteps(ctx, tx, steps, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return tx, err
		}

		// create schedule task
		tasks, err := utils.MapCreateScheduleTasks(ctx, req.Steps, steps, *schedule.ID, userID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return tx, err
		}

		if tx, err = s.repo.CreateScheduleTasks(ctx, tx, tasks, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return tx, err
		}
	}

	return tx, nil
}

// CreateSchedule
func (s *TicketService) CreateScheduleNoRef(ctx *fiber.Ctx, req dtos.CreateScheduleNoRefRequest, userID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("TicketService-CreateScheduleNoRef", opentracing.ChildOf(span.Context()))

	schedule, err := utils.MapCreateScheduleNoRef(ctx, req, userID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return tx, err
	}

	if tx, err = s.repo.CreateSchedule(ctx, tx, &schedule, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return tx, err
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return tx, err
	}

	propJson := map[string]interface{}{}

	// files, err := ctx.FormFile("files")
	files := form.File["files"]
	log.Println("files 1", files)
	if len(files) > 0 {
		// handle new files upload
		newFiles, err := utils.MapNewSalesOrderFiles(ctx, files, *schedule.ID, userID, propJson, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return tx, err
		}

		// create new letters
		if tx, err = s.repo.CreateTicketFiles(ctx, tx, newFiles, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return tx, err
		}
	}

	// create steps
	if len(req.Steps) > 0 {
		steps, err := utils.MapCreateScheduleSteps(ctx, req.Steps, *schedule.ID, userID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return tx, err
		}

		if tx, err = s.repo.CreateScheduleSteps(ctx, tx, steps, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return tx, err
		}

		// create schedule task
		tasks, err := utils.MapCreateScheduleTasks(ctx, req.Steps, steps, *schedule.ID, userID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return tx, err
		}

		if tx, err = s.repo.CreateScheduleTasks(ctx, tx, tasks, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return tx, err
		}
	}

	return tx, nil
}

func (s *TicketService) GetTicketByID(ctx *fiber.Ctx, params *dtos.GetTicketParams, tx *gorm.DB, span opentracing.Span) (*dtos.TicketDetailDTO, error) {
	childSpan := opentracing.StartSpan("TicketService-GetTicketByID", opentracing.ChildOf(span.Context()))

	ticket, err := s.repo.GetTicketByID(ctx, params, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return ticket, nil
}

// get schedule by sales order id
func (s *TicketService) GetScheduleByTicketID(ctx *fiber.Ctx, params *dtos.GetTicketParams, tx *gorm.DB, span opentracing.Span) (*dtos.ScheduleDetailDTO, error) {
	childSpan := opentracing.StartSpan("TicketService-GetScheduleByTicketID", opentracing.ChildOf(span.Context()))

	schedule, err := s.repo.GetScheduleByTicketID(ctx, params, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	scheduleIDs := []uint{schedule.ID}
	filters := map[string]string{}

	scheduleTasks, err := s.repo.GetScheduleTasksByScheduleID(ctx, filters, scheduleIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	// map steps
	steps, err := utils.MapGetScheduleStepsTasks(ctx, scheduleTasks, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	schedule.Steps = steps

	return schedule, nil
}

// get schedule by sales order id
func (s *TicketService) GetScheduleByID(ctx *fiber.Ctx, params *dtos.GetTicketParams, tx *gorm.DB, span opentracing.Span) (*dtos.ScheduleSingleDetailDTO, error) {
	childSpan := opentracing.StartSpan("TicketService-GetScheduleByID", opentracing.ChildOf(span.Context()))

	schedule, err := s.repo.GetScheduleByID(ctx, params, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	scheduleIDs := []uint{schedule.ID}
	filters := map[string]string{}

	scheduleTasks, err := s.repo.GetScheduleTasksByScheduleID(ctx, filters, scheduleIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	// map steps
	steps, err := utils.MapGetScheduleStepsTasks(ctx, scheduleTasks, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	schedule.Steps = steps

	// scheduleIDs := []uint{uint(schedule.ID)}
	// filters := map[string]string{}

	// scheduleTasks, err := s.repo.GetScheduleTasksByScheduleID(ctx, filters, scheduleIDs, childSpan)
	// if err != nil {
	// 	defer childSpan.Finish()
	// 	return nil, err
	// }

	return schedule, nil
}

func (s *TicketService) UpdateTicket(ctx *fiber.Ctx, req dtos.FormTicketRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.Ticket, error) {
	childSpan := opentracing.StartSpan("TicketService-UpdateTicket", opentracing.ChildOf(span.Context()))

	customerSoCreatedThisMonthNumber, err := s.repo.GetCustomerTicketCreatedThisMonth(ctx, tx, *req.CustomerID, childSpan)
	globalSoCreatedThisMonthNumber, err := s.repo.GetGlobalTicketCreatedThisMonth(ctx, tx, *req.CustomerID, childSpan)

	ticket, err := utils.MapCreateUpdateTicket(ctx, req, req.ID, userID, branchID, customerSoCreatedThisMonthNumber, globalSoCreatedThisMonthNumber, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	if err := s.repo.UpdateTicket(tx, &ticket, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	// update desc
	if len(req.IssueAttachments) > 0 {
		attachments := utils.MapUpdateSalesOrderAttachments(ctx, req.IssueAttachments, *ticket.ID, userID, childSpan)

		if tx, err := s.repo.UpdateAttachmentsDesc(ctx, tx, attachments, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(req.DeletedIssueFiles) > 0 {
		if err := s.repo.DeleteTicketFilesByIDs(ctx, tx, req.DeletedIssueFiles, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	// update desc
	if len(req.SolutionAttachments) > 0 {
		attachments := utils.MapUpdateSalesOrderAttachments(ctx, req.SolutionAttachments, *ticket.ID, userID, childSpan)

		if tx, err := s.repo.UpdateAttachmentsDesc(ctx, tx, attachments, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(req.DeletedSolutionFiles) > 0 {
		if err := s.repo.DeleteTicketFilesByIDs(ctx, tx, req.DeletedSolutionFiles, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	// files, err := ctx.FormFile("files")
	issueFiles := form.File["issue_files"]
	log.Println("issueFiles 1", issueFiles)
	if len(issueFiles) > 0 {
		propJson := map[string]interface{}{}
		propJson["attachment_type"] = "issue"
		propJson["ref_type"] = "tickets"
		// handle new issueFiles upload
		newFiles, err := utils.MapNewSalesOrderFiles(ctx, issueFiles, *ticket.ID, userID, propJson, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}

		// create new letters
		if tx, err = s.repo.CreateTicketFiles(ctx, tx, newFiles, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	// files, err := ctx.FormFile("files")
	solutionFiles := form.File["solution_files"]
	if len(solutionFiles) > 0 {
		propJson := map[string]interface{}{}
		propJson["attachment_type"] = "solution"
		propJson["ref_type"] = "tickets"
		// handle new solutionFiles upload
		newFiles, err := utils.MapNewSalesOrderFiles(ctx, solutionFiles, *ticket.ID, userID, propJson, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}

		// create new letters
		if tx, err = s.repo.CreateTicketFiles(ctx, tx, newFiles, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	return &ticket, nil
}

func (s *TicketService) DeleteTicket(ctx *fiber.Ctx, params *dtos.GetTicketParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TicketService-DeleteTicket", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteTicket(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *TicketService) DeleteScheduleTasksByScheduleID(ctx *fiber.Ctx, params *dtos.GetTicketParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TicketService-DeleteScheduleTasksByScheduleID", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteScheduleTasksByScheduleID(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *TicketService) DeleteScheduleByID(ctx *fiber.Ctx, params *dtos.GetTicketParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TicketService-DeleteScheduleByID", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteScheduleByID(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *TicketService) RestoreTicket(ctx *fiber.Ctx, params *dtos.GetTicketParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TicketService-RestoreTicket", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreTicket(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *TicketService) ExcelGetTickets(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("TicketService-ExcelGetTickets", opentracing.ChildOf(span.Context()))

	tickets, _, err := s.GetTickets(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "tickets"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("Tickets", "A1", &[]string{"ID", "Branch", "Code", "Factory Code", "Name", "Sku", "Barcode", "Unit", "Specification", "Desc", "Remark", "Price Sell", "Price Buy"})

	for i, ticket := range tickets {
		row := []interface{}{
			ticket.ID,
			utils.GetPtrVal(ticket.Remark),
		}
		file.SetSheetRow("Tickets", fmt.Sprintf("A%d", i+2), &row)
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
func (s *TicketService) CsvGetTickets(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("TicketService-CsvGetTickets", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	tickets, _, err := s.GetTickets(ctx, filters, childSpan)
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
	csv += "Master Ticket\n"
	csv += "\n"

	csv += "ID,Branch,Code,Factory Code,Name,Sku,Barcode,Unit,Specification,Desc,Remark,Price Sell,Price Buy\n"
	// Build CSV rows
	for _, ticket := range tickets {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			ticket.ID,
			utils.GetPtrVal(ticket.Remark),
		)
	}

	return []byte(csv), nil
}

// Lock all quotation table update
func (s *TicketService) LockTicketTable(ctx *fiber.Ctx, tx *gorm.DB, req dtos.FormTicketRequest, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("TicketService-LockTicketTable", opentracing.ChildOf(span.Context()))

	if req.ID != nil {
		if err := s.repo.LockTicketHeader(ctx, tx, req, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	return tx, nil
}

func (s *TicketService) UpdateTicketSchedule(ctx *fiber.Ctx, req dtos.UpdateSalesOrderScheduleRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.Schedule, error) {
	childSpan := opentracing.StartSpan("TicketService-UpdateTicketSchedule", opentracing.ChildOf(span.Context()))

	schedule, err := utils.MapUpdateSalesOrderSchedule(ctx, req, userID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	// delete schedule
	if schedule.ID == nil || *schedule.ID == 0 && req.IsDelete != nil && *req.IsDelete == 1 {
		if err := s.repo.DeleteTicketScheduleByTicketID(ctx, tx, req.SalesOrderID, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}

		return &schedule, nil
	}

	if schedule.ID == nil || *schedule.ID == 0 {
		ticket := &models.Ticket{
			ID: &req.SalesOrderID,
		}

		reqSchedule, err := utils.MapReqCreateSchedule(ctx, req, userID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}

		if tx, err := s.CreateSchedule(ctx, reqSchedule, ticket, userID, tx, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	log.Println("schedule.ID 1", schedule.ID)

	if schedule.ID != nil && *schedule.ID > 0 {
		if err := s.repo.UpdateTicketSchedule(tx, &schedule, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}

		form, err := ctx.MultipartForm()
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}

		// files, err := ctx.FormFile("files")
		issueFiles := form.File["issue_files"]
		if len(issueFiles) > 0 {
			propJson := map[string]interface{}{}
			propJson["attachment_type"] = "issue"
			propJson["ref_type"] = "tickets"
			// handle new issueFiles upload
			newFiles, err := utils.MapNewSalesOrderFiles(ctx, issueFiles, schedule.SalesOrderID, userID, propJson, childSpan)
			if err != nil {
				defer childSpan.Finish()
				tx.Rollback()
				return nil, err
			}

			// create new letters
			if tx, err = s.repo.CreateTicketFiles(ctx, tx, newFiles, childSpan); err != nil {
				defer childSpan.Finish()
				tx.Rollback()
				return nil, err
			}
		}

		// files, err := ctx.FormFile("files")
		solutionFiles := form.File["solution_files"]
		if len(solutionFiles) > 0 {
			propJson := map[string]interface{}{}
			propJson["attachment_type"] = "solution"
			propJson["ref_type"] = "tickets"
			// handle new solutionFiles upload
			newFiles, err := utils.MapNewSalesOrderFiles(ctx, solutionFiles, schedule.SalesOrderID, userID, propJson, childSpan)
			if err != nil {
				defer childSpan.Finish()
				tx.Rollback()
				return nil, err
			}

			// create new letters
			if tx, err = s.repo.CreateTicketFiles(ctx, tx, newFiles, childSpan); err != nil {
				defer childSpan.Finish()
				tx.Rollback()
				return nil, err
			}
		}

		if len(req.IssueAttachments) > 0 {
			attachments := utils.MapUpdateSalesOrderAttachments(ctx, req.IssueAttachments, schedule.SalesOrderID, userID, childSpan)

			if tx, err := s.repo.UpdateAttachmentsDesc(ctx, tx, attachments, childSpan); err != nil {
				defer childSpan.Finish()
				tx.Rollback()
				return nil, err
			}
		}

		if len(req.SolutionAttachments) > 0 {
			attachments := utils.MapUpdateSalesOrderAttachments(ctx, req.SolutionAttachments, schedule.SalesOrderID, userID, childSpan)

			if tx, err := s.repo.UpdateAttachmentsDesc(ctx, tx, attachments, childSpan); err != nil {
				defer childSpan.Finish()
				tx.Rollback()
				return nil, err
			}
		}

		if len(req.DeletedIssueFiles) > 0 {
			if err := s.repo.DeleteTicketFilesByIDs(ctx, tx, req.DeletedIssueFiles, childSpan); err != nil {
				defer childSpan.Finish()
				tx.Rollback()
				return nil, err
			}
		}

		if len(req.DeletedSolutionFiles) > 0 {
			if err := s.repo.DeleteTicketFilesByIDs(ctx, tx, req.DeletedSolutionFiles, childSpan); err != nil {
				defer childSpan.Finish()
				tx.Rollback()
				return nil, err
			}
		}

		// Bulk/Create Update Batch Steps
		scheduleSteps, err := utils.MapUpdateScheduleSteps(ctx, req.Steps, userID, *schedule.ID, childSpan)

		tx, err = s.BulkCreateUpdateScheduleSteps(ctx, req, &schedule, userID, scheduleSteps, *schedule.ID, tx, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
		log.Println("schedule.ID 2", schedule.ID)
		updatedTicketScheduleIDs := make([]uint, 0)
		updatedTicketScheduleIDs = append(updatedTicketScheduleIDs, *schedule.ID)

		// get updated steps
		updatedScheduleSteps, err := s.repo.GetUpdatedScheduleStepsByTicketScheduleIDs(ctx, tx, updatedTicketScheduleIDs, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
		log.Println("schedule.ID 3", schedule.ID)

		// Bulk/Create Update Batch Tasks
		err = s.BulkCreateUpdateScheduleTasks(ctx, updatedScheduleSteps, req, *schedule.ID, tx, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	return &schedule, nil
}

func (s *TicketService) UpdateSchedule(ctx *fiber.Ctx, req dtos.UpdateScheduleRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.Schedule, error) {
	childSpan := opentracing.StartSpan("TicketService-UpdateSchedule", opentracing.ChildOf(span.Context()))

	ticketSchedule, err := utils.MapUpdateSchedule(ctx, req, userID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	if ticketSchedule.ID != nil && *ticketSchedule.ID > 0 {
		if err := s.repo.UpdateTicketSchedule(tx, &ticketSchedule, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}

		// Bulk/Create Update Batch Steps
		scheduleSteps, err := utils.MapUpdateScheduleSteps(ctx, req.Steps, userID, *ticketSchedule.ID, childSpan)

		tx, err = s.BulkCreateUpdateSingleScheduleSteps(ctx, req, &ticketSchedule, userID, scheduleSteps, *ticketSchedule.ID, tx, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}

		updatedTicketScheduleIDs := make([]uint, 0)
		updatedTicketScheduleIDs = append(updatedTicketScheduleIDs, *ticketSchedule.ID)

		// get updated steps
		updatedScheduleSteps, err := s.repo.GetUpdatedScheduleStepsByTicketScheduleIDs(ctx, tx, updatedTicketScheduleIDs, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}

		// Bulk/Create Update Batch Tasks
		err = s.BulkCreateUpdateSingleScheduleTasks(ctx, updatedScheduleSteps, req, *ticketSchedule.ID, tx, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	return &ticketSchedule, nil
}

func (s *TicketService) UpdateTicketScheduleApp(ctx *fiber.Ctx, req dtos.UpdateSalesOrderScheduleAppRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TicketService-UpdateTicketScheduleApp", opentracing.ChildOf(span.Context()))

	// updatedTicketScheduleIDs := make([]uint, 0)
	// updatedTicketScheduleIDs = append(updatedTicketScheduleIDs, req.TicketID)

	// Bulk/Create Update Batch Tasks
	err := s.BulkCreateUpdateScheduleTasksApp(ctx, req, req.ScheduleID, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	// update desc
	if len(req.Attachments) > 0 {
		attachments := utils.MapUpdateSalesOrderAttachments(ctx, req.Attachments, req.ScheduleID, userID, childSpan)

		if tx, err := s.repo.UpdateAttachmentsDesc(ctx, tx, attachments, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(req.IssueAttachments) > 0 {
		attachments := utils.MapUpdateSalesOrderAttachments(ctx, req.IssueAttachments, req.ScheduleID, userID, childSpan)

		if tx, err := s.repo.UpdateAttachmentsDesc(ctx, tx, attachments, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(req.SolutionAttachments) > 0 {
		attachments := utils.MapUpdateSalesOrderAttachments(ctx, req.SolutionAttachments, req.ScheduleID, userID, childSpan)

		if tx, err := s.repo.UpdateAttachmentsDesc(ctx, tx, attachments, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(req.DeletedFiles) > 0 {
		if err := s.repo.DeleteTicketFilesByIDs(ctx, tx, req.DeletedFiles, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(req.DeletedIssueFiles) > 0 {
		if err := s.repo.DeleteTicketFilesByIDs(ctx, tx, req.DeletedIssueFiles, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(req.DeletedSolutionFiles) > 0 {
		if err := s.repo.DeleteTicketFilesByIDs(ctx, tx, req.DeletedSolutionFiles, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	// form, err := ctx.MultipartForm()
	// if err != nil {
	// 	defer childSpan.Finish()
	// 	tx.Rollback()
	// 	return nil
	// }

	// files := form.File["files"]
	// if len(files) > 0 {
	// 	// handle new files upload
	// 	newFiles, err := utils.MapNewSalesOrderFiles(ctx, files, req.TicketID, userID, childSpan)
	// 	if err != nil {
	// 		defer childSpan.Finish()
	// 		tx.Rollback()
	// 		return nil
	// 	}

	// 	// create new letters
	// 	if tx, err = s.repo.CreateTicketFiles(ctx, tx, newFiles, childSpan); err != nil {
	// 		defer childSpan.Finish()
	// 		tx.Rollback()
	// 		return nil
	// 	}
	// }

	return nil
}

func (s *TicketService) UpdateTicketScheduleAppUpload(ctx *fiber.Ctx, req dtos.UpdateSalesOrderScheduleAppRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TicketService-UpdateTicketScheduleAppUpload", opentracing.ChildOf(span.Context()))

	form, err := ctx.MultipartForm()
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil
	}

	files := form.File["files"]
	log.Println("files 1", files)
	if len(files) > 0 {
		log.Println("files > 0", files, req)

		propJson := map[string]interface{}{
			"ref_type": "tickets",
		}

		// handle new files upload
		newFiles, err := utils.MapNewSalesOrderFilesApp(ctx, files, req, req.ScheduleID, userID, propJson, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil
		}

		// create new letters
		if tx, err = s.repo.CreateTicketFiles(ctx, tx, newFiles, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil
		}
	}

	return nil
}

// bulk create/update boms for a quotation
func (s *TicketService) BulkCreateUpdateScheduleSteps(ctx *fiber.Ctx, req dtos.UpdateSalesOrderScheduleRequest, updatedTicketSchedule *models.Schedule, userID uint, steps []*models.ScheduleTask, scheduleID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("TicketService-BulkCreateUpdateScheduleSteps", opentracing.ChildOf(span.Context()))

	// Bulk/Create Update Batch ScheduleSteps
	scheduleSteps, err := utils.MapUpdateScheduleSteps(ctx, req.Steps, userID, *updatedTicketSchedule.ID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	// filter without ID to bulk create
	bulkCreateScheduleSteps := []*models.ScheduleTask{}
	// filter with ID to bulk update
	bulkUpdateScheduleSteps := []*models.ScheduleTask{}
	// get all ids
	scheduleStepIDs := []uint{}

	for _, scheduleStep := range scheduleSteps {
		if scheduleStep.ID == 0 {
			scheduleStep.CreatedByID = &userID
			scheduleStep.CreatedAt = time.Now()
			bulkCreateScheduleSteps = append(bulkCreateScheduleSteps, scheduleStep)
		} else {
			bulkUpdateScheduleSteps = append(bulkUpdateScheduleSteps, scheduleStep)
			scheduleStepIDs = append(scheduleStepIDs, scheduleStep.ID)
		}
	}

	// delete scheduleSteps that are not in the list
	if len(scheduleStepIDs) > 0 {
		if tx, err := s.repo.DeleteScheduleStepsWhereNotIn(ctx, tx, scheduleID, scheduleStepIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(bulkCreateScheduleSteps) > 0 {
		if tx, err := s.repo.CreateScheduleSteps(ctx, tx, bulkCreateScheduleSteps, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(bulkUpdateScheduleSteps) > 0 {
		if tx, err := s.repo.UpdateScheduleSteps(ctx, tx, bulkUpdateScheduleSteps, userID, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	return tx, nil
}

// bulk create/update boms for a quotation
func (s *TicketService) BulkCreateUpdateSingleScheduleSteps(ctx *fiber.Ctx, req dtos.UpdateScheduleRequest, updatedTicketSchedule *models.Schedule, userID uint, steps []*models.ScheduleTask, scheduleID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("TicketService-BulkCreateUpdateSingleScheduleSteps", opentracing.ChildOf(span.Context()))

	// Bulk/Create Update Batch ScheduleSteps
	scheduleSteps, err := utils.MapUpdateScheduleSteps(ctx, req.Steps, userID, *updatedTicketSchedule.ID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	// filter without ID to bulk create
	bulkCreateScheduleSteps := []*models.ScheduleTask{}
	// filter with ID to bulk update
	bulkUpdateScheduleSteps := []*models.ScheduleTask{}
	// get all ids
	scheduleStepIDs := []uint{}

	for _, scheduleStep := range scheduleSteps {
		if scheduleStep.ID == 0 {
			scheduleStep.CreatedByID = &userID
			scheduleStep.CreatedAt = time.Now()
			bulkCreateScheduleSteps = append(bulkCreateScheduleSteps, scheduleStep)
		} else {
			bulkUpdateScheduleSteps = append(bulkUpdateScheduleSteps, scheduleStep)
			scheduleStepIDs = append(scheduleStepIDs, scheduleStep.ID)
		}
	}

	// delete scheduleSteps that are not in the list
	if len(scheduleStepIDs) > 0 {
		if tx, err := s.repo.DeleteScheduleStepsWhereNotIn(ctx, tx, scheduleID, scheduleStepIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(bulkCreateScheduleSteps) > 0 {
		if tx, err := s.repo.CreateScheduleSteps(ctx, tx, bulkCreateScheduleSteps, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(bulkUpdateScheduleSteps) > 0 {
		if tx, err := s.repo.UpdateScheduleSteps(ctx, tx, bulkUpdateScheduleSteps, userID, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	return tx, nil
}

// bulk create/update boms for a quotation
func (s *TicketService) BulkCreateUpdateScheduleTasks(ctx *fiber.Ctx, steps []dtos.UpdatedScheduleStepListDTO, req dtos.UpdateSalesOrderScheduleRequest, scheduleID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TicketService-BulkCreateUpdateScheduleTasks", opentracing.ChildOf(span.Context()))

	bulkCreateScheduleTasks, bulkUpdateScheduleTasks, taskIDs, err := utils.MapFilterUpdateScheduleTasksToSteps(ctx, steps, req, scheduleID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()

		return err
	}

	// delete soDts that are not in the list
	if len(taskIDs) > 0 {
		if err := s.repo.DeleteScheduleTasksWhereNotIn(ctx, tx, scheduleID, taskIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(bulkCreateScheduleTasks) > 0 {
		if tx, err := s.repo.CreateScheduleTasks(ctx, tx, bulkCreateScheduleTasks, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(bulkUpdateScheduleTasks) > 0 {
		if err := s.repo.UpdateScheduleTasks(tx, bulkUpdateScheduleTasks, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	return nil
}

// bulk create/update boms for a quotation
func (s *TicketService) BulkCreateUpdateSingleScheduleTasks(ctx *fiber.Ctx, steps []dtos.UpdatedScheduleStepListDTO, req dtos.UpdateScheduleRequest, scheduleID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TicketService-BulkCreateUpdateScheduleTasks", opentracing.ChildOf(span.Context()))

	bulkCreateScheduleTasks, bulkUpdateScheduleTasks, taskIDs, err := utils.MapFilterUpdateSingleScheduleTasksToSteps(ctx, steps, req, scheduleID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()

		return err
	}

	// delete soDts that are not in the list
	if len(taskIDs) > 0 {
		if err := s.repo.DeleteScheduleTasksWhereNotIn(ctx, tx, scheduleID, taskIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(bulkCreateScheduleTasks) > 0 {
		if tx, err := s.repo.CreateScheduleTasks(ctx, tx, bulkCreateScheduleTasks, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(bulkUpdateScheduleTasks) > 0 {
		if err := s.repo.UpdateScheduleTasks(tx, bulkUpdateScheduleTasks, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	return nil
}

// bulk create/update boms for a quotation
func (s *TicketService) BulkCreateUpdateScheduleTasksApp(ctx *fiber.Ctx, req dtos.UpdateSalesOrderScheduleAppRequest, scheduleID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("TicketService-BulkCreateUpdateScheduleTasksApp", opentracing.ChildOf(span.Context()))

	bulkUpdateScheduleTasks, err := utils.MapFilterUpdateScheduleTasksApp(ctx, req, scheduleID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()

		return err
	}

	// // delete soDts that are not in the list
	// if len(taskIDs) > 0 {
	// 	if err := s.repo.DeleteScheduleTasksWhereNotIn(ctx, tx, scheduleID, taskIDs, childSpan); err != nil {
	// 		defer childSpan.Finish()
	// 		tx.Rollback()
	// 		return err
	// 	}
	// }

	if len(bulkUpdateScheduleTasks) > 0 {
		if err := s.repo.UpdateScheduleTasks(tx, bulkUpdateScheduleTasks, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	scheduleTask, err := s.repo.GetScheduleTaskTotalDoneByID(ctx, tx, req.ScheduleID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return err
	}

	// // update schedule total done
	schedule, err := utils.MapUpdateScheduleApp(ctx, req, scheduleTask, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	if tx, err := s.repo.UpdateTicketScheduleApp(tx, schedule, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// GetAttachmentsByTicketID
func (s *TicketService) GetAttachmentsByTicketID(ctx *fiber.Ctx, tx *gorm.DB, ticketID uint, scheduleID uint, attachmentType string, span opentracing.Span) ([]dtos.SalesOrderAttachmentsDTO, error) {
	childSpan := opentracing.StartSpan("TicketService-GetAttachmentsByTicketID", opentracing.ChildOf(span.Context()))

	var combined []dtos.SalesOrderAttachmentsDTO
	attachments, err := s.repo.GetAttachmentsByTicketID(ctx, tx, ticketID, attachmentType, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	if len(attachments) > 0 {
		combined = append(combined, attachments...)
	}

	attachmentsSchedule, err := s.repo.GetAttachmentsByScheduleID(ctx, tx, scheduleID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	log.Println("attachments1", attachments)

	log.Println("ticketID, scheduleID", ticketID, scheduleID, attachmentsSchedule)

	// combine to attachments
	if len(attachmentsSchedule) > 0 {
		log.Println("attachmentsSchedule>0, attachments", attachments)
		attachmentsScheduleMapped := utils.MapAttachmentsScheduleToAttachments(attachmentsSchedule)
		log.Println("attachmentsScheduleMapped", attachmentsScheduleMapped)

		combined = append(combined, attachmentsScheduleMapped...)
	}

	log.Println("attachments", attachments)
	log.Println("combined", combined)

	return combined, nil
}

func (s *TicketService) GetCalendars(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.CalendarListDTO, int, error) {
	childSpan := opentracing.StartSpan("TicketService-GetCalendars", opentracing.ChildOf(span.Context()))

	tickets, total, err := s.repo.GetCalendars(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return tickets, total, nil
}

func (s *TicketService) CreateScheduleSingle(ctx *fiber.Ctx, req dtos.CreateScheduleNoRefRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("TicketService-CreateScheduleSingle", opentracing.ChildOf(span.Context()))

	tx, err := s.CreateScheduleNoRef(ctx, req, userID, tx, childSpan)

	if err != nil {
		childSpan.Finish()
		tx.Rollback()
		return tx, err
	}

	return tx, nil
}

// GetAttachmentsByTicketID
func (s *TicketService) GetAttachmentsByScheduleID(ctx *fiber.Ctx, tx *gorm.DB, scheduleID uint, span opentracing.Span) ([]dtos.ScheduleAttachmentsDTO, error) {
	childSpan := opentracing.StartSpan("TicketService-GetAttachmentsByScheduleID", opentracing.ChildOf(span.Context()))

	attachments, err := s.repo.GetAttachmentsByScheduleID(ctx, tx, scheduleID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	attachments = utils.MapAttachmentsToURL(attachments)

	return attachments, nil
}

func (s *TicketService) GetWidgetTickets(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.TicketStatusWidget, int, error) {
	childSpan := opentracing.StartSpan("TicketService-GetWidgetTickets", opentracing.ChildOf(span.Context()))

	tickets, total, err := s.repo.GetWidgetTickets(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return tickets, total, nil
}
