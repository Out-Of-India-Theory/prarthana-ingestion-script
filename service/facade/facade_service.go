package facade

import (
	"context"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/configuration"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/service/deity_ingestion"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/service/prarthana_ingestion"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/service/shlok_ingestion"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/service/stotra_ingestion"
	"go.uber.org/zap"
)

type FacadeService struct {
	logger                    *zap.Logger
	configuration             *configuration.Configuration
	shlokIngestionService     shlok_ingestion.Service
	stotraIngestionService    stotra_ingestion.Service
	prarthanaIngestionService prarthana_ingestion.Service
	deityIngestionService     deity_ingestion.Service
}

func InitFacadeService(
	ctx context.Context,
	configuration *configuration.Configuration,
	shlokIngestionService shlok_ingestion.Service,
	stotraIngestionService stotra_ingestion.Service,
	prarthanaIngestionService prarthana_ingestion.Service,
	deityIngestionService deity_ingestion.Service,
) *FacadeService {
	return &FacadeService{
		logger:                    logging.WithContext(ctx),
		configuration:             configuration,
		shlokIngestionService:     shlokIngestionService,
		stotraIngestionService:    stotraIngestionService,
		prarthanaIngestionService: prarthanaIngestionService,
		deityIngestionService:     deityIngestionService,
	}
}

func (s *FacadeService) ShlokIngestionService() shlok_ingestion.Service {
	return s.shlokIngestionService
}

func (s *FacadeService) StotraIngestionService() stotra_ingestion.Service {
	return s.stotraIngestionService
}

func (s *FacadeService) PrarthanaIngestionService() prarthana_ingestion.Service {
	return s.prarthanaIngestionService
}

func (s *FacadeService) DeityIngestionService() deity_ingestion.Service {
	return s.deityIngestionService
}
