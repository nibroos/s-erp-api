package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type EmailService struct {
	repo     *repository.EmailRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewEmailService(repo *repository.EmailRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *EmailService {
	return &EmailService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *EmailService) CreateEmail(ctx *fiber.Ctx, unit *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("EmailService-CreateEmail", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateEmail(tx, unit, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return unit, nil
}

// func (s *EmailService) GetEmails(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.EmailListDTO, int, error) {
// 	childSpan := opentracing.StartSpan("EmailService-GetEmails", opentracing.ChildOf(span.Context()))

// 	units, total, err := s.repo.GetEmails(ctx, filters, childSpan)
// 	if err != nil {
// 		defer childSpan.Finish()
// 		return nil, 0, err
// 	}
// 	return units, total, nil
// }

// func (s *EmailService) GetEmailByID(ctx *fiber.Ctx, params *dtos.GetEmailParams, span opentracing.Span) (*dtos.EmailDetailDTO, error) {
// 	childSpan := opentracing.StartSpan("EmailService-GetEmailByID", opentracing.ChildOf(span.Context()))

// 	unit, err := s.repo.GetEmailByID(ctx, params, childSpan)
// 	if err != nil {
// 		defer childSpan.Finish()
// 		return nil, err
// 	}
// 	return unit, nil
// }

// func (s *EmailService) UpdateEmail(ctx *fiber.Ctx, unit *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
// 	childSpan := opentracing.StartSpan("EmailService-UpdateEmail", opentracing.ChildOf(span.Context()))

// 	if err := s.repo.UpdateEmail(tx, unit, childSpan); err != nil {
// 		defer childSpan.Finish()
// 		tx.Rollback()
// 		return nil, err
// 	}

// 	return unit, nil
// }

// func (s *EmailService) DeleteEmail(ctx *fiber.Ctx, params *dtos.GetEmailParams, tx *gorm.DB, span opentracing.Span) error {
// 	childSpan := opentracing.StartSpan("EmailService-DeleteEmail", opentracing.ChildOf(span.Context()))

// 	if err := s.repo.DeleteEmail(tx, params, childSpan); err != nil {
// 		defer childSpan.Finish()
// 		tx.Rollback()
// 		return err
// 	}

// 	return nil
// }

// func (s *EmailService) RestoreEmail(ctx *fiber.Ctx, params *dtos.GetEmailParams, tx *gorm.DB, span opentracing.Span) error {
// 	childSpan := opentracing.StartSpan("EmailService-RestoreEmail", opentracing.ChildOf(span.Context()))

// 	if err := s.repo.RestoreEmail(tx, params, childSpan); err != nil {
// 		defer childSpan.Finish()
// 		tx.Rollback()
// 		return err
// 	}

// 	return nil
// }

// // github.com/xuri/excelize/v2
// func (s *EmailService) ExcelGetEmails(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
// 	childSpan := opentracing.StartSpan("EmailService-ExcelGetEmails", opentracing.ChildOf(span.Context()))

// 	units, _, err := s.GetEmails(ctx, filters, childSpan)
// 	if err != nil {
// 		defer childSpan.Finish()
// 		return nil, err
// 	}

// 	file := excelize.NewFile()

// 	// Create a new sheet
// 	sheetName := "units"
// 	index, err := file.NewSheet(sheetName)
// 	if err != nil {
// 		return nil, err
// 	}

// 	file.SetSheetRow("Emails", "A1", &[]string{"ID", "Name", "Description", "Created At", "Updated At"})

// 	for i, unit := range units {
// 		row := []interface{}{
// 			unit.ID,
// 			unit.Name,
// 			unit.Description,
// 			unit.CreatedAt,
// 			unit.UpdatedAt,
// 		}
// 		file.SetSheetRow("Emails", fmt.Sprintf("A%d", i+2), &row)
// 	}

// 	// Set active sheet of the workbook
// 	file.SetActiveSheet(index)

// 	// Save the file
// 	if err := file.SaveAs("output.xlsx"); err != nil {
// 		fmt.Println("Error saving file:", err)
// 		return nil, err
// 	}

// 	buffer, err := file.WriteToBuffer()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return buffer.Bytes(), nil
// }

// // github.com/xuri/excelize/v2
// func (s *EmailService) CsvGetEmails(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
// 	childSpan := opentracing.StartSpan("EmailService-CsvGetEmails", opentracing.ChildOf(span.Context()))

// 	// filters is_csv
// 	filters["is_csv"] = "1"
// 	units, _, err := s.GetEmails(ctx, filters, childSpan)
// 	if err != nil {
// 		defer childSpan.Finish()
// 		return nil, err
// 	}

// 	// get company profile
// 	companyProfileParams := dtos.GetCompanyProfileParams{ID: 1}
// 	companyProfile, err := s.utilRepo.GetCompanyProfileByID(ctx, &companyProfileParams)
// 	appName := "App"
// 	if err != nil {
// 		defer childSpan.Finish()
// 	} else {
// 		appName = *companyProfile.CompanyName
// 	}

// 	csv := fmt.Sprintf("%s\n", appName)
// 	csv += "\n"
// 	csv += "Emails\n"
// 	csv += "\n"

// 	csv += "ID,Name,Description,Remark,Created At,Updated At\n"
// 	// Build CSV rows
// 	for _, unit := range units {
// 		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s\n",
// 			unit.ID,
// 			unit.Name,
// 			utils.GetPtrVal(unit.Description),
// 			utils.GetPtrVal(unit.Remark),
// 			utils.GetPtrVal(unit.CreatedAt),
// 			utils.GetPtrVal(unit.UpdatedAt),
// 		)
// 	}

// 	return []byte(csv), nil
// }
