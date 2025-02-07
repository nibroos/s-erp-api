package service

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
)

type ItemSubGroupService struct {
	repo *repository.ItemSubGroupRepository
}

func NewItemSubGroupService(repo *repository.ItemSubGroupRepository) *ItemSubGroupService {
	return &ItemSubGroupService{repo: repo}
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

func (s *ItemSubGroupService) CreateItemSubGroup(ctx context.Context, itemSubGroup *models.MixValue) (*models.MixValue, error) {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return nil, err
	}

	// Create itemSubGroup
	if err := s.repo.CreateItemSubGroup(tx, itemSubGroup); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return itemSubGroup, nil
}

func (s *ItemSubGroupService) GetItemSubGroupByID(ctx context.Context, id uint) (*dtos.ItemSubGroupDetailDTO, error) {
	itemSubGroupChan := make(chan *dtos.ItemSubGroupDetailDTO, 1)
	errChan := make(chan error, 1)

	go func() {
		itemSubGroup, err := s.repo.GetItemSubGroupByID(ctx, id)
		if err != nil {
			errChan <- err
			return
		}
		itemSubGroupChan <- itemSubGroup
	}()

	select {
	case itemSubGroup := <-itemSubGroupChan:
		return itemSubGroup, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *ItemSubGroupService) UpdateItemSubGroup(ctx context.Context, itemSubGroup *models.MixValue) (*models.MixValue, error) {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return nil, err
	}

	// Update itemSubGroup
	if err := s.repo.UpdateItemSubGroup(tx, itemSubGroup); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return itemSubGroup, nil
}

func (s *ItemSubGroupService) DeleteItemSubGroup(ctx context.Context, id uint) error {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return err
	}

	// Delete itemSubGroup
	if err := s.repo.DeleteItemSubGroup(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (s *ItemSubGroupService) RestoreItemSubGroup(ctx context.Context, id uint) error {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return err
	}

	// Restore itemSubGroup
	if err := s.repo.RestoreItemSubGroup(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}
