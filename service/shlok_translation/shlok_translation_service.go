package shlok_translation

import (
	"context"
	"errors"
	"fmt"
	"log"

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
		return err
	}
	if len(response.Records) == 0 {
		return errors.New("no records found")
	}
	if startId < 0 || endId < 0 || startId > endId {
		return errors.New("invalid range of Id's")
	}

	translatedRecords := entity.ShlokaSheetResponse{}
	languages := []string{"english", "kannada", "hindi", "telugu", "bengali", "marathi", "tamil", "gujarati", "odiya", "malayalam", "assamese", "punjabi"}

	for i := startId - 1; i < len(response.Records); i++ {
		record := response.Records[i]
		newRecord := make(map[string]interface{})
		newRecord["ID"] = record["ID"]
		newRecord["Name (Optional)"] = record["Name (Optional)"]
		newRecord["text_sanskrit"] = record["text_sanskrit"]
		textSanskrit := record["text_sanskrit"].(string)

		for _, lang := range languages {
			textKey := "text_" + lang
			translationKey := "translation_" + lang
			translated := s.GetTranslation(textSanskrit, lang, true)
			textKeyValue := s.GetTranslation(textSanskrit, lang, false)
			newRecord[translationKey] = translated

			newRecord[textKey] = textKeyValue
		}

		translatedRecords.Records = append(translatedRecords.Records, newRecord)
	}

	fmt.Printf("Translated records: %v\n", translatedRecords.Records)

	err = s.zohoService.SetSheetData(ctx, "shloka", translatedRecords)
	return err
}

//func GenerateTranslation() {
//	// Input and output CSV file names
//	inputFile := "shlok.csv"
//	outputFile := "output.csv"
//
//	// Define language order
//	languages := []string{"english", "kannada", "hindi", "telugu", "bengali", "marathi", "tamil", "gujarati", "odiya", "malayalam", "assamese", "punjabi"}
//
//	// Construct headers
//	headers := []string{"ID", "Name (Optional)", "text_sanskrit"}
//	for _, lang := range languages {
//		headers = append(headers, "text_"+lang, "translation_"+lang)
//	}
//	writer.Write(headers)
//
//	// Process each row in CSV (excluding header)
//	for _, row := range records[1:] {
//		id := row[0]
//		name := ""
//		if len(row) > 1 {
//			name = row[1]
//		}
//
//		// Get actual Sanskrit text
//		textSanskrit := row[2]
//
//		// Add Sanskrit text
//		rowData := []string{id, name, textSanskrit}
//
//		// Append transliteration and translation in desired order
//		for _, lang := range languages {
//			transliterated := getTranslation(textSanskrit, lang, false)
//			translated := getTranslation(textSanskrit, lang, true)
//			rowData = append(rowData, transliterated, translated)
//		}
//
//		writer.Write(rowData)
//	}
//
//	fmt.Println("CSV file successfully created:", outputFile)
//}
