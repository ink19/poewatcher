package notify

import "context"

type Client interface {
	SendTextMsg(ctx context.Context, msg string) error
}
