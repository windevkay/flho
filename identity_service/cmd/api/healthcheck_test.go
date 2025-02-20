package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthcheckHandler(t *testing.T) {
	tests := []struct {
		name     string
		wantCode int
	}{
		{
			name:     "it should return a 200 on success",
			wantCode: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/v1/healthcheck", nil)

			testApp.healthcheckHandler(rr, req)

			if rr.Code != test.wantCode {
				t.Errorf("got status code %d, want %d", rr.Code, test.wantCode)
			}
		})
	}
}
