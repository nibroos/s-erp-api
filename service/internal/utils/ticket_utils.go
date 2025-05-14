package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/config"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/opentracing/opentracing-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

func GenTicketNo(ctx *fiber.Ctx, req dtos.FormTicketRequest, orderedNumber int, globalOrderedNumber int, span opentracing.Span) string {
	if req.TicketNo != nil {
		return *req.TicketNo
	}

	// SURNAME-YEAR-MONTH-ORDER-REV-(NUM) -> SURNAME-2001-12-20-REV-1
	// surname := req.CustomerCode
	surname := "Yubi"
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	order := fmt.Sprintf("%d", orderedNumber)
	orderGlobal := fmt.Sprintf("%d", globalOrderedNumber)

	// str := fmt.Sprintf("%s-%s-%s-%s", surname, year, month, order)
	str := fmt.Sprintf("SO/%s/%s-%s-%s-%s", orderGlobal, surname, year, month, order)

	return str
}

func MapCreateUpdateTicket(ctx *fiber.Ctx, req dtos.FormTicketRequest, id *uint, userID uint, branchID uint, customerSoCreatedThisMonthNumber int, globalSoCreatedThisMonthNumber int, span opentracing.Span) (models.Ticket, error) {
	ticketNo := ""
	revNo := 0

	ticket := models.Ticket{
		CustomerID:    req.CustomerID,
		ProductID:     req.ProductID,
		Title:         req.Title,
		IssueDesc:     req.IssueDesc,
		IssueSolution: req.IssueSolution,
		ReportedAt:    req.ReportedAt,
		PriorityType:  req.PriorityType,
		RevNo:         &revNo,
		TicketNo:      &ticketNo,
		Remark:        req.Remark,
		Status:        req.Status,
		BranchID:      &branchID,
	}

	if id != nil {
		if req.RevNo != nil {
			revNo = *req.RevNo
		}
		revNo++
		// GenerateTicketNoOnCreateTicket
		ticketNo = GenerateTicketNoOnUpdateTicket(ctx, req, req.TicketNo, revNo, span)
		ticket.ID = id
		ticket.TicketNo = &ticketNo
		ticket.RevNo = &revNo
		ticket.UpdatedByID = &userID
	} else {
		customerSoCreatedThisMonthNumber++
		globalSoCreatedThisMonthNumber++
		ticketNo = GenerateTicketNoOnCreateTicket(ctx, req, customerSoCreatedThisMonthNumber, globalSoCreatedThisMonthNumber, span)
		ticket.TicketNoOri = &ticketNo
		ticket.CreatedByID = &userID
	}

	return ticket, nil
}

func GenerateTicketNoOnUpdateTicket(ctx *fiber.Ctx, req dtos.FormTicketRequest, ticketNo *string, revNo int, span opentracing.Span) string {

	// get before REV-number, full string is SURNAME-YEAR-MONTH-ORDER-REV-(NUM) -> SURNAME-2001-12-20-REV-1 or SURNAME-2001-12-20
	// check if "REV" string exist (random), if not add "REV-1" else replace REV-1 change the number to increment REV-revNo
	if !strings.Contains(*ticketNo, "REV") {
		*ticketNo = fmt.Sprintf("%s/REV-1", *ticketNo)
	} else {
		// remove after /REV
		// use split to get the first part of string
		*ticketNo = strings.Split(*ticketNo, "/REV")[0]
		*ticketNo = fmt.Sprintf("%s/REV-%d", *ticketNo, revNo)
	}

	return *ticketNo
}

func GenerateTicketNoOnCreateTicket(ctx *fiber.Ctx, req dtos.FormTicketRequest, orderedNumber int, globalOrderedNumber int, span opentracing.Span) string {
	if req.TicketNo != nil {
		return *req.TicketNo
	}

	// SURNAME-YEAR-MONTH-ORDER-REV-(NUM) -> SURNAME-2001-12-20-REV-1
	// surname := req.CustomerCode
	surname := "Yubi"
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	order := fmt.Sprintf("%d", orderedNumber)
	orderGlobal := fmt.Sprintf("%d", globalOrderedNumber)

	// str := fmt.Sprintf("%s-%s-%s-%s", surname, year, month, order)
	str := fmt.Sprintf("%s/%s/%s-%s-%s", surname, orderGlobal, year, month, order)

	return str
}

func MapCreateUpdateSchedule(ctx *fiber.Ctx, req dtos.UpdateScheduleRequest, userID uint, span opentracing.Span) (models.Schedule, error) {
	totalTaskStep4Done := 0
	totalAllTasksDone := 0
	totalTasks := 0
	totalTasks4 := 0

	for iStep, reqStep := range req.Steps {
		for _, reqTask := range reqStep.Tasks {
			totalTasks += 1

			if reqTask.IsChecked != nil && *reqTask.IsChecked == 1 {
				totalAllTasksDone += 1
			}

			if iStep == 3 {
				totalTasks4 += 1
			}

			if iStep == 3 && reqTask.IsChecked != nil && *reqTask.IsChecked == 1 {
				totalTaskStep4Done += 1
			}
		}
	}

	scheduleTask := models.Schedule{
		// ID:                 &req.ID,
		AssigneeID:         req.AssigneeID,
		SalesOrderID:       *req.SalesOrderID,
		CustomerID:         req.CustomerID,
		UUID:               req.UUID,
		Title:              req.Title,
		ModuleType:         req.ModuleType,
		Remark:             req.Remark,
		Status:             "WAITING",
		StartAt:            req.StartAt,
		EndAt:              req.EndAt,
		Color:              req.Color,
		CreatedByID:        &userID,
		TotalTaskStep4Done: &totalTaskStep4Done,
		TotalAllTasksDone:  &totalAllTasksDone,
		TotalTasks:         &totalTasks,
		TotalTasks4:        &totalTasks4,
	}

	if req.ID != nil {
		scheduleTask.ID = req.ID
	}

	return scheduleTask, nil
}

// type UpdateTicketAttachmentsDTO struct {
// 	ID       *uint   `json:"id" db:"id"`
// 	RefID    *uint   `json:"ref_id" db:"ref_id"`
// 	RefType  *string `json:"ref_type" db:"ref_type"`
// 	FileType *string `json:"file_type" db:"file_type"`
// 	FileUrl  *string `json:"file_url" db:"file_url"`
// 	FileName *string `json:"file_name" db:"file_name"`
// 	Remark   *string `json:"remark" db:"remark"`
// }

func PublishSendEmailSolutionTicket(ctx *fiber.Ctx, rabbitmq *config.RabbitMQ, data dtos.FormTicketRequest) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return rabbitmq.Channel.PublishWithContext(ctx.Context(),
		"",                                 // exchange
		"send_email_solution_ticket_queue", // routing key (queue name)
		false,                              // mandatory
		false,                              // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
}
