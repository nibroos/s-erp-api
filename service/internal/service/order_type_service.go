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

type OrderTypeService struct {
	repo     *repository.OrderTypeRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewOrderTypeService(repo *repository.OrderTypeRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *OrderTypeService {
	return &OrderTypeService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *OrderTypeService) GetOrderTypes(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.OrderTypeListDTO, int, error) {
	childSpan := opentracing.StartSpan("OrderTypeService-GetOrderTypes", opentracing.ChildOf(span.Context()))

	orderTypes, total, err := s.repo.GetOrderTypes(ctx.Context(), filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return orderTypes, total, nil
}

func (s *OrderTypeService) CreateOrderType(ctx *fiber.Ctx, orderType *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("OrderTypeService-CreateOrderType", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateOrderType(tx, orderType, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return orderType, nil
}

func (s *OrderTypeService) GetOrderTypeByID(ctx *fiber.Ctx, params *dtos.GetOrderTypeParams, span opentracing.Span) (*dtos.OrderTypeDetailDTO, error) {
	childSpan := opentracing.StartSpan("OrderTypeService-GetOrderTypeByID", opentracing.ChildOf(span.Context()))

	orderType, err := s.repo.GetOrderTypeByID(ctx.Context(), params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return orderType, nil
}

func (s *OrderTypeService) UpdateOrderType(ctx *fiber.Ctx, orderType *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("OrderTypeService-UpdateOrderType", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateOrderType(tx, orderType, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return orderType, nil
}

func (s *OrderTypeService) DeleteOrderType(ctx *fiber.Ctx, params *dtos.GetOrderTypeParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("OrderTypeService-DeleteOrderType", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteOrderType(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *OrderTypeService) RestoreOrderType(ctx *fiber.Ctx, params *dtos.GetOrderTypeParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("OrderTypeService-RestoreOrderType", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreOrderType(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *OrderTypeService) ExcelGetOrderTypes(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("OrderTypeService-ExcelGetOrderTypes", opentracing.ChildOf(span.Context()))

	orderTypes, _, err := s.GetOrderTypes(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "orderTypes"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("OrderTypes", "A1", &[]string{"ID", "Name", "Description", "Created At", "Updated At"})

	for i, orderType := range orderTypes {
		row := []interface{}{
			orderType.ID,
			orderType.Name,
			orderType.Description,
			orderType.CreatedAt,
			orderType.UpdatedAt,
		}
		file.SetSheetRow("OrderTypes", fmt.Sprintf("A%d", i+2), &row)
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
func (s *OrderTypeService) CsvGetOrderTypes(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("OrderTypeService-CsvGetOrderTypes", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	orderTypes, _, err := s.GetOrderTypes(ctx, filters, childSpan)
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
		log.Println("CsvGetOrderTypes error:", err)
	} else {
		appName = *companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "OrderTypes\n"
	csv += "\n"

	csv += "ID,Name,Description,Remark,Created At,Updated At\n"
	// Build CSV rows
	for _, orderType := range orderTypes {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s\n",
			orderType.ID,
			orderType.Name,
			utils.GetPtrVal(orderType.Description),
			utils.GetPtrVal(orderType.Remark),
			utils.GetPtrVal(orderType.CreatedAt),
			utils.GetPtrVal(orderType.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
