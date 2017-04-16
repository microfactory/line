package session

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/microfactory/line"
)

//FromContext will return a aws session from the lambda context or panic if it isn't available
func FromContext(ctx context.Context) *session.Session {
	sess, ok := ctx.Value(sessionKey).(*session.Session)
	if !ok {
		panic("no aws session available in context")
	}

	return sess
}

type key int

const sessionKey key = 0

//LineSession will include a aws.Session into the context with line specific environment
func LineSession() func(line.Handler) line.Handler {
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

	return func(h line.Handler) line.Handler {
		return line.HandlerFunc(
			func(ctx context.Context, msg json.RawMessage) (interface{}, error) {
				ctx = context.WithValue(ctx, sessionKey, sess)
				return h.HandleEvent(ctx, msg)
			})
	}
}
