package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/windevkay/flho/internal/assert"
)

func TestHealthcheckHandler(t *testing.T) {
	// Arrange
	req, _ := http.NewRequest(http.MethodGet, "/healthcheck", nil)

	rr := httptest.NewRecorder()

	app := newTestApplication()

	handler := http.HandlerFunc(app.healthcheckHandler)

	// Act
	handler.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, rr.Code, http.StatusOK)

	// Check the response body is what we expect.
	// Prepare the expected response body.
	expectedMap := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": "test",
			"version":     "1.0.0",
		},
	}
	expectedBytes, _ := json.MarshalIndent(expectedMap, "", "\t")
	expected := string(expectedBytes) + "\n"

	assert.Equal(t, rr.Body.String(), expected)
}
