package verse_ingestion

import (
	"context"
)

type Service interface {
	VerseIngestion(ctx context.Context, startID, endID int) error
}
