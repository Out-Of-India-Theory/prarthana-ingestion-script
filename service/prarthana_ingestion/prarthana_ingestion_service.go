package prarthana_ingestion

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/entity"
	mongoRepo "github.com/Out-Of-India-Theory/prarthana-automated-script/repository/mongo/prarthana_data"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/service/util"
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
}

func InitPrathanaIngestionService(ctx context.Context,
	prarthanaMongoRepository mongoRepo.MongoRepository,
) *PrarthanaIngestionService {
	return &PrarthanaIngestionService{
		logger:                   logging.WithContext(ctx),
		prarthanaMongoRepository: prarthanaMongoRepository,
	}
}

func (s *PrarthanaIngestionService) PrarthanaIngestion(ctx context.Context, prarthanaToDeityCsvFilePath string, deityCsvFilePath string, stotraCsvFilePath string, adhyayaCsvFilePath string, variantCsvFilePath string, PrarthanaCsvFilePath string, startID, endID int) (map[string]string, error) {
	deityIdMap, err := s.DeityIngestion(ctx, startID, endID)
	if err != nil {
		return nil, err
	}
	prarthanaToDeityMap, _ := PreparePrarthanaToDeityMap(prarthanaToDeityCsvFilePath)
	stotraMap, err := s.prarthanaMongoRepository.GetAllStotras(ctx)
	chapterMap, err := prepareChapterMap(adhyayaCsvFilePath, stotraMap)
	if err != nil {
		log.Fatalf("Failed to prepare chapter map: %v", err)
	}

	variantMap, err := prepareVariantMap(variantCsvFilePath, chapterMap)
	if err != nil {
		log.Fatalf("Failed to prepare chapter map: %v", err)
	}

	file, err := os.Open(PrarthanaCsvFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading header: %w", err)
	}

	fieldMap := make(map[string]int)
	for i, field := range header {
		fieldMap[field] = i
	}
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error:", err)
		return nil, fmt.Errorf("Error: %w", err)
	}
	idTemplateMap, tmpIdToIdMap, err := s.prarthanaMongoRepository.GetTmpIdToPrarthanaIds(ctx)
	if err != nil {
		log.Fatal(err)
	}
	oldReversedPrarthanaIds := tmpIdToIdMap
	prarthanaIdMap := make(map[string]string)
	prarthanas := make([]entity.Prarthana, 0)
	for i, record := range records {
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
		deityIds := make([]string, 0)
		name := record[fieldMap["Name (Mandatory)"]]
		re := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
		if re.MatchString(name) {
			return nil, fmt.Errorf("the name '%s' contains special characters. Please remove them", name)
		}
		tmpId := record[fieldMap["ID"]]
		if val, found := deityIdMap[prarthanaToDeityMap[tmpId]]; found {
			deityIds = []string{val}
		}
		prarthanaUuid := uuid.NewString()
		if val, found := oldReversedPrarthanaIds[tmpId]; found {
			prarthanaUuid = val
		}
		templateNumber := "template_1"
		if val, found := idTemplateMap[tmpId]; found {
			templateNumber = val
		}
		fmt.Sprintf("%s", templateNumber)
		audioURL := fmt.Sprintf("https://d161fa2zahtt3z.cloudfront.net/audio/stitched_audio/%s.wav", strings.ToLower(strings.ReplaceAll(name, " ", "_")))
		albumArtURL := fmt.Sprintf("https://d161fa2zahtt3z.cloudfront.net/prarthanas/album_art/%s.png", record[fieldMap["Album Art"]])
		if !util.UrlExists(audioURL) {
			return nil, fmt.Errorf("audio URL does not exist: %s", audioURL)
		}
		if !util.UrlExists(albumArtURL) {
			return nil, fmt.Errorf("album art URL does not exist: %s", albumArtURL)
		}
		prarthana := entity.Prarthana{
			TmpId: tmpId,
			Id:    prarthanaUuid,
			Title: map[string]string{
				"default": name,
			},
			Days: util.GetDaysFromTitle(name),
			AudioInfo: entity.AudioInfo{AudioUrl: audioURL,
				IsAudioAvailable: true},
			Variants:      []entity.Variant{variantMap[record[fieldMap["Prarthana Variant ID (Comma separated - Ordered)"]]]},
			Description:   map[string]string{"default": record[fieldMap["Short description"]]},
			Importance:    map[string]string{},
			Instruction:   map[string]string{},
			ItemsRequired: map[string][]string{},
		}
		prarthana.UiInfo = entity.PrarthanaUIInfo{
			AlbumArt:        fmt.Sprintf("https://d161fa2zahtt3z.cloudfront.net/prarthanas/album_art/%s.png", record[fieldMap["Album Art"]]),
			DefaultImageUrl: fmt.Sprintf("https://d161fa2zahtt3z.cloudfront.net/prarthanas/album_art/%s.png", record[fieldMap["Album Art"]]),
			TemplateNumber:  fmt.Sprintf("template_%s", record[fieldMap["Template Number"]]),
		}

		prarthana.DeityIds = deityIds
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
		prarthanaIdMap[record[fieldMap["ID"]]] = prarthana.Id
	}
	return prarthanaIdMap, s.prarthanaMongoRepository.InsertManyPrarthanas(ctx, prarthanas)
}

func prepareChapterMap(AdhyayaCsvFilePath string, stotraMap map[string]entity.Stotra) (map[string]entity.Chapter, error) {
	file, err := os.Open(AdhyayaCsvFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		fmt.Println("Error:", err)
		return nil, fmt.Errorf("error: %w", err)
	}
	fieldMap := make(map[string]int)
	for i, field := range header {
		fieldMap[field] = i
	}
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error:", err)
		return nil, fmt.Errorf("error reading records: %w", err)
	}
	chapterMap := make(map[string]entity.Chapter)
	for _, record := range records {
		stotraIds := util.GetSplittedString(record[fieldMap["Stotra ID (Comma separated - Ordered)"]])
		duration := 0
		for _, id := range stotraIds {
			if sto, ok := stotraMap[id]; ok {
				duration += sto.DurationInSeconds
			}
		}
		minutes := int(math.Max(1, math.Round((float64(duration) / float64(60)))))
		durationStr := fmt.Sprintf("%dm", minutes)
		chapter := entity.Chapter{
			Order:     1,
			Timestamp: "1m",
			Duration:  durationStr,
			Title: map[string]string{
				"default": record[fieldMap["Name (Mandatory)"]],
			},
			DurationInSec: duration,
			StotraIds:     stotraIds,
		}
		chapterMap[record[fieldMap["ID"]]] = chapter
	}
	return chapterMap, nil
}

func prepareVariantMap(variantCsvFilePath string, chapterMap map[string]entity.Chapter) (map[string]entity.Variant, error) {
	file, err := os.Open(variantCsvFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		fmt.Println("Error:", err)
		return nil, fmt.Errorf("error reading records: %w", err)
	}
	fieldMap := make(map[string]int)
	for i, field := range header {
		fieldMap[field] = i
	}
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error:", err)
		return nil, fmt.Errorf("error: %w", err)
	}
	variantMap := make(map[string]entity.Variant)
	for _, record := range records {
		duration := 0
		chapterIds := util.GetSplittedString(record[fieldMap["Adhyaya ID (Comma separated - Ordered)"]])
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
		variantMap[record[fieldMap["ID"]]] = variant
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
