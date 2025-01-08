package deity_ingestion

import (
	"context"
)

type Service interface {
	DeityIngestion(ctx context.Context, prarthanaToDeityCsvFilePath string, deityCsvFilePath string, startID, endID int) (map[string]string, error)
}
