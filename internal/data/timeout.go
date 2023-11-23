package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Timeout int32

var ErrInvalidTimeoutFormat = errors.New("invalid timeout format")

func (t Timeout) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", t)
	quotedJSONValue := strconv.Quote(jsonValue)

	return []byte(quotedJSONValue), nil
}

func (t *Timeout) UnmarshalJSON(jsonValue []byte) error {
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidTimeoutFormat
	}

	parts := strings.Split(unquotedJSONValue, " ")

	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidTimeoutFormat
	}

	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidTimeoutFormat
	}

	*t = Timeout(i)

	return nil
}
