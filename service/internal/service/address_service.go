package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
)

type AddressService struct {
	repo *repository.AddressRepository
}

func NewAddressService(repo *repository.AddressRepository) *AddressService {
	return &AddressService{repo: repo}
}

func (s *AddressService) ListAddresses(ctx *fiber.Ctx, filters map[string]string) ([]dtos.AddressListDTO, int, error) {
	addresses, total, err := s.repo.ListAddresses(ctx, filters)
	if err != nil {
		return nil, 0, err
	}
	return addresses, total, nil
}

func (s *AddressService) CreateAddress(ctx *fiber.Ctx, address *models.Address) (*models.Address, error) {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return nil, err
	}

	// Create address
	if err := s.repo.CreateAddress(tx, address); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return address, nil
}

func (s *AddressService) GetAddressByID(ctx *fiber.Ctx, params *dtos.GetAddressParams) (*dtos.AddressDetailDTO, error) {
	address, err := s.repo.GetAddressByID(ctx, params)
	if err != nil {
		return nil, err
	}

	return address, nil
}

func (s *AddressService) UpdateAddress(ctx *fiber.Ctx, address *models.Address) (*models.Address, error) {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return nil, err
	}

	// Update address
	if err := s.repo.UpdateAddress(tx, address); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return address, nil
}

func (s *AddressService) DeleteAddress(ctx *fiber.Ctx, id uint) error {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return err
	}

	// Delete address
	if err := s.repo.DeleteAddress(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (s *AddressService) RestoreAddress(ctx *fiber.Ctx, id uint) error {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return err
	}

	// Restore address
	if err := s.repo.RestoreAddress(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}
