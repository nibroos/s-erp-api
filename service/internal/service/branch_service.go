package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type BranchService struct {
	repo   *repository.BranchRepository
	tracer opentracing.Tracer
}

func NewBranchService(repo *repository.BranchRepository, tracer opentracing.Tracer) *BranchService {
	return &BranchService{repo: repo, tracer: tracer}
}

func (s *BranchService) GetBranches(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.BranchListDTO, int, error) {
	childSpan := opentracing.StartSpan("BranchService-GetBranches")

	branches, total, err := s.repo.GetBranches(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return branches, total, nil
}

func (s *BranchService) CreateBranch(ctx *fiber.Ctx, branch *models.Branch, tx *gorm.DB, span opentracing.Span) (*models.Branch, error) {
	childSpan := opentracing.StartSpan("BranchService-CreateBranch", opentracing.ChildOf(span.Context()))

	// Create branch
	if err := s.repo.CreateBranch(tx, branch, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return branch, nil
}

func (s *BranchService) GetBranchByID(ctx *fiber.Ctx, params *dtos.GetBranchParams, span opentracing.Span) (*dtos.BranchDetailDTO, error) {
	childSpan := opentracing.StartSpan("BranchService-GetBranchByID", opentracing.ChildOf(span.Context()))

	branch, err := s.repo.GetBranchByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return branch, nil
}

func (s *BranchService) UpdateBranch(ctx *fiber.Ctx, branch *models.Branch, tx *gorm.DB, span opentracing.Span) (*models.Branch, error) {
	childSpan := opentracing.StartSpan("BranchService-UpdateBranch", opentracing.ChildOf(span.Context()))

	// Update branch
	if err := s.repo.UpdateBranch(tx, branch, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return branch, nil
}

func (s *BranchService) DeleteBranch(ctx *fiber.Ctx, params *dtos.GetBranchParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("BranchService-DeleteBranch", opentracing.ChildOf(span.Context()))
	// Delete branch
	if err := s.repo.DeleteBranch(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *BranchService) RestoreBranch(ctx *fiber.Ctx, params *dtos.GetBranchParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("BranchService-RestoreBranch", opentracing.ChildOf(span.Context()))
	// Restore branch
	if err := s.repo.RestoreBranch(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}
