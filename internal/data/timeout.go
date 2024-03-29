package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Timeout int

var ErrInvalidTimeoutFormat = errors.New("invalid timeout format")

// transform custom struct field when writing to JSON
func (t Timeout) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d seconds", t)
	quotedJSONValue := strconv.Quote(jsonValue)

	return []byte(quotedJSONValue), nil
}

// transform incoming JSON field to match struct field type
func (t *Timeout) UnmarshalJSON(jsonValue []byte) error {
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidTimeoutFormat
	}

	parts := strings.Split(unquotedJSONValue, " ")

	if len(parts) != 2 || parts[1] != "seconds" {
		return ErrInvalidTimeoutFormat
	}

	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidTimeoutFormat
	}

	*t = Timeout(i)

	return nil
}
