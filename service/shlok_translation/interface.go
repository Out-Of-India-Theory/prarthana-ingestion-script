package shlok_translation

import "context"

type Service interface {
	GenerateShlokaTranslation(ctx context.Context, startId, endId int) error
}
