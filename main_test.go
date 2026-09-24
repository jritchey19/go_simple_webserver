package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"encoding/json"
)

func TestHomeHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/home", nil)
	rec := httptest.NewRecorder()

	homeHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200; got %d", rec.Code)
	}

	if rec.Body.String() != "This is a GET request to the home page\n" {
		t.Errorf("expected body 'This is a GET request to the home page'; got '%s'", rec.Body.String())
	}
}

func TestGetAllUsers(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rec := httptest.NewRecorder()

	apiHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200; got %d", rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected 'application/json'; got '%s'", rec.Header().Get("Content-Type"))
	}
}

func TestCreateUserValidation(t *testing.T) {
	tests := []struct {
		name               string
		body               string
		expectedStatus     int
		expectedHeaders    map[string]string
		expectedBody       map[string]any
	}{
		{
			name:            "valid user",
			body:            `{"name":"testA","email":"testA@test.com"}`,
			expectedStatus:  http.StatusCreated,
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
			expectedBody:    map[string]any{
				"id": int64(1),
				"name": "testA",
				"email": "testA@test.com",
			},
		},
		{
			name:            "missing email",
			body:            `{"name":"testB","email":""}`,
			expectedStatus:  http.StatusBadRequest,
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
				"X-Error-Reason": "Missing Input",
			},
			expectedBody:    map[string]any{},
		},
		{
			name:            "missing name",
			body:            `{"name":"","email":"testC@test.com"}`,
			expectedStatus:  http.StatusBadRequest,
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
				"X-Error-Reason": "Missing Input",
			},
			expectedBody:    map[string]any{},
		},
		{
			name:            "malformed json",
			body:            `{"name":"dave",}`,
			expectedStatus:   http.StatusBadRequest,
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
				"X-Error-Reason": "Invalid input",
			},
			expectedBody:    map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/data", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			apiHandler(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			for key, want := range tt.expectedHeaders {
				got := rec.Header().Get(key)
				if got != want {
					t.Errorf("header %s: expected %q, got %q", key, want, got)
				}
			}

			if tt.expectedStatus == http.StatusCreated {
				var got User
				if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}

				if got.Name != tt.expectedBody["name"] {
					t.Errorf("expected name %q, got %q", tt.expectedBody["name"], got.Name)
				}
				if got.Email != tt.expectedBody["email"] {
					t.Errorf("expected email %q, got %q", tt.expectedBody["email"], got.Email)
				}
				if got.Id == 0 {
					t.Errorf("expected a non-zero ID")
				}
			}
		})
	}
}
