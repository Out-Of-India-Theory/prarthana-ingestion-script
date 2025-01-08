package stotra_ingestion

import (
	"context"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/entity"
)

type Service interface {
	StotraIngestion(ctx context.Context, csvFilePath string, startID, endID int) (map[string]entity.Stotra, error)
}
