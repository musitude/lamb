package lamb

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rs/zerolog"
)

// S3Handler is the S3 event handler func for AWS lambda
type S3Handler func(ctx context.Context, r events.S3Event) error

// S3HandlerFunc is the lamb handler that users of this library implement. It gives access to convenience methods via `ctx`
type S3HandlerFunc func(ctx *S3Context) error

// APIGatewayProxyContext provides convenience methods for working with API Gateway requests and responses
type S3Context struct {
	Context context.Context
	Logger  zerolog.Logger
	Event   events.S3EventRecord
}

// NewS3Handler adapts the lamb APIGatewayProxyHandlerFunc to the AWS lambda handler that is passed to lambda.Start
func NewS3Handler(handlerFunc S3HandlerFunc) func(ctx context.Context, r events.S3Event) error {
	return func(ctx context.Context, e events.S3Event) error {
		for _, record := range e.Records {
			logger := zerolog.New(os.Stdout).With().
				Timestamp().
				Caller().
				Logger()

			c := &S3Context{
				Context: ctx,
				Event:   record,
				Logger:  logger,
			}

			if err := handlerFunc(c); err != nil {
				logUnhandledError(c.Logger, err)
				return err
			}
		}
		return nil
	}
}
