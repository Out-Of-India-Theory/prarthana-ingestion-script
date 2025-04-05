package stotra_ingestion

import (
	"context"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/entity"
)

type Service interface {
	StotraIngestion(ctx context.Context, startID, endID int) (map[string]entity.Stotra, error)
}
