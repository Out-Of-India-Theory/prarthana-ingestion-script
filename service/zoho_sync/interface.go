package zoho_sync

import "context"

type Service interface {
	SyncPrarthanaToSheet(ctx context.Context, prarthanaID string) error
}
