package repository

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type PurchaseOrderRepository struct {
	db       *gorm.DB
	sqlDB    *sqlx.DB
	utilRepo *UtilRepository
	tracer   opentracing.Tracer
}

func NewPurchaseOrderRepository(db *gorm.DB, sqlDB *sqlx.DB, utilRepo *UtilRepository, tracer opentracing.Tracer) *PurchaseOrderRepository {
	return &PurchaseOrderRepository{
		db:       db,
		sqlDB:    sqlDB,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (r *PurchaseOrderRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *PurchaseOrderRepository) Rollback() *gorm.DB {
	return r.db.Rollback()
}

func (r *PurchaseOrderRepository) CreatePurchaseOrder(tx *gorm.DB, po *models.PurchaseOrder, poDts []models.PurchaseOrderDt, poDtBoms []models.PurchaseOrderDtBom, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-CreatePurchaseOrder", opentracing.ChildOf(span.Context()))

	if err := tx.Create(po).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	productItems := []models.PurchaseOrderDt{}
	// roItems := []models.PurchaseOrderDt{}
	// soItems := []models.PurchaseOrderDt{}

	for i := range poDts {
		poDts[i].PoID = &po.ID

		if poDts[i].RefType != nil {
			switch *poDts[i].RefType {
			case "product":
				productItems = append(productItems, poDts[i])
			// case "request_order":
			//     roItems = append(roItems, poDts[i])
			// case "sales_order":
			//     soItems = append(soItems, poDts[i])
			default:
				err := fmt.Errorf("unsupported reference type: %s", *poDts[i].RefType)
				utils.LogErrors(childSpan, err)
				return nil, err
			}
		} else {
			err := fmt.Errorf("reference type is required for all items")
			utils.LogErrors(childSpan, err)
			return nil, err
		}
	}

	if len(productItems) > 0 {
		_, err := r.CreatePurchaseOrderFromProduct(tx, po, productItems, poDtBoms, childSpan)
		if err != nil {
			return nil, err
		}
	}

	// if len(roItems) > 0 {
	//     tx, err := r.CreatePurchaseOrderFromRO(tx, po, roItems, poDtBoms, childSpan)
	//     if err != nil {
	//         return nil, err
	//     }
	// }

	// if len(soItems) > 0 {
	//     tx, err := r.CreatePurchaseOrderFromSO(tx, po, soItems, poDtBoms, childSpan)
	//     if err != nil {
	//         return nil, err
	//     }
	// }

	return tx, nil
}

func (r *PurchaseOrderRepository) CreatePurchaseOrderFromProduct(tx *gorm.DB, po *models.PurchaseOrder, poDts []models.PurchaseOrderDt, poDtBoms []models.PurchaseOrderDtBom, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderRepository-CreatePurchaseOrderFromProduct", opentracing.ChildOf(span.Context()))

	var discountType string
	if po.DiscountAmount != nil && po.DiscountPercentage != nil &&
		*po.DiscountAmount > 0 && *po.DiscountPercentage > 0 {
		discountType = "all"
	} else if po.DiscountAmount != nil && *po.DiscountAmount > 0 {
		discountType = "amount"
	} else if po.DiscountPercentage != nil && *po.DiscountPercentage > 0 {
		discountType = "percentage"
	}
	po.DiscountType = &discountType

	for i := range poDts {
		poDts[i].PoID = &po.ID

		if poDts[i].NeedQty == nil {
			var needQty float64 = 0
			poDts[i].NeedQty = &needQty
		}

		genCode := fmt.Sprintf("%d/%02d/%02d/%d/%d",
			time.Now().Year(),
			time.Now().Month(),
			time.Now().Day(),
			*poDts[i].RefID,
			*poDts[i].ProductID)
		poDts[i].GenCode = &genCode

		var discountType string
		if poDts[i].DiscountAmount != nil && poDts[i].DiscountPercentage != nil &&
			*poDts[i].DiscountAmount > 0 && *poDts[i].DiscountPercentage > 0 {
			discountType = "all"
		} else if poDts[i].DiscountAmount != nil && *poDts[i].DiscountAmount > 0 {
			discountType = "amount"
		} else if poDts[i].DiscountPercentage != nil && *poDts[i].DiscountPercentage > 0 {
			discountType = "percentage"
		}
		poDts[i].DiscountType = &discountType
	}

	if err := tx.Create(&poDts).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	for i := range poDtBoms {
		poDtBoms[i].PoID = &po.ID
		poDtBoms[i].PoDtID = &poDts[0].ID

		if poDtBoms[i].BomID != nil && *poDtBoms[i].BomID != 0 {
			var bom models.Bom
			if err := tx.First(&bom, poDtBoms[i].BomID).Error; err != nil {
				utils.LogErrors(childSpan, err)
				return nil, err
			}
			poDtBoms[i].ProductID = bom.ProductID
		} else {
			poDtBoms[i].BomID = nil
		}

		genCode := fmt.Sprintf("%d/%02d/%02d/%d/%d",
			time.Now().Year(),
			time.Now().Month(),
			time.Now().Day(),
			*poDts[0].RefID,
			poDtBoms[i].ProductID)
		poDtBoms[i].GenCode = &genCode
	}

	if len(poDtBoms) > 0 {
		if err := tx.Create(&poDtBoms).Error; err != nil {
			utils.LogErrors(childSpan, err)
			return nil, err
		}
	}

	if err := tx.Save(po).Error; err != nil {
		utils.LogErrors(childSpan, err)
		return nil, err
	}

	return tx, nil
}
