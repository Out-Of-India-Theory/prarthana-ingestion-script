package deity_ingestion

import (
	"context"
)

type Service interface {
	DeityIngestion(ctx context.Context, prarthanaToDeityCsvFilePath string, deityCsvFilePath string, stotraCsvFilePath string, adhyayaCsvFilePath string, variantCsvFilePath string, PrarthanaCsvFilePath string, startID, endID int) (map[string]string, error)
}
