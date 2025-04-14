package shlok_translation

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/entity"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/repository/openai"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/zoho"
	"go.uber.org/zap"
)

const maxWorkers = 5

type ShlokTranslationService struct {
	logger                 *zap.Logger
	zohoService            zoho.Service
	openaiClientRepository openai.ClientRepository
}

func InitShlokTranslationService(ctx context.Context,
	zohoService zoho.Service,
	openaiClientRepository openai.ClientRepository,
) *ShlokTranslationService {
	return &ShlokTranslationService{
		logger:                 logging.WithContext(ctx),
		zohoService:            zohoService,
		openaiClientRepository: openaiClientRepository,
	}
}

func (s *ShlokTranslationService) GetTranslation(text, lang string, isTranslation bool) string {
	translatedText, err := s.openaiClientRepository.TranslateText(text, lang, isTranslation)
	if err != nil {
		log.Printf("Error translating to %s: %v", lang, err)
		return text
	}
	return translatedText
}

func (s *ShlokTranslationService) GenerateShlokaTranslation(ctx context.Context, startId, endId int) error {
	var response entity.ShlokaSheetResponse
	err := s.zohoService.GetSheetData(ctx, "shloka", &response)
	if err != nil {
		return fmt.Errorf("failed to get sheet data: %w", err)
	}
	if len(response.Records) == 0 {
		return errors.New("no records found")
	}
	if startId < 0 || endId < 0 || startId > endId {
		return errors.New("invalid range of Id's")
	}

	languages := []string{"english", "kannada", "hindi", "telugu", "bengali", "marathi", "tamil", "gujarati", "odiya", "malayalam", "assamese", "punjabi"}
	batchSize := 5

	for batchStart := startId; batchStart <= endId; batchStart += batchSize {
		batchEnd := batchStart + batchSize - 1
		if batchEnd > endId {
			batchEnd = endId
		}

		for i := batchStart - 1; i < batchEnd && i < len(response.Records); i++ {
			record := response.Records[i]
			textSanskrit, ok := record["text_sanskrit"].(string)
			if !ok || strings.TrimSpace(textSanskrit) == "" {
				log.Printf("invalid text_sanskrit at row %d", i+1)
				continue
			}
			newRecord := []interface{}{}
			newRecord = append(newRecord, record["ID"], record["Name (Optional)"], textSanskrit)
			var wg sync.WaitGroup
			mu := sync.Mutex{}
			texts := make(map[string]string)
			for _, lang := range languages {
				wg.Add(2)
				go func(lang string) {
					defer wg.Done()
					text := s.GetTranslation(textSanskrit, lang, false)
					if strings.TrimSpace(text) == "" {
						text = "[MISSING]"
					}
					mu.Lock()
					texts["text_"+lang] = text
					mu.Unlock()
				}(lang)

				go func(lang string) {
					defer wg.Done()
					trans := s.GetTranslation(textSanskrit, lang, true)
					if strings.TrimSpace(trans) == "" {
						trans = "[MISSING]"
					}
					mu.Lock()
					texts["translation_"+lang] = trans
					mu.Unlock()
				}(lang)
			}
			wg.Wait()
			for _, lang := range languages {
				newRecord = append(newRecord, texts["text_"+lang], texts["translation_"+lang])
			}
			columnIndexes := make([]int, len(newRecord))
			for j := range newRecord {
				columnIndexes[j] = j + 1
			}
			err := s.zohoService.SetSheetData(ctx, "shloka", i+2, columnIndexes, newRecord)
			if err != nil {
				return fmt.Errorf("failed to write row %d: %w", i+1, err)
			}
			//log.Printf("✅ Successfully wrote row ID %d", i+1)
			time.Sleep(500 * time.Millisecond)
		}
	}
	return nil
}
