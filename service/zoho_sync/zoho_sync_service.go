package zoho_sync

import (
	"context"
	"fmt"
	"strings"

	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/entity"
	mongoRepo "github.com/Out-Of-India-Theory/prarthana-ingestion-script/repository/mongo/prarthana_data"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/zoho"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/util"
	"go.uber.org/zap"
)

const prarthanaSheetName = "prarthanas"

// 1-indexed column positions in the prarthana sheet (see plan).
const (
	colNameDefault      = 5
	colNameHindi        = 6
	colNameKannada      = 7
	colNameMarathi      = 8
	colNameTamil        = 9
	colNameTelugu       = 10
	colNameGujarati     = 11
	colFestivalIds      = 17
	colShortDescDefault = 18
	colShortDescHindi   = 19
	colShortDescKannada = 20
	colShortDescMarathi = 21
	colShortDescTamil   = 22
	colShortDescTelugu  = 23
	colShortDescGujarati = 24
)

type ZohoSyncService struct {
	logger   *zap.Logger
	prodRepo mongoRepo.MongoRepository
	zohoSvc  zoho.Service
}

func InitZohoSyncService(ctx context.Context, prodRepo mongoRepo.MongoRepository, zohoSvc zoho.Service) *ZohoSyncService {
	return &ZohoSyncService{
		logger:   logging.WithContext(ctx),
		prodRepo: prodRepo,
		zohoSvc:  zohoSvc,
	}
}

func (s *ZohoSyncService) SyncPrarthanaToSheet(ctx context.Context, prarthanaID string) error {
	accessToken, err := s.zohoSvc.RefreshAccessToken()
	if err != nil {
		return fmt.Errorf("failed to refresh zoho access token: %w", err)
	}
	ctx = util.SetZohoAccessTokenInContext(ctx, accessToken)

	prarthana, err := s.prodRepo.FindPrarthanaById(ctx, prarthanaID)
	if err != nil {
		return fmt.Errorf("failed to fetch prod prarthana %s: %w", prarthanaID, err)
	}

	row, err := s.findSheetRowByUUID(ctx, prarthanaID)
	if err != nil {
		return err
	}

	columnIndexes := []int{
		colNameDefault, colNameHindi, colNameKannada, colNameMarathi, colNameTamil, colNameTelugu, colNameGujarati,
		colFestivalIds,
		colShortDescDefault, colShortDescHindi, colShortDescKannada, colShortDescMarathi, colShortDescTamil, colShortDescTelugu, colShortDescGujarati,
	}
	dataArray := []interface{}{
		prarthana.Title["default"], prarthana.Title["hi"], prarthana.Title["kn"], prarthana.Title["mr"], prarthana.Title["ta"], prarthana.Title["te"], prarthana.Title["gu"],
		strings.Join(prarthana.FestivalIds, ","),
		prarthana.Description["default"], prarthana.Description["hi"], prarthana.Description["kn"], prarthana.Description["mr"], prarthana.Description["ta"], prarthana.Description["te"], prarthana.Description["gu"],
	}

	if err := s.zohoSvc.SetSheetData(ctx, prarthanaSheetName, row, columnIndexes, dataArray); err != nil {
		return fmt.Errorf("failed to write prarthana %s to sheet row %d: %w", prarthanaID, row, err)
	}
	s.logger.Info("synced prarthana to sheet", zap.String("prarthana_id", prarthanaID), zap.Int("row", row))
	return nil
}

func (s *ZohoSyncService) findSheetRowByUUID(ctx context.Context, prarthanaID string) (int, error) {
	var response entity.ShlokaSheetResponse
	if err := s.zohoSvc.GetSheetData(ctx, prarthanaSheetName, &response); err != nil {
		return 0, fmt.Errorf("failed to fetch prarthana sheet: %w", err)
	}
	for _, record := range response.Records {
		uuidVal, _ := record["UUID"].(string)
		if uuidVal == prarthanaID {
			rowFloat, ok := record["row_index"].(float64)
			if !ok {
				return 0, fmt.Errorf("row_index missing for prarthana %s", prarthanaID)
			}
			return int(rowFloat), nil
		}
	}
	return 0, fmt.Errorf("prarthana %s not found in sheet by UUID", prarthanaID)
}
