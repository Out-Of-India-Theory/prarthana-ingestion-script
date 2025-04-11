package prarthana

import "github.com/Out-Of-India-Theory/prarthana-ingestion-script/entity"

type ESRepository interface {
	InsertDeitySearchDocument(doc entity.DeitySearchData) error
	InsertPrarthanaSearchDocument(doc entity.PrarthanaSearchData) error
}
