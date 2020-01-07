package lamb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rs/zerolog"
)

type APIGatewayProxyHandler = func(r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

type Context struct {
	Logger   zerolog.Logger
	Request  events.APIGatewayProxyRequest
	Response events.APIGatewayProxyResponse
}

func Handle(handlerFunc func(ctx *Context) error) APIGatewayProxyHandler {
	return func(r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		log := zerolog.New(os.Stdout).With().
			Timestamp().
			Caller().
			Logger()

		c := &Context{
			Logger:   log,
			Request:  r,
			Response: events.APIGatewayProxyResponse{},
		}

		err := handlerFunc(c)
		return c.Response, err
	}
}

type Validatable interface {
	Validate() error
}

var ErrInternalServer = Err{
	Status: http.StatusInternalServerError,
	Code:   "INTERNAL_SERVER_ERROR",
	Detail: "Internal server error",
}

var ErrInvalidBody = Err{
	Status: http.StatusBadRequest,
	Code:   "INVALID_REQUEST_BODY",
	Detail: "Invalid request body",
}

type Err struct {
	Status int         `json:"-"`
	Code   string      `json:"code"`
	Detail string      `json:"detail"`
	Params interface{} `json:"params,omitempty"`
}

func (err Err) Error() string {
	errorParts := []string{
		fmt.Sprintf("Code: %s; Status: %d; Detail: %s", err.Code, err.Status, err.Detail),
	}
	return strings.Join(errorParts, "; ")
}

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

func (c *Context) Error(err error) error {
	var newErr Err
	switch err := err.(type) {
	case Err:
		newErr = err
	default:
		newErr = Err{
			Status: ErrInternalServer.Status,
			Code:   ErrInternalServer.Code,
			Detail: ErrInternalServer.Detail,
		}
		fmt.Printf("Unhandled error: %s", err.Error())
	}

	return c.JSON(newErr.Status, newErr)
}

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

func (c *Context) Header(k, v string) {
	if c.Response.Headers == nil {
		c.Response.Headers = map[string]string{k: v}
		return
	}
	c.Response.Headers[k] = v
}

func (c *Context) Created(location string) error {
	c.Header("Location", location)
	return c.JSON(http.StatusCreated, nil)
}

func (c *Context) OK(body interface{}) error {
	return c.JSON(http.StatusOK, body)
}
