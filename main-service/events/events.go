package events

import "context"

type Producer interface {
	Publish(context.Context, string, []byte) error
}
