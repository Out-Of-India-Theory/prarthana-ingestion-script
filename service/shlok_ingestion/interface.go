package shlok_ingestion

import (
	"context"
)

type Service interface {
	ShlokIngestion(ctx context.Context, csvFilePath string, startID, endID int) error
}
