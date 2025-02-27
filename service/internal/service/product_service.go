package service

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type ProductService struct {
	repo     *repository.ProductRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewProductService(repo *repository.ProductRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *ProductService {
	return &ProductService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *ProductService) GetProducts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.ProductListDTO, int, error) {
	childSpan := opentracing.StartSpan("ProductService-GetProducts", opentracing.ChildOf(span.Context()))

	products, total, err := s.repo.GetProducts(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return products, total, nil
}

func (s *ProductService) CreateProduct(ctx *fiber.Ctx, product *models.Product, tx *gorm.DB, span opentracing.Span) (*models.Product, error) {
	childSpan := opentracing.StartSpan("ProductService-CreateProduct", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateProduct(tx, product, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return product, nil
}

func (s *ProductService) GetProductByID(ctx *fiber.Ctx, params *dtos.GetProductParams, span opentracing.Span) (*dtos.ProductDetailDTO, error) {
	childSpan := opentracing.StartSpan("ProductService-GetProductByID", opentracing.ChildOf(span.Context()))

	product, err := s.repo.GetProductByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return product, nil
}

func (s *ProductService) UpdateProduct(ctx *fiber.Ctx, product *models.Product, tx *gorm.DB, span opentracing.Span) (*models.Product, error) {
	childSpan := opentracing.StartSpan("ProductService-UpdateProduct", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateProduct(tx, product, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return product, nil
}

func (s *ProductService) DeleteProduct(ctx *fiber.Ctx, params *dtos.GetProductParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ProductService-DeleteProduct", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteProduct(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *ProductService) RestoreProduct(ctx *fiber.Ctx, params *dtos.GetProductParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ProductService-RestoreProduct", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreProduct(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *ProductService) ExcelGetProducts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("ProductService-ExcelGetProducts", opentracing.ChildOf(span.Context()))

	products, _, err := s.GetProducts(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "products"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("Products", "A1", &[]string{"ID", "Branch", "Code", "Factory Code", "Name", "Sku", "Barcode", "Unit", "Specification", "Desc", "Remark", "Price Sell", "Price Buy"})

	for i, product := range products {
		row := []interface{}{
			product.ID,
			*product.BranchName,
			utils.GetPtrVal(product.Code),
			utils.GetPtrVal(product.FactoryCode),
			product.Name,
			utils.GetPtrVal(product.Sku),
			utils.GetPtrVal(product.Barcode),
			utils.GetPtrVal(product.UnitName),
			utils.GetPtrVal(product.Specification),
			utils.GetPtrVal(product.Description),
			utils.GetPtrVal(product.Remark),
			utils.GetPtrVal(product.PriceSell),
			utils.GetPtrVal(product.PriceBuy),
		}
		file.SetSheetRow("Products", fmt.Sprintf("A%d", i+2), &row)
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
func (s *ProductService) CsvGetProducts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("ProductService-CsvGetProducts", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	products, _, err := s.GetProducts(ctx, filters, childSpan)
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
	csv += "Master Product\n"
	csv += "\n"

	csv += "ID,Branch,Code,Factory Code,Name,Sku,Barcode,Unit,Specification,Desc,Remark,Price Sell,Price Buy\n"
	// Build CSV rows
	for _, product := range products {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			product.ID,
			*product.BranchName,
			utils.GetPtrVal(product.Code),
			utils.GetPtrVal(product.FactoryCode),
			product.Name,
			utils.GetPtrVal(product.Sku),
			utils.GetPtrVal(product.Barcode),
			utils.GetPtrVal(product.UnitName),
			utils.GetPtrVal(product.Specification),
			utils.GetPtrVal(product.Description),
			utils.GetPtrVal(product.Remark),
			utils.GetPtrVal(product.PriceSell),
			utils.GetPtrVal(product.PriceBuy),
		)
	}

	return []byte(csv), nil
}

// product *models.Product
func (s *ProductService) CreateBoms(ctx *fiber.Ctx, boms []*models.Bom, productID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ProductService-CreateBoms", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateBoms(tx, boms, productID, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// bulk create/update boms for a product
func (s *ProductService) BulkCreateUpdateBoms(ctx *fiber.Ctx, boms []*models.Bom, productID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ProductService-BulkCreateBoms", opentracing.ChildOf(span.Context()))

	// filter without ID to bulk create
	bulkCreateBoms := []*models.Bom{}
	// filter with ID to bulk update
	bulkUpdateBoms := []*models.Bom{}
	// get all ids
	bomIDs := []uint{}

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	for _, bom := range boms {
		if bom.ID == nil {
			bom.CreatedByID = &userID
			bom.CreatedAt = time.Now()
			bulkCreateBoms = append(bulkCreateBoms, bom)
		} else {
			bulkUpdateBoms = append(bulkUpdateBoms, bom)
			bomIDs = append(bomIDs, *bom.ID)
		}
	}

	// delete boms that are not in the list
	if len(bomIDs) > 0 {
		if err := s.repo.DeleteBomsWhereNotIn(tx, productID, bomIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(bulkCreateBoms) > 0 {
		if err := s.repo.CreateBoms(tx, bulkCreateBoms, productID, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(bulkUpdateBoms) > 0 {
		if err := s.repo.UpdateBoms(tx, bulkUpdateBoms, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	return nil
}

func (s *ProductService) GetBomsByProductID(ctx *fiber.Ctx, productID uint, span opentracing.Span) ([]dtos.ProductBomListDTO, error) {
	childSpan := opentracing.StartSpan("ProductService-GetBomsByProductID", opentracing.ChildOf(span.Context()))

	boms, err := s.repo.GetBomsByProductID(ctx, productID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return boms, nil
}
