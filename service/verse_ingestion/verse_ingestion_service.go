package verse_ingestion

import (
	"context"
	"errors"
	"fmt"

	// "strconv"

	// "fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/entity"
	mongoRepo "github.com/Out-Of-India-Theory/prarthana-ingestion-script/repository/mongo/prarthana_data"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/zoho"

	// "github.com/Out-Of-India-Theory/prarthana-ingestion-script/util"
	// "github.com/go-audio/wav"
	// "github.com/hajimehoshi/go-mp3"
	"go.uber.org/zap"
	// "io"
	"log"
	// "math"
	// "net/http"
	// "os"
	// "path/filepath"
	// "regexp"
	// "strconv"
	// "strings"
	// "sync"
	// "time"
)

type VerseIngestionService struct {
	logger                   *zap.Logger
	prarthanaMongoRepository mongoRepo.MongoRepository
	zohoService              zoho.Service
}

func InitVerseIngestionService(ctx context.Context,
	prarthanaMongoRepository mongoRepo.MongoRepository,
	zohoService zoho.Service,
) *VerseIngestionService {
	return &VerseIngestionService{
		logger:                   logging.WithContext(ctx),
		prarthanaMongoRepository: prarthanaMongoRepository,
		zohoService:              zohoService,
	}
}

func (v *VerseIngestionService) VerseIngestion(ctx context.Context, startId int, endId int) error {
	var response entity.ShlokaSheetResponse
	err := v.zohoService.GetSheetData(ctx, "verses", &response)
	if err != nil {
		return err
	}
	if len(response.Records) == 0 {
		return errors.New("no records found")
	}

	var verses []entity.Verse
	langs := []string{"default", "kn", "hi", "te", "mr", "ta", "gu"}
	for i, record := range response.Records {
		log.Printf("Processing record %d\n", i+1) // Log the current record number
		idf, ok := record["Id"].(float64)
		if !ok {
			return fmt.Errorf("invalid Id")
		}
		id := int(idf)
		if id < startId || id > endId {
			continue
		}
		date, ok := record["Date"].(string)
		if !ok {
			date = ""
		}
		cn, ok := record["Chapter"].(float64)
		var chapterNo int64 = 0
		if ok {
			chapterNo = int64(cn)
		}
		vn, ok := record["verse"].(float64)
		var verseNo int64 = 0
		if ok {
			verseNo = int64(vn)
		}
		verse := entity.Verse{
			ID:                id,
			Date:              date,
			ScriptureName:     make(map[string]string),
			Chapter:           chapterNo,
			VerseNumber:       verseNo,
			Verse:             make(map[string]string),
			SimplifiedMeaning: make(map[string]string),
			Lesson:            make(map[string]string),
		}
		for _, lang := range langs {
			value, exists := record[fmt.Sprintf("scripture_name_%s", lang)].(string)
			if !exists || value == "" {
				log.Printf("Warning: Missing scripture name for language '%s' in record %d\n", lang, i+1)
				continue
			}
			verse.ScriptureName[lang] = value
		}

		for _, lang := range langs {
			value, exists := record[fmt.Sprintf("verse_%s", lang)].(string)
			if !exists || value == "" {
				log.Printf("Warning: Missing Verse for language '%s' in record %d\n", lang, i+1)
				continue
			}
			verse.Verse[lang] = value
		}

		for _, lang := range langs {
			value, exists := record[fmt.Sprintf("simplified_meaning_%s", lang)].(string)
			if !exists || value == "" {
				log.Printf("Warning: Missing Simplified Meaning for language '%s' in record %d\n", lang, i+1)
				continue
			}
			verse.SimplifiedMeaning[lang] = value
		}

		for _, lang := range langs {
			value, exists := record[fmt.Sprintf("lesson_%s", lang)].(string)
			if !exists || value == "" {
				log.Printf("Warning: Missing Lesson for language '%s' in record %d\n", lang, i+1)
				continue
			}
			verse.Lesson[lang] = value
		}
		verses = append(verses, verse)
	}
	if len(verses) == 0 {
		return errors.New("no verses to ingest")
	}
	return v.prarthanaMongoRepository.InsertManyVerses(ctx, verses)
}
