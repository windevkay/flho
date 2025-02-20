package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/windevkay/flho/identity_service/internal/services"
	"github.com/windevkay/flhoutils/helpers"
)

func TestRegisterIdentityHandler(t *testing.T) {
	tests := []struct {
		name     string
		input    services.RegisterIdentityInput
		wantCode int
		wantErr  bool
	}{
		{
			name: "valid registration",
			input: services.RegisterIdentityInput{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Test User",
			},
			wantCode: http.StatusCreated,
			wantErr:  false,
		},
		// {
		// 	name: "invalid email format",
		// 	input: services.RegisterIdentityInput{
		// 		Email:    "invalid-email",
		// 		Password: "password123",
		// 		Name:     "Test User",
		// 	},
		// 	wantCode: http.StatusUnprocessableEntity,
		// 	wantErr:  true,
		// },
		// {
		// 	name: "missing required fields",
		// 	input: services.RegisterIdentityInput{
		// 		Email: "test@example.com",
		// 	},
		// 	wantCode: http.StatusUnprocessableEntity,
		// 	wantErr:  true,
		// },
		// {
		// 	name: "duplicate email",
		// 	input: services.RegisterIdentityInput{
		// 		Email:    "duplicate@example.com",
		// 		Password: "password123",
		// 		Name:     "Test User",
		// 	},
		// 	wantCode: http.StatusUnprocessableEntity,
		// 	wantErr:  true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.input)

			rr := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			testApp.registerIdentityHandler(rr, req)

			if rr.Code != tt.wantCode {
				t.Errorf("got status code %d, want %d", rr.Code, tt.wantCode)
			}

			var response helpers.Envelope
			_ = json.NewDecoder(rr.Body).Decode(&response)

			if tt.wantErr {
				if response["error"] == nil {
					t.Error("expected error in response, got none")
				}
			} else {
				if response["identity"] == nil {
					t.Error("expected identity in response, got none")
				}
			}
		})
	}
}

// func TestActivateIdentityHandler(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		input    services.ActivateIdentityInput
// 		wantCode int
// 		wantErr  bool
// 	}{
// 		{
// 			name: "valid activation",
// 			input: services.ActivateIdentityInput{
// 				TokenPlaintext: "valid-token",
// 			},
// 			wantCode: http.StatusOK,
// 			wantErr:  false,
// 		},
// 		{
// 			name: "invalid token",
// 			input: services.ActivateIdentityInput{
// 				TokenPlaintext: "",
// 			},
// 			wantCode: http.StatusUnprocessableEntity,
// 			wantErr:  true,
// 		},
// 		{
// 			name: "token not found",
// 			input: services.ActivateIdentityInput{
// 				TokenPlaintext: "non-existent-token",
// 			},
// 			wantCode: http.StatusUnprocessableEntity,
// 			wantErr:  true,
// 		},
// 		{
// 			name: "already activated",
// 			input: services.ActivateIdentityInput{
// 				TokenPlaintext: "already-activated-token",
// 			},
// 			wantCode: http.StatusConflict,
// 			wantErr:  true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			jsonData, _ := json.Marshal(tt.input)

// 			rr := httptest.NewRecorder()
// 			req, _ := http.NewRequest(http.MethodPut, "/v1/users/activate", bytes.NewBuffer(jsonData))
// 			req.Header.Set("Content-Type", "application/json")

// 			testApp.activateIdentityHandler(rr, req)

// 			if rr.Code != tt.wantCode {
// 				t.Errorf("got status code %d, want %d", rr.Code, tt.wantCode)
// 			}

// 			var response helpers.Envelope
// 			_ = json.NewDecoder(rr.Body).Decode(&response)

// 			if tt.wantErr {
// 				if response["error"] == nil {
// 					t.Error("expected error in response, got none")
// 				}
// 			} else {
// 				if response["identity"] == nil {
// 					t.Error("expected identity in response, got none")
// 				}

// 				// Verify the identity is activated when successful
// 				if !tt.wantErr {
// 					identity, ok := response["identity"].(map[string]interface{})
// 					if !ok {
// 						t.Error("could not assert identity type")
// 						return
// 					}

// 					activated, ok := identity["activated"].(bool)
// 					if !ok {
// 						t.Error("could not assert activated field type")
// 						return
// 					}

// 					if !activated {
// 						t.Error("expected identity to be activated")
// 					}
// 				}
// 			}
// 		})
// 	}
// }
