package shlok_translation

import (
	"context"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/zoho"
	"go.uber.org/zap"
)

type ShlokTranslationService struct {
	logger      *zap.Logger
	zohoService zoho.Service
}

func InitShlokTranslationService(ctx context.Context,
	zohoService zoho.Service,
) *ShlokTranslationService {
	return &ShlokTranslationService{
		logger:      logging.WithContext(ctx),
		zohoService: zohoService,
	}
}

func (s *ShlokTranslationService) GenerateShlokaTranslation(ctx context.Context, startId, endId int) error {
	var err error
	return err
}
