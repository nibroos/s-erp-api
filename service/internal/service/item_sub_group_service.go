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

type ItemSubGroupService struct {
	repo     *repository.ItemSubGroupRepository
	utilRepo *repository.UtilRepository
}

func NewItemSubGroupService(repo *repository.ItemSubGroupRepository, utilRepo *repository.UtilRepository) *ItemSubGroupService {
	return &ItemSubGroupService{repo: repo, utilRepo: utilRepo}
}

func (s *ItemSubGroupService) GetItemSubGroups(ctx context.Context, filters map[string]string) ([]dtos.ItemSubGroupListDTO, int, error) {

	resultChan := make(chan dtos.GetItemSubGroupsResult, 1)

	go func() {
		itemSubGroups, total, err := s.repo.GetItemSubGroups(ctx, filters)
		resultChan <- dtos.GetItemSubGroupsResult{ItemSubGroups: itemSubGroups, Total: total, Err: err}
	}()

	select {
	case res := <-resultChan:
		return res.ItemSubGroups, res.Total, res.Err
	case <-ctx.Done():
		return nil, 0, ctx.Err()
	}
}

func (s *ItemSubGroupService) CreateItemSubGroup(ctx context.Context, itemSubGroup *models.MixValue, tx *gorm.DB) (*models.MixValue, error) {
	if err := s.repo.CreateItemSubGroup(tx, itemSubGroup); err != nil {
		tx.Rollback()
		return nil, err
	}

	return itemSubGroup, nil
}

func (s *ItemSubGroupService) GetItemSubGroupByID(ctx context.Context, params *dtos.GetItemSubGroupParams) (*dtos.ItemSubGroupDetailDTO, error) {
	itemSubGroup, err := s.repo.GetItemSubGroupByID(ctx, params)
	if err != nil {
		return nil, err
	}
	return itemSubGroup, nil
}

func (s *ItemSubGroupService) UpdateItemSubGroup(ctx context.Context, itemSubGroup *models.MixValue, tx *gorm.DB) (*models.MixValue, error) {
	if err := s.repo.UpdateItemSubGroup(tx, itemSubGroup); err != nil {
		tx.Rollback()
		return nil, err
	}

	return itemSubGroup, nil
}

func (s *ItemSubGroupService) DeleteItemSubGroup(ctx context.Context, id uint, tx *gorm.DB) error {
	if err := s.repo.DeleteItemSubGroup(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *ItemSubGroupService) RestoreItemSubGroup(ctx context.Context, id uint, tx *gorm.DB) error {
	if err := s.repo.RestoreItemSubGroup(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *ItemSubGroupService) ExcelGetItemSubGroups(ctx context.Context, filters map[string]string) ([]byte, error) {
	itemSubGroups, _, err := s.GetItemSubGroups(ctx, filters)
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

	file.SetSheetRow("ItemSubGroups", "A1", &[]string{"ID", "Name", "Description", "Created At", "Updated At"})

	for i, itemSubGroup := range itemSubGroups {
		row := []interface{}{
			itemSubGroup.ID,
			itemSubGroup.Name,
			itemSubGroup.Description,
			itemSubGroup.CreatedAt,
			itemSubGroup.UpdatedAt,
		}
		file.SetSheetRow("ItemSubGroups", fmt.Sprintf("A%d", i+2), &row)
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
func (s *ItemSubGroupService) CsvGetItemSubGroups(ctx context.Context, filters map[string]string) ([]byte, error) {
	// filters is_csv
	filters["is_csv"] = "1"
	itemSubGroups, _, err := s.GetItemSubGroups(ctx, filters)
	if err != nil {
		return nil, err
	}

	// get company profile
	companyProfileParams := dtos.GetCompanyProfileParams{ID: 1}
	companyProfile, err := s.utilRepo.GetCompanyProfileByID(ctx, &companyProfileParams)
	appName := "App"
	if err != nil {
		log.Println("CsvGetItemSubGroups error:", err)
	} else {
		appName = companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "Item Sub Groups\n"
	csv += "\n"

	csv += "ID,Name,Description,Remark,Created At,Updated At\n"
	// Build CSV rows
	for _, itemSubGroup := range itemSubGroups {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s\n",
			itemSubGroup.ID,
			itemSubGroup.Name,
			utils.GetPtrVal(&itemSubGroup.Description),
			utils.GetPtrVal(itemSubGroup.Remark),
			utils.GetPtrVal(itemSubGroup.CreatedAt),
			utils.GetPtrVal(itemSubGroup.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
