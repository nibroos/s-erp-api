package service

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type CustomerTypeService struct {
	repo     *repository.CustomerTypeRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewCustomerTypeService(repo *repository.CustomerTypeRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *CustomerTypeService {
	return &CustomerTypeService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *CustomerTypeService) GetCustomerTypes(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.CustomerTypeListDTO, int, error) {
	childSpan := opentracing.StartSpan("CustomerTypeService-GetCustomerTypes", opentracing.ChildOf(span.Context()))

	customerTypes, total, err := s.repo.GetCustomerTypes(ctx.Context(), filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return customerTypes, total, nil
}

func (s *CustomerTypeService) CreateCustomerType(ctx *fiber.Ctx, customerType *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("CustomerTypeService-CreateCustomerType", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateCustomerType(tx, customerType, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return customerType, nil
}

func (s *CustomerTypeService) GetCustomerTypeByID(ctx *fiber.Ctx, params *dtos.GetCustomerTypeParams, span opentracing.Span) (*dtos.CustomerTypeDetailDTO, error) {
	childSpan := opentracing.StartSpan("CustomerTypeService-GetCustomerTypeByID", opentracing.ChildOf(span.Context()))

	customerType, err := s.repo.GetCustomerTypeByID(ctx.Context(), params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return customerType, nil
}

func (s *CustomerTypeService) UpdateCustomerType(ctx *fiber.Ctx, customerType *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("CustomerTypeService-UpdateCustomerType", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateCustomerType(tx, customerType, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return customerType, nil
}

func (s *CustomerTypeService) DeleteCustomerType(ctx *fiber.Ctx, params *dtos.GetCustomerTypeParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CustomerTypeService-DeleteCustomerType", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteCustomerType(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *CustomerTypeService) RestoreCustomerType(ctx *fiber.Ctx, params *dtos.GetCustomerTypeParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CustomerTypeService-RestoreCustomerType", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreCustomerType(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *CustomerTypeService) ExcelGetCustomerTypes(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("CustomerTypeService-ExcelGetCustomerTypes", opentracing.ChildOf(span.Context()))

	customerTypes, _, err := s.GetCustomerTypes(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "customerTypes"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("CustomerTypes", "A1", &[]string{"ID", "Name", "Description", "Created At", "Updated At"})

	for i, customerType := range customerTypes {
		row := []interface{}{
			customerType.ID,
			customerType.Name,
			customerType.Description,
			customerType.CreatedAt,
			customerType.UpdatedAt,
		}
		file.SetSheetRow("CustomerTypes", fmt.Sprintf("A%d", i+2), &row)
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
func (s *CustomerTypeService) CsvGetCustomerTypes(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("CustomerTypeService-CsvGetCustomerTypes", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	customerTypes, _, err := s.GetCustomerTypes(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	// get company profile
	companyProfileParams := dtos.GetCompanyProfileParams{ID: 1}
	companyProfile, err := s.utilRepo.GetCompanyProfileByID(ctx.Context(), &companyProfileParams)
	appName := "App"
	if err != nil {
		defer childSpan.Finish()
		log.Println("CsvGetCustomerTypes error:", err)
	} else {
		appName = companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "CustomerTypes\n"
	csv += "\n"

	csv += "ID,Name,Description,Remark,Created At,Updated At\n"
	// Build CSV rows
	for _, customerType := range customerTypes {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s\n",
			customerType.ID,
			customerType.Name,
			utils.GetPtrVal(customerType.Description),
			utils.GetPtrVal(customerType.Remark),
			utils.GetPtrVal(customerType.CreatedAt),
			utils.GetPtrVal(customerType.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
