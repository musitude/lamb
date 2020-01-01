# lamb

[![Build Status](https://travis-ci.org/musitude/lamb.svg?branch=master)](https://travis-ci.org/musitude/lamb)
[![Coverage Status](https://coveralls.io/repos/github/musitude/lamb/badge.svg?branch=master)](https://coveralls.io/github/musitude/lamb?branch=master)

Provides the following utilities to simplify working with AWS lambda and Api Gateway.

* HTTP request parsing with JSON support and request body validation
* HTTP response writer with JSON support
* Custom error type with JSON support

## Request body parsing

Use the bind method to unmarshal the response body to a struct

```go
type requestBody struct {
	Name   string `json:"name"`
}

handler := func(r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var b requestBody
	err := lamb.Bind(r.Body, &b)

	...
}
```

## Request body validation

Implement the `Validate` method on the struct

```go
type requestBody struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (b body) Validate() error {
	if b.Status == "" {
		return errors.New("status empty")
	}
	return nil
}

handler := func(r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var b requestBody
	err := lamb.Bind(r.Body, &b)
	
	// work with requestBody.Name or err == "status empty"
	
	return lamb.JSON(http.StatusOK, responseBody)
}
```

## Response writer

There are several methods provided to simplify writing HTTP responses. 

```go
handler := func(r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	...
	lamb.JSON(http.StatusOK, responseBody)
}
```

`lamb.OK(responseBody)` sets the HTTP status code to `http.StatusOK` and marshals `responseBody` as JSON.

## Errors

### Go Errors

Passing Go errors to the error response writer will log the error and respond with an internal server error

```go
handler := func(r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return lamb.JSON(http.StatusOK, myStruct))
}
```

### Custom Errors

You can pass custom `lamb` errors and also map then to HTTP status codes

```go
handler := func(r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return lamb.ErrorResponse(&lamb.Error{
		Status: http.StatusBadRequest,
		Code:   "INVALID_QUERY_PARAM",
		Detail: "Invalid query param",
		Params: map[string]string{
			"custom": "content",
		},
	})
}
```

Writes the the following response

```json
{
  "code": "INVALID_QUERY_PARAM",
  "detail": "Invalid query param",
  "params": {
    "custom": "content"
  }
}
```

where params is type `interface{}` to support arbitrary data in responses.
