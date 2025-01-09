package prarthana_ingestion

import "context"

type Service interface {
	PrarthanaIngestion(ctx context.Context, startID, endID int) (map[string]string, error)
}
