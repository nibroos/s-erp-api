package service

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type RoleService struct {
	repo     *repository.RoleRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewRoleService(repo *repository.RoleRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *RoleService {
	return &RoleService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *RoleService) GetRoles(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RoleListDTO, int, error) {
	childSpan := opentracing.StartSpan("RoleService-GetRoles", opentracing.ChildOf(span.Context()))

	roles, total, err := s.repo.GetRoles(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return roles, total, nil
}

func (s *RoleService) CreateRole(ctx *fiber.Ctx, role *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("RoleService-CreateRole", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateRole(tx, role, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return role, nil
}

func (s *RoleService) GetRoleByID(ctx *fiber.Ctx, params *dtos.GetRoleParams, span opentracing.Span) (*dtos.RoleDetailDTO, error) {
	childSpan := opentracing.StartSpan("RoleService-GetRoleByID", opentracing.ChildOf(span.Context()))

	role, err := s.repo.GetRoleByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return role, nil
}

func (s *RoleService) UpdateRole(ctx *fiber.Ctx, role *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("RoleService-UpdateRole", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateRole(tx, role, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return role, nil
}

func (s *RoleService) DeleteRole(ctx *fiber.Ctx, params *dtos.GetRoleParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("RoleService-DeleteRole", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteRole(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *RoleService) RestoreRole(ctx *fiber.Ctx, params *dtos.GetRoleParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("RoleService-RestoreRole", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreRole(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *RoleService) ExcelGetRoles(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("RoleService-ExcelGetRoles", opentracing.ChildOf(span.Context()))

	roles, _, err := s.GetRoles(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "roles"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("Roles", "A1", &[]string{"ID", "Name", "Description", "Created At", "Updated At"})

	for i, role := range roles {
		row := []interface{}{
			role.ID,
			role.Name,
			role.Description,
			role.CreatedAt,
			role.UpdatedAt,
		}
		file.SetSheetRow("Roles", fmt.Sprintf("A%d", i+2), &row)
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
func (s *RoleService) CsvGetRoles(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("RoleService-CsvGetRoles", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	roles, _, err := s.GetRoles(ctx, filters, childSpan)
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
	csv += "Item Groups\n"
	csv += "\n"

	csv += "ID,Name,Description,Remark,Created At,Updated At\n"
	// Build CSV rows
	for _, role := range roles {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s\n",
			role.ID,
			role.Name,
			utils.GetPtrVal(role.Description),
			utils.GetPtrVal(role.Remark),
			utils.GetPtrVal(role.CreatedAt),
			utils.GetPtrVal(role.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
