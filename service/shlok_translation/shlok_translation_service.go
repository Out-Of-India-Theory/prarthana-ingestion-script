package shlok_translation

import (
	"context"
	"errors"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/entity"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/repository/openai"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/zoho"
	"go.uber.org/zap"
	"log"
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
