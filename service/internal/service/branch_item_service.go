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

type BranchItemService struct {
	repo     *repository.BranchItemRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewBranchItemService(repo *repository.BranchItemRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *BranchItemService {
	return &BranchItemService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *BranchItemService) GetBranchItems(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.BranchItemListDTO, int, error) {
	childSpan := opentracing.StartSpan("BranchItemService-GetBranchItems", opentracing.ChildOf(span.Context()))

	branchItems, total, err := s.repo.GetBranchItems(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return branchItems, total, nil
}

func (s *BranchItemService) CreateBranchItem(ctx *fiber.Ctx, branchItem *models.BranchItem, tx *gorm.DB, span opentracing.Span) (*models.BranchItem, error) {
	childSpan := opentracing.StartSpan("BranchItemService-CreateBranchItem", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateBranchItem(tx, branchItem, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return branchItem, nil
}

func (s *BranchItemService) GetBranchItemByID(ctx *fiber.Ctx, params *dtos.GetBranchItemParams, span opentracing.Span) (*dtos.BranchItemDetailDTO, error) {
	childSpan := opentracing.StartSpan("BranchItemService-GetBranchItemByID", opentracing.ChildOf(span.Context()))

	branchItem, err := s.repo.GetBranchItemByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return branchItem, nil
}

func (s *BranchItemService) UpdateBranchItem(ctx *fiber.Ctx, branchItem *models.BranchItem, tx *gorm.DB, span opentracing.Span) (*models.BranchItem, error) {
	childSpan := opentracing.StartSpan("BranchItemService-UpdateBranchItem", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateBranchItem(tx, branchItem, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return branchItem, nil
}

func (s *BranchItemService) DeleteBranchItem(ctx *fiber.Ctx, params *dtos.GetBranchItemParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("BranchItemService-DeleteBranchItem", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteBranchItem(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *BranchItemService) RestoreBranchItem(ctx *fiber.Ctx, params *dtos.GetBranchItemParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("BranchItemService-RestoreBranchItem", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreBranchItem(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *BranchItemService) ExcelGetBranchItems(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("BranchItemService-ExcelGetBranchItems", opentracing.ChildOf(span.Context()))

	branchItems, _, err := s.GetBranchItems(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "branchItems"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("BranchItems", "A1", &[]string{"ID", "Name", "Branch Name", "Unit Name", "Specification", "Description", "TPB Code", "Minimum Stock", "Price Sell", "Price Buy"})

	for i, branchItem := range branchItems {
		row := []interface{}{
			branchItem.ID,
			utils.GetPtrVal(branchItem.Name),
			utils.GetPtrVal(branchItem.BranchName),
			utils.GetPtrVal(branchItem.UnitName),
			utils.GetPtrVal(branchItem.Specification),
			utils.GetPtrVal(branchItem.Description),
			utils.GetPtrVal(branchItem.TpbCode),
			utils.GetFloatPtrVal(branchItem.MinimumStock),
			utils.GetFloatPtrVal(branchItem.PriceSell),
			utils.GetFloatPtrVal(branchItem.PriceBuy),
			branchItem.CreatedAt,
			branchItem.UpdatedAt,
		}
		file.SetSheetRow("BranchItems", fmt.Sprintf("A%d", i+2), &row)
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
func (s *BranchItemService) CsvGetBranchItems(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("BranchItemService-CsvGetBranchItems", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	branchItems, _, err := s.GetBranchItems(ctx, filters, childSpan)
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
	csv += "Master Item\n"
	csv += "\n"

	csv += "ID,Name,Branch Name,Unit Name,Specification,Description,TPB Code,Minimum Stock,Price Sell,Price Buy,Created At,Updated At\n"
	// Build CSV rows
	for _, branchItem := range branchItems {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%f,%f,%f,%s,%s\n",
			branchItem.ID,
			utils.GetPtrVal(branchItem.Name),
			utils.GetPtrVal(branchItem.BranchName),
			utils.GetPtrVal(branchItem.UnitName),
			utils.GetPtrVal(branchItem.Specification),
			utils.GetPtrVal(branchItem.Description),
			utils.GetPtrVal(branchItem.TpbCode),
			utils.GetFloatPtrVal(branchItem.MinimumStock),
			utils.GetFloatPtrVal(branchItem.PriceSell),
			utils.GetFloatPtrVal(branchItem.PriceBuy),
			utils.GetPtrVal(branchItem.CreatedAt),
			utils.GetPtrVal(branchItem.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
