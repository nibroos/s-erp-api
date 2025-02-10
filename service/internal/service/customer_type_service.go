package service

import (
	"context"
	"fmt"
	"log"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type CustomerTypeService struct {
	repo     *repository.CustomerTypeRepository
	utilRepo *repository.UtilRepository
}

func NewCustomerTypeService(repo *repository.CustomerTypeRepository, utilRepo *repository.UtilRepository) *CustomerTypeService {
	return &CustomerTypeService{
		repo:     repo,
		utilRepo: utilRepo,
	}
}

func (s *CustomerTypeService) GetCustomerTypes(ctx context.Context, filters map[string]string) ([]dtos.CustomerTypeListDTO, int, error) {
	customerTypes, total, err := s.repo.GetCustomerTypes(ctx, filters)
	if err != nil {
		return nil, 0, err
	}
	return customerTypes, total, nil
}

func (s *CustomerTypeService) CreateCustomerType(ctx context.Context, customerType *models.MixValue, tx *gorm.DB) (*models.MixValue, error) {
	if err := s.repo.CreateCustomerType(tx, customerType); err != nil {
		tx.Rollback()
		return nil, err
	}

	return customerType, nil
}

func (s *CustomerTypeService) GetCustomerTypeByID(ctx context.Context, params *dtos.GetCustomerTypeParams) (*dtos.CustomerTypeDetailDTO, error) {
	customerType, err := s.repo.GetCustomerTypeByID(ctx, params)
	if err != nil {
		return nil, err
	}
	return customerType, nil
}

func (s *CustomerTypeService) UpdateCustomerType(ctx context.Context, customerType *models.MixValue, tx *gorm.DB) (*models.MixValue, error) {
	if err := s.repo.UpdateCustomerType(tx, customerType); err != nil {
		tx.Rollback()
		return nil, err
	}

	return customerType, nil
}

func (s *CustomerTypeService) DeleteCustomerType(ctx context.Context, params *dtos.GetCustomerTypeParams, tx *gorm.DB) error {
	if err := s.repo.DeleteCustomerType(tx, params); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *CustomerTypeService) RestoreCustomerType(ctx context.Context, params *dtos.GetCustomerTypeParams, tx *gorm.DB) error {
	if err := s.repo.RestoreCustomerType(tx, params); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *CustomerTypeService) ExcelGetCustomerTypes(ctx context.Context, filters map[string]string) ([]byte, error) {
	customerTypes, _, err := s.GetCustomerTypes(ctx, filters)
	if err != nil {
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "customer-types"
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
func (s *CustomerTypeService) CsvGetCustomerTypes(ctx context.Context, filters map[string]string) ([]byte, error) {
	// filters is_csv
	filters["is_csv"] = "1"
	customerTypes, _, err := s.GetCustomerTypes(ctx, filters)
	if err != nil {
		return nil, err
	}

	// get company profile
	companyProfileParams := dtos.GetCompanyProfileParams{ID: 1}
	companyProfile, err := s.utilRepo.GetCompanyProfileByID(ctx, &companyProfileParams)
	appName := "App"
	if err != nil {
		log.Println("CsvGetCustomerTypes error:", err)
	} else {
		appName = companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "Item Groups\n"
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
