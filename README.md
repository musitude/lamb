# lamb

[![Build Status](https://travis-ci.org/musitude/lamb.svg?branch=master)](https://travis-ci.org/musitude/lamb)
[![Coverage Status](https://coveralls.io/repos/github/musitude/lamb/badge.svg?branch=master)](https://coveralls.io/github/musitude/lamb?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/musitude/lamb)](https://goreportcard.com/report/github.com/musitude/lamb)

Provides the following utilities to simplify working with AWS lambda and Api Gateway.

* HTTP request parsing with JSON support and request body validation
* HTTP response writer with JSON support
* Custom error type with JSON support
* Logging using zerolog

## Request body parsing

Use the bind method to unmarshal the response body to a struct

```go
type requestBody struct {
	Name   string `json:"name"`
}

handler := lamb.Handle(func(c *lamb.Context) error {
	var b requestBody
	err := c.Bind(&b)

	...
})
```

## Request body validation

Implement the `Validate` method on the struct

```go
type requestBody struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (b requestBody) Validate() error {
	if b.Status == "" {
		return errors.New("status empty")
	}
	return nil
}

handler := lamb.Handle(func(c *lamb.Context) error {
	var b requestBody
	err := c.Bind(&b)
	
	// work with requestBody.Name or err == "status empty"
	
	return c.JSON(http.StatusOK, responseBody)
}
```

## Response writer

There are several methods provided to simplify writing HTTP responses. 

```go
handler := lamb.Handle(func(c *lamb.Context) error {
	...
	return c.JSON(http.StatusOK, responseBody)
})
```

`lamb.OK(responseBody)` sets the HTTP status code to `http.StatusOK` and marshals `responseBody` as JSON.

## Errors

### Custom Errors

You can pass custom `lamb` errors and also map then to HTTP status codes

```go
handler := lamb.Handle(func(c *lamb.Context) error {
	return c.Error(lamb.Err{
		Status: http.StatusBadRequest,
		Code:   "INVALID_QUERY_PARAM",
		Detail: "Invalid query param",
		Params: map[string]string{
			"custom": "content",
		},
	})
})
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

### Bubbling errors up

Go errors returned in the handler are automatically marshalled to a generic json HTTP response and the status code is set to 500. These errors are also logged. If you wrap the source error with [eris](https://github.com/rotisserie/eris) a stack trace is included in the log.

### Access the logger

```go
func(c *lamb.Context) error {
    c.Logger.Log().Str("my_custom_field", "33").Msg("It worked!") // {"my_custom_field":"33","time":"2020-01-08T09:27:07Z","caller":"/path/to/file.go:125","message":"It worked!"}
}
```
