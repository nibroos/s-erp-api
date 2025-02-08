package service

import (
	"context"
	"fmt"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/xuri/excelize/v2"
)

type ItemGroupService struct {
	repo *repository.ItemGroupRepository
}

func NewItemGroupService(repo *repository.ItemGroupRepository) *ItemGroupService {
	return &ItemGroupService{repo: repo}
}

func (s *ItemGroupService) GetItemGroups(ctx context.Context, filters map[string]string) ([]dtos.ItemGroupListDTO, int, error) {

	resultChan := make(chan dtos.GetItemGroupsResult, 1)

	go func() {
		itemGroups, total, err := s.repo.GetItemGroups(ctx, filters)
		resultChan <- dtos.GetItemGroupsResult{ItemGroups: itemGroups, Total: total, Err: err}
	}()

	select {
	case res := <-resultChan:
		return res.ItemGroups, res.Total, res.Err
	case <-ctx.Done():
		return nil, 0, ctx.Err()
	}
}

func (s *ItemGroupService) CreateItemGroup(ctx context.Context, itemGroup *models.MixValue) (*models.MixValue, error) {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return nil, err
	}

	// Create itemGroup
	if err := s.repo.CreateItemGroup(tx, itemGroup); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return itemGroup, nil
}

func (s *ItemGroupService) GetItemGroupByID(ctx context.Context, params *dtos.GetItemGroupParams) (*dtos.ItemGroupDetailDTO, error) {
	itemGroupChan := make(chan *dtos.ItemGroupDetailDTO, 1)
	errChan := make(chan error, 1)

	go func() {
		itemGroup, err := s.repo.GetItemGroupByID(ctx, params)
		if err != nil {
			errChan <- err
			return
		}
		itemGroupChan <- itemGroup
	}()

	select {
	case itemGroup := <-itemGroupChan:
		return itemGroup, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *ItemGroupService) UpdateItemGroup(ctx context.Context, itemGroup *models.MixValue) (*models.MixValue, error) {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return nil, err
	}

	// Update itemGroup
	if err := s.repo.UpdateItemGroup(tx, itemGroup); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return itemGroup, nil
}

func (s *ItemGroupService) DeleteItemGroup(ctx context.Context, params *dtos.GetItemGroupParams) error {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return err
	}

	// Delete itemGroup
	if err := s.repo.DeleteItemGroup(tx, params); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (s *ItemGroupService) RestoreItemGroup(ctx context.Context, params *dtos.GetItemGroupParams) error {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return err
	}

	// Restore itemGroup
	if err := s.repo.RestoreItemGroup(tx, params); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *ItemGroupService) GetItemGroupsExcel(ctx context.Context, filters map[string]string) ([]byte, error) {
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
