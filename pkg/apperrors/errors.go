package apperrors

import (
	"fmt"
	"net/http"

	"github.com/windevkay/flho/pkg/utils"
)

// ErrorResponse writes an error response to the http.ResponseWriter.
func ErrorResponse(w http.ResponseWriter, status int, message any) {
	env := utils.Envelope{"error": message}
	utils.WriteJSON(w, status, env, nil)
}

// ServerErrorResponse sends a server error response to the client.
func ServerErrorResponse(w http.ResponseWriter, err error) {
	message := "The server encountered a problem and could not process your request: " + err.Error()
	ErrorResponse(w, http.StatusInternalServerError, message)
}

// NotFoundResponse sends a HTTP 404 Not Found response to the client with the specified message.
func NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "The requested resource could not be found"
	ErrorResponse(w, http.StatusNotFound, message)
}

// MethodNotAllowedResponse sends a HTTP 405 Method Not Allowed response to the client.
func MethodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	ErrorResponse(w, http.StatusMethodNotAllowed, message)
}

// BadRequestResponse sends a HTTP 400 Bad Request response with the given error message.
func BadRequestResponse(w http.ResponseWriter, err error) {
	ErrorResponse(w, http.StatusBadRequest, err.Error())
}

// FailedValidationResponse sends a failed validation response with the specified errors.
func FailedValidationResponse(w http.ResponseWriter, errors map[string]string) {
	ErrorResponse(w, http.StatusUnprocessableEntity, errors)
}
