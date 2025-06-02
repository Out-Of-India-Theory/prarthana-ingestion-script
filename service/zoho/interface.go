package zoho

import (
	"context"
)

type Service interface {
	RefreshAccessToken() (string, error)
	GetSheetData(ctx context.Context, sheetName string, response interface{}) error
	SetSheetData(ctx context.Context, sheetName string, row int, columnIndexes []int, dataArray []interface{}) error
	AddUUIDToSheet(ctx context.Context, sheetName string, uuid string, row int) error
}
