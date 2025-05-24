package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/opentracing/opentracing-go"
)

func MapUpdateCustomerContracts(ctx *fiber.Ctx, req dtos.FormCrmCustomerRequest, updatedCustomer *models.Customer, userID uint, span opentracing.Span) ([]models.CustomerContract, []uint, error) {

	customerContractsModel := []models.CustomerContract{}
	customerContractIDs := []uint{}

	for _, contract := range req.CustomerContracts {
		if contract.ProductID == nil {
			continue
		}

		customerContractID := uint(0)

		customerContractModel := models.CustomerContract{
			// ID:              customerContractID,
			CustomerID:     &updatedCustomer.ID,
			ProductID:      contract.ProductID,
			PaymentTypeID:  contract.PaymentTypeID,
			AgreeAt:        contract.AgreeAt,
			DueAt:          contract.DueAt,
			Price:          contract.Price,
			Qty:            contract.Qty,
			InstallationAt: contract.InstallationAt,
			WarrantyAt:     contract.WarrantyAt,
			Remark:         contract.Remark,

			// CreatedByID: &userID,
		}

		if contract.ID != nil {
			customerContractID = *contract.ID
			customerContractModel.ID = customerContractID
			customerContractModel.UpdatedByID = &userID

			customerContractIDs = append(customerContractIDs, customerContractID)
		} else {
			customerContractModel.CreatedByID = &userID
		}

		customerContractsModel = append(customerContractsModel, customerContractModel)

	}

	return customerContractsModel, customerContractIDs, nil
}

func MapUpdatePicEmails(ctx *fiber.Ctx, req dtos.FormCrmCustomerRequest, updatedCustomer *models.Customer, userID uint, span opentracing.Span) ([]models.PicEmail, []uint, error) {

	picEmailsModel := []models.PicEmail{}
	picEmailIDs := []uint{}

	for _, contract := range req.PicEmails {
		picEmailID := uint(0)

		picEmailModel := models.PicEmail{
			CustomerID: &updatedCustomer.ID,
			Name:       contract.Name,
			IsMain:     contract.IsMain,
		}

		if contract.ID != nil {
			picEmailID = *contract.ID
			picEmailModel.ID = picEmailID
			picEmailModel.UpdatedByID = &userID

			picEmailIDs = append(picEmailIDs, picEmailID)
		} else {
			picEmailModel.CreatedByID = &userID
		}

		picEmailsModel = append(picEmailsModel, picEmailModel)

	}

	return picEmailsModel, picEmailIDs, nil
}
