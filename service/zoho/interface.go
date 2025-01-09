package zoho

import "context"

type Service interface {
	GetAuthorizationURL(state string) string
	ExchangeCodeForTokens(ctx context.Context, code string) error
	RefreshAccessToken() error
	GetSheetData(sheetId, sheetName string) ([]byte, error)
	IsTokenExpired() bool
}
