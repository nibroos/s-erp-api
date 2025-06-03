package service

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gorm.io/gorm"
)

type RequestOrderService struct {
	repo     *repository.RequestOrderRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewRequestOrderService(repo *repository.RequestOrderRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *RequestOrderService {
	return &RequestOrderService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *RequestOrderService) GetRequestOrders(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RequestOrderListDTO, int, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-GetRequestOrders", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	requestOrders, total, err := s.repo.GetRequestOrders(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}
	return requestOrders, total, nil
}

func (s *RequestOrderService) CreateRequestOrder(ctx *fiber.Ctx, req dtos.CreateRequestOrderRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.RequestOrder, *gorm.DB, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-CreateRequestOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	requestOrderCreatedThisMonthNumber, err := s.repo.GetRequestOrderCreatedThisMonth(ctx, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, tx, err
	}

	orderedNumber := requestOrderCreatedThisMonthNumber + 1

	requestOrder, err := utils.MapCreateRequestOrder(ctx, req, userID, branchID, orderedNumber, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	tx, err = s.repo.CreateRequestOrder(tx, &requestOrder, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, tx, err
	}

	tx, _, err = s.CreateRequestOrderDts(ctx, req, userID, &requestOrder, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, tx, err
	}

	return &requestOrder, tx, nil
}

func (s *RequestOrderService) GetRequestOrderByID(ctx *fiber.Ctx, params *dtos.GetRequestOrderParams, tx *gorm.DB, span opentracing.Span) (*dtos.RequestOrderDetailDTO, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-GetRequestOrderByID", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	requestOrder, err := s.repo.GetRequestOrderByID(ctx, params, tx, childSpan)
	if err != nil {
		return nil, err
	}

	requestOrderDts, err := s.repo.GetRequestOrderDts(ctx, requestOrder.ID, params.IsDeleted, childSpan)
	if err != nil {
		return nil, err
	}

	requestOrder.RequestOrderDts = requestOrderDts

	return requestOrder, nil
}

func (s *RequestOrderService) UpdateRequestOrder(ctx *fiber.Ctx, req dtos.UpdateRequestOrderRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.RequestOrder, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-UpdateRequestOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	params := dtos.GetRequestOrderParams{ID: req.ID}
	existingRequestOrder, err := s.GetRequestOrderByID(ctx, &params, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := s.repo.LockRequestOrder(tx, req.ID, childSpan); err != nil {
		tx.Rollback()
		return nil, err
	}

	requestOrder, err := utils.MapUpdateRequestOrder(ctx, req, userID, branchID, existingRequestOrder.RevNo, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx, err = s.repo.UpdateRequestOrder(tx, &requestOrder, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	existingDtMap := make(map[string]*dtos.RequestOrderDtListDTO)
	for i, dt := range existingRequestOrder.RequestOrderDts {
		if dt.ID != nil {
			key := fmt.Sprintf("%d-%d-%d",
				utils.GetValueOrDefault(dt.ProductID, 0),
				utils.GetValueOrDefault(dt.RefID, 0),
				utils.GetValueOrDefault(dt.ItemID, 0))
			existingDtMap[key] = &existingRequestOrder.RequestOrderDts[i]
		}
	}

	requestOrderDts, err := utils.MapUpdateRequestOrderDts(ctx, req, &requestOrder, userID, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	var createRequestOrderDts []models.RequestOrderDt
	var updateRequestOrderDts []models.RequestOrderDt
	var deleteRequestOrderDtIDs []uint

	processedExistingDts := make(map[uint]bool)

	for i, dt := range requestOrderDts {
		key := fmt.Sprintf("%d-%d-%d",
			utils.GetValueOrDefault(dt.ProductID, 0),
			utils.GetValueOrDefault(dt.RefID, 0),
			utils.GetValueOrDefault(dt.ItemID, 0))

		if existingDt, exists := existingDtMap[key]; exists {
			requestOrderDts[i].ID = *existingDt.ID
			requestOrderDts[i].UpdatedByID = &userID
			updateRequestOrderDts = append(updateRequestOrderDts, requestOrderDts[i])
			processedExistingDts[*existingDt.ID] = true
		} else {
			requestOrderDts[i].CreatedByID = &userID
			requestOrderDts[i].CreatedAt = time.Now()
			createRequestOrderDts = append(createRequestOrderDts, requestOrderDts[i])
		}
	}

	for _, dt := range existingRequestOrder.RequestOrderDts {
		if dt.ID != nil && !processedExistingDts[*dt.ID] {
			deleteRequestOrderDtIDs = append(deleteRequestOrderDtIDs, *dt.ID)
		}
	}

	if len(deleteRequestOrderDtIDs) > 0 {
		tx, err = s.repo.DeleteRequestOrderDtsByIDs(tx, deleteRequestOrderDtIDs, userID, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if len(createRequestOrderDts) > 0 {
		tx, err = s.repo.BulkCreateRequestOrderDts(tx, createRequestOrderDts, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if len(updateRequestOrderDts) > 0 {
		tx, err = s.repo.BulkUpdateRequestOrderDts(tx, updateRequestOrderDts, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	return &requestOrder, nil
}

func (s *RequestOrderService) DeleteRequestOrder(ctx *fiber.Ctx, requestOrderID uint, userID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("RequestOrderService-DeleteRequestOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	params := dtos.GetRequestOrderParams{ID: requestOrderID}
	_, err := s.GetRequestOrderByID(ctx, &params, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := s.repo.LockRequestOrder(tx, requestOrderID, childSpan); err != nil {
		tx.Rollback()
		return err
	}

	tx, err = s.repo.DeleteRequestOrder(tx, requestOrderID, userID, childSpan)
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *RequestOrderService) RestoreRequestOrder(ctx *fiber.Ctx, params *dtos.GetRequestOrderParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("RequestOrderService-RestoreRequestOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	isDeleted := 1
	getParams := &dtos.GetRequestOrderParams{
		ID:        params.ID,
		IsDeleted: &isDeleted,
	}

	_, err := s.GetRequestOrderByID(ctx, getParams, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := s.repo.RestoreRequestOrder(ctx, params, tx, childSpan); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *RequestOrderService) CreateRequestOrderDts(ctx *fiber.Ctx, req dtos.CreateRequestOrderRequest, userID uint, createdRequestOrder *models.RequestOrder, tx *gorm.DB, span opentracing.Span) (*gorm.DB, []models.RequestOrderDt, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-CreateRequestOrderDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	requestOrderDts, err := utils.MapCreateRequestOrderDts(ctx, req, createdRequestOrder, userID, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	tx, createdRequestOrderDts, err := s.repo.CreateRequestOrderDts(tx, requestOrderDts, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	return tx, createdRequestOrderDts, nil
}

func (s *RequestOrderService) GetRefSalesOrderDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefSalesOrderForRequestOrderListDTO, int, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-GetRefSalesOrderDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	soDts, total, err := s.repo.GetRefSalesOrderDts(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}

	return soDts, total, nil
}

func (s *RequestOrderService) GetRefProductForRequestOrder(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefProductForRequestOrderListDTO, int, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-GetRefProductForRequestOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	products, total, err := s.repo.GetRefProductForRequestOrder(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (s *RequestOrderService) GetWidgetRequestOrders(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RequestOrderStatusWidget, int, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-GetWidgetRequestOrders", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	requestOrders, total, err := s.repo.GetWidgetRequestOrders(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}
	return requestOrders, total, nil
}

func (s *RequestOrderService) BeginTransaction() *gorm.DB {
	return s.repo.BeginTransaction()
}

func (s *RequestOrderService) Commit(tx *gorm.DB) error {
	return s.repo.Commit(tx)
}

func (s *RequestOrderService) Rollback(tx *gorm.DB) *gorm.DB {
	return tx.Rollback()
}

func (s *RequestOrderService) Pdf(ctx *fiber.Ctx, req dtos.RequestOrderDetailDTO, tx *gorm.DB, span opentracing.Span) (*string, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-Pdf", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	form := req
	var requestOrder *dtos.RequestOrderDetailDTO
	var err error

	var params dtos.GetRequestOrderParams
	if req.IsIDOnly != nil && *req.IsIDOnly == 1 {
		params.ID = req.ID

		requestOrder, err = s.repo.GetRequestOrderByID(ctx, &params, tx, childSpan)
		if err != nil {
			return nil, err
		}

		requestOrderDts, err := s.repo.GetRequestOrderDts(ctx, requestOrder.ID, nil, childSpan)
		if err != nil {
			utils.LogErrors(childSpan, err)
			log.Printf("Failed to fetch requestOrderDts: %v", err)
		}

		if requestOrder.CompanyProfileID != nil {
			companyParams := &dtos.GetCompanyProfileParams{ID: uint(*requestOrder.CompanyProfileID)}
			company, err := s.utilRepo.GetCompanyProfileByID(ctx, companyParams)
			if err != nil {
				utils.LogErrors(childSpan, err)
				log.Printf("Failed to fetch company: %v", err)
			} else if company != nil {
				requestOrder.Company = *company
			}
		}

		requestOrder.RequestOrderDts = requestOrderDts
		form = *requestOrder
		req.RequestNo = requestOrder.RequestNo
	}

	var num string
	if req.RequestNo != nil {
		num = *req.RequestNo
	} else {
		num = ""
	}

	data := dtos.RequestOrderPDFData{
		Num:  num,
		Form: form,
	}

	htmlFileName := "request-order-detail"
	log.Println("Pdf-htmlFileName-ro", htmlFileName)

	templateFile, err := templateFS.Open(fmt.Sprintf("templates/%s.html", htmlFileName))
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Printf("Failed to open embedded template: %v", err)
		return nil, err
	}

	templateContent, err := io.ReadAll(templateFile)
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Printf("Failed to read template content: %v", err)
		return nil, err
	}

	htmlFile, err := os.CreateTemp("", fmt.Sprintf("%s-*.html", htmlFileName))
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error writing htmlFile:", err)
		return nil, err
	}
	defer os.Remove(htmlFile.Name())

	funcMap := template.FuncMap{
		"formatNumber": func(n float64, args ...int) string {
			decimals := 2
			if len(args) > 0 {
				decimals = args[0]
			}

			format := fmt.Sprintf("%%.%df", decimals)
			p := message.NewPrinter(language.English)
			return p.Sprintf(format, n)
		},
		"inc": func(i int) int {
			return i + 1
		},
	}

	tmpl, err := template.New(fmt.Sprintf("%s.html", htmlFileName)).Funcs(funcMap).Parse(string(templateContent))
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Printf("Failed to parse template: %v", err)
		return nil, err
	}

	if err := tmpl.Execute(htmlFile, data); err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error Execute:", err)
		return nil, err
	}

	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error pdfg:", err)
		return nil, err
	}

	headerContent, err := templateFS.ReadFile("templates/header.html")
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	footerContent, err := templateFS.ReadFile("templates/footer.html")
	if err != nil {
		return nil, fmt.Errorf("failed to read footer: %w", err)
	}

	headerPath, err := createTempFileFromEmbed(string(headerContent))
	if err != nil {
		return nil, fmt.Errorf("failed to create header temp file: %w", err)
	}
	defer os.Remove(headerPath)

	footerPath, err := createTempFileFromEmbed(string(footerContent))
	if err != nil {
		return nil, fmt.Errorf("failed to create footer temp file: %w", err)
	}
	defer os.Remove(footerPath)

	page := wkhtmltopdf.NewPage(htmlFile.Name())
	page.EnableLocalFileAccess.Set(true)
	page.HeaderHTML.Set("file://" + headerPath)
	page.FooterHTML.Set("file://" + footerPath)
	page.FooterSpacing.Set(10)

	pdfg.AddPage(page)
	pdfg.MarginLeft.Set(0)
	pdfg.MarginRight.Set(0)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)

	if err := pdfg.Create(); err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error pdfg create:", err)
		return nil, err
	}

	uploadDir := "./public/generated_pdfs"

	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		utils.LogErrors(childSpan, err)
		log.Println("Error mkdirall:", err)
		return nil, err
	}

	fileName := fmt.Sprintf("request-order-%s.pdf", time.Now().Format("20060102150405"))
	pdfPath := filepath.Join(uploadDir, fileName)
	if err := pdfg.WriteFile(pdfPath); err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error pdf path:", err)
		return nil, err
	}

	return &pdfPath, nil
}

// github.com/xuri/excelize/v2
func (s *RequestOrderService) Csv(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-Csv", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"

	exportType := utils.GetStringOrDefault(filters["export_type"], "all")

	if exportType == "detail" {
		return s.CsvGetDetail(ctx, filters, childSpan)
	}

	return s.CsvGetAll(ctx, filters, childSpan)
}

// CsvGetAll
func (s *RequestOrderService) CsvGetAll(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-CsvGetAll", opentracing.ChildOf(span.Context()))
	salesOrders, _, err := s.GetRequestOrders(ctx, filters, childSpan)
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
	csv += "Purchase Orders\n"
	csv += "\n"

	utils.BuildRequestOrderAllCSVRows(salesOrders, &csv)

	return []byte(csv), nil
}

func (s *RequestOrderService) CsvGetDetail(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-CsvGetDetail", opentracing.ChildOf(span.Context()))
	quotations, _, err := s.GetRequestOrdersDetails(ctx, filters, childSpan)
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
	csv += "Purchase Orders\n"
	csv += "\n"

	// csv += "ID,Sales Order No,Order Type,Customer,Expired Date,Quot Date,Currency,Total,Status,Created By,Updated By\n"

	// // Build CSV rows
	// for _, quotation := range quotations {
	// 	csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s,%.2f,%s,%s,%s\n",
	// 		quotation.ID,
	// 		utils.GetPtrVal(quotation.Remark),
	// 	)
	// }
	utils.BuildRequestOrderDetailCSVRows(quotations, &csv)

	return []byte(csv), nil
}

func (s *RequestOrderService) GetRequestOrdersDetails(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RequestOrderDetailDTO, int, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-GetRequestOrdersDetails", opentracing.ChildOf(span.Context()))

	salesOrders, total, err := s.repo.GetRequestOrderDetails(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}

	salesOrderIDs := utils.GetRequestOrderDetailsIDs(salesOrders)

	// Get QuoDts by salesOrder IDs
	soDts, _, err := s.repo.GetRequestOrderDetailsDts(ctx, filters, salesOrderIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}

	salesOrders = utils.MapGetRequestOrderDetails(salesOrders, soDts)

	return salesOrders, total, nil
}
