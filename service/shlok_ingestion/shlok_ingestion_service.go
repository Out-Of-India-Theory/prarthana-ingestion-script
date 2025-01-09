package shlok_ingestion

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/entity"
	mongoRepo "github.com/Out-Of-India-Theory/prarthana-automated-script/repository/mongo/prarthana_data"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/service/zoho"
	"go.uber.org/zap"
)

type ShlokIngestionService struct {
	logger                   *zap.Logger
	prarthanaMongoRepository mongoRepo.MongoRepository
	zohoService              zoho.Service
}

func InitShlokIngestionService(ctx context.Context,
	prarthanaMongoRepository mongoRepo.MongoRepository,
	zohoService zoho.Service,
) *ShlokIngestionService {
	return &ShlokIngestionService{
		logger:                   logging.WithContext(ctx),
		prarthanaMongoRepository: prarthanaMongoRepository,
		zohoService:              zohoService,
	}
}

func (s *ShlokIngestionService) ShlokIngestion(ctx context.Context, startID, endID int) error {
	var response entity.ShlokaSheetResponse
	err := s.zohoService.GetSheetData(ctx, "shloka", &response)
	if err != nil {
		return err
	}
	if len(response.Records) == 0 {
		return errors.New("no records found")
	}

	var shloks []entity.Shlok
	langs := []string{"sanskrit", "english", "kannada", "hindi", "telugu", "bengali", "marathi", "tamil", "gujarati", "odiya", "malayalam", "assamese", "punjabi"}
	for i, record := range response.Records {
		log.Printf("Processing record %d\n", i+1) // Log the current record number
		idf, ok := record["ID"].(float64)
		if !ok {
			return fmt.Errorf("invalid ID")
		}
		id := int(idf)
		if id < startID || id > endID {
			continue
		}
		name, ok := record["Name (Optional)"].(string)
		if !ok {
			name = ""
		}
		shlok := entity.Shlok{
			ID:    strconv.Itoa(id),
			IntId: id,
			Title: map[string]string{
				"default": name,
			},
			Explanation: make(map[string]string),
			Shlok:       make(map[string]string),
		}

		for _, lang := range langs {
			tmpLang := lang
			if lang == "english" {
				tmpLang = "default"
			}
			value, ok1 := record[fmt.Sprintf("translation_%s", tmpLang)].(string)
			if !ok1 {
				fmt.Println("no text present for language", tmpLang)
				continue
			}
			shlok.Explanation[tmpLang] = value
		}

		for _, lang := range langs {
			tmpLang := lang
			if lang == "english" {
				tmpLang = "default"
			}
			value, ok1 := record[fmt.Sprintf("text_%s", tmpLang)].(string)
			if !ok1 {
				fmt.Println("no text present for language", tmpLang)
				continue
			}
			shlok.Shlok[tmpLang] = value
		}
		shloks = append(shloks, shlok)
	}
	if len(shloks) == 0 {
		return errors.New("no shloks to ingest")
	}
	return s.prarthanaMongoRepository.InsertManyShloks(ctx, shloks)
}
