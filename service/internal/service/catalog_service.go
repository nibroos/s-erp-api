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

type CatalogService struct {
	repo     *repository.CatalogRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewCatalogService(repo *repository.CatalogRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *CatalogService {
	return &CatalogService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *CatalogService) GetCatalogs(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.CatalogListDTO, int, error) {
	childSpan := opentracing.StartSpan("CatalogService-GetCatalogs", opentracing.ChildOf(span.Context()))

	catalogs, total, err := s.repo.GetCatalogs(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return catalogs, total, nil
}

func (s *CatalogService) CreateCatalog(ctx *fiber.Ctx, catalog *models.Catalog, tx *gorm.DB, span opentracing.Span) (*models.Catalog, error) {
	childSpan := opentracing.StartSpan("CatalogService-CreateCatalog", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateCatalog(tx, catalog, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return catalog, nil
}

func (s *CatalogService) GetCatalogByID(ctx *fiber.Ctx, params *dtos.GetCatalogParams, span opentracing.Span) (*dtos.CatalogDetailDTO, error) {
	childSpan := opentracing.StartSpan("CatalogService-GetCatalogByID", opentracing.ChildOf(span.Context()))

	catalog, err := s.repo.GetCatalogByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return catalog, nil
}

func (s *CatalogService) UpdateCatalog(ctx *fiber.Ctx, catalog *models.Catalog, tx *gorm.DB, span opentracing.Span) (*models.Catalog, error) {
	childSpan := opentracing.StartSpan("CatalogService-UpdateCatalog", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateCatalog(tx, catalog, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return catalog, nil
}

func (s *CatalogService) DeleteCatalog(ctx *fiber.Ctx, params *dtos.GetCatalogParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CatalogService-DeleteCatalog", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteCatalog(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *CatalogService) RestoreCatalog(ctx *fiber.Ctx, params *dtos.GetCatalogParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CatalogService-RestoreCatalog", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreCatalog(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *CatalogService) ExcelGetCatalogs(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("CatalogService-ExcelGetCatalogs", opentracing.ChildOf(span.Context()))

	catalogs, _, err := s.GetCatalogs(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "catalogs"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("Catalogs", "A1", &[]string{"ID", "Name", "Sub Group", "Unit", "Specification", "TPB Code", "Price Sell", "Price Buy", "Minimum Stock", "Created At", "Updated At"})

	for i, catalog := range catalogs {
		row := []interface{}{
			catalog.ID,
			catalog.Name,
			utils.GetPtrVal(catalog.UnitName),
			utils.GetPtrVal(catalog.Specification),
			utils.GetPtrVal(catalog.PriceSell),
			utils.GetPtrVal(catalog.PriceBuy),
			catalog.CreatedAt,
			catalog.UpdatedAt,
		}
		file.SetSheetRow("Catalogs", fmt.Sprintf("A%d", i+2), &row)
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
func (s *CatalogService) CsvGetCatalogs(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("CatalogService-CsvGetCatalogs", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	catalogs, _, err := s.GetCatalogs(ctx, filters, childSpan)
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
		log.Println("CsvGetCatalogs error:", err)
	} else {
		appName = *companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "Master Item\n"
	csv += "\n"

	csv += "ID,Name,Sub Group,Unit,Specification,TPB Code,Price Sell,Price Buy,Minimum Stock\n"
	// Build CSV rows
	for _, catalog := range catalogs {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s,%v,%v,%v\n",
			catalog.ID,
			catalog.Name,
			utils.GetPtrVal(catalog.UnitName),
			utils.GetPtrVal(catalog.Specification),
			utils.GetPtrVal(catalog.PriceSell),
			utils.GetPtrVal(catalog.PriceBuy),
		)
	}

	return []byte(csv), nil
}

// catalog *models.Catalog
func (s *CatalogService) CreateBoms(ctx *fiber.Ctx, boms []*models.Bom, catalogID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CatalogService-CreateBoms", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateBoms(tx, boms, catalogID, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}
