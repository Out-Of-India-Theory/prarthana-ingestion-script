package sync

import "context"

type Service interface {
	SyncPrarthanas(ctx context.Context, startID, endID int) error
	SyncStotras(ctx context.Context, startID, endID int) error
	SyncShloks(ctx context.Context, startID, endID int) error
}
