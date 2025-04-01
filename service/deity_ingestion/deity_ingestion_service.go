package deity_ingestion

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/entity"
	mongoRepo "github.com/Out-Of-India-Theory/prarthana-ingestion-script/repository/mongo/prarthana_data"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/zoho"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/util"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DeityIngestionService struct {
	logger                   *zap.Logger
	prarthanaMongoRepository mongoRepo.MongoRepository
	zohoService              zoho.Service
}

func InitDeityIngestionService(ctx context.Context,
	prarthanaMongoRepository mongoRepo.MongoRepository,
	zohoService zoho.Service,
) *DeityIngestionService {
	return &DeityIngestionService{
		logger:                   logging.WithContext(ctx),
		prarthanaMongoRepository: prarthanaMongoRepository,
		zohoService:              zohoService,
	}
}

func (s *DeityIngestionService) DeityIngestion(ctx context.Context, startID, endID int) (map[string]string, error) {
	var err error
	_, deityToPrarthanaMap, err := s.preparePrarthanaToDeityMap(ctx)
	if err != nil {
		log.Fatalf("Error generating TmpId to ID map: %v", err)
	}
	prarthanaIdMap, err := s.prarthanaMongoRepository.GeneratePrarthanaTmpIdToIdMap(ctx)
	if err != nil {
		log.Fatalf("Error generating TmpId to ID map: %v", err)
	}
	var response entity.ShlokaSheetResponse
	err = s.zohoService.GetSheetData(ctx, "deities", &response)
	if err != nil {
		return nil, err
	}
	if len(response.Records) == 0 {
		return nil, errors.New("no records found")
	}

	var deities []entity.DeityDocument
	deityIdMap := make(map[string]string)

	tmpIdToDeityIdMap, err := s.prarthanaMongoRepository.GetTmpIdToDeityIdMap(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for i, record := range response.Records {
		log.Printf("Processing record %d\n", i+1)

		idf, ok := record["ID"].(float64)
		if !ok {
			return nil, errors.New("Invalid ID")
		}
		id := int(idf)

		if id < startID || id > endID {
			continue
		}
		deityNameDefault := record["Title (Default)"].(string)
		re := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
		if re.MatchString(deityNameDefault) {
			return nil, fmt.Errorf("the name '%s' contains special characters. Please remove them", deityNameDefault)
		}
		deityNameHindi := record["Title (Hindi)"].(string)
		deityNameKannada := record["Title (Kannada)"].(string)
		deityNameMarathi := record["Title (Marathi)"].(string)
		deityNameTamil := record["Title (Tamil)"].(string)
		deityNameTelugu := record["Title (Telugu)"].(string)
		deityNameGujarati := record["Title (Gujarati)"].(string)

		deityUuid := record["UUID"].(string)
		if strings.TrimSpace(deityUuid) == "" {
			deityUuid = uuid.NewString()
		}
		tmpId := fmt.Sprintf("%d", id)
		if val, found := tmpIdToDeityIdMap[tmpId]; found {
			deityUuid = val
		}
		deityImageNameStr, ok := record["Deity Image"].(string)
		if !ok {
			return nil, errors.New("Invalid Deity Image")
		}
		deityImageName := strings.ToLower(util.SanitizeString(deityImageNameStr))
		defaultImage := fmt.Sprintf("https://d161fa2zahtt3z.cloudfront.net/prarthanas/deities/list-image/%s.png", deityImageName)
		if !util.UrlExists(defaultImage) {
			return nil, fmt.Errorf("deity image does not exist: %s", defaultImage)
		}
		backgroundImage := fmt.Sprintf("https://d161fa2zahtt3z.cloudfront.net/prarthanas/deities/bg-image/%s.png", deityImageName)
		if !util.UrlExists(backgroundImage) {
			return nil, fmt.Errorf("deity background image does not exist: %s", backgroundImage)
		}
		aliases, ok := record["Also known as"].(string)
		if !ok {
			aliases = ""
		}
		descriptionDefault, ok := record["Description (Default)"].(string)
		if !ok {
			return nil, fmt.Errorf("description unavailable for row : %d", id)
		}
		descriptionHindi, ok := record["Description (Hindi)"].(string)
		descriptionKannada, ok := record["Description (Kannada)"].(string)
		descriptionMarathi, ok := record["Description (Marathi)"].(string)
		descriptionTamil, ok := record["Description (Tamil)"].(string)
		descriptionTelugu, ok := record["Description (Telugu)"].(string)
		descriptionGujarati, ok := record["Description (Gujarati)"].(string)
		deity := entity.DeityDocument{
			TmpId: tmpId,
			Id:    deityUuid,
			Title: map[string]string{
				"default": deityNameDefault,
				"hi":      deityNameHindi,
				"kn":      deityNameKannada,
				"mr":      deityNameMarathi,
				"ta":      deityNameTamil,
				"te":      deityNameTelugu,
				"gu":      deityNameGujarati,
			},
			Slug:    strings.ToLower(strings.ReplaceAll(deityNameDefault, " ", "_")),
			Aliases: util.GetSplittedString(aliases),
			Description: map[string]string{
				"default": descriptionDefault,
				"hi":      descriptionHindi,
				"kn":      descriptionKannada,
				"mr":      descriptionMarathi,
				"ta":      descriptionTamil,
				"te":      descriptionTelugu,
				"gu":      descriptionGujarati,
			},
			UIInfo: entity.DeityUIInfo{
				DefaultImage:    defaultImage,
				BackgroundImage: backgroundImage,
			},
		}
		deityIdMap[tmpId] = deity.Id
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
	return deityIdMap, s.prarthanaMongoRepository.InsertManyDeities(ctx, deities)
}

func (s *DeityIngestionService) preparePrarthanaToDeityMap(ctx context.Context) (map[string]string, map[string][]string, error) {
	var response entity.ShlokaSheetResponse
	err := s.zohoService.GetSheetData(ctx, "deity to prarthana mapping", &response)
	if err != nil {
		return nil, nil, err
	}
	if len(response.Records) == 0 {
		return nil, nil, errors.New("no records found")
	}
	pdmap := make(map[string]string)
	dpMap := make(map[string][]string)
	for _, record := range response.Records {
		prarthanaIdf, ok := record["Prarthana ID"].(float64)
		if !ok {
			return nil, nil, errors.New("prarthana ID is not a float")
		}
		deityIdString := fmt.Sprintf("%v", record["Diety ID"])

		//deityIdf, ok := record["Diety ID"].(float64)
		if len(deityIdString) == 0 {
			return nil, nil, errors.New("diety ID is not a float")
		}
		deityIds := util.GetSplittedString(deityIdString)
		prarthanaId := strconv.FormatFloat(prarthanaIdf, 'f', -1, 64)
		//deityId := fmt.Sprintf("%f", deityIdf)
		for _, id := range deityIds {
			pdmap[prarthanaId] = id
			dpMap[id] = append(dpMap[id], prarthanaId)
		}

	}
	return pdmap, dpMap, nil
}
