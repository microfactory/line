package line

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
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

//SessionFromContext will return a aws session from the lambda context or panic if it isn't available
func SessionFromContext(ctx context.Context) *session.Session {
	sess, ok := ctx.Value(sessionKey).(*session.Session)
	if !ok {
		panic("no aws session available in context")
	}

	return sess
}

const sessionKey key = 0

//WithSession will include a aws.Session into the context with line specific environment
func WithSession() func(Handler) Handler {
	region := os.Getenv("LINE_AWS_REGION")
	accessKey := os.Getenv("LINE_AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("LINE_AWS_SECRET_ACCESS_KEY")
	if region == "" || accessKey == "" || secretKey == "" {
		panic("cannot use line session without all of the following environment variables: LINE_AWS_REGION, LINE_AWS_ACCESS_KEY_ID, LINE_AWS_SECRET_ACCESS_KEY")
	}

	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(region),
			Credentials: credentials.NewStaticCredentials(
				accessKey,
				secretKey,
				"",
			),
		},
	)
	if err != nil {
		panic("failed to setup AWS session: " + err.Error())
	}

	return func(h Handler) Handler {
		return HandlerFunc(
			func(ctx context.Context, msg json.RawMessage) (interface{}, error) {
				ctx = context.WithValue(ctx, sessionKey, sess)
				return h.HandleEvent(ctx, msg)
			})
	}
}
