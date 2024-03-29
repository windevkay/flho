package main

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/windevkay/flho/internal/assert"
)

func TestGenerateWorkflowUniqueId(t *testing.T) {
	// Arrange
	app := newTestApplication()
	expectedLength := 15

	// Act
	generatedId := app.generateWorkflowUniqueId()

	// Assert
	assert.Equal(t, len(generatedId), expectedLength)
}

func TestReadIDParam(t *testing.T) {
	// Arrange
	app := newTestApplication()
	req, _ := http.NewRequest(http.MethodGet, "", nil)

	tests := []struct {
		name string
		args string
		want int64
		err  error
	}{
		{name: "valid param", args: "1", want: 1, err: nil},
		{name: "invalid param", args: "0", want: 0, err: errors.New("invalid ID parameter")},
	}

	// Act
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := httprouter.Params{{Key: "id", Value: tt.args}}
			ctx := context.WithValue(req.Context(), httprouter.ParamsKey, params)
			req = req.WithContext(ctx)

			id, err := app.readIDParam(req)

			// Assert
			if tt.err != nil {
				assert.Equal(t, err.Error(), tt.err.Error())
			} else {
				assert.Equal(t, id, tt.want)
			}
		})
	}
}

func TestWriteJSON(t *testing.T) {}
