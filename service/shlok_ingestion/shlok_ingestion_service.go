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
	// Open the provided CSV file
	file, err := os.Open(csvFilePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Create a CSV reader
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields per record

	// Read the CSV header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("error reading header: %w", err)
	}

	// Map CSV header to field indices in the Shlok struct
	fieldMap := make(map[string]int)
	for i, field := range header {
		fieldMap[field] = i
	}

	// Create a slice to store Shlok objects
	var shloks []entity.Shlok

	// Read remaining records from the CSV file
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading records: %w", err)
	}

	// Iterate over each record in the CSV file
	for i, record := range records {
		log.Printf("Processing record %d\n", i+1) // Log the current record number

		// Defensive check to avoid index out of range errors
		if len(record) <= fieldMap["ID"] {
			log.Printf("Skipping record %d: Missing ID field\n", i+1)
			continue
		}
		// Convert the ID from string to an integer
		id, err := strconv.Atoi(record[fieldMap["ID"]])
		if err != nil {
			log.Printf("Skipping record %d: Invalid ID format\n", i+1)
			continue
		}

		// Check if the ID is within the specified range
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

		// Extract language keys for Explanation and Shlok
		explanationKeys := util.ExtractLanguageKeys(fieldMap, "translation_")
		shlokKeys := util.ExtractLanguageKeys(fieldMap, "text_")

		// Fill in the Explanation map
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

		// Fill in the Shlok map
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
