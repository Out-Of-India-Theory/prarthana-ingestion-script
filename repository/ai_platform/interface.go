package ai_platform

import (
	"context"
)

type ClientRepository interface {
	BatchTranslateText(ctx context.Context, text []string, lang string, isTransliterate bool) ([]TranslateText, error)
}
