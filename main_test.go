package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
)

func TestStartHttpServerAndCheckers(t *testing.T) {
	testCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go startHTTPServerAndCheckers(testCtx, flags{})

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

	t.Run("Server should support a websocket connecting and listening for events.", func(t *testing.T) {
		t.Parallel()

		ws, _, wsErr := websocket.DefaultDialer.Dial("ws://localhost:9111/ws", http.Header{})
		if wsErr != nil {
			t.Error(wsErr)
			t.FailNow()
		}
		defer ws.Close()

		var results []statusCheckResult
		readErr := ws.ReadJSON(&results)

		if readErr != nil {
			t.Error(readErr)
			t.FailNow()
		}
	})

	time.Sleep(time.Millisecond * 50)
}

func TestServerShouldShutdownWhenCtxIsCancelled(t *testing.T) {
	testCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startHTTPServerAndCheckers(testCtx, flags{})
	time.Sleep(time.Millisecond * 25)

	_, _, wsErr := websocket.DefaultDialer.Dial("ws://localhost:9111/ws", http.Header{})
	if wsErr != nil {
		t.Error(wsErr)
		t.FailNow()
	}
}

func TestServerOpts(t *testing.T) {
	if _, err := os.Stat("config.json"); errors.Is(err, nil) {
		contents, readErr := os.ReadFile("config.json")
		if readErr != nil {
			t.Error(readErr)
			t.FailNow()
		}

		t.Cleanup(func() {
			os.WriteFile("config.json", []byte(contents), 0666)
		})
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Error(err)
		t.FailNow()
	}

	t.Run("Generating default config with no config in current dir.", func(t *testing.T) {
		if remErr := os.Remove("config.json"); remErr != nil {
			t.Error(remErr)
			t.FailNow()
		}

		writeErr := os.WriteFile("config.json", []byte("KERMIT"), 0666)
		if writeErr != nil {
			t.Error(writeErr)
			t.FailNow()
		}

		startHTTPServerAndCheckers(context.Background(), flags{genConfigFlag: true})

		contents, readErr := os.ReadFile("config.json")
		if readErr != nil {
			t.Error(readErr)
			t.FailNow()
		}

		if string(contents) == "KERMIT" {
			t.Error("did not generate a default config file")
		}
	})

	t.Run("passing the --status flag should check the status of the services and exit.", func(t *testing.T) {
		if remErr := os.Remove("config.json"); remErr != nil {
			t.Error(remErr)
			t.FailNow()
		}

		if err := storeDefaultConfigIn("config.json"); err != nil {
			t.Error(err)
			t.FailNow()
		}

		startHTTPServerAndCheckers(context.Background(), flags{getStatusOnly: true})
	})
}
