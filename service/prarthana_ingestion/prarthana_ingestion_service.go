package prarthana_ingestion

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/entity"
	mongoRepo "github.com/Out-Of-India-Theory/prarthana-automated-script/repository/mongo/prarthana_data"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/service/zoho"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/util"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type PrarthanaIngestionService struct {
	logger                   *zap.Logger
	prarthanaMongoRepository mongoRepo.MongoRepository
	zohoService              zoho.Service
}

func InitPrathanaIngestionService(ctx context.Context,
	prarthanaMongoRepository mongoRepo.MongoRepository,
	zohoService zoho.Service,
) *PrarthanaIngestionService {
	return &PrarthanaIngestionService{
		logger:                   logging.WithContext(ctx),
		prarthanaMongoRepository: prarthanaMongoRepository,
		zohoService:              zohoService,
	}
}

func (s *PrarthanaIngestionService) PrarthanaIngestion(ctx context.Context, startID, endID int) (map[string]string, error) {
	stotraMap, err := s.prarthanaMongoRepository.GetAllStotras(ctx)
	if err != nil {
		return nil, err
	}
	chapterMap, err := s.prepareChapterMap(ctx, stotraMap)
	if err != nil {
		log.Fatalf("Failed to prepare chapter map: %v", err)
	}

	variantMap, err := s.prepareVariantMap(ctx, chapterMap)
	if err != nil {
		log.Fatalf("Failed to prepare chapter map: %v", err)
	}

	var response entity.ShlokaSheetResponse
	err = s.zohoService.GetSheetData(ctx, "prarthanas", &response)
	if err != nil {
		return nil, err
	}
	if len(response.Records) == 0 {
		return nil, errors.New("no records found")
	}
	prarthanaIdMap := make(map[string]string)
	prarthanas := make([]entity.Prarthana, 0)
	for i, record := range response.Records {
		fmt.Println("Processing record : ", i+1)
		idf, ok := record["ID"].(float64)
		if !ok {
			return nil, errors.New("Invalid ID")
		}
		id := int(idf)

		if id < startID || id > endID {
			continue
		}

		name, ok := record["Name (Mandatory)"].(string)
		if !ok {
			return nil, errors.New("Missing prarthana name")
		}

		re := regexp.MustCompile(`[^a-zA-Z0-9\s\-\(\)]+`)
		if re.MatchString(name) {
			return nil, fmt.Errorf("the name '%s' contains special characters. Please remove them", name)
		}
		tmpId := strconv.Itoa(id)

		extId, ok := record["UUID"].(string)
		if !ok {
			extId = uuid.NewString()
			//return nil, errors.New("Missing UUID")
		}

		albumArt, ok := record["Album Art"].(string)
		if !ok {
			return nil, errors.New("Missing prarthana album art")
		}
		audioName := strings.ToLower(util.SanitizeString(name))

		audioURL := fmt.Sprintf("https://d161fa2zahtt3z.cloudfront.net/audio/stitched_audio/%s.wav", audioName)
		audioURLMp3 := fmt.Sprintf("https://d161fa2zahtt3z.cloudfront.net/audio/stitched_audio/%s.mp3", audioName)
		if !util.UrlExists(audioURL) {
			if !util.UrlExists(audioURLMp3) {
				return nil, fmt.Errorf("audio URL does not exist: %s", audioURL)
			}
			audioURL = audioURLMp3
		}

		albumArtURL := fmt.Sprintf("https://d161fa2zahtt3z.cloudfront.net/prarthanas/album_art/%s.png", albumArt)
		if !util.UrlExists(albumArtURL) {
			return nil, fmt.Errorf("album art URL does not exist: %s", albumArtURL)
		}

		studioRecorded := false
		studioRecordedStr, ok := record["Studio Recorded(yes/no)"].(string)
		if ok && studioRecordedStr == "yes" {
			studioRecorded = true
		}
		festivalIdsStr, ok := record["Festival Ids"].(string)
		festivalIds := []string{}
		if ok && len(festivalIdsStr) != 0 {
			festivalIds = util.GetSplittedString(festivalIdsStr)
		}

		shortDescription, ok := record["Short Description"].(string)
		if !ok {
			return nil, fmt.Errorf("Missing prarthana short description : %d", id)
		}
		//variantIds, ok := record["Prarthana Variant ID (Comma separated - Ordered)"].(string)
		variantIds := fmt.Sprintf("%v", record["Prarthana Variant ID (Comma separated - Ordered)"])
		prarthana := entity.Prarthana{
			TmpId: tmpId,
			Id:    extId,
			Title: map[string]string{
				"default": name,
			},
			FestivalIds: festivalIds,
			Days:        util.GetDaysFromTitle(name),
			AudioInfo: entity.AudioInfo{AudioUrl: audioURL,
				IsAudioAvailable: true,
				IsStudioRecorded: studioRecorded},
			Variants:      []entity.Variant{variantMap[variantIds]},
			Description:   map[string]string{"default": shortDescription},
			Importance:    map[string]string{},
			Instruction:   map[string]string{},
			ItemsRequired: map[string][]string{},
		}
		templateNumberS, ok := record["Template Number"].(string)
		if !ok {
			return nil, errors.New("Missing prarthana template number")
		}
		templateNumber, err := strconv.Atoi(templateNumberS)
		if err != nil {
			return nil, err
		}
		prarthana.UiInfo = entity.PrarthanaUIInfo{
			AlbumArt:        fmt.Sprintf("https://d161fa2zahtt3z.cloudfront.net/prarthanas/album_art/%s.png", albumArt),
			DefaultImageUrl: fmt.Sprintf("https://d161fa2zahtt3z.cloudfront.net/prarthanas/album_art/%s.png", albumArt),
			TemplateNumber:  fmt.Sprintf("template_%s", templateNumber),
		}

		prarthana.AvailableLanguages = []entity.KeyValue{
			{"default", "Default (Sanskrit)"},
			{"hindi", "हिंदी"},
			{"kannada", "ಕನ್ನಡ"},
			{"english", "English"},
			{"telugu", "తెలుగు"},
			{"bengali", "বাংলা"},
			{"marathi", "मराठी"},
			{"tamil", "தமிழ்"},
			{"gujarati", "ગુજરાતી"},
			{"odiya", "ଓଡିଆ"},
			{"malayalam", "മലയാളം"},
			{"assamese", "অসমীয়া"},
			{"punjabi", "ਪੰਜਾਬੀ"},
		}
		prarthanas = append(prarthanas, prarthana)
		prarthanaIdMap[tmpId] = prarthana.Id
	}
	return prarthanaIdMap, s.prarthanaMongoRepository.InsertManyPrarthanas(ctx, prarthanas)
}

func (s *PrarthanaIngestionService) prepareChapterMap(ctx context.Context, stotraMap map[string]entity.Stotra) (map[string]entity.Chapter, error) {
	var response entity.ShlokaSheetResponse
	err := s.zohoService.GetSheetData(ctx, "adhyaya", &response)
	if err != nil {
		return nil, err
	}
	if len(response.Records) == 0 {
		return nil, errors.New("no records found")
	}
	chapterMap := make(map[string]entity.Chapter)
	for _, record := range response.Records {
		stotraIds := util.GetSplittedString(fmt.Sprintf("%v", record["Stotra ID (Comma separated - Ordered)"]))
		if len(stotraIds) == 0 {
			return nil, errors.New("no stotra ID")
		}
		duration := 0
		for _, id := range stotraIds {
			if sto, ok := stotraMap[id]; ok {
				duration += sto.DurationInSeconds
			}
		}
		minutes := int(math.Max(1, math.Round((float64(duration) / float64(60)))))
		durationStr := fmt.Sprintf("%dm", minutes)
		name, ok := record["Name (Mandatory)"].(string)
		if !ok {
			return nil, errors.New("no name found")
		}
		id, ok := record["ID"].(float64)
		if !ok {
			return nil, errors.New("no ID found")
		}
		chapter := entity.Chapter{
			Order:     1,
			Timestamp: "1m",
			Duration:  durationStr,
			Title: map[string]string{
				"default": name,
			},
			DurationInSec: duration,
			StotraIds:     stotraIds,
		}
		chapterMap[fmt.Sprintf("%f", id)] = chapter
	}
	return chapterMap, nil
}

func (s *PrarthanaIngestionService) prepareVariantMap(ctx context.Context, chapterMap map[string]entity.Chapter) (map[string]entity.Variant, error) {
	var response entity.ShlokaSheetResponse
	err := s.zohoService.GetSheetData(ctx, "prarthana variant", &response)
	if err != nil {
		return nil, err
	}
	if len(response.Records) == 0 {
		return nil, errors.New("no records found")
	}
	variantMap := make(map[string]entity.Variant)
	for _, record := range response.Records {
		duration := 0
		//chapterIds := util.GetSplittedString(record[fieldMap["Adhyaya ID (Comma separated - Ordered)"]])
		chapterIds := util.GetSplittedString(fmt.Sprintf("%v", record["Adhyaya ID (Comma separated - Ordered)"]))
		if len(chapterIds) == 0 {
			return nil, errors.New("no stotra ID")
		}
		chapters := make([]entity.Chapter, 0)
		for _, id := range chapterIds {
			chapter := chapterMap[id]
			duration += chapter.DurationInSec
			chapters = append(chapters, chapter)
		}
		minutes := int(math.Max(1, math.Round((float64(duration) / float64(60)))))
		durationStr := fmt.Sprintf("%dm", minutes)
		variant := entity.Variant{
			Duration:  durationStr,
			Chapters:  chapters,
			IsDefault: true,
		}
		id, ok := record["ID"].(float64)
		if !ok {
			return nil, errors.New("no ID found")
		}
		variantMap[fmt.Sprintf("%f", id)] = variant
	}
	return variantMap, nil
}

func PreparePrarthanaToDeityMap(csvFilePath string) (map[string]string, map[string][]string) {
	file, err := os.Open(csvFilePath)
	if err != nil {
		return nil, nil
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		fmt.Println("Error:", err)
		return nil, nil
	}
	fieldMap := make(map[string]int)
	for i, field := range header {
		fieldMap[field] = i
	}
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error:", err)
		return nil, nil
	}
	pdmap := make(map[string]string)
	dpMap := make(map[string][]string)
	for _, record := range records {
		pdmap[record[fieldMap["Prarthana ID"]]] = record[fieldMap["Diety ID"]]
		dpMap[record[fieldMap["Diety ID"]]] = append(dpMap[record[fieldMap["Diety ID"]]], record[fieldMap["Prarthana ID"]])
	}
	return pdmap, dpMap
}
