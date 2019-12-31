# lamb

Provides the following utilities to simplify working with AWS lambda and Api Gateway.

* HTTP request parsing with JSON support and request body validation
* HTTP response writer with JSON support
* Custom error type with JSON support

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
		
        ...

		return lamb.OK(nil)
	}
```
