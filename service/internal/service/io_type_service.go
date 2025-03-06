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

type IOTypeService struct {
	repo     *repository.IOTypeRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewIOTypeService(repo *repository.IOTypeRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *IOTypeService {
	return &IOTypeService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *IOTypeService) GetIOTypes(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.IOTypeListDTO, int, error) {
	childSpan := opentracing.StartSpan("IOTypeService-GetIOTypes", opentracing.ChildOf(span.Context()))

	ioTypes, total, err := s.repo.GetIOTypes(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return ioTypes, total, nil
}

func (s *IOTypeService) GetIOTypeByID(ctx *fiber.Ctx, params *dtos.GetIOTypeParams, span opentracing.Span) (*dtos.IOTypeDetailDTO, error) {
	childSpan := opentracing.StartSpan("IOTypeService-GetIOTypeByID", opentracing.ChildOf(span.Context()))

	term, err := s.repo.GetIOTypeByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return term, nil
}

func (s *IOTypeService) CreateIOType(ctx *fiber.Ctx, term *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("IOTypeService-CreateIOType", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateIOType(tx, term, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return term, nil
}

func (s *IOTypeService) UpdateIOType(ctx *fiber.Ctx, term *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("IOTypeService-UpdateIOType", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateIOType(tx, term, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return term, nil
}

func (s *IOTypeService) DeleteIOType(ctx *fiber.Ctx, params *dtos.GetIOTypeParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("IOTypeService-DeleteIOType", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteIOType(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *IOTypeService) RestoreIOType(ctx *fiber.Ctx, params *dtos.GetIOTypeParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("IOTypeService-RestoreIOType", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreIOType(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *IOTypeService) ExcelGetIOTypes(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("IOTypeService-ExcelGetIOTypes", opentracing.ChildOf(span.Context()))

	ioTypes, _, err := s.GetIOTypes(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	sheetName := "IOTypes"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("IOTypes", "A1", &[]string{"ID", "Name", "Code", "Type", "Description", "Created At", "Updated At"})

	for i, term := range ioTypes {
		row := []interface{}{
			term.ID,
			term.Name,
			term.Code,
			term.Type,
			term.Description,
			term.CreatedAt,
			term.UpdatedAt,
		}
		file.SetSheetRow("IOTypes", fmt.Sprintf("A%d", i+2), &row)
	}

	file.SetActiveSheet(index)

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

func (s *IOTypeService) CsvGetIOTypes(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("IOTypeService-CsvGetIOTypes", opentracing.ChildOf(span.Context()))

	filters["is_csv"] = "1"
	ioTypes, _, err := s.GetIOTypes(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

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
	csv += "IO Types\n"
	csv += "\n"

	csv += "ID,Name,Code,Type,Description,Remark,Created At,Updated At\n"
	for _, term := range ioTypes {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s\n",
			term.ID,
			term.Name,
			term.Code,
			term.Type,
			utils.GetPtrVal(term.Description),
			utils.GetPtrVal(term.Remark),
			utils.GetPtrVal(term.CreatedAt),
			utils.GetPtrVal(term.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
