package utils

import (
	"log"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
)

func MapQuoDtBomsToQuoDts(quoDtBoms []dtos.QuotationQuoDtBomListDTO, quoDts []dtos.QuotationQuoDtListDTO) []dtos.QuotationQuoDtListDTO {
	combinedQuoDts := []dtos.QuotationQuoDtListDTO{}

	log.Println("init quodtbom", quoDtBoms)

	// test for quoDtBoms
	for _, quoDtBom := range quoDtBoms {
		log.Println("quoDtBomabc", quoDtBom)
	}

	for _, quoDt := range quoDts {
		quoDtBoms := make([]dtos.QuotationQuoDtBomListDTO, 0)
		for _, quoDtBom := range quoDtBoms {
			log.Println("ptr quoDt.ID", *quoDt.ID, "quoDtBom.QuoDtID", quoDtBom.QuoDtID)
			if quoDtBom.QuoDtID == quoDt.ID {
				quoDtBoms = append(quoDtBoms, quoDtBom)
				log.Println("quoDtBom-match", quoDtBom)
			}
		}
		log.Println("quoDtBoms", quoDtBoms)
		quoDt.QuoDtsBoms = quoDtBoms
		log.Println("quoDt", quoDt)

		test := "asdasd"
		quoDt.GenCode = &test

		combinedQuoDts = append(combinedQuoDts, quoDt)

		log.Println("quoDt.GenCode", *quoDt.GenCode)
	}

	log.Println("combinedQuoDts", combinedQuoDts)

	// test changed
	for _, quoDt := range combinedQuoDts {
		log.Println("quoDt.GenCode", *quoDt.GenCode)
	}

	return combinedQuoDts
}
