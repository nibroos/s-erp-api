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

type MsItemService struct {
	repo     *repository.MsItemRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewMsItemService(repo *repository.MsItemRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *MsItemService {
	return &MsItemService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *MsItemService) GetMsItems(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.MsItemListDTO, int, error) {
	childSpan := opentracing.StartSpan("MsItemService-GetMsItems", opentracing.ChildOf(span.Context()))

	msItems, total, err := s.repo.GetMsItems(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return msItems, total, nil
}

func (s *MsItemService) CreateMsItem(ctx *fiber.Ctx, msItem *models.MsItem, tx *gorm.DB, span opentracing.Span) (*models.MsItem, error) {
	childSpan := opentracing.StartSpan("MsItemService-CreateMsItem", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateMsItem(tx, msItem, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return msItem, nil
}

func (s *MsItemService) GetMsItemByID(ctx *fiber.Ctx, params *dtos.GetMsItemParams, span opentracing.Span) (*dtos.MsItemDetailDTO, error) {
	childSpan := opentracing.StartSpan("MsItemService-GetMsItemByID", opentracing.ChildOf(span.Context()))

	msItem, err := s.repo.GetMsItemByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return msItem, nil
}

func (s *MsItemService) UpdateMsItem(ctx *fiber.Ctx, msItem *models.MsItem, tx *gorm.DB, span opentracing.Span) (*models.MsItem, error) {
	childSpan := opentracing.StartSpan("MsItemService-UpdateMsItem", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateMsItem(tx, msItem, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return msItem, nil
}

func (s *MsItemService) DeleteMsItem(ctx *fiber.Ctx, params *dtos.GetMsItemParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("MsItemService-DeleteMsItem", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteMsItem(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *MsItemService) RestoreMsItem(ctx *fiber.Ctx, params *dtos.GetMsItemParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("MsItemService-RestoreMsItem", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreMsItem(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *MsItemService) ExcelGetMsItems(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("MsItemService-ExcelGetMsItems", opentracing.ChildOf(span.Context()))

	msItems, _, err := s.GetMsItems(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "msItems"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("MsItems", "A1", &[]string{"ID", "Name", "Sub Group", "Unit", "Specification", "TPB Code", "Price Sell", "Price Buy", "Minimum Stock", "Created At", "Updated At"})

	for i, msItem := range msItems {
		row := []interface{}{
			msItem.ID,
			msItem.Name,
			utils.GetPtrVal(msItem.ItemSubGroupName),
			utils.GetPtrVal(msItem.ItemSubGroupName),
			utils.GetPtrVal(msItem.UnitName),
			utils.GetPtrVal(msItem.Specification),
			utils.GetPtrVal(msItem.TpbCode),
			utils.GetPtrVal(msItem.PriceSell),
			utils.GetPtrVal(msItem.PriceBuy),
			utils.GetPtrVal(msItem.MinimumStock),
			msItem.CreatedAt,
			msItem.UpdatedAt,
		}
		file.SetSheetRow("MsItems", fmt.Sprintf("A%d", i+2), &row)
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
func (s *MsItemService) CsvGetMsItems(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("MsItemService-CsvGetMsItems", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	msItems, _, err := s.GetMsItems(ctx, filters, childSpan)
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
		log.Println("CsvGetMsItems error:", err)
	} else {
		appName = *companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "Master Item\n"
	csv += "\n"

	csv += "ID,Name,Sub Group,Unit,Specification,TPB Code,Price Sell,Price Buy,Minimum Stock\n"
	// Build CSV rows
	for _, msItem := range msItems {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s\n",
			msItem.ID,
			msItem.Name,
			utils.GetPtrVal(msItem.ItemSubGroupName),
			utils.GetPtrVal(msItem.UnitName),
			utils.GetPtrVal(msItem.Specification),
			utils.GetPtrVal(msItem.TpbCode),
			utils.GetPtrVal(msItem.PriceSell),
			utils.GetPtrVal(msItem.PriceBuy),
			utils.GetPtrVal(msItem.MinimumStock),
		)
	}

	return []byte(csv), nil
}

// msItem *models.MsItem
func (s *MsItemService) CreateItemUnits(ctx *fiber.Ctx, itemUnits []dtos.CreateMsItemUnitsRequest, msItemID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("MsItemService-CreateItemUnits", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateItemUnits(tx, itemUnits, msItemID, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}
