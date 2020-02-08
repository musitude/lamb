package lamb

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/rs/zerolog"
)

// DynamoDBHandler is the DynamoDB event handler func for AWS lambda
type DynamoDBHandler func(ctx context.Context, r events.DynamoDBEvent) error

// DynamoDBHandlerFunc is the lamb handler that users of this library implement. It gives access to convenience methods via `ctx`
type DynamoDBHandlerFunc func(ctx *DynamoDBContext) error

// APIGatewayProxyContext provides convenience methods for working with API Gateway requests and responses
type DynamoDBContext struct {
	Context context.Context
	Logger  zerolog.Logger
	Event   events.DynamoDBEventRecord
}

// NewDynamoDBHandler adapts the lamb APIGatewayProxyHandlerFunc to the AWS lambda handler that is passed to lambda.Start
func NewDynamoDBHandler(handlerFunc DynamoDBHandlerFunc) func(ctx context.Context, r events.DynamoDBEvent) error {
	return func(ctx context.Context, e events.DynamoDBEvent) error {
		for _, record := range e.Records {
			logger := zerolog.New(os.Stdout).With().
				Timestamp().
				Caller().
				Logger()

			c := &DynamoDBContext{
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

// Bind attempts to populate the provided struct with data from the HTTP request body.
// It also performs validation if the provided struct implements `Validatable`
func (c *DynamoDBContext) Bind(v interface{}) error {
	if c.EventType() == events.DynamoDBOperationTypeRemove {
		return unmarshalStreamImage(c.Event.Change.Keys, v)
	}
	return unmarshalStreamImage(c.Event.Change.NewImage, v)
}

func unmarshalStreamImage(attribute map[string]events.DynamoDBAttributeValue, out interface{}) error {
	dbAttrMap := make(map[string]*dynamodb.AttributeValue)
	for k, v := range attribute {
		var dbAttr dynamodb.AttributeValue
		bytes, marshalErr := v.MarshalJSON(); if marshalErr != nil {
			return marshalErr
		}
		if err := json.Unmarshal(bytes, &dbAttr);err != nil {
			return err
		}
		dbAttrMap[k] = &dbAttr
	}
	return dynamodbattribute.UnmarshalMap(dbAttrMap, out)
}

func (c *DynamoDBContext) EventType() events.DynamoDBOperationType {
	return events.DynamoDBOperationType(c.Event.EventName)
}