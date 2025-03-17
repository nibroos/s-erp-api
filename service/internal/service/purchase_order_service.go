package service

import (
	"fmt"
	"math"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type PurchaseOrderService struct {
	repo     *repository.PurchaseOrderRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewPurchaseOrderService(repo *repository.PurchaseOrderRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *PurchaseOrderService {
	return &PurchaseOrderService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *PurchaseOrderService) CreatePurchaseOrder(ctx *fiber.Ctx, po *models.PurchaseOrder, poDts []models.PurchaseOrderDt, poDtBoms []models.PurchaseOrderDtBom, tx *gorm.DB, span opentracing.Span) (*models.PurchaseOrder, *gorm.DB, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-CreatePurchaseOrder", opentracing.ChildOf(span.Context()))

	if err := s.validateCalculations(po, poDts, poDtBoms); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	tx, err := s.repo.CreatePurchaseOrder(tx, po, poDts, poDtBoms, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	return po, tx, nil
}

func (r *PurchaseOrderService) validateCalculations(po *models.PurchaseOrder, poDts []models.PurchaseOrderDt, poDtBoms []models.PurchaseOrderDtBom) error {
	for _, dt := range poDts {
		if dt.Qty != nil && dt.Price != nil && dt.Subtotal != nil {
			calculatedSubtotal := *dt.Qty * *dt.Price
			if math.Abs(calculatedSubtotal-*dt.Subtotal) > 0.01 {
				return fmt.Errorf("subtotal mismatch for product %d", *dt.ProductID)
			}
		}

		if dt.DiscountPercentage != nil && dt.Price != nil && dt.DiscountPercentageNum != nil && *dt.DiscountPercentage > 0 {
			calculatedDiscountPercentageNum := *dt.Price * (*dt.DiscountPercentage / 100)
			if math.Abs(calculatedDiscountPercentageNum-*dt.DiscountPercentageNum) > 0.01 {
				return fmt.Errorf("discount percentage num mismatch for product %d", *dt.ProductID)
			}
		}

		if dt.DiscountPercentage != nil && dt.Subtotal != nil && dt.DiscountPercentageAmount != nil && *dt.DiscountPercentage > 0 {
			calculatedDiscountAmount := *dt.Subtotal * (*dt.DiscountPercentage / 100)
			if math.Abs(calculatedDiscountAmount-*dt.DiscountPercentageAmount) > 0.01 {
				return fmt.Errorf("discount percentage amount mismatch for product %d", *dt.ProductID)
			}
		}

		if dt.DiscountFinal != nil {
			var calculatedDiscountFinal float64
			if dt.DiscountAmount != nil {
				calculatedDiscountFinal += *dt.DiscountAmount
			}
			if dt.DiscountPercentageAmount != nil {
				calculatedDiscountFinal += *dt.DiscountPercentageAmount
			}
			if math.Abs(calculatedDiscountFinal-*dt.DiscountFinal) > 0.01 {
				return fmt.Errorf("discount final mismatch for product %d", *dt.ProductID)
			}
		}

		if dt.VatPercentage != nil && dt.Subtotal != nil && dt.VatPercentageAmount != nil && *dt.VatPercentage > 0 {
			calculatedVatAmount := *dt.Subtotal * (*dt.VatPercentage / 100)
			if math.Abs(calculatedVatAmount-*dt.VatPercentageAmount) > 0.01 {
				return fmt.Errorf("VAT amount mismatch for product %d", *dt.ProductID)
			}
		}

		if dt.Subtotal != nil && dt.TotalAmount != nil {
			calculatedTotal := *dt.Subtotal
			if dt.VatPercentageAmount != nil {
				calculatedTotal += *dt.VatPercentageAmount
			}
			if dt.DiscountFinal != nil {
				calculatedTotal -= *dt.DiscountFinal
			}
			if math.Abs(calculatedTotal-*dt.TotalAmount) > 0.01 {
				return fmt.Errorf("total amount mismatch for product %d", *dt.ProductID)
			}
		}
	}

	for _, bom := range poDtBoms {
		if bom.Qty != nil && bom.Price != nil && bom.Subtotal != nil && bom.ProductID != nil {
			calculatedBomSubtotal := *bom.Qty * *bom.Price
			if math.Abs(calculatedBomSubtotal-*bom.Subtotal) > 0.01 {
				return fmt.Errorf("BOM subtotal mismatch for product %d", *bom.ProductID)
			}
		}
	}

	var calculatedProductDiscountTotal float64
	for _, dt := range poDts {
		if dt.DiscountFinal != nil {
			calculatedProductDiscountTotal += *dt.DiscountFinal
		}
	}

	if po.DiscountAmountProduct != nil {
		if math.Abs(calculatedProductDiscountTotal-*po.DiscountAmountProduct) > 0.01 {
			return fmt.Errorf("product discount amount mismatch")
		}
	}

	var calculatedHeaderDiscount float64
	if po.DiscountAmount != nil {
		calculatedHeaderDiscount += *po.DiscountAmount
	}
	if po.DiscountPercentage != nil && po.TotalAmountProduct != nil && po.DiscountPercentageAmount != nil && *po.DiscountPercentage > 0 {
		calculatedDiscountPercentageAmount := *po.TotalAmountProduct * (*po.DiscountPercentage / 100)
		if math.Abs(calculatedDiscountPercentageAmount-*po.DiscountPercentageAmount) > 0.01 {
			return fmt.Errorf("header discount percentage amount mismatch")
		}
		calculatedHeaderDiscount += calculatedDiscountPercentageAmount
	}

	if po.DiscountFinalHeader != nil {
		if math.Abs(calculatedHeaderDiscount-*po.DiscountFinalHeader) > 0.01 {
			return fmt.Errorf("header final discount mismatch")
		}
	}

	if po.TotalDiscount != nil {
		calculatedTotalDiscount := calculatedProductDiscountTotal + calculatedHeaderDiscount
		if math.Abs(calculatedTotalDiscount-*po.TotalDiscount) > 0.01 {
			return fmt.Errorf("total discount mismatch")
		}
	}

	var calculatedHeaderSubtotal float64
	for _, dt := range poDts {
		if dt.Subtotal != nil {
			calculatedHeaderSubtotal += *dt.Subtotal
		}
	}

	if po.Subtotal != nil {
		if math.Abs(calculatedHeaderSubtotal-*po.Subtotal) > 0.01 {
			return fmt.Errorf("header subtotal mismatch")
		}
	}

	var calculatedTotalAmount float64
	var calculatedTotalQty float64
	for _, dt := range poDts {
		if dt.TotalAmount != nil {
			calculatedTotalAmount += *dt.TotalAmount
		}
		if dt.Qty != nil {
			calculatedTotalQty += *dt.Qty
		}
	}

	if po.TotalAmountProduct != nil {
		if math.Abs(calculatedTotalAmount-*po.TotalAmountProduct) > 0.01 {
			return fmt.Errorf("total amount products mismatch")
		}
	}

	if po.TotalQty != nil {
		if math.Abs(calculatedTotalQty-*po.TotalQty) > 0.01 {
			return fmt.Errorf("total quantity mismatch")
		}
	}

	if po.VatPercentage != nil && po.TotalAmountProduct != nil && *po.VatPercentage > 0 {
		calculatedVatAmount := *po.TotalAmountProduct * (*po.VatPercentage / 100)
		if po.VatPercentageAmount != nil {
			if math.Abs(calculatedVatAmount-*po.VatPercentageAmount) > 0.01 {
				return fmt.Errorf("VAT percentage amount mismatch")
			}
		}
		if po.TotalVat != nil {
			if math.Abs(calculatedVatAmount-*po.TotalVat) > 0.01 {
				return fmt.Errorf("total VAT mismatch")
			}
		}
	}

	if po.Pph23Percentage != nil && po.TotalAmountProduct != nil && *po.Pph23Percentage > 0 {
		calculatedPph23Amount := *po.TotalAmountProduct * (*po.Pph23Percentage / 100)
		if po.Pph23PercentageAmount != nil {
			if math.Abs(calculatedPph23Amount-*po.Pph23PercentageAmount) > 0.01 {
				return fmt.Errorf("PPh23 percentage amount mismatch")
			}
		}
		if po.TotalPph23 != nil {
			if math.Abs(calculatedPph23Amount-*po.TotalPph23) > 0.01 {
				return fmt.Errorf("total PPh23 mismatch")
			}
		}
	}

	if po.GrandTotal != nil && po.TotalAmountProduct != nil {
		calculatedGrandTotal := *po.TotalAmountProduct
		if po.TotalPph23 != nil {
			calculatedGrandTotal += *po.TotalPph23
		}
		if po.TotalVat != nil {
			calculatedGrandTotal += *po.TotalVat
		}
		if po.TotalDiscount != nil {
			calculatedGrandTotal -= *po.TotalDiscount
		}
		if math.Abs(calculatedGrandTotal-*po.GrandTotal) > 0.01 {
			return fmt.Errorf("grand total mismatch")
		}
	}

	return nil
}
