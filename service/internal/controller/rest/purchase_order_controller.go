package rest

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/middleware"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/nibroos/s-erp-api/service/internal/validators/form_requests"
	"github.com/opentracing/opentracing-go"
)

type PurchaseOrderController struct {
	service *service.PurchaseOrderService
	repo    *repository.PurchaseOrderRepository
	tracer  opentracing.Tracer
}

func NewPurchaseOrderController(service *service.PurchaseOrderService, repo *repository.PurchaseOrderRepository, tracer opentracing.Tracer) *PurchaseOrderController {
	return &PurchaseOrderController{service: service, repo: repo, tracer: tracer}
}

func (c *PurchaseOrderController) CreatePurchaseOrder(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("PurchaseOrderController-CreatePurchaseOrder", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.CreatePurchaseOrderRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	reqValidator, isValid := form_requests.NewPurchaseOrderStoreRequest().Validate(&req, ctx)
	if !isValid {
		return utils.ErrValidResponse(ctx, apiSpan, "Failed to create purchase order", reqValidator)
	}

	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		utils.LogErrors(parentSpan, err)
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	userID := uint(claims["user_id"].(float64))

	purchaseOrder := models.PurchaseOrder{
		CustomerID:               req.CustomerID,
		PurchaseTypeID:           req.PurchaseTypeID,
		CurrencyID:               req.CurrencyID,
		VatID:                    req.VatID,
		VatPercentage:            req.VatPercentage,
		VatPercentageAmount:      req.VatPercentageAmount,
		PaymentTermID:            req.PaymentTermID,
		ShippingTermID:           req.ShippingTermID,
		Pph23ID:                  req.Pph23ID,
		Pph23Percentage:          req.Pph23Percentage,
		Pph23PercentageAmount:    req.Pph23PercentageAmount,
		BranchID:                 req.BranchID,
		PoNo:                     req.PoNo,
		PoDate:                   req.PoDate,
		DeliveryDate:             req.DeliveryDate,
		ShippingDestination:      req.ShippingDestination,
		Remark:                   req.Remark,
		ExchangeRate:             req.ExchangeRate,
		DiscountAmount:           req.DiscountAmount,
		DiscountPercentage:       req.DiscountPercentage,
		DiscountPercentageAmount: req.DiscountPercentageAmount,
		DiscountFinalHeader:      req.DiscountFinalHeader,
		DiscountAmountProduct:    req.DiscountAmountProduct,
		Subtotal:                 req.Subtotal,
		TotalAmountProduct:       req.TotalAmountProduct,
		TotalQty:                 req.TotalQty,
		TotalDiscount:            req.TotalDiscount,
		TotalPph23:               req.TotalPph23,
		TotalVat:                 req.TotalVat,
		GrandTotal:               req.GrandTotal,
		Status:                   req.Status,
		CreatedByID:              &userID,
	}

	var poDts []models.PurchaseOrderDt
	for _, dt := range req.PoDts {
		poDt := models.PurchaseOrderDt{
			ItemUnitID:               dt.ItemUnitID,
			VatID:                    dt.VatID,
			ProductID:                &dt.ProductID,
			RefID:                    &dt.ProductID,
			RefType:                  dt.RefType,
			RefJSON:                  dt.RefJSON,
			ProductType:              dt.ProductType,
			ProductJSON:              dt.ProductJSON,
			Remark:                   dt.Remark,
			NeedQty:                  dt.NeedQty,
			Qty:                      dt.Qty,
			Price:                    dt.Price,
			Subtotal:                 dt.Subtotal,
			DiscountAmount:           dt.DiscountAmount,
			DiscountPercentage:       dt.DiscountPercentage,
			DiscountPercentageNum:    dt.DiscountPercentageNum,
			DiscountPercentageAmount: dt.DiscountPercentageAm,
			DiscountFinal:            dt.DiscountFinal,
			VatPercentage:            dt.VatPercentage,
			VatPercentageAmount:      dt.VatPercentageAmount,
			TotalAmount:              dt.TotalAmount,
			CreatedByID:              &userID,
		}
		poDts = append(poDts, poDt)
	}

	var poDtBoms []models.PurchaseOrderDtBom
	for _, dt := range req.PoDts {
		for _, bom := range dt.PoDtBoms {
			poDtBom := models.PurchaseOrderDtBom{
				ProductID:   &bom.ProductID,
				BomID:       &bom.BomID,
				ItemUnitID:  bom.ItemUnitID,
				ProductJSON: bom.ProductJSON,
				Remark:      bom.Remark,
				Qty:         bom.Qty,
				Price:       bom.Price,
				Subtotal:    bom.Subtotal,
				CreatedByID: &userID,
			}
			poDtBoms = append(poDtBoms, poDtBom)
		}
	}

	tx := c.repo.BeginTransaction()

	createdPO, tx, err := c.service.CreatePurchaseOrder(ctx, &purchaseOrder, poDts, poDtBoms, tx, parentSpan)
	if err != nil {
		return utils.ErrTrxResponse(ctx, tx, apiSpan, err, "Failed to create purchase order", http.StatusInternalServerError)
	}

	tx.Commit()

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{createdPO}, paginationMeta, "Purchase order created successfully", http.StatusCreated, nil, nil)
}
