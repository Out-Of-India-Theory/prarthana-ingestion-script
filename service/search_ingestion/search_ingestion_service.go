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
	prarthanaDocs := s.prarthanaMongoRepository.PullPrarthanaDocs(ctx)

	for _, doc := range prarthanaDocs {
		if len(doc.Variants) == 0 {
			continue
		}

		durationStr := doc.Variants[0].Duration
		pDuration, err := time.ParseDuration(durationStr)
		if err != nil {
			return fmt.Errorf("invalid duration format for Prarthana ID '%s': %w", doc.Id, err)
		}

		for _, lang := range languages {
			title, exists := doc.Title[lang]
			if !exists {
				continue
			}

			output := entity.PrarthanaSearchData{
				ID:               doc.Id,
				Title:            title,
				Duration:         fmt.Sprintf("%.0f min", pDuration.Minutes()),
				ImageURL:         doc.UiInfo.DefaultImageUrl,
				IsAudioAvailable: doc.AudioInfo.IsAudioAvailable,
			}

			if err := s.prarthanaESRepository.InsertPrarthanaSearchDocument(output); err != nil {
				return fmt.Errorf("failed to index prarthana document for ID '%s', lang '%s': %w", doc.Id, lang, err)
			}
		}
	}

	return nil
}
