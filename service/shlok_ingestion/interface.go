package shlok_ingestion

import (
	"context"
)

type Service interface {
	ShlokIngestion(ctx context.Context, startID, endID int) error
}
