package tests

import (
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	app := GetTestApp(t)
	req := httptest.NewRequest("GET", "/api/health", nil)

	resp, err := app.Test(req)

	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, but got %d", resp.StatusCode)
	}

	defer resp.Body.Close()
}
