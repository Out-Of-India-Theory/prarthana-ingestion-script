package zoho

import (
	"context"

	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/entity"
)

type Service interface {
	RefreshAccessToken() (string, error)
	GetSheetData(ctx context.Context, sheetName string, response interface{}) error
	SetSheetData(ctx context.Context, sheetName string, data entity.ShlokaSheetResponse, startId int) error
	AddUUIDToSheet(ctx context.Context, sheetName string, uuid string, row int) error
}
