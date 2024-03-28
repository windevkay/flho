// Package main provides helper functions for the API.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

const (
	// lowerChars string = "abcdefghijklmnopqrstuvwxyz"
	upperChars string = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits     string = "0123456789"
)

// envelope is a helper type for creating JSON envelopes.
type envelope map[string]interface{}

// generateWorkflowUniqueId generates a unique ID for a workflow.
func (app *application) generateWorkflowUniqueId() string {
	charset := digits + upperChars
	generatedId := make([]byte, 15)

	for index := range generatedId {
		generatedId[index] = charset[rand.Intn(len(charset))]
	}

	return string(generatedId)
}

// readIDParam extracts and parses the "id" parameter from the request URL.
// It returns the parsed ID as an int64 value, or an error if the ID is invalid.
func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)

	if err != nil || id < 1 {
		return 0, errors.New("invalid ID parameter")
	}

	return id, nil
}

// writeJSON writes the given data as a JSON response with the specified status code and headers.
// It returns an error if the data cannot be serialized to JSON or if there is an error writing the response.
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

// readJSON reads and decodes the JSON request body into the provided destination object.
// It enforces a maximum request body size of 1MB and disallows unknown fields in the JSON.
// It returns an error if the JSON is malformed, the body is empty, the JSON contains unknown fields,
// or if the body exceeds the maximum size.
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	maxBytes := 1_048_576 // 1MB max request body
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}
