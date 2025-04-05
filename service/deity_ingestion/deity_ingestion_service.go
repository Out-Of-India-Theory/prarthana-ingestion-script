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
		log.Fatalf("Error generating Prarthana TmpId to Deity ID map: %v", err)
	}
	prarthanaIdMap, err := s.prarthanaMongoRepository.GeneratePrarthanaTmpIdToIdMap(ctx)
	if err != nil {
		log.Fatalf("Error generating Prarthana TmpId to Prarthana ID map: %v", err)
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
		deityNameAssamese := record["Title (Assamese)"].(string)
		deityNamePunjabi := record["Title (Punjabi)"].(string)
		deityNameMalayalam := record["Title (Malayalam)"].(string)
		deityNameOdia := record["Title (Odia)"].(string)
		deityNameBengali := record["Title (Bengali)"].(string)

		deityUuid := record["UUID"].(string)
		if strings.TrimSpace(deityUuid) == "" {
			deityUuid = uuid.NewString()
		}
		tmpId := fmt.Sprintf("%d", id)
		if val, found := tmpIdToDeityIdMap[tmpId]; found {
			deityUuid = val
		}
		deityImageName, ok := record["Deity Image"].(string)
		if !ok {
			return nil, errors.New("Invalid Deity Image")
		}
		defaultImage := fmt.Sprintf("https://d161fa2zahtt3z.cloudfront.net/prarthanas/deities/list-image/%s.png", deityImageName)
		if !util.UrlExists(defaultImage) {
			return nil, fmt.Errorf("deity image does not exist: %s", defaultImage)
		}
		backgroundImage := fmt.Sprintf("https://d161fa2zahtt3z.cloudfront.net/prarthanas/deities/bg-image/%s.png", deityImageName)
		if !util.UrlExists(backgroundImage) {
			return nil, fmt.Errorf("deity background image does not exist: %s", backgroundImage)
		}
		formattedtitle := strings.ToLower(strings.ReplaceAll(deityNameDefault, " ", "_"))
		var heroImageAlbum []entity.HeroImageAlbum
		heroImageCount, ok := record["Hero Image Count"].(float64)
		if ok && heroImageCount > 0 {
			for i := 0; i < int(heroImageCount); i++ { // Convert float64 to int directly
				imageIndex := ""
				if i > 0 {
					imageIndex = strconv.Itoa(i)
				}
				heroImageAlbum = append(heroImageAlbum, entity.HeroImageAlbum{
					FullImage:      fmt.Sprintf("https://d161fa2zahtt3z.cloudfront.net/prarthanas/deities/hero_image_album/full_image/%s%s.png", formattedtitle, imageIndex),
					ThumbnailImage: fmt.Sprintf("https://d161fa2zahtt3z.cloudfront.net/prarthanas/deities/hero_image_album/full_image/%s%s.png", formattedtitle, imageIndex),
					ShareImage:     fmt.Sprintf("https://d161fa2zahtt3z.cloudfront.net/prarthanas/deities/hero_image_album/share_image/%s%s.png", formattedtitle, imageIndex),
				})
			}
		}

		var deityOfTheDay string
		if dodFlag, ok := record["DOD Flag"].(bool); ok && dodFlag {
			deityOfTheDay = fmt.Sprintf("https://d161fa2zahtt3z.cloudfront.net/prarthanas/deities/hero_image_album/dod_image/%s.png", formattedtitle)
		}

		aliases, ok := record["Also known as"].(string)
		if !ok {
			aliases = ""
		}
		festivalIdsStr, ok := record["Festival Ids"].(string)
		festivalIds := []string{}
		if ok && len(festivalIdsStr) != 0 {
			festivalIds = util.GetSplittedString(festivalIdsStr)
		}
		regionsStr, ok := record["Region"].(string)
		regions := []string{}
		if ok && len(regionsStr) != 0 {
			regions = util.GetSplittedString(regionsStr)
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
		aliasesV1Default, ok := record["Aliases_v1 (Default)"].(string)
		aliasesV1Hindi, ok := record["Aliases_v1 (Hindi)"].(string)
		aliasesV1Kannada, ok := record["Aliases_v1 (Kannada)"].(string)
		aliasesV1Marathi, ok := record["Aliases_v1 (Marathi)"].(string)
		aliasesV1Tamil, ok := record["Aliases_v1 (Tamil)"].(string)
		aliasesV1Telugu, ok := record["Aliases_v1 (Telugu)"].(string)
		aliasesV1Gujarati, ok := record["Aliases_v1 (Gujarati)"].(string)
		aliasesV1 := make(map[string][]string)

		if aliasesV1Default != "" {
			aliasesV1["default"] = util.GetSplittedString(aliasesV1Default)
		}
		if aliasesV1Hindi != "" {
			aliasesV1["hi"] = util.GetSplittedString(aliasesV1Hindi)
		}
		if aliasesV1Kannada != "" {
			aliasesV1["kn"] = util.GetSplittedString(aliasesV1Kannada)
		}
		if aliasesV1Marathi != "" {
			aliasesV1["mr"] = util.GetSplittedString(aliasesV1Marathi)
		}
		if aliasesV1Tamil != "" {
			aliasesV1["ta"] = util.GetSplittedString(aliasesV1Tamil)
		}
		if aliasesV1Telugu != "" {
			aliasesV1["te"] = util.GetSplittedString(aliasesV1Telugu)
		}
		if aliasesV1Gujarati != "" {
			aliasesV1["gu"] = util.GetSplittedString(aliasesV1Gujarati)
		}
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
				"as":      deityNameAssamese,
				"pa":      deityNamePunjabi,
				"bn":      deityNameBengali,
				"od":      deityNameOdia,
				"ml":      deityNameMalayalam,
			},
			Region:    regions,
			Slug:      strings.ToLower(strings.ReplaceAll(deityNameDefault, " ", "_")),
			Aliases:   util.GetSplittedString(aliases),
			AliasesV1: aliasesV1,
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
				HeroImageAlbum:  heroImageAlbum,
				DeityOfTheDay:   deityOfTheDay,
			},
			FestivalIds: festivalIds,
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
