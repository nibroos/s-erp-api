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

type PaymentTermService struct {
	repo     *repository.PaymentTermRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewPaymentTermService(repo *repository.PaymentTermRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *PaymentTermService {
	return &PaymentTermService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *PaymentTermService) GetPaymentTerms(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.PaymentTermListDTO, int, error) {
	childSpan := opentracing.StartSpan("PaymentTermService-GetPaymentTerms", opentracing.ChildOf(span.Context()))

	paymentTerms, total, err := s.repo.GetPaymentTerms(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return paymentTerms, total, nil
}

func (s *PaymentTermService) GetPaymentTermByID(ctx *fiber.Ctx, params *dtos.GetPaymentTermParams, span opentracing.Span) (*dtos.PaymentTermDetailDTO, error) {
	childSpan := opentracing.StartSpan("PaymentTermService-GetPaymentTermByID", opentracing.ChildOf(span.Context()))

	term, err := s.repo.GetPaymentTermByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return term, nil
}

func (s *PaymentTermService) CreatePaymentTerm(ctx *fiber.Ctx, term *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("PaymentTermService-CreatePaymentTerm", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreatePaymentTerm(tx, term, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return term, nil
}

func (s *PaymentTermService) UpdatePaymentTerm(ctx *fiber.Ctx, term *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("PaymentTermService-UpdatePaymentTerm", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdatePaymentTerm(tx, term, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return term, nil
}

func (s *PaymentTermService) DeletePaymentTerm(ctx *fiber.Ctx, params *dtos.GetPaymentTermParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PaymentTermService-DeletePaymentTerm", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeletePaymentTerm(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *PaymentTermService) RestorePaymentTerm(ctx *fiber.Ctx, params *dtos.GetPaymentTermParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PaymentTermService-RestorePaymentTerm", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestorePaymentTerm(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *PaymentTermService) ExcelGetPaymentTerms(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("PaymentTermService-ExcelGetPaymentTerms", opentracing.ChildOf(span.Context()))

	paymentTerms, _, err := s.GetPaymentTerms(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	sheetName := "PaymentTerms"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("PaymentTerms", "A1", &[]string{"ID", "Name", "Description", "Created At", "Updated At"})

	for i, term := range paymentTerms {
		row := []interface{}{
			term.ID,
			term.Name,
			term.Description,
			term.CreatedAt,
			term.UpdatedAt,
		}
		file.SetSheetRow("PaymentTerms", fmt.Sprintf("A%d", i+2), &row)
	}

	file.SetActiveSheet(index)

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

func (s *PaymentTermService) CsvGetPaymentTerms(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("PaymentTermService-CsvGetPaymentTerms", opentracing.ChildOf(span.Context()))

	filters["is_csv"] = "1"
	paymentTerms, _, err := s.GetPaymentTerms(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

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
	csv += "Payment Terms\n"
	csv += "\n"

	csv += "ID,Name,Description,Remark,Created At,Updated At\n"
	for _, term := range paymentTerms {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s\n",
			term.ID,
			term.Name,
			utils.GetPtrVal(term.Description),
			utils.GetPtrVal(term.Remark),
			utils.GetPtrVal(term.CreatedAt),
			utils.GetPtrVal(term.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
