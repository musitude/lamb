package lamb_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/PaesslerAG/jsonpath"
	"github.com/aws/aws-lambda-go/events"
	adapter "github.com/gaw508/lambda-proxy-http-adapter"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"

	"github.com/musitude/lamb"
)

type body struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (b body) Validate() error {
	if b.Status == "" {
		return errors.New("status empty")
	}
	return nil
}

func TestBind(t *testing.T) {
	h := handler(lamb.Handle(func(c *lamb.Context) error {
		var b body
		err := c.Bind(&b)
		if err != nil {
			t.Fatalf("expected err to be nil. %s", err)
		}

		if b.Name != "Mei" {
			t.Fatal("expected body to contain 'name==Mei'")
		}

		if b.Status != "ACTIVE" {
			t.Fatal("expected body to contain 'status==ACTIVE'")
		}

		return c.OK(nil)
	}))

	apitest.New().
		Handler(h).
		Get("/").
		JSON(`{
			"name": "Mei",
			"status": "ACTIVE"
		}`).
		Expect(t).
		Status(http.StatusOK).
		End()
}

func TestBind_Validate(t *testing.T) {
	h := handler(lamb.Handle(func(c *lamb.Context) error {
		var b body
		err := c.Bind(&b)
		if err == nil {
			t.Fatal("expected err from validation")
		}

		if err.Error() != "status empty" {
			t.Fatalf("unexpected validation message: %s", err.Error())
		}

		return c.OK(nil)
	}))

	apitest.New().
		Handler(h).
		Get("/").
		JSON(`{
			"name": "Mei"
		}`).
		Expect(t).
		Status(http.StatusOK).
		End()
}

func TestBind_HandlesInvalidRequestJSON(t *testing.T) {
	h := handler(lamb.Handle(func(c *lamb.Context) error {
		var b body
		err := c.Bind(&b)
		if err == nil {
			t.Fatal("expected err from validation")
		}

		if err != lamb.ErrInvalidBody {
			t.Fatal("expected ErrInvalidBody from validation")
		}

		return c.JSON(http.StatusBadRequest, nil)
	}))

	apitest.New().
		Handler(h).
		Get("/").
		JSON(`not json`).
		Expect(t).
		Status(http.StatusBadRequest).
		End()
}

func TestBind_HandlesInvalidResponseJSON(t *testing.T) {
	h := handler(lamb.Handle(func(c *lamb.Context) error {
		forceErrorValue := make(chan int)
		return c.OK(forceErrorValue)
	}))

	apitest.New().
		Handler(h).
		Get("/").
		Expect(t).
		Status(http.StatusInternalServerError).
		End()
}

func TestCreated(t *testing.T) {
	h := handler(lamb.Handle(func(c *lamb.Context) error {
		c.Logger.Log().Str("my_custom_field", "33").Msg("It worked!")

		c.Header("Custom", "54321")
		return c.Created("12345")
	}))

	apitest.New().
		Handler(h).
		Get("/").
		Expect(t).
		Status(http.StatusCreated).
		Header("Location", "12345").
		Header("Custom", "54321").
		End()
}

func TestErrorResponse_InternalServer(t *testing.T) {
	h := handler(lamb.Handle(func(c *lamb.Context) error {
		return errors.New("error")
	}))

	apitest.New().
		Handler(h).
		Get("/").
		Expect(t).
		Status(http.StatusInternalServerError).
		Body(`{"code":"INTERNAL_SERVER_ERROR", "detail":"Internal server error"}`).
		End()
}

func TestErrorResponse_CustomError(t *testing.T) {
	h := handler(lamb.Handle(func(c *lamb.Context) error {
		return lamb.Err{
			Status: http.StatusBadRequest,
			Code:   "INVALID_QUERY_PARAM",
			Detail: "Invalid query param",
		}
	}))

	apitest.New().
		Handler(h).
		Get("/").
		Expect(t).
		Status(http.StatusBadRequest).
		Body(`{"code":"INVALID_QUERY_PARAM", "detail":"Invalid query param"}`).
		End()
}

func TestErrorResponse_SupportsParams(t *testing.T) {
	h := handler(lamb.Handle(func(c *lamb.Context) error {
		return lamb.Err{
			Status: http.StatusBadRequest,
			Code:   "INVALID_QUERY_PARAM",
			Detail: "Invalid query param",
			Params: map[string]string{
				"custom": "content",
			},
		}
	}))

	apitest.New().
		Handler(h).
		Get("/").
		Expect(t).
		Status(http.StatusBadRequest).
		Body(`{"code":"INVALID_QUERY_PARAM", "detail":"Invalid query param", "params": {"custom": "content"}}`).
		End()
}

func TestError_Error(t *testing.T) {
	err := lamb.Err{
		Status: http.StatusBadRequest,
		Code:   "INVALID_QUERY_PARAM",
		Detail: "Invalid query param",
	}

	if err.Error() != "Code: INVALID_QUERY_PARAM; Status: 400; Detail: Invalid query param" {
		t.Fatalf("unexpected error: '%s'", err.Error())
	}
}

func TestAPIGatewayProxyHandler_UnhandledErrorWithStackTrace(t *testing.T) {
	logCaptor := &bytes.Buffer{}
	h := handler(lamb.Handle(func(c *lamb.Context) error {
		c.Logger = zerolog.New(logCaptor)
		return eris.Wrap(errors.New("source err"), "add context")
	}))

	apitest.New().
		Handler(h).
		Get("/").
		Expect(t).
		Status(http.StatusInternalServerError).
		Body(`{
			"code": "INTERNAL_SERVER_ERROR",
			"detail": "Internal server error"
		}`).
		End()

	logMessage := logCaptor.Bytes()
	assert.Equal(t, "Unhandled error", jsonPath("$.message", logMessage))
	assert.Equal(t, "source err", jsonPath("$.error.root.message", logMessage))
	assert.NotEmpty(t, jsonPath("$.error.root.stack", logMessage))
	assert.Equal(t, "add context", jsonPath("$.error.wrap[0].message", logMessage))
	assert.NotEmpty(t, jsonPath("$.error.wrap[0].stack", logMessage))
}

func handler(handler lamb.APIGatewayProxyHandler) http.Handler {
	h := func(r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return handler(context.Background(), r)
	}
	return adapter.GetHttpHandler(h, "/", nil)
}

func jsonPath(path string, jsonData []byte) interface{} {
	v := interface{}(nil)
	err := json.Unmarshal(jsonData, &v)
	if err != nil {
		panic(err)
	}

	value, err := jsonpath.Get(path, v)
	if err != nil {
		panic(err)
	}
	return value
}
