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

type CustomerService struct {
	repo     *repository.CustomerRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewCustomerService(repo *repository.CustomerRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *CustomerService {
	return &CustomerService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *CustomerService) GetCustomers(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.CustomerListDTO, int, error) {
	childSpan := opentracing.StartSpan("CustomerService-GetCustomers", opentracing.ChildOf(span.Context()))

	customers, total, err := s.repo.GetCustomers(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return customers, total, nil
}

func (s *CustomerService) CreateCustomer(ctx *fiber.Ctx, customer *models.Customer, tx *gorm.DB, span opentracing.Span) (*models.Customer, error) {
	childSpan := opentracing.StartSpan("CustomerService-CreateCustomer", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateCustomer(tx, customer, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return customer, nil
}

func (s *CustomerService) GetCustomerByID(ctx *fiber.Ctx, params *dtos.GetCustomerParams, span opentracing.Span) (*dtos.CustomerDetailDTO, error) {
	childSpan := opentracing.StartSpan("CustomerService-GetCustomerByID", opentracing.ChildOf(span.Context()))

	customer, err := s.repo.GetCustomerByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return customer, nil
}

func (s *CustomerService) GetCustomerPicEmails(ctx *fiber.Ctx, params *dtos.GetCustomerParams, span opentracing.Span) ([]dtos.FormCustomerPICEmailsRequest, error) {
	childSpan := opentracing.StartSpan("CustomerService-GetCustomerPicEmails", opentracing.ChildOf(span.Context()))

	customerEmails, err := s.repo.GetCustomerPicEmails(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return customerEmails, nil
}

func (s *CustomerService) GetCustomerContracts(ctx *fiber.Ctx, params *dtos.GetCustomerParams, span opentracing.Span) ([]dtos.FormCustomerContractsRequest, error) {
	childSpan := opentracing.StartSpan("CustomerService-GetCustomerContracts", opentracing.ChildOf(span.Context()))

	customerEmails, err := s.repo.GetCustomerContracts(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return customerEmails, nil
}

func (s *CustomerService) UpdateCustomer(ctx *fiber.Ctx, req dtos.FormCrmCustomerRequest, customer *models.Customer, userID uint, tx *gorm.DB, span opentracing.Span) (*models.Customer, error) {
	childSpan := opentracing.StartSpan("CustomerService-UpdateCustomer", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateCustomer(tx, customer, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	// create/update customer emails
	customerEmails, customerEmailIDs, err := utils.MapUpdatePicEmails(ctx, req, customer, userID, span)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}
	err = s.repo.DeletePicEmails(tx, customerEmailIDs, customer.ID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	if len(customerEmails) > 0 {
		if err := s.utilRepo.BatchUpsertModels(tx, customerEmails, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	// create/update customer contracts
	customerContracts, customerContractIDs, err := utils.MapUpdateCustomerContracts(ctx, req, customer, userID, span)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	err = s.repo.DeleteCustomerContracts(tx, customerContractIDs, customer.ID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	if len(customerContracts) > 0 {
		if err := s.utilRepo.BatchUpsertModels(tx, customerContracts, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	return customer, nil
}

func (s *CustomerService) DeleteCustomer(ctx *fiber.Ctx, params *dtos.GetCustomerParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CustomerService-DeleteCustomer", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteCustomer(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *CustomerService) DeleteCrmCustomer(ctx *fiber.Ctx, params *dtos.GetCustomerParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CustomerService-DeleteCrmCustomer", opentracing.ChildOf(span.Context()))

	// customerUpdateCrm := []map[string]interface{}{}
	// customerUpdateCrm = append(customerUpdateCrm, map[string]interface{}{
	// 	"id":     params.ID,
	// 	"is_crm": 0,
	// })

	// if err := s.utilRepo.Upsert(tx, "customers", "id", customerUpdateCrm, childSpan); err != nil {
	// 	defer childSpan.Finish()
	// 	tx.Rollback()
	// 	return err
	// }

	customerUpdateCrm := map[string]interface{}{
		"id":     params.ID,
		"is_crm": 0,
	}

	// customerUpdateCrm := &models.Customer{
	// 	ID:     params.ID,
	// 	IsCrm:  0,
	// }

	if err := s.repo.UpdateDeleteCrmCustomer(tx, customerUpdateCrm, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *CustomerService) RestoreCustomer(ctx *fiber.Ctx, params *dtos.GetCustomerParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CustomerService-RestoreCustomer", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreCustomer(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *CustomerService) ExcelGetCustomers(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("CustomerService-ExcelGetCustomers", opentracing.ChildOf(span.Context()))

	customers, _, err := s.GetCustomers(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "customers"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("Customers", "A1", &[]string{"ID", "Code", "Name", "Customer Type", "Agent", "Address", "Phone", "Email", "PIC", "Created At", "Updated At"})

	for i, customer := range customers {
		row := []interface{}{
			customer.ID,
			utils.GetPtrVal(customer.Code),
			customer.Name,
			utils.GetPtrVal(customer.CustomerTypeName),
			utils.GetPtrVal(customer.AgentName),
			utils.GetPtrVal(customer.Address),
			utils.GetPtrVal(customer.Phone),
			utils.GetPtrVal(customer.Email),
			utils.GetPtrVal(customer.Pic),
			customer.CreatedAt,
			customer.UpdatedAt,
		}
		file.SetSheetRow("Customers", fmt.Sprintf("A%d", i+2), &row)
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
func (s *CustomerService) CsvGetCustomers(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("CustomerService-CsvGetCustomers", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	customers, _, err := s.GetCustomers(ctx, filters, childSpan)
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
	csv += "Customers\n"
	csv += "\n"

	csv += "ID,Code,Name,Customer Type,Agent,Address,Phone,Email,PIC,Created At,Updated At\n"
	// Build CSV rows
	for _, customer := range customers {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s,%s,%v,%v\n",
			customer.ID,
			utils.GetPtrVal(customer.Code),
			customer.Name,
			utils.GetPtrVal(customer.CustomerTypeName),
			utils.GetPtrVal(customer.AgentName),
			utils.GetPtrVal(customer.Address),
			utils.GetPtrVal(customer.Phone),
			utils.GetPtrVal(customer.Email),
			utils.GetPtrVal(customer.Pic),
			customer.CreatedAt,
			customer.UpdatedAt,
		)
	}

	return []byte(csv), nil
}
