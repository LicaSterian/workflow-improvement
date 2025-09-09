package cache

import "context"

type Cache interface {
	Get(context.Context, string) (string, error)
	Set(context.Context, string, string) error
}
