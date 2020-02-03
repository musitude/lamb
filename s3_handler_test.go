package lamb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"

	"github.com/musitude/lamb"
)

func TestS3Handler_Success(t *testing.T) {
	timesCalled := 0
	h := lamb.NewS3Handler(func(c *lamb.S3Context) error {
		timesCalled++

		assert.Equal(t, fmt.Sprintf("key-%d", timesCalled), c.Event.S3.Object.Key)

		return nil
	})

	err := h(context.Background(), events.S3Event{
		Records: []events.S3EventRecord{
			{
				S3: events.S3Entity{
					Object: events.S3Object{
						Key: "key-1",
					},
				},
			},
			{
				S3: events.S3Entity{
					Object: events.S3Object{
						Key: "key-2",
					},
				},
			},
		}})

	assert.Nil(t, err)
	assert.Equal(t, 2, timesCalled)
}

func TestS3Handler_HandlerError(t *testing.T) {
	timesCalled := 0
	handlerFunc := func(c *lamb.S3Context) error {
		timesCalled++
		return assert.AnError
	}

	h := lamb.NewS3Handler(handlerFunc)
	err := h(context.Background(), events.S3Event{
		Records: []events.S3EventRecord{
			{
				S3: events.S3Entity{
					Object: events.S3Object{
						Key: "key-2",
					},
				},
			},
		}})

	assert.Equal(t, assert.AnError, err)
	assert.Equal(t, 1, timesCalled)
}
