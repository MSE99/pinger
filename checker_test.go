package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
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
			HttpReporters: []httpReportingDef{},
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
			HttpReporters: []httpReportingDef{},
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
			HttpReporters: []httpReportingDef{
				{
					Url:  alerterServer.URL + "/",
					Body: struct{}{},
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
			HttpReporters: []httpReportingDef{
				{
					Url:  alerterServer.URL + "/",
					Body: struct{}{},
				},
				{
					Url:  badAlerterServer.URL + "/",
					Body: struct{}{},
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

func TestCheckOnAllEmptyDefs(t *testing.T) {
	results := checkOnAll([]appDef{})

	if !reflect.DeepEqual(results, []statusCheckResult{}) {
		t.Error("did not return an empty results slice")
	}
}

func TestCheckOnAllWithBadStatusAndAlert(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	def := appDef{
		AppName:       "Mohamed's App",
		StatusURL:     server.URL + "/",
		HttpReporters: []httpReportingDef{},
		CheckInterval: 100,
	}

	results := checkOnAll([]appDef{
		def,
	})

	if !reflect.DeepEqual(results, []statusCheckResult{{App: def.AppName, IsOK: false}}) {
		t.Error("did not return a results slice")
	}
}

func TestCheckOnAllWithGoodAndBadStatuses(t *testing.T) {
	badStatusMux := http.NewServeMux()
	badStatusMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	badStatusServer := httptest.NewServer(badStatusMux)
	defer badStatusServer.Close()

	goodStatusMux := http.NewServeMux()
	goodStatusMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	goodStatusServer := httptest.NewServer(goodStatusMux)
	defer goodStatusServer.Close()

	badStatusDef := appDef{
		AppName:       "Mohamed's App",
		StatusURL:     badStatusServer.URL + "/",
		HttpReporters: []httpReportingDef{},
		CheckInterval: 100,
	}

	goodStatusDef := appDef{
		AppName:       "Mohamed's Other app",
		StatusURL:     goodStatusServer.URL + "/",
		HttpReporters: []httpReportingDef{},
		CheckInterval: 100,
	}

	checkOnAll([]appDef{
		badStatusDef,
		goodStatusDef,
	})
}
