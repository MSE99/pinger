package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestStartHttpServerAndCheckers(t *testing.T) {
	t.Parallel()

	testCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go startHTTPServerAndCheckers(testCtx)

	time.Sleep(time.Millisecond * 50)

	t.Run("HTTP server should check on all services and respond with 200.", func(t *testing.T) {
		t.Parallel()

		resp, err := http.Get("http://localhost:9111/")
		if err != nil {
			t.Error(err)
			t.FailNow()
		} else if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d but got %d", http.StatusOK, resp.StatusCode)
			t.FailNow()
		}

		if resp.Header.Get("Content-Type") != "text/html; charset=utf-8" {
			t.Errorf("Expected content type to be %s but got %s", "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
			t.FailNow()
		}
	})

	time.Sleep(time.Millisecond * 50)
}
