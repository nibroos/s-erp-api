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

func SetupTicketRoutes(tickets fiber.Router, gormDB *gorm.DB, sqlDB *sqlx.DB, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) {
	ticketRepo := repository.NewTicketRepository(gormDB, sqlDB, utilRepo, tracer)
	ticketService := service.NewTicketService(ticketRepo, utilRepo, tracer)
	ticketController := rest.NewTicketController(ticketService, ticketRepo, tracer)

	// tickets.Post("/index-ticket", middleware.PermissionMiddleware("read_masters"), ticketController.GetTickets)
	// tickets.Post("/index-project-app", ticketController.GetProjectsApp)
	tickets.Post("/show-project-app", ticketController.GetTicketByID)
	tickets.Post("/index-ticket", ticketController.GetTickets)
	tickets.Post("/widget-ticket", ticketController.GetWidgetTickets)
	tickets.Post("/create-ticket", ticketController.CreateTicket)
	tickets.Post("/update-ticket", ticketController.UpdateTicket)
	tickets.Post("/delete-ticket", ticketController.DeleteTicket)
	tickets.Post("/restore-ticket", ticketController.RestoreTicket)
	tickets.Post("/excel-ticket", ticketController.ExcelGetTickets)
	tickets.Post("/csv-ticket", ticketController.CsvGetTickets)

	tickets.Post("/index-calendar", ticketController.GetCalendars)
	tickets.Post("/index-schedule-app", ticketController.GetCalendars)
	tickets.Post("/create-schedule", ticketController.CreateScheduleSingle)
	tickets.Post("/show-schedule", ticketController.GetScheduleByID)
	tickets.Post("/delete-schedule", ticketController.DeleteSchedule)
	tickets.Post("/show-ticket", ticketController.GetTicketByID)
	tickets.Post("/show-schedule-app", ticketController.GetScheduleAppByID)
	tickets.Post("/update-schedule", ticketController.UpdateScheduleTicket)
	tickets.Post("/update-ticket-schedule", ticketController.UpdateScheduleTicket)
	tickets.Post("/update-ticket-schedule-app", ticketController.UpdateScheduleTicketApp)
	tickets.Post("/update-ticket-schedule-app-upload", ticketController.UpdateScheduleTicketAppUpload)
	tickets.Post("/update-schedule-app", ticketController.UpdateScheduleTicketApp)
	tickets.Post("/update-schedule-app-upload", ticketController.UpdateScheduleTicketAppUpload)
}
