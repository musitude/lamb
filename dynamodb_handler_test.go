package lamb_test

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"

	"github.com/musitude/lamb"
)

type record struct {
	ArtistID   string `json:"pk"`
	ProfileKey string `json:"sk"`
	ImageURL   string `json:"image_url"`
}

func TestDynamoDBHandler(t *testing.T) {
	h := lamb.NewDynamoDBHandler(func(c *lamb.DynamoDBContext) error {
		var rec record
		if err := c.Bind(c.Event.Change.NewImage, &rec); err != nil {
			return err
		}

		assert.Equal(t, record{
			ArtistID:   "ARTIST#1234567",
			ProfileKey: "PROFILE#",
			ImageURL:   "https://site.com/image.jpg",
		}, rec)
		assert.Equal(t, events.DynamoDBOperationTypeInsert, c.EventType())

		return nil
	})

	err := h(context.Background(), events.DynamoDBEvent{
		Records: []events.DynamoDBEventRecord{
			{
				EventName: "INSERT",
				Change: events.DynamoDBStreamRecord{
					Keys: map[string]events.DynamoDBAttributeValue{
						"pk": events.NewStringAttribute("ARTIST#1234567"),
						"sk": events.NewStringAttribute("PROFILE#"),
					},
					NewImage: map[string]events.DynamoDBAttributeValue{
						"pk":        events.NewStringAttribute("ARTIST#1234567"),
						"sk":        events.NewStringAttribute("PROFILE#"),
						"image_url": events.NewStringAttribute("https://site.com/image.jpg"),
					},
				},
			},
		},
	})

	assert.NoError(t, err)
}

func TestDynamoDBHandler_Remove(t *testing.T) {
	h := lamb.NewDynamoDBHandler(func(c *lamb.DynamoDBContext) error {
		var rec record
		if err := c.Bind(c.Event.Change.Keys, &rec); err != nil {
			return err
		}

		assert.Equal(t, record{
			ArtistID:   "ARTIST#1234567",
			ProfileKey: "PROFILE#",
		}, rec)
		assert.Equal(t, events.DynamoDBOperationTypeRemove, c.EventType())

		return nil
	})

	err := h(context.Background(), events.DynamoDBEvent{
		Records: []events.DynamoDBEventRecord{
			{
				EventName: "REMOVE",
				Change: events.DynamoDBStreamRecord{
					Keys: map[string]events.DynamoDBAttributeValue{
						"pk": events.NewStringAttribute("ARTIST#1234567"),
						"sk": events.NewStringAttribute("PROFILE#"),
					},
				},
			},
		},
	})

	assert.NoError(t, err)
}
