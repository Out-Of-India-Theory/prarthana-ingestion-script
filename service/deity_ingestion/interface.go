package deity_ingestion

import (
	"context"
)

type Service interface {
	DeityIngestion(ctx context.Context, startID, endID int) (map[string]string, error)
}
