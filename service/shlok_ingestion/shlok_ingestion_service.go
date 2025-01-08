package shlok_ingestion

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/entity"
	mongoRepo "github.com/Out-Of-India-Theory/prarthana-automated-script/repository/mongo/prarthana_data"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/service/util"
	"go.uber.org/zap"
	"log"
	"os"
	"strconv"
)

type ShlokIngestionService struct {
	logger                   *zap.Logger
	prarthanaMongoRepository mongoRepo.MongoRepository
}

func InitShlokIngestionService(ctx context.Context,
	prarthanaMongoRepository mongoRepo.MongoRepository,
) *ShlokIngestionService {
	return &ShlokIngestionService{
		logger:                   logging.WithContext(ctx),
		prarthanaMongoRepository: prarthanaMongoRepository,
	}
}

func (s *ShlokIngestionService) ShlokIngestion(ctx context.Context, csvFilePath string, startID, endID int) error {
	file, err := os.Open(csvFilePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("error reading header: %w", err)
	}

	fieldMap := make(map[string]int)
	for i, field := range header {
		fieldMap[field] = i
	}

	var shloks []entity.Shlok

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading records: %w", err)
	}
	for i, record := range records {
		log.Printf("Processing record %d\n", i+1) // Log the current record number

		if len(record) <= fieldMap["ID"] {
			log.Printf("Skipping record %d: Missing ID field\n", i+1)
			continue
		}
		id, err := strconv.Atoi(record[fieldMap["ID"]])
		if err != nil {
			log.Printf("Skipping record %d: Invalid ID format\n", i+1)
			continue
		}
		if id < startID || id > endID {
			continue
		}

		if len(record) <= fieldMap["Name (Optional)"] {
			log.Printf("Skipping record %d: Missing Name field\n", i+1)
			continue
		}

		shlok := entity.Shlok{
			ID: record[fieldMap["ID"]],
			Title: map[string]string{
				"default": record[fieldMap["Name (Optional)"]],
			},
			Explanation: make(map[string]string),
			Shlok:       make(map[string]string),
		}

		explanationKeys := util.ExtractLanguageKeys(fieldMap, "translation_")
		shlokKeys := util.ExtractLanguageKeys(fieldMap, "text_")
		for lang, index := range explanationKeys {
			if index < len(record) && record[index] != "" {
				if lang == "english" {
					shlok.Explanation["default"] = record[index]
				} else {
					shlok.Explanation[lang] = record[index]
				}
			} else {
				log.Printf("Warning: Missing translation for language '%s' in record %d\n", lang, i+1)
			}
		}
		for lang, index := range shlokKeys {
			if index < len(record) && record[index] != "" {
				if lang == "sanskrit" {
					shlok.Shlok["default"] = record[index]
				} else {
					shlok.Shlok[lang] = record[index]
				}
			} else {
				log.Printf("Warning: Missing prarthana_data for language '%s' in record %d\n", lang, i+1)
			}
		}
		shloks = append(shloks, shlok)
	}
	return s.prarthanaMongoRepository.InsertManyShloks(ctx, shloks)
}
