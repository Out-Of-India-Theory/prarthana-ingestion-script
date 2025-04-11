package shlok_translation

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/entity"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/repository/openai"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/zoho"
	"go.uber.org/zap"
)

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

	translatedRecords := entity.ShlokaSheetResponse{}
	languages := []string{"english", "kannada", "hindi", "telugu", "bengali", "marathi", "tamil", "gujarati", "odiya", "malayalam", "assamese", "punjabi"}

	for i := startId - 1; i < endId && i < len(response.Records); i++ {
		record := response.Records[i]
		newRecord := make(map[string]interface{})
		newRecord["ID"] = record["ID"]
		newRecord["Name (Optional)"] = record["Name (Optional)"]
		newRecord["text_sanskrit"] = record["text_sanskrit"]
		textSanskrit, ok := record["text_sanskrit"].(string)
		if !ok || strings.TrimSpace(textSanskrit) == "" {
			return fmt.Errorf("missing or invalid 'text_sanskrit' at row %d", i+1)
		}
		for _, lang := range languages {
			textKey := "text_" + lang
			translationKey := "translation_" + lang
			textVal := s.GetTranslation(textSanskrit, lang, false)
			translationVal := s.GetTranslation(textSanskrit, lang, true)
			if strings.TrimSpace(textVal) == "" {
				textVal = "[MISSING]"
			}
			if strings.TrimSpace(translationVal) == "" {
				translationVal = "[MISSING]"
			}
			newRecord[textKey] = textVal
			newRecord[translationKey] = translationVal
		}
		translatedRecords.Records = append(translatedRecords.Records, newRecord)
	}
	err = s.zohoService.SetSheetData(ctx, "shloka", translatedRecords, startId)
	if err != nil {
		return fmt.Errorf("failed to set sheet data: %w", err)
	}
	return nil
}
