package search_ingestion

import "context"

type Service interface {
	InsertDeitySearchData(ctx context.Context) error
	InsertPrarthanaSearchData(ctx context.Context) error
}
