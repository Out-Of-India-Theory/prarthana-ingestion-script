package deity_ingestion

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
	"os"
	"regexp"
	"strconv"
	"strings"
)

type DeityIngestionService struct {
	logger                   *zap.Logger
	prarthanaMongoRepository mongoRepo.MongoRepository
}

func InitDeityIngestionService(ctx context.Context,
	prarthanaMongoRepository mongoRepo.MongoRepository,
) *DeityIngestionService {
	return &DeityIngestionService{
		logger:                   logging.WithContext(ctx),
		prarthanaMongoRepository: prarthanaMongoRepository,
	}
}

func (s *DeityIngestionService) DeityIngestion(ctx context.Context, prarthanaToDeityCsvFilePath string, deityCsvFilePath string, stotraCsvFilePath string, adhyayaCsvFilePath string, variantCsvFilePath string, PrarthanaCsvFilePath string, startID, endID int) (map[string]string, error) {
	_, deityToPrarthanaMap := preparePrarthanaToDeityMap(prarthanaToDeityCsvFilePath)
	prarthanaIdMap, _ := s.PrarthanaIngestion(ctx, prarthanaToDeityCsvFilePath, deityCsvFilePath, stotraCsvFilePath, adhyayaCsvFilePath, variantCsvFilePath, PrarthanaCsvFilePath, startID, endID)
	file, err := os.Open(deityCsvFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	fieldMap := make(map[string]int)
	for i, field := range header {
		fieldMap[field] = i
	}

	var deities []entity.DeityDocument
	deityIdMap := make(map[string]string)

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	tmpIdToDeityIdMap, err := s.prarthanaMongoRepository.GetTmpIdToDeityIdMap(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for i, record := range records {
		log.Printf("Processing record %d\n", i+1)

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
		deityName := record[fieldMap["Deity Name"]]
		re := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
		if re.MatchString(deityName) {
			return nil, fmt.Errorf("the name '%s' contains special characters. Please remove them", deityName)
		}
		deityUuid := record[fieldMap["UUID"]]
		if strings.TrimSpace(deityUuid) == "" {
			deityUuid = uuid.NewString()
		}
		tmpId := record[fieldMap["ID"]]
		if strings.TrimSpace(tmpId) == "" {
			tmpId = record[fieldMap["ID"]]
		}
		if val, found := tmpIdToDeityIdMap[tmpId]; found {
			deityUuid = val
		}
		deity := entity.DeityDocument{
			TmpId: tmpId,
			Id:    deityUuid,
			Title: map[string]string{
				"default": deityName,
			},
			Slug:    strings.ToLower(strings.ReplaceAll(deityName, " ", "_")),
			Aliases: util.GetSplittedString(record[fieldMap["Also known as"]]),
			Description: map[string]string{
				"default": record[fieldMap["Description"]],
			},
			UIInfo: entity.DeityUIInfo{
				DefaultImage:    fmt.Sprintf("https://d161fa2zahtt3z.cloudfront.net/prarthanas/deities/list-image/%s.png", strings.ToLower(strings.ReplaceAll(deityName, " ", "_"))),
				BackgroundImage: fmt.Sprintf("https://d161fa2zahtt3z.cloudfront.net/prarthanas/deities/bg-image/%s.png", strings.ToLower(strings.ReplaceAll(deityName, " ", "_")))},
		}
		deityIdMap[record[fieldMap["ID"]]] = deity.Id
		deities = append(deities, deity)
	}
	for i, deity := range deities {
		ids := deityToPrarthanaMap[deity.TmpId]
		var prarthanaIds []string
		for _, id := range ids {
			prarthanaIds = append(prarthanaIds, prarthanaIdMap[id])
		}
		deities[i].Prarthanas = prarthanaIds
	}
	return deityIdMap, err
}

func preparePrarthanaToDeityMap(csvFilePath string) (map[string]string, map[string][]string) {
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
