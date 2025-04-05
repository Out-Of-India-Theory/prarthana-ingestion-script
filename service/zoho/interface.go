package zoho

import (
	"context"
)

type Service interface {
	RefreshAccessToken() (string, error)
	GetSheetData(ctx context.Context, sheetName string, response interface{}) error
}
