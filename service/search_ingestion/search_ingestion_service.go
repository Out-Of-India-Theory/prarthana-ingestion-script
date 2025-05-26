package search_ingestion

import (
	"context"
	"fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/entity"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/repository/es/prarthana"
	mongoRepo "github.com/Out-Of-India-Theory/prarthana-ingestion-script/repository/mongo/prarthana_data"
	"go.uber.org/zap"
	"time"
)

const DuratiomInMin = "%s.0f min"

type SearchIngestionService struct {
	logger                   *zap.Logger
	prarthanaMongoRepository mongoRepo.MongoRepository
	prarthanaESRepository    prarthana.ESRepository
}

func InitSearchIngestionService(ctx context.Context,
	prarthanaMongoRepository mongoRepo.MongoRepository,
	prarthanaESRepository prarthana.ESRepository,
) *SearchIngestionService {
	return &SearchIngestionService{
		logger:                   logging.WithContext(ctx),
		prarthanaMongoRepository: prarthanaMongoRepository,
		prarthanaESRepository:    prarthanaESRepository,
	}
}

func (s *SearchIngestionService) InsertDeitySearchData(ctx context.Context) error {
	languages := []string{"default", "hi", "mr", "ta", "te", "kn", "gu"}
	deityDocs := s.prarthanaMongoRepository.PullDeityDocs(ctx)

	for _, doc := range deityDocs {
		for _, lang := range languages {
			title, exists := doc.Title[lang]
			if !exists {
				continue
			}
			aliases := doc.AliasesV1[lang]
			data := entity.DeitySearchData{
				ID:       doc.Id,
				Title:    title,
				Aliases:  aliases,
				ImageURL: doc.UIInfo.DefaultImage,
			}
			if err := s.prarthanaESRepository.InsertDeitySearchDocument(data); err != nil {
				return fmt.Errorf("failed to index deity document for ID '%s', lang '%s': %w", doc.Id, lang, err)
			}
		}
	}
	return nil
}

func (s *SearchIngestionService) InsertPrarthanaSearchData(ctx context.Context) error {
	languages := []string{"default", "hi", "mr", "ta", "te", "kn", "gu"}
	langMap := map[string]string{
		"default": "english",
		"hi":      "hindi",
		"kn":      "kannada",
		"mr":      "marathi",
		"ta":      "tamil",
		"te":      "telugu",
		"gu":      "gujarati",
	}
	prarthanaDocs := s.prarthanaMongoRepository.PullPrarthanaDocs(ctx)

	//var outputs []entity.PrarthanaSearchData
	for _, doc := range prarthanaDocs {
		for _, language := range languages {
			duration := doc.Variants[0].Duration
			pDuration, _ := time.ParseDuration(duration)
			output := entity.PrarthanaSearchData{
				ID:       doc.ID,
				Title:    doc.Title[language],
				Duration: fmt.Sprintf(DuratiomInMin, pDuration.Minutes()),
			}

			for _, deityDoc := range doc.Deity {
				output.DeityNames = append(output.DeityNames, deityDoc.Title[language])
				// Append non-empty aliases to DeityNames
				if language == "default" {
					for _, alias := range deityDoc.Aliases {
						if alias != "" {
							output.DeityNames = append(output.DeityNames, alias)
						}
					}
				}
			}
			output.ImageURL = doc.UIDetails.DefaultImageUrl
			output.IsAudioAvailable = doc.AudioInfo.IsAudioAvailable
			for _, shlokDoc := range doc.ShlokDocs {
				output.Shloks = append(output.Shloks, shlokDoc.Shlok[langMap[language]])
			}
			if err := s.prarthanaESRepository.InsertPrarthanaSearchDocument(output); err != nil {
				return fmt.Errorf("failed to index prarthana document for ID '%s', lang '%s': %w", doc.ID, language, err)
			}
		}
	}

	return nil
}
