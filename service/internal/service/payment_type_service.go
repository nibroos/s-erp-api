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

type PaymentTypeService struct {
	repo     *repository.PaymentTypeRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewPaymentTypeService(repo *repository.PaymentTypeRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *PaymentTypeService {
	return &PaymentTypeService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *PaymentTypeService) GetPaymentTypes(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.PaymentTypeListDTO, int, error) {
	childSpan := opentracing.StartSpan("PaymentTypeService-GetPaymentTypes", opentracing.ChildOf(span.Context()))

	paymentTypes, total, err := s.repo.GetPaymentTypes(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return paymentTypes, total, nil
}

func (s *PaymentTypeService) CreatePaymentType(ctx *fiber.Ctx, paymentType *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("PaymentTypeService-CreatePaymentType", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreatePaymentType(tx, paymentType, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return paymentType, nil
}

func (s *PaymentTypeService) GetPaymentTypeByID(ctx *fiber.Ctx, params *dtos.GetPaymentTypeParams, span opentracing.Span) (*dtos.PaymentTypeDetailDTO, error) {
	childSpan := opentracing.StartSpan("PaymentTypeService-GetPaymentTypeByID", opentracing.ChildOf(span.Context()))

	paymentType, err := s.repo.GetPaymentTypeByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return paymentType, nil
}

func (s *PaymentTypeService) UpdatePaymentType(ctx *fiber.Ctx, paymentType *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("PaymentTypeService-UpdatePaymentType", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdatePaymentType(tx, paymentType, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return paymentType, nil
}

func (s *PaymentTypeService) DeletePaymentType(ctx *fiber.Ctx, params *dtos.GetPaymentTypeParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PaymentTypeService-DeletePaymentType", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeletePaymentType(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *PaymentTypeService) RestorePaymentType(ctx *fiber.Ctx, params *dtos.GetPaymentTypeParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PaymentTypeService-RestorePaymentType", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestorePaymentType(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *PaymentTypeService) ExcelGetPaymentTypes(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("PaymentTypeService-ExcelGetPaymentTypes", opentracing.ChildOf(span.Context()))

	paymentTypes, _, err := s.GetPaymentTypes(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "paymentTypes"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("PaymentTypes", "A1", &[]string{"ID", "Name", "Description", "Created At", "Updated At"})

	for i, paymentType := range paymentTypes {
		row := []interface{}{
			paymentType.ID,
			paymentType.Name,
			paymentType.Description,
			paymentType.CreatedAt,
			paymentType.UpdatedAt,
		}
		file.SetSheetRow("PaymentTypes", fmt.Sprintf("A%d", i+2), &row)
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
func (s *PaymentTypeService) CsvGetPaymentTypes(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("PaymentTypeService-CsvGetPaymentTypes", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	paymentTypes, _, err := s.GetPaymentTypes(ctx, filters, childSpan)
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
	csv += "PaymentTypes\n"
	csv += "\n"

	csv += "ID,Name,Description,Remark,Created At,Updated At\n"
	// Build CSV rows
	for _, paymentType := range paymentTypes {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s\n",
			paymentType.ID,
			paymentType.Name,
			utils.GetPtrVal(paymentType.Description),
			utils.GetPtrVal(paymentType.Remark),
			utils.GetPtrVal(paymentType.CreatedAt),
			utils.GetPtrVal(paymentType.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
