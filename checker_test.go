package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckerStatusOK(t *testing.T) {
	hit := false

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		hit = true
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startChecker(
		ctx,
		appDef{
			AppName:       "Mohamed's App",
			StatusURL:     server.URL + "/",
			OnError:       []errorHandlingDef{},
			CheckInterval: 10,
		},
	)
	time.Sleep(time.Millisecond * 80)

	if !hit {
		t.Error("checker did not send http request")
	}
}

func TestCheckerWithBadStatusAndNoAlerters(t *testing.T) {
	hit := false

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		hit = true
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startChecker(
		ctx,
		appDef{
			AppName:       "Mohamed's App",
			StatusURL:     server.URL + "/",
			OnError:       []errorHandlingDef{},
			CheckInterval: 10,
		},
	)
	time.Sleep(time.Millisecond * 80)

	if !hit {
		t.Error("checker did not send http request")
	}
}

func TestCheckerWithBadStatusAndAnAlerter(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	alertSent := false
	alerterMux := http.NewServeMux()
	alerterMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		alertSent = true
	})
	alerterServer := httptest.NewServer(alerterMux)
	defer alerterServer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startChecker(
		ctx,
		appDef{
			AppName:   "Mohamed's App",
			StatusURL: server.URL + "/",
			OnError: []errorHandlingDef{
				{
					AlertURL: alerterServer.URL + "/",
					Body:     struct{}{},
				},
			},
			CheckInterval: 10,
		},
	)
	time.Sleep(time.Millisecond * 150)

	if !alertSent {
		t.Error("checker did not send an alert")
	}
}

func TestCheckerWithBadStatusAndABadReporter(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	alertSent := false
	alerterMux := http.NewServeMux()
	alerterMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		alertSent = true
	})
	alerterServer := httptest.NewServer(alerterMux)
	defer alerterServer.Close()

	badAlertSent := false
	badAlerterMux := http.NewServeMux()
	badAlerterMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		badAlertSent = true
	})
	badAlerterServer := httptest.NewServer(badAlerterMux)
	defer badAlerterServer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startChecker(
		ctx,
		appDef{
			AppName:   "Mohamed's App",
			StatusURL: server.URL + "/",
			OnError: []errorHandlingDef{
				{
					AlertURL: alerterServer.URL + "/",
					Body:     struct{}{},
				},
				{
					AlertURL: badAlerterServer.URL + "/",
					Body:     struct{}{},
				},
			},
			CheckInterval: 10,
		},
	)
	time.Sleep(time.Millisecond * 150)

	if !alertSent {
		t.Error("checker did not send an alert")
	}

	if !badAlertSent {
		t.Error("checker did not send an alert")
	}
}
