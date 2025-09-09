package store

import "context"

type Store interface {
	CreateItem(ctx context.Context, name string) (int64, error)
}
