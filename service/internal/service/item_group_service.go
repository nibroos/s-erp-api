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

type ItemGroupService struct {
	repo     *repository.ItemGroupRepository
	utilRepo *repository.UtilRepository
}

func NewItemGroupService(repo *repository.ItemGroupRepository, utilRepo *repository.UtilRepository) *ItemGroupService {
	return &ItemGroupService{
		repo:     repo,
		utilRepo: utilRepo,
	}
}

func (s *ItemGroupService) GetItemGroups(ctx context.Context, filters map[string]string) ([]dtos.ItemGroupListDTO, int, error) {
	itemGroups, total, err := s.repo.GetItemGroups(ctx, filters)
	if err != nil {
		return nil, 0, err
	}
	return itemGroups, total, nil
}

func (s *ItemGroupService) CreateItemGroup(ctx context.Context, itemGroup *models.MixValue, tx *gorm.DB) (*models.MixValue, error) {
	if err := s.repo.CreateItemGroup(tx, itemGroup); err != nil {
		tx.Rollback()
		return nil, err
	}

	return itemGroup, nil
}

func (s *ItemGroupService) GetItemGroupByID(ctx context.Context, params *dtos.GetItemGroupParams) (*dtos.ItemGroupDetailDTO, error) {
	itemGroup, err := s.repo.GetItemGroupByID(ctx, params)
	if err != nil {
		return nil, err
	}
	return itemGroup, nil
}

func (s *ItemGroupService) UpdateItemGroup(ctx context.Context, itemGroup *models.MixValue, tx *gorm.DB) (*models.MixValue, error) {
	if err := s.repo.UpdateItemGroup(tx, itemGroup); err != nil {
		tx.Rollback()
		return nil, err
	}

	return itemGroup, nil
}

func (s *ItemGroupService) DeleteItemGroup(ctx context.Context, params *dtos.GetItemGroupParams, tx *gorm.DB) error {
	if err := s.repo.DeleteItemGroup(tx, params); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *ItemGroupService) RestoreItemGroup(ctx context.Context, params *dtos.GetItemGroupParams, tx *gorm.DB) error {
	if err := s.repo.RestoreItemGroup(tx, params); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *ItemGroupService) ExcelGetItemGroups(ctx context.Context, filters map[string]string) ([]byte, error) {
	itemGroups, _, err := s.GetItemGroups(ctx, filters)
	if err != nil {
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "item-groups"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("ItemGroups", "A1", &[]string{"ID", "Name", "Description", "Created At", "Updated At"})

	for i, itemGroup := range itemGroups {
		row := []interface{}{
			itemGroup.ID,
			itemGroup.Name,
			itemGroup.Description,
			itemGroup.CreatedAt,
			itemGroup.UpdatedAt,
		}
		file.SetSheetRow("ItemGroups", fmt.Sprintf("A%d", i+2), &row)
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
func (s *ItemGroupService) CsvGetItemGroups(ctx context.Context, filters map[string]string) ([]byte, error) {
	// filters is_csv
	filters["is_csv"] = "1"
	itemGroups, _, err := s.GetItemGroups(ctx, filters)
	if err != nil {
		return nil, err
	}

	// get company profile
	companyProfileParams := dtos.GetCompanyProfileParams{ID: 1}
	companyProfile, err := s.utilRepo.GetCompanyProfileByID(ctx, &companyProfileParams)
	appName := "App"
	if err != nil {
		log.Println("CsvGetItemGroups error:", err)
	} else {
		appName = companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "Item Groups\n"
	csv += "\n"

	csv += "ID,Name,Description,Remark,Created At,Updated At\n"
	// Build CSV rows
	for _, itemGroup := range itemGroups {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s\n",
			itemGroup.ID,
			itemGroup.Name,
			utils.GetPtrVal(itemGroup.Description),
			utils.GetPtrVal(itemGroup.Remark),
			utils.GetPtrVal(itemGroup.CreatedAt),
			utils.GetPtrVal(itemGroup.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
