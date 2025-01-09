package zoho

import (
	"context"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/entity"
)

type Service interface {
	GetAuthorizationURL(state string) string
	ExchangeCodeForTokens(ctx context.Context, code string) error
	RefreshAccessToken() (string, error)
	GetSheetData(ctx context.Context, sheetName string) (*entity.SheetResponse, error)
	IsTokenExpired() bool
}
