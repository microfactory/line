package line

import (
	"context"
	"encoding/json"
	"time"
)

//EarlyTimeout will cancel the lambda context millisecond before actual timeout to give handlers time to shutdown cleanly
func EarlyTimeout(shutdownTimeMillis int64) func(Handler) Handler {
	return func(h Handler) Handler {
		return HandlerFunc(
			func(ctx context.Context, msg json.RawMessage) (interface{}, error) {
				lctx, ok := InvocationFromContext(ctx)
				if !ok {
					return h.HandleEvent(ctx, msg)
				}

				ctx, cancel := context.WithTimeout(ctx, time.Millisecond*time.Duration(lctx.RemainingTimeInMillis()-5000))
				defer cancel()
				return h.HandleEvent(ctx, msg)
			})
	}
}
