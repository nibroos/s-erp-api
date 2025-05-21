package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nibroos/s-erp-api/service/internal/config"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/opentracing/opentracing-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

func GetInvoiceMaintenanceIDs(req dtos.UpdateInvoiceMaintenanceRequest) ([]*uint, []*uint, []*uint) {
	invoiceMaintenanceDtIDs := []*uint{}
	productIDs := []*uint{}
	itemUnitIDs := []*uint{}

	for _, reqInvoiceMaintenanceDt := range req.InvoiceMaintenanceDts {
		if reqInvoiceMaintenanceDt.InvoiceMaintenanceDtID != nil && *reqInvoiceMaintenanceDt.InvoiceMaintenanceDtID > 0 {
			invoiceMaintenanceDtIDs = append(invoiceMaintenanceDtIDs, reqInvoiceMaintenanceDt.InvoiceMaintenanceDtID)
		}
	}

	return invoiceMaintenanceDtIDs, productIDs, itemUnitIDs
}

func GetLockInvoiceMaintenanceSalesOrderIDs(req dtos.CreateInvoiceMaintenanceRequest) ([]*uint, []*uint) {
	soIDs := []*uint{}
	soDtIDs := []*uint{}

	soIDsMap := make(map[uint]bool)

	for _, reqInvoiceMaintenanceDt := range req.InvoiceMaintenanceDts {
		if reqInvoiceMaintenanceDt.RefType != nil && *reqInvoiceMaintenanceDt.RefType == "so" {
			if reqInvoiceMaintenanceDt.RefID != nil && *reqInvoiceMaintenanceDt.RefID > 0 {
				if !soIDsMap[*reqInvoiceMaintenanceDt.RefID] {
					soIDsMap[*reqInvoiceMaintenanceDt.RefID] = true
					soIDs = append(soIDs, reqInvoiceMaintenanceDt.RefID)
				}
			}

			if reqInvoiceMaintenanceDt.RefDtID != nil && *reqInvoiceMaintenanceDt.RefDtID > 0 {
				soDtIDs = append(soDtIDs, reqInvoiceMaintenanceDt.RefDtID)
			}
		}
	}

	return soIDs, soDtIDs
}

func MapCreateInvoiceMaintenanceDts(ctx *fiber.Ctx, req dtos.CreateInvoiceMaintenanceRequest, createdInvoiceMaintenance *models.InvoiceMaintenance, userID uint, span opentracing.Span) ([]models.InvoiceMaintenanceDt, error) {
	invoiceMaintenanceDtsModel := []models.InvoiceMaintenanceDt{}

	for _, invoiceMaintenanceDt := range req.InvoiceMaintenanceDts {
		refJSONStr := "{}"
		productJSONStr := "{}"

		refJSON := json.RawMessage(refJSONStr)
		productJSON := json.RawMessage(productJSONStr)

		invoiceMaintenanceDtModel := models.InvoiceMaintenanceDt{
			ProductUuid:          invoiceMaintenanceDt.ProductUuid,
			InvoiceMaintenanceID: &createdInvoiceMaintenance.ID,
			ItemUnitID:           invoiceMaintenanceDt.ItemUnitID,
			VatID:                invoiceMaintenanceDt.VatID,
			Pph23ID:              invoiceMaintenanceDt.Pph23ID,
			RefID:                invoiceMaintenanceDt.RefID,
			RefDtID:              invoiceMaintenanceDt.RefDtID,
			ProductID:            invoiceMaintenanceDt.ProductID,
			RefType:              invoiceMaintenanceDt.RefType,
			ProductType:          invoiceMaintenanceDt.ProductType,
			RefJSON:              &refJSON,
			ProductJSON:          &productJSON,
			Remark:               invoiceMaintenanceDt.Remark,
			IsVat:                invoiceMaintenanceDt.IsVat,
			IsPph23:              invoiceMaintenanceDt.IsPph23,
			Qty:                  invoiceMaintenanceDt.Qty,
			Price:                invoiceMaintenanceDt.Price,
			Subtotal:             invoiceMaintenanceDt.Subtotal,
			Discount:             invoiceMaintenanceDt.Discount,
			TotalAmount:          invoiceMaintenanceDt.TotalAmount,
			TotalDp:              invoiceMaintenanceDt.TotalDp,
			TotalBalance:         invoiceMaintenanceDt.TotalBalance,
			CreatedByID:          &userID,
		}
		invoiceMaintenanceDtsModel = append(invoiceMaintenanceDtsModel, invoiceMaintenanceDtModel)
	}

	return invoiceMaintenanceDtsModel, nil
}

func MapUpdateInvoiceMaintenanceDts(ctx *fiber.Ctx, req dtos.UpdateInvoiceMaintenanceRequest, updatedInvoiceMaintenance *models.InvoiceMaintenance, userID uint, span opentracing.Span) ([]models.InvoiceMaintenanceDt, error) {
	invoiceMaintenanceDtsModel := []models.InvoiceMaintenanceDt{}

	refJSONStr := "{}"
	productJSONStr := "{}"

	refJSON := json.RawMessage(refJSONStr)
	productJSON := json.RawMessage(productJSONStr)

	for _, reqInvoiceMaintenanceDt := range req.InvoiceMaintenanceDts {
		invoiceMaintenanceDtID := uint(0)
		if reqInvoiceMaintenanceDt.InvoiceMaintenanceDtID != nil {
			invoiceMaintenanceDtID = *reqInvoiceMaintenanceDt.InvoiceMaintenanceDtID
		}

		invoiceMaintenanceDtModel := models.InvoiceMaintenanceDt{
			ID:                   invoiceMaintenanceDtID,
			ProductUuid:          reqInvoiceMaintenanceDt.ProductUuid,
			InvoiceMaintenanceID: &updatedInvoiceMaintenance.ID,
			ItemUnitID:           reqInvoiceMaintenanceDt.ItemUnitID,
			VatID:                reqInvoiceMaintenanceDt.VatID,
			Pph23ID:              reqInvoiceMaintenanceDt.Pph23ID,
			RefID:                reqInvoiceMaintenanceDt.RefID,
			RefDtID:              reqInvoiceMaintenanceDt.RefDtID,
			ProductID:            reqInvoiceMaintenanceDt.ProductID,
			RefType:              reqInvoiceMaintenanceDt.RefType,
			ProductType:          reqInvoiceMaintenanceDt.ProductType,
			RefJSON:              &refJSON,
			ProductJSON:          &productJSON,
			Remark:               reqInvoiceMaintenanceDt.Remark,
			IsVat:                reqInvoiceMaintenanceDt.IsVat,
			IsPph23:              reqInvoiceMaintenanceDt.IsPph23,
			Qty:                  reqInvoiceMaintenanceDt.Qty,
			Price:                reqInvoiceMaintenanceDt.Price,
			Subtotal:             reqInvoiceMaintenanceDt.Subtotal,
			Discount:             reqInvoiceMaintenanceDt.Discount,
			TotalAmount:          reqInvoiceMaintenanceDt.TotalAmount,
			TotalDp:              reqInvoiceMaintenanceDt.TotalDp,
			TotalBalance:         reqInvoiceMaintenanceDt.TotalBalance,
			CreatedByID:          &userID,
		}
		invoiceMaintenanceDtsModel = append(invoiceMaintenanceDtsModel, invoiceMaintenanceDtModel)
	}

	return invoiceMaintenanceDtsModel, nil
}

func GenInvoiceMaintenanceNo(ctx *fiber.Ctx, req dtos.CreateInvoiceMaintenanceRequest, orderedNumber int, span opentracing.Span) string {
	prefix := "IMT"
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	day := time.Now().Format("02")
	order := fmt.Sprintf("%d", orderedNumber)

	str := fmt.Sprintf("%s/%s/%s-%s-%s", prefix, order, year, month, day)

	return str
}

func GenerateInvoiceMaintenanceNoOnUpdate(ctx *fiber.Ctx, req dtos.UpdateInvoiceMaintenanceRequest, revNo *int, span opentracing.Span) string {
	if req.InvoiceNo == nil {
		prefix := "IMT"
		year := time.Now().Format("2006")
		month := time.Now().Format("01")
		day := time.Now().Format("02")

		return fmt.Sprintf("%s/REV-%d/%s-%s-%s", prefix, *revNo, year, month, day)
	}

	invoiceNo := *req.InvoiceNo

	if !strings.Contains(invoiceNo, "REV") {
		return fmt.Sprintf("%s/REV-%d", invoiceNo, *revNo)
	} else {
		basePart := strings.Split(invoiceNo, "/REV")[0]
		return fmt.Sprintf("%s/REV-%d", basePart, *revNo)
	}
}

func MapCreateInvoiceMaintenance(ctx *fiber.Ctx, req dtos.CreateInvoiceMaintenanceRequest, userID uint, branchID uint, orderedNumber int, span opentracing.Span) (models.InvoiceMaintenance, error) {
	invoiceNo := GenInvoiceMaintenanceNo(ctx, req, orderedNumber, span)

	var invoiceDate *time.Time
	if req.InvoiceDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.InvoiceDate)
		if err != nil {
			return models.InvoiceMaintenance{}, err
		}
		invoiceDate = &parsedTime
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return models.InvoiceMaintenance{}, err
		}
		dueDate = &parsedTime
	}

	invoiceMaintenance := models.InvoiceMaintenance{
		CustomerID:               req.CustomerID,
		CurrencyID:               req.CurrencyID,
		PaymentTermID:            req.PaymentTermID,
		VatID:                    req.VatID,
		Pph23ID:                  req.Pph23ID,
		BranchID:                 &branchID,
		BankID:                   req.BankID,
		Title:                    req.Title,
		InvoiceNo:                &invoiceNo,
		InvoiceDate:              invoiceDate,
		DueDate:                  dueDate,
		ExchangeRate:             req.ExchangeRate,
		Remark:                   req.Remark,
		Status:                   req.Status,
		ApprovedStatus:           req.ApprovedStatus,
		RevNo:                    new(int),
		Pph23Percentage:          req.Pph23Percentage,
		VatPercentage:            req.VatPercentage,
		DiscountAmount:           req.DiscountAmount,
		DiscountPercentage:       req.DiscountPercentage,
		DiscountPercentageAmount: req.DiscountPercentageAmount,
		DiscountFinal:            req.DiscountFinal,
		DiscountType:             req.DiscountType,
		TotalAmountProducts:      req.TotalAmountProducts,
		TotalDpProducts:          req.TotalDpProducts,
		TotalBalanceProducts:     req.TotalBalanceProducts,
		Subtotal:                 req.Subtotal,
		TotalQty:                 req.TotalQty,
		TotalDiscount:            req.TotalDiscount,
		TotalPph23:               req.TotalPph23,
		TotalVat:                 req.TotalVat,
		GrandTotal:               req.GrandTotal,
		CreatedByID:              &userID,
	}
	*invoiceMaintenance.RevNo = 0

	return invoiceMaintenance, nil
}

func MapUpdateInvoiceMaintenance(ctx *fiber.Ctx, req dtos.UpdateInvoiceMaintenanceRequest, userID uint, branchID uint, existingRevNo *int, span opentracing.Span) (models.InvoiceMaintenance, error) {
	revNo := 0
	if existingRevNo != nil {
		revNo = *existingRevNo + 1
	} else {
		revNo = 1
	}

	invoiceNo := GenerateInvoiceMaintenanceNoOnUpdate(ctx, req, &revNo, span)

	var invoiceDate *time.Time
	if req.InvoiceDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.InvoiceDate)
		if err != nil {
			return models.InvoiceMaintenance{}, err
		}
		invoiceDate = &parsedTime
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return models.InvoiceMaintenance{}, err
		}
		dueDate = &parsedTime
	}

	invoiceMaintenance := models.InvoiceMaintenance{
		ID:                       req.ID,
		CustomerID:               req.CustomerID,
		CurrencyID:               req.CurrencyID,
		PaymentTermID:            req.PaymentTermID,
		VatID:                    req.VatID,
		Pph23ID:                  req.Pph23ID,
		BranchID:                 &branchID,
		BankID:                   req.BankID,
		Title:                    req.Title,
		InvoiceNo:                &invoiceNo,
		InvoiceDate:              invoiceDate,
		DueDate:                  dueDate,
		ExchangeRate:             req.ExchangeRate,
		Remark:                   req.Remark,
		Status:                   req.Status,
		ApprovedStatus:           req.ApprovedStatus,
		RevNo:                    &revNo,
		Pph23Percentage:          req.Pph23Percentage,
		VatPercentage:            req.VatPercentage,
		DiscountAmount:           req.DiscountAmount,
		DiscountPercentage:       req.DiscountPercentage,
		DiscountPercentageAmount: req.DiscountPercentageAmount,
		DiscountFinal:            req.DiscountFinal,
		DiscountType:             req.DiscountType,
		TotalAmountProducts:      req.TotalAmountProducts,
		TotalDpProducts:          req.TotalDpProducts,
		TotalBalanceProducts:     req.TotalBalanceProducts,
		Subtotal:                 req.Subtotal,
		TotalQty:                 req.TotalQty,
		TotalDiscount:            req.TotalDiscount,
		TotalPph23:               req.TotalPph23,
		TotalVat:                 req.TotalVat,
		GrandTotal:               req.GrandTotal,
		UpdatedByID:              &userID,
	}

	return invoiceMaintenance, nil
}

func GetSalesOrderDtIDsForInvoiceMaintenance(soDts []dtos.RefSalesOrderForInvoiceMaintenanceListDTO) []uint {
	salesOrderIDs := []uint{}

	for _, soDt := range soDts {
		if soDt.SalesOrderID != nil {
			salesOrderIDs = append(salesOrderIDs, *soDt.SalesOrderID)
		}
	}

	return salesOrderIDs
}

func MapRefSoDtBomsToSoDtsForInvoiceMaintenance(soDtBoms []dtos.SalesOrderSoDtBomListDTO, soDts []dtos.RefSalesOrderForInvoiceMaintenanceListDTO) []dtos.RefSalesOrderForInvoiceMaintenanceListDTO {
	combinedSoDts := []dtos.RefSalesOrderForInvoiceMaintenanceListDTO{}

	for _, soDt := range soDts {
		newSoDtBoms := make([]dtos.SalesOrderSoDtBomListDTO, 0)
		for _, soDtBom := range soDtBoms {
			if *soDtBom.SoDtID == *soDt.ID {
				newSoDtBoms = append(newSoDtBoms, soDtBom)
			}
		}

		soDt.SoDtsBoms = newSoDtBoms
		combinedSoDts = append(combinedSoDts, soDt)
	}

	return combinedSoDts
}

func MapUpdateSalesOrderStatusForInvoiceMaintenance(salesOrderStatusUpdate map[string]interface{}) dtos.UpdateSalesOrderStatusForInvoiceMaintenanceRequest {
	params := dtos.UpdateSalesOrderStatusForInvoiceMaintenanceRequest{}

	params.ID = salesOrderStatusUpdate["id"].(uint)
	params.Status = salesOrderStatusUpdate["status"].(string)

	return params
}

func MapRepeatInvoiceMaintenance(ctx *fiber.Ctx, originalInvoice *models.InvoiceMaintenance, item dtos.RepeatInvoiceMaintenanceItem, userID uint, branchID uint, orderedNumber int, span opentracing.Span) (models.InvoiceMaintenance, error) {
	invoiceNo := GenInvoiceMaintenanceNo(ctx, dtos.CreateInvoiceMaintenanceRequest{}, orderedNumber, span)

	var invoiceDate *time.Time
	if item.InvoiceDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *item.InvoiceDate)
		if err != nil {
			return models.InvoiceMaintenance{}, fmt.Errorf("invalid invoice date format: %v", err)
		}
		invoiceDate = &parsedTime
	} else if originalInvoice.InvoiceDate != nil {
		// Use original invoice date if not provided
		invoiceDate = originalInvoice.InvoiceDate
	}

	var dueDate *time.Time
	if item.DueDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *item.DueDate)
		if err != nil {
			return models.InvoiceMaintenance{}, fmt.Errorf("invalid due date format: %v", err)
		}
		dueDate = &parsedTime
	} else if originalInvoice.DueDate != nil {
		// Use original due date if not provided
		dueDate = originalInvoice.DueDate
	}

	// Use provided title or original title
	title := originalInvoice.Title
	if item.Title != nil {
		title = item.Title
	}

	// Use provided remark or original one
	remark := originalInvoice.Remark
	if item.Remark != nil {
		remark = item.Remark
	}

	// Set default status to "UNPAID"
	defaultStatus := "UNPAID"
	defaultApprovedStatus := "PENDING"
	defaultRevNo := 0

	invoiceMaintenance := models.InvoiceMaintenance{
		CustomerID:               originalInvoice.CustomerID,
		CurrencyID:               originalInvoice.CurrencyID,
		PaymentTermID:            originalInvoice.PaymentTermID,
		VatID:                    originalInvoice.VatID,
		Pph23ID:                  originalInvoice.Pph23ID,
		BranchID:                 &branchID,
		BankID:                   originalInvoice.BankID,
		Title:                    title,
		InvoiceNo:                &invoiceNo,
		InvoiceDate:              invoiceDate,
		DueDate:                  dueDate,
		ExchangeRate:             originalInvoice.ExchangeRate,
		Remark:                   remark,
		Status:                   &defaultStatus,
		ApprovedStatus:           &defaultApprovedStatus,
		RevNo:                    &defaultRevNo,
		Pph23Percentage:          originalInvoice.Pph23Percentage,
		VatPercentage:            originalInvoice.VatPercentage,
		DiscountAmount:           originalInvoice.DiscountAmount,
		DiscountPercentage:       originalInvoice.DiscountPercentage,
		DiscountPercentageAmount: originalInvoice.DiscountPercentageAmount,
		DiscountFinal:            originalInvoice.DiscountFinal,
		DiscountType:             originalInvoice.DiscountType,
		TotalAmountProducts:      originalInvoice.TotalAmountProducts,
		TotalDpProducts:          originalInvoice.TotalDpProducts,
		TotalBalanceProducts:     originalInvoice.TotalBalanceProducts,
		Subtotal:                 originalInvoice.Subtotal,
		TotalQty:                 originalInvoice.TotalQty,
		TotalDiscount:            originalInvoice.TotalDiscount,
		TotalPph23:               originalInvoice.TotalPph23,
		TotalVat:                 originalInvoice.TotalVat,
		GrandTotal:               originalInvoice.GrandTotal,
		CreatedByID:              &userID,
	}

	return invoiceMaintenance, nil
}

func MapRepeatInvoiceMaintenanceDts(ctx *fiber.Ctx, originalDts []models.InvoiceMaintenanceDt, newInvoiceMaintenanceID uint, userID uint, span opentracing.Span) ([]models.InvoiceMaintenanceDt, error) {
	newDts := make([]models.InvoiceMaintenanceDt, 0, len(originalDts))

	for _, dt := range originalDts {
		newProductUuid := uuid.New().String()

		newDt := models.InvoiceMaintenanceDt{
			ProductUuid:          newProductUuid,
			InvoiceMaintenanceID: &newInvoiceMaintenanceID,
			ItemUnitID:           dt.ItemUnitID,
			VatID:                dt.VatID,
			Pph23ID:              dt.Pph23ID,
			RefID:                dt.RefID,
			RefDtID:              dt.RefDtID,
			ProductID:            dt.ProductID,
			RefType:              dt.RefType,
			ProductType:          dt.ProductType,
			RefJSON:              dt.RefJSON,
			ProductJSON:          dt.ProductJSON,
			Remark:               dt.Remark,
			IsVat:                dt.IsVat,
			IsPph23:              dt.IsPph23,
			Qty:                  dt.Qty,
			Price:                dt.Price,
			Subtotal:             dt.Subtotal,
			Discount:             dt.Discount,
			TotalAmount:          dt.TotalAmount,
			TotalDp:              dt.TotalDp,
			TotalBalance:         dt.TotalBalance,
			CreatedByID:          &userID,
		}
		newDts = append(newDts, newDt)
	}

	return newDts, nil
}

func MapBulkSendEmailApprovedSingle(emailObject *dtos.FormSentEmailRequest, userID uint, branchID uint, span opentracing.Span) models.SentEmail {

	email := models.SentEmail{}
	email.ID = *emailObject.ID
	email.RefID = emailObject.ID
	email.SenderID = &userID
	email.FromEmail = *emailObject.FromEmail
	email.ToEmail = emailObject.ToEmail
	email.Subject = *emailObject.Subject
	email.Remark = emailObject.Remark
	email.ErrorMessage = emailObject.ErrorMessage
	email.UpdatedByID = &userID

	return email
}

func MapBulkSendEmailApproved(invoiceMaintenances []dtos.InvoiceMaintenanceListDTO, req dtos.BulkSendEmailApprovedInvoiceMaintenancesRequest, userID uint, branchID uint, span opentracing.Span) ([]models.SentEmail, []uint, error) {
	status := "PROCESS"
	refType := "invoice_maintenances"

	refIDs := []uint{}
	emails := []models.SentEmail{}
	for _, im := range invoiceMaintenances {
		// logJson, err := json.Marshal(invoiceMaintenances)
		// if err != nil {
		// 	return nil, nil, err
		// }
		// logJsonStr := string(logJson)
		mail := models.SentEmail{
			SenderID: &userID,
			RefID:    &im.ID,
			RefType:  refType,
			Subject:  *im.Title,
			Status:   status,
			ToEmail:  *im.CustomerEmail,
			// LogJson:     &logJsonStr,
			CreatedByID: &userID,
		}

		emails = append(emails, mail)
		refIDs = append(refIDs, im.ID)
	}

	return emails, refIDs, nil
}

func MapBulkSendEmailApprovedModelToDTO(ctx *fiber.Ctx, newEmails []models.SentEmail, userID uint, branchID uint, span opentracing.Span) ([]dtos.FormSentEmailRequest, error) {
	emails := []dtos.FormSentEmailRequest{}
	for _, newMail := range newEmails {
		mail := dtos.FormSentEmailRequest{
			ID:           &newMail.ID,
			RefID:        newMail.RefID,
			RefType:      &newMail.RefType,
			Status:       &newMail.Status,
			SenderID:     newMail.SenderID,
			FromEmail:    &newMail.FromEmail,
			ToEmail:      newMail.ToEmail,
			Subject:      &newMail.Subject,
			Remark:       newMail.Remark,
			ErrorMessage: newMail.ErrorMessage,
			LogJson:      newMail.LogJson,
			CreatedByID:  newMail.CreatedByID,
			UpdatedByID:  newMail.UpdatedByID,
		}
		if newMail.ID > 0 {
			mail.ID = &newMail.ID
			mail.UpdatedByID = &userID
		} else {
			mail.CreatedByID = &userID
		}
		emails = append(emails, mail)
	}

	return emails, nil
}

func PublishBulkSendEmailApprovedInvoiceMaintenance(ctx *fiber.Ctx, rabbitmq *config.RabbitMQ, data dtos.BulkSendEmailApprovedInvoiceMaintenancesRequest) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// PublishBulkSendEmailApprovedInvoiceMaintenance
	return rabbitmq.Channel.PublishWithContext(ctx.Context(),
		"", // exchange
		"bulk_send_email_approved_invoice_maintenance_queue", // routing key (queue name)
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
}

func MapGetInvoiceMaintenancesIDs(req []dtos.InvoiceMaintenanceListDTO) []uint {
	invoiceMaintenanceIDs := []uint{}

	for _, reqInvoiceMaintenance := range req {
		if reqInvoiceMaintenance.ID > 0 {
			invoiceMaintenanceIDs = append(invoiceMaintenanceIDs, uint(reqInvoiceMaintenance.ID))
		}
	}

	return invoiceMaintenanceIDs
}

// MapInvoiceMaintenancesDts
func MapInvoiceMaintenancesDts(parents []dtos.InvoiceMaintenanceListDTO, dts []dtos.InvoiceMaintenanceDtListNoBomDTO) []dtos.InvoiceMaintenanceListDTO {
	invoiceMaintenanceDts := []dtos.InvoiceMaintenanceListDTO{}

	for _, parent := range parents {
		newDts := make([]*dtos.InvoiceMaintenanceDtListNoBomDTO, 0)
		for _, dt := range dts {
			if parent.ID == *dt.InvoiceMaintenanceID {
				dtCopy := dt
				newDts = append(newDts, &dtCopy)
			}
		}

		// parent.InvoiceMaintenanceDts = newDts
		invoiceMaintenanceDts = append(invoiceMaintenanceDts, parent)
	}

	return invoiceMaintenanceDts
}

func GetSelectedEmailInvoiceMaintenance(invoiceMaintenance dtos.InvoiceMaintenanceListDTO, req dtos.BulkSendEmailApprovedInvoiceMaintenancesRequest) dtos.FormSentEmailRequest {
	var selectedEmail dtos.FormSentEmailRequest
	log.Println("GetSelectedEmailInvoiceMaintenance", invoiceMaintenance.ID, req.SentEmails)

	for i, email := range req.SentEmails {
		log.Println("req.SentEmails1", email.ToEmail, "abc", email, "idx", i, email.ToEmail)
		log.Println("req.SentEmails1.2", *email.RefID)
		// if email.RefID != nil && *email.RefID == invoiceMaintenance.ID && email.RefType != nil && *email.RefType == "invoice_maintenances" {
		if *email.RefID == invoiceMaintenance.ID {
			log.Println("req.SentEmails2", email, "idx", i)
			selectedEmail = email
		}
	}

	return selectedEmail
}

func GetSelectedDtsInvoiceMaintenance(parent dtos.InvoiceMaintenanceListDTO, invoiceMaintenanceDts []dtos.InvoiceMaintenanceDtListNoBomDTO) []dtos.InvoiceMaintenanceDtListNoBomDTO {
	var selectedDts []dtos.InvoiceMaintenanceDtListNoBomDTO
	for _, dt := range invoiceMaintenanceDts {
		if dt.InvoiceMaintenanceID != nil && *dt.InvoiceMaintenanceID == parent.ID && dt.ID != nil && *dt.ID > 0 {
			selectedDts = append(selectedDts, dt)
		}
	}

	return selectedDts
}
