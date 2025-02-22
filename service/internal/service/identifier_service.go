package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
)

type IdentifierService struct {
	repo *repository.IdentifierRepository
}

func NewIdentifierService(repo *repository.IdentifierRepository) *IdentifierService {
	return &IdentifierService{repo: repo}
}

func (s *IdentifierService) ListIdentifiers(ctx *fiber.Ctx, filters map[string]string) ([]dtos.IdentifierListDTO, int, error) {
	identifiers, total, err := s.repo.ListIdentifiers(ctx, filters)
	if err != nil {
		return nil, 0, err
	}
	return identifiers, total, nil
}

func (s *IdentifierService) CreateIdentifier(ctx *fiber.Ctx, identifier *models.Identifier) (*models.Identifier, error) {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return nil, err
	}

	// Create identifier
	if err := s.repo.CreateIdentifier(tx, identifier); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return identifier, nil
}

func (s *IdentifierService) GetIdentifierByID(ctx *fiber.Ctx, params *dtos.GetIdentifierParams) (*dtos.IdentifierDetailDTO, error) {
	identifier, err := s.repo.GetIdentifierByID(ctx, params)
	if err != nil {
		return nil, err
	}
	return identifier, nil
}

func (s *IdentifierService) UpdateIdentifier(ctx *fiber.Ctx, identifier *models.Identifier) (*models.Identifier, error) {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return nil, err
	}

	// Update identifier
	if err := s.repo.UpdateIdentifier(tx, identifier); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return identifier, nil
}

func (s *IdentifierService) DeleteIdentifier(ctx *fiber.Ctx, id uint) error {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return err
	}

	// Delete identifier
	if err := s.repo.DeleteIdentifier(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (s *IdentifierService) RestoreIdentifier(ctx *fiber.Ctx, id uint) error {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return err
	}

	// Restore identifier
	if err := s.repo.RestoreIdentifier(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (s *IdentifierService) ListIdentifiersByAuthUser(ctx *fiber.Ctx, filters map[string]string) ([]dtos.IdentifierListDTO, int, error) {
	identifiers, total, err := s.repo.ListIdentifiers(ctx, filters)
	if err != nil {
		return nil, 0, err
	}
	return identifiers, total, nil
}
