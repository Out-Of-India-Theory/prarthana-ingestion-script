package prarthana_data

import (
	"context"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/entity"
)

type MongoRepository interface {
	InsertManyShloks(ctx context.Context, shloks []entity.Shlok) error
	InsertManyStotras(ctx context.Context, stotras []entity.Stotra) error
	InsertManyDeities(ctx context.Context, deities []entity.DeityDocument) error
	InsertManyPrarthanas(ctx context.Context, prarthanas []entity.Prarthana) error
	GetTmpIdToPrarthanaIds(ctx context.Context) (map[string]string, map[string]string, error)
	GetTmpIdToDeityIdMap(ctx context.Context) (map[string]string, error)
	GetAllStotras(ctx context.Context) (map[string]entity.Stotra, error)
	GetAllDeities(ctx context.Context) ([]entity.DeityDocument, error)
}
