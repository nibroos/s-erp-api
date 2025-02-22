package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
)

type ContactService struct {
	repo *repository.ContactRepository
}

func NewContactService(repo *repository.ContactRepository) *ContactService {
	return &ContactService{repo: repo}
}

func (s *ContactService) ListContacts(ctx *fiber.Ctx, filters map[string]string) ([]dtos.ContactListDTO, int, error) {
	contacts, total, err := s.repo.ListContacts(ctx, filters)
	if err != nil {
		return nil, 0, err
	}
	return contacts, total, nil
}

func (s *ContactService) CreateContact(ctx *fiber.Ctx, contact *models.Contact) (*models.Contact, error) {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return nil, err
	}

	// Create contact
	if err := s.repo.CreateContact(tx, contact); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return contact, nil
}

func (s *ContactService) GetContactByID(ctx *fiber.Ctx, params *dtos.GetContactParams) (*dtos.ContactDetailDTO, error) {
	contact, err := s.repo.GetContactByID(ctx, params)
	if err != nil {
		return nil, err
	}
	return contact, nil
}

func (s *ContactService) UpdateContact(ctx *fiber.Ctx, contact *models.Contact) (*models.Contact, error) {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return nil, err
	}

	// Update contact
	if err := s.repo.UpdateContact(tx, contact); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return contact, nil
}

func (s *ContactService) DeleteContact(ctx *fiber.Ctx, id uint) error {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return err
	}

	// Delete contact
	if err := s.repo.DeleteContact(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (s *ContactService) RestoreContact(ctx *fiber.Ctx, id uint) error {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return err
	}

	// Restore contact
	if err := s.repo.RestoreContact(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}
