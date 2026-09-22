package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
