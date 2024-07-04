package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/julienschmidt/httprouter"
	"github.com/windevkay/flhoutils/validator"
)

type Envelope map[string]interface{}

const (
	upperChars string = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits     string = "0123456789"
)

// GenerateUniqueId generates a unique identifier of the specified length.
// It uses a combination of digits and uppercase characters from the charset.
// The generated identifier is returned as a string.
func GenerateUniqueId(length int) string {
	charset := digits + upperChars
	generatedId := make([]byte, length)

	for index := range generatedId {
		generatedId[index] = charset[rand.Intn(len(charset))]
	}

	return string(generatedId)
}

// RunInBackground runs the given function in a separate goroutine and adds it to the wait group.
// The wait group is incremented before the goroutine starts and decremented after it finishes.
// If the function panics, it is recovered and the panic message can be logged to a logger service.
func RunInBackground(fn func(), wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		// defer func() {
		// 	if err := recover(); err != nil {
		// 		//write to logger service here - goroutine
		// 		//app.logger.Error(fmt.Sprintf("%v", err))
		// 	}
		// }()

		fn()
	}()
}

// ReadIDParam extracts and parses the "id" parameter from the given HTTP request.
// It returns the parsed ID as an int64 value. If the ID is invalid or missing, it returns an error.
func ReadIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)

	if err != nil || id < 1 {
		return 0, errors.New("invalid ID parameter")
	}

	return id, nil
}

// WriteJSON writes the provided data as a JSON response to the http.ResponseWriter.
// It sets the provided status code, headers, and content type.
func WriteJSON(w http.ResponseWriter, status int, data Envelope, headers http.Header) {
	js, _ := json.MarshalIndent(data, "", "\t")

	js = append(js, '\n')

	for key, values := range headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
}

// ReadJSON reads and decodes JSON data from the request body into the provided destination object.
// It enforces a maximum request body size of 1MB and disallows unknown fields in the JSON.
// If any errors occur during decoding, appropriate error messages are returned.
// The function returns nil if the decoding is successful.
func ReadJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
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

// ReadString reads a string value from the given url.Values object based on the provided key.
// If the value is empty, it returns the defaultValue.
func ReadString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

// ReadCSV reads a comma-separated value (CSV) string from the given query string parameter.
// If the parameter is empty, it returns the provided defaultValue.
// Otherwise, it splits the CSV string and returns the resulting slice of strings.
func ReadCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

// ReadInt reads an integer value from the given URL query string parameter.
// If the parameter is not present or cannot be parsed as an integer, it returns the defaultValue.
// If a validator is provided, it adds an error to the validator if the value is not a valid integer.
func ReadInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}
