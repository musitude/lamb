package lamb

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rs/zerolog"
)

// APIGatewayProxyHandler is the handler function for lambda events that originate from API Gateway
type APIGatewayProxyHandler = func(ctx context.Context, r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

// APIGatewayProxyHandlerFunc is the lamb handler that users of this library implement. It gives access to convenience methods via `ctx`
type APIGatewayProxyHandlerFunc func(ctx *APIGatewayProxyContext) error

// APIGatewayProxyContext provides convenience methods for working with API Gateway requests and responses
type APIGatewayProxyContext struct {
	Context  context.Context
	Logger   zerolog.Logger
	Request  events.APIGatewayProxyRequest
	Response events.APIGatewayProxyResponse
}

// NewAPIGatewayProxyHandler adapts the lamb APIGatewayProxyHandlerFunc to the AWS lambda handler that is passed to lambda.Start
func NewAPIGatewayProxyHandler(handlerFunc APIGatewayProxyHandlerFunc) APIGatewayProxyHandler {
	return func(context context.Context, r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		log := zerolog.New(os.Stdout).With().
			Timestamp().
			Caller().
			Logger()

		c := &APIGatewayProxyContext{
			Context:  context,
			Logger:   log,
			Request:  r,
			Response: events.APIGatewayProxyResponse{},
		}

		err := handlerFunc(c)
		if err != nil {
			c.handleError(err)
		}

		return c.Response, nil
	}
}

// Bind attempts to populate the provided struct with data from the HTTP request body.
// It also performs validation if the provided struct implements `Validatable`
func (c *APIGatewayProxyContext) Bind(v interface{}) error {
	return bind([]byte(c.Request.Body), v)
}

func (c *APIGatewayProxyContext) handleError(err error) {
	var newErr Err
	switch err := err.(type) {
	case Err:
		newErr = err
	default:
		newErr = ErrInternalServer
		logUnhandledError(c.Logger, err)
	}

	_ = c.JSON(newErr.Status, newErr)
}

// JSON writes the provided body and status code to the API Gateway response
func (c *APIGatewayProxyContext) JSON(statusCode int, body interface{}) error {
	var b []byte
	var err error
	if body != nil {
		b, err = json.Marshal(body)
		if err != nil {
			statusCode = http.StatusInternalServerError
			b, _ = json.Marshal(ErrInternalServer)
		}
	}

	c.Response.StatusCode = statusCode
	c.Response.Body = string(b)
	return nil
}

// Header writes the provided header to the API Gateway response
func (c *APIGatewayProxyContext) Header(k, v string) {
	if c.Response.Headers == nil {
		c.Response.Headers = map[string]string{k: v}
		return
	}
	c.Response.Headers[k] = v
}

// Created is a convenient method for writing HTTP Status 201 API Gateway responses.
// It is opinionated in that it sets the resource location header. If you do not want this use
//
//     c.JSON(http.StatusCreated, nil)
//
// instead.
func (c *APIGatewayProxyContext) Created(location string) error {
	c.Header("Location", location)
	return c.JSON(http.StatusCreated, nil)
}

// OK is a convenient method for writing the provided body and HTTP Status 200 to the API Gateway responses.
func (c *APIGatewayProxyContext) OK(body interface{}) error {
	return c.JSON(http.StatusOK, body)
}
