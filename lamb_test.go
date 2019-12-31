package lamb_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	adapter "github.com/gaw508/lambda-proxy-http-adapter"
	"github.com/steinfletcher/apitest"

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
	h := func(r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		var b body
		err := lamb.Bind(r.Body, &b)
		if err != nil {
			t.Fatalf("expected err to be nil. %s", err)
		}

		if b.Name != "Mei" {
			t.Fatal("expected body to contain 'name==Mei'")
		}

		if b.Status != "ACTIVE" {
			t.Fatal("expected body to contain 'status==ACTIVE'")
		}

		return lamb.OK(nil)
	}

	apitest.New().
		Handler(adapter.GetHttpHandler(h, "/", nil)).
		Get("/").
		Body(`{
			"name": "Mei",
			"status": "ACTIVE"
		}`).
		Expect(t).
		Status(http.StatusOK).
		End()
}

func TestBind_Validate(t *testing.T) {
	h := func(r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		var b body
		err := lamb.Bind(r.Body, &b)
		if err == nil {
			t.Fatal("expected err from validation")
		}

		if err.Error() != "status empty" {
			t.Fatalf("unexpected validation message: %s", err.Error())
		}

		return lamb.OK(nil)
	}

	apitest.New().
		Handler(adapter.GetHttpHandler(h, "/", nil)).
		Get("/").
		Body(`{
			"name": "Mei"
		}`).
		Expect(t).
		Status(http.StatusOK).
		End()
}
