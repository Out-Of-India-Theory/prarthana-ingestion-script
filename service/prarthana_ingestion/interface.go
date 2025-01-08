package prarthana_ingestion

import "context"

type Service interface {
	PrarthanaIngestion(ctx context.Context, prarthanaToDeityCsvFilePath string, deityCsvFilePath string, stotraCsvFilePath string, adhyayaCsvFilePath string, variantCsvFilePath string, PrarthanaCsvFilePath string, startID, endID int) (map[string]string, error)
}
