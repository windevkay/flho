package errors

import (
	"fmt"
	"net/http"

	"github.com/windevkay/flhoutils/helpers"
)

// ErrorResponse writes an error response to the http.ResponseWriter.
// It takes the http.ResponseWriter, http.Request, status code, and error message as input parameters.
// It creates an envelope with the error message and writes it as JSON to the response writer.
func ErrorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := helpers.Envelope{"error": message}
	//write to logger service here - goroutine
	//app.logger.Error(err.Error(), "method", r.Method, "uri", r.URL.RequestURI())

	helpers.WriteJSON(w, status, env, nil)
}

// ServerErrorResponse sends a server error response to the client.
// It takes the http.ResponseWriter, http.Request, and an error as parameters.
// The function sets the HTTP status code to 500 (Internal Server Error)
// and sends a message indicating that the server encountered a problem
// and could not process the request.
func ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	message := "The server encountered a problem and could not process your request: " + err.Error()
	ErrorResponse(w, r, http.StatusInternalServerError, message)
}

// NotFoundResponse sends a HTTP 404 Not Found response to the client with the specified message.
func NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "The requested resource could not be found"
	ErrorResponse(w, r, http.StatusNotFound, message)
}

// MethodNotAllowedResponse sends a HTTP 405 Method Not Allowed response to the client.
// It takes the http.ResponseWriter and http.Request as parameters.
// It generates an error message indicating that the specified HTTP method is not supported for the requested resource,
// and calls the ErrorResponse function to send the error response to the client.
func MethodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	ErrorResponse(w, r, http.StatusMethodNotAllowed, message)
}

// BadRequestResponse sends a HTTP 400 Bad Request response with the given error message.
func BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	ErrorResponse(w, r, http.StatusBadRequest, err.Error())
}

// FailedValidationResponse sends a failed validation response with the specified errors.
// It writes the response to the given http.ResponseWriter and http.Request.
// The HTTP status code used is http.StatusUnprocessableEntity.
// The errors parameter is a map where the keys represent the field names and the values represent the error messages.
func FailedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	ErrorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// EditConflictResponse handles the response for an edit conflict (mainly arising from race conditions).
// It sends an error response with the specified message and HTTP status code.
func EditConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "Unable to update the record, please try again"
	ErrorResponse(w, r, http.StatusConflict, message)
}

// RateLimitExceededResponse sends a rate limit exceeded response to the client.
// It takes in the http.ResponseWriter and http.Request as parameters.
// It calls the ErrorResponse function to send the response with the appropriate status code and message.
func RateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "Rate limit exceeded"
	ErrorResponse(w, r, http.StatusTooManyRequests, message)
}

// InvalidCredentialsResponse sends an HTTP response with a status code of 401 (Unauthorized)
// and a message indicating invalid authentication credentials.
func InvalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "Invalid authentication credentials"
	ErrorResponse(w, r, http.StatusUnauthorized, message)
}

// InvalidAuthenticationTokenResponse sends a response indicating that the authentication token is invalid or missing.
// It sets the "WWW-Authenticate" header to "Bearer" to remind clients that a bearer token is required.
// It also sends an error response with the specified message and HTTP status code.
func InvalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	message := "Invalid or missing authentication token"
	ErrorResponse(w, r, http.StatusUnauthorized, message)
}

// AuthenticationRequiredResponse sends an authentication required response to the client.
// It sets the HTTP status code to 401 Unauthorized and includes the provided message in the response body.
func AuthenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "You must be authenticated to access this resource"
	ErrorResponse(w, r, http.StatusUnauthorized, message)
}

// InactiveAccountResponse sends a response indicating that the user account is inactive.
// It takes the http.ResponseWriter and http.Request as parameters.
func InactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "Your user account must be activated to access this resource"
	ErrorResponse(w, r, http.StatusForbidden, message)
}
