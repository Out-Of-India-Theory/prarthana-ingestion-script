package sync

import (
	"context"
	"fmt"

	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	mongoRepo "github.com/Out-Of-India-Theory/prarthana-ingestion-script/repository/mongo/prarthana_data"
	"go.uber.org/zap"
)

type SyncService struct {
	logger      *zap.Logger
	stagingRepo mongoRepo.MongoRepository
	prodRepo    mongoRepo.MongoRepository
}

func InitSyncService(ctx context.Context, stagingRepo, prodRepo mongoRepo.MongoRepository) *SyncService {
	return &SyncService{
		logger:      logging.WithContext(ctx),
		stagingRepo: stagingRepo,
		prodRepo:    prodRepo,
	}
}

func (s *SyncService) SyncPrarthanas(ctx context.Context, startID, endID int) error {
	prarthanas, err := s.stagingRepo.GetPrarthanasByTmpIdRange(ctx, startID, endID)
	if err != nil {
		return fmt.Errorf("error fetching staging prarthanas: %w", err)
	}
	if len(prarthanas) == 0 {
		s.logger.Info("no prarthanas found in range", zap.Int("start", startID), zap.Int("end", endID))
		return nil
	}
	// share_link is environment-specific. The Prarthana entity in this repo
	// has no share_link field, so the bson $set built by InsertManyPrarthanas
	// will not include or overwrite the prod doc's existing share_link.
	if err := s.prodRepo.InsertManyPrarthanas(ctx, prarthanas); err != nil {
		return fmt.Errorf("error syncing prarthanas to prod: %w", err)
	}
	s.logger.Info("synced prarthanas to prod", zap.Int("count", len(prarthanas)))
	return nil
}

func (s *SyncService) SyncStotras(ctx context.Context, startID, endID int) error {
	stotras, err := s.stagingRepo.GetStotrasByIntIdRange(ctx, startID, endID)
	if err != nil {
		return fmt.Errorf("error fetching staging stotras: %w", err)
	}
	if len(stotras) == 0 {
		s.logger.Info("no stotras found in range", zap.Int("start", startID), zap.Int("end", endID))
		return nil
	}
	if err := s.prodRepo.InsertManyStotras(ctx, stotras); err != nil {
		return fmt.Errorf("error syncing stotras to prod: %w", err)
	}
	s.logger.Info("synced stotras to prod", zap.Int("count", len(stotras)))
	return nil
}

func (s *SyncService) SyncShloks(ctx context.Context, startID, endID int) error {
	shloks, err := s.stagingRepo.GetShloksByIntIdRange(ctx, startID, endID)
	if err != nil {
		return fmt.Errorf("error fetching staging shloks: %w", err)
	}
	if len(shloks) == 0 {
		s.logger.Info("no shloks found in range", zap.Int("start", startID), zap.Int("end", endID))
		return nil
	}
	if err := s.prodRepo.InsertManyShloks(ctx, shloks); err != nil {
		return fmt.Errorf("error syncing shloks to prod: %w", err)
	}
	s.logger.Info("synced shloks to prod", zap.Int("count", len(shloks)))
	return nil
}
