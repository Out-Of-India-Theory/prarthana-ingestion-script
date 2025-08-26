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
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/repository/ai_platform"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/zoho"
	"go.uber.org/zap"
)

const maxWorkers = 5

type ShlokTranslationService struct {
	logger                     *zap.Logger
	zohoService                zoho.Service
	platformaiClientRepository ai_platform.ClientRepository
}

func InitShlokTranslationService(ctx context.Context,
	zohoService zoho.Service,
	platformaiClientRepository ai_platform.ClientRepository,
) *ShlokTranslationService {
	return &ShlokTranslationService{
		logger:                     logging.WithContext(ctx),
		zohoService:                zohoService,
		platformaiClientRepository: platformaiClientRepository,
	}
}

func (s *ShlokTranslationService) GetTranslation(ctx context.Context, text []string, lang string, isTransliterate bool) []ai_platform.TranslateText {
	translatedText, err := s.platformaiClientRepository.BatchTranslateText(ctx, text, lang, isTransliterate)
	if err != nil {
		log.Printf("Error translating to %s: %v", lang, err)
		return nil
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

			newRecord := []interface{}{record["ID"], record["Name (Optional)"], textSanskrit}

			var wg sync.WaitGroup
			mu := sync.Mutex{}
			texts := make(map[string]string)

			for _, lang := range languages {
				wg.Add(2)
				langCopy := lang
				go func(lang string) {
					defer wg.Done()
					textInput := []string{textSanskrit}
					raw := s.GetTranslation(ctx, textInput, lang, true)
					translated := "[MISSING]"
					if len(raw) > 0 && strings.TrimSpace(raw[0].Translation) != "" {
						translated = raw[0].Translation
					}
					mu.Lock()
					texts["text_"+lang] = translated
					mu.Unlock()
				}(langCopy)
			
				go func(lang string) {
					defer wg.Done()
					textInput := []string{textSanskrit}
					trans := s.GetTranslation(ctx, textInput, lang, false)
					translated := "[MISSING]"
					if len(trans) > 0 && strings.TrimSpace(trans[0].Translation) != "" {
						translated = trans[0].Translation
					}
					mu.Lock()
					texts["translation_"+lang] = translated
					mu.Unlock()
				}(langCopy)
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
			time.Sleep(500 * time.Millisecond)
		}
	}
	return nil
}
