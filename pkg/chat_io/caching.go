package chat_io

import "context"

type Caching interface {
	Store(ctx context.Context, cancle context.CancelFunc, payload Transporter)
}
