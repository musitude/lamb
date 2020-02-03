package lamb

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
)

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

func logUnhandledError(logger zerolog.Logger, err error) {
	if isErisErr := eris.Unpack(err).ExternalErr == ""; isErisErr {
		logger.Error().
			Fields(map[string]interface{}{
				"error": eris.ToJSON(err, true),
			}).
			Msg("Unhandled error")
	} else {
		logger.Error().Msgf("Unhandled error: %+v", err)
	}
}
