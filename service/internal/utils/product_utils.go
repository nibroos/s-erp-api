package utils

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/auth"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/opentracing/opentracing-go"
)

func MapCreateUpdateItemUnits(ctx *fiber.Ctx, req []dtos.CreateMsItemUnitsRequest, productID uint, childSpan opentracing.Span) ([]map[string]interface{}, []map[string]interface{}, []uint, error) {
	// Create a slice to hold the item units
	createItemUnits := []map[string]interface{}{}
	updateItemUnits := []map[string]interface{}{}
	IDs := make([]uint, 0)

	claims, _ := auth.GetAuthUser(ctx)
	userID := claims["user_id"]

	// Iterate over the item units in the request
	for _, itemUnit := range req {
		itemUnitModel := map[string]interface{}{
			"product_id": productID,
			"price_sell": itemUnit.PriceSell,
			"price_buy":  itemUnit.PriceBuy,
			"unit_id":    itemUnit.UnitID,
			"conversion": itemUnit.Conversion,
			"status":     1,
		}

		// log.Println("itemUnitModel", itemUnitModel)
		// Check if the item unit has an ID
		if itemUnit.ItemUnitID == nil {
			itemUnitModel["id"] = 0
			itemUnitModel["created_by_id"] = userID
			itemUnitModel["created_at"] = time.Now()
			createItemUnits = append(createItemUnits, itemUnitModel)
		} else {
			IDs = append(IDs, *itemUnit.ItemUnitID)
			itemUnitModel["id"] = *itemUnit.ItemUnitID
			itemUnitModel["updated_by_id"] = userID
			itemUnitModel["updated_at"] = time.Now()
			updateItemUnits = append(updateItemUnits, itemUnitModel)
		}
	}

	log.Println("createItemUnits", createItemUnits)
	log.Println("updateItemUnits", updateItemUnits)

	return createItemUnits, updateItemUnits, IDs, nil
}

func MapGetUnitIDs(ctx *fiber.Ctx, req []dtos.CreateMsItemUnitsRequest, childSpan opentracing.Span) ([]uint, error) {
	// Create a slice to hold the unit IDs
	unitIDs := make([]uint, 0)

	// Iterate over the item units in the request
	for _, itemUnit := range req {
		// Check if the item unit has an ID
		if itemUnit.ItemUnitID != nil {
			unitIDs = append(unitIDs, *itemUnit.ItemUnitID)
		}
	}

	log.Println("unitIDs", unitIDs)

	return unitIDs, nil
}

func MapItemUnitsByProductIDsAndUnitIDs(ctx *fiber.Ctx, productID uint, oldItemUnits []dtos.ItemUnitDetailDTO, newItemUnit []dtos.CreateMsItemUnitsRequest, childSpan opentracing.Span) ([]dtos.CreateMsItemUnitsRequest, error) {
	itemUnits := make([]dtos.CreateMsItemUnitsRequest, 0)

	for _, reqItemUnit := range newItemUnit {
		if reqItemUnit.ID == nil || *reqItemUnit.ID == 0 {
			log.Println("MapItemUnitsByProductIDsAndUnitIDs-reqItemUnit.ID1", reqItemUnit)

			isUpdate := false
			for _, oldItemUnit := range oldItemUnits {
				if oldItemUnit.UnitID == &reqItemUnit.UnitID {
					reqItemUnit.ID = &oldItemUnit.ID
					itemUnits = append(itemUnits, reqItemUnit)
					isUpdate = true
					log.Println("MapItemUnitsByProductIDsAndUnitIDs-reqItemUnit-isUpdate", reqItemUnit)
				}
			}

			if !isUpdate {
				log.Println("MapItemUnitsByProductIDsAndUnitIDs-reqItemUnit-!isUpdate", reqItemUnit)
				itemUnits = append(itemUnits, reqItemUnit)
			}
		} else {
			log.Println("MapItemUnitsByProductIDsAndUnitIDs-reqItemUnit.ID2", reqItemUnit.ID, reqItemUnit.PriceSell)
			itemUnits = append(itemUnits, reqItemUnit)
		}
	}
	log.Println("MapItemUnitsByProductIDsAndUnitIDs-itemUnits", itemUnits)

	return itemUnits, nil
}
