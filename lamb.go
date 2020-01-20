package lamb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
)

// APIGatewayProxyHandler is the handler function for lambda events that originate from API Gateway
type APIGatewayProxyHandler = func(ctx context.Context, r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

// Handler is the lamb handler that uses of this library implement. It gives access to convenience methods via `ctx`
type Handler func(ctx *Context) error

// Context provides convenience methods for working with API Gateway requests and responses
type Context struct {
	Context  context.Context
	Logger   zerolog.Logger
	Request  events.APIGatewayProxyRequest
	Response events.APIGatewayProxyResponse
}

// Handle adapts the lamb Handler to the AWS lambda handler that is passed to lambda.Start
func Handle(handlerFunc Handler) APIGatewayProxyHandler {
	return func(context context.Context, r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		log := zerolog.New(os.Stdout).With().
			Timestamp().
			Caller().
			Logger()

		c := &Context{
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

// Validatable is implemented by the request body struct.
// Example:
//
//      type requestBody struct {
// 	      Name   string `json:"name"`
// 	      Status string `json:"status"`
//      }
//
//      func (b body) Validate() error {
// 	      if b.Status == "" {
// 		    return errors.New("status empty")
// 	      }
// 	      return nil
//      }
//
// This will then be validated in `ctx.Bind`
type Validatable interface {
	Validate() error
}

// ErrInternalServer is a standard error to represent server failures
var ErrInternalServer = Err{
	Status: http.StatusInternalServerError,
	Code:   "INTERNAL_SERVER_ERROR",
	Detail: "Internal server error",
}

// ErrInvalidBody is a standard error to represent an invalid request body
var ErrInvalidBody = Err{
	Status: http.StatusBadRequest,
	Code:   "INVALID_REQUEST_BODY",
	Detail: "Invalid request body",
}

// Err is the error type returned to consumers of the API
type Err struct {
	Status int         `json:"-"`
	Code   string      `json:"code"`
	Detail string      `json:"detail"`
	Params interface{} `json:"params,omitempty"`
}

// Error implements Go's error condition
func (err Err) Error() string {
	errorParts := []string{
		fmt.Sprintf("Code: %s; Status: %d; Detail: %s", err.Code, err.Status, err.Detail),
	}
	return strings.Join(errorParts, "; ")
}

// Bind attempts to populate the provided struct with data from the HTTP request body.
// It also performs validation if the provided struct implements `Validatable`
func (c *Context) Bind(v interface{}) error {
	if err := json.Unmarshal([]byte(c.Request.Body), v); err != nil {
		return ErrInvalidBody
	}

	if validatable, ok := v.(Validatable); ok {
		err := validatable.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Context) handleError(err error) {
	var newErr Err
	switch err := err.(type) {
	case Err:
		newErr = err
	default:
		newErr = ErrInternalServer
		if isErisErr := eris.Unpack(err).ExternalErr == ""; isErisErr {
			c.Logger.Error().
				Fields(map[string]interface{}{
					"error": eris.ToJSON(err, true),
				}).
				Msg("Unhandled error")
		} else {
			c.Logger.Error().Msgf("Unhandled error: %+v", err)
		}
	}

	_ = c.JSON(newErr.Status, newErr)
}

// JSON writes the provided body and status code to the API Gateway response
func (c *Context) JSON(statusCode int, body interface{}) error {
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
func (c *Context) Header(k, v string) {
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
func (c *Context) Created(location string) error {
	c.Header("Location", location)
	return c.JSON(http.StatusCreated, nil)
}

// OK is a convenient method for writing the provided body and HTTP Status 200 to the API Gateway responses.
func (c *Context) OK(body interface{}) error {
	return c.JSON(http.StatusOK, body)
}
