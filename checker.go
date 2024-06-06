package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type statusCheckResult struct {
	App  string `json:"app"`
	IsOK bool   `json:"isOk"`
}

func checkOnAll(defs []appDef, ctx context.Context) []statusCheckResult {
	group := &sync.WaitGroup{}
	guard := &sync.Mutex{}
	results := []statusCheckResult{}

	for _, def := range defs {
		group.Add(1)

		appDef := def

		go func() {
			defer group.Done()
			checkErr := hit(ctx, appDef)

			guard.Lock()
			defer guard.Unlock()

			results = append(results, statusCheckResult{App: appDef.AppName, IsOK: checkErr == nil})
		}()
	}

	group.Wait()

	return results
}

func startChecker(ctx context.Context, def appDef) {
	interval, err := time.ParseDuration(def.CheckInterval)
	if err != nil {
		log.Panic(err)
		return
	}

	go func() {
		log.Printf("Starting checker for %s", def.AppName)

		for {
			timeChan := time.After(interval)

			select {
			case <-ctx.Done():
				log.Printf("Shutting down checker for %s", def.AppName)
				return
			case <-timeChan:
				log.Printf("Checking on %s", def.AppName)
				err := hit(ctx, def)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}()
}

func hit(ctx context.Context, def appDef) error {
	resp, err := http.Get(def.StatusURL)

	if err != nil {
		log.Printf("Gotten `%v` error from checking on the status of %s (Reporting...)", err, def.AppName)

		reportingErr := reportError(ctx, def, map[string]string{})
		if reportingErr == nil {
			return err
		}
		return reportingErr
	} else if resp.StatusCode != 200 {
		log.Printf("Gotten `%v` response status from checking on the status of %s (Reporting...)", resp.StatusCode, def.AppName)
		reportingErr := reportError(ctx, def, map[string]string{"{status}": fmt.Sprintf("%d", resp.StatusCode)})
		if reportingErr == nil {
			return fmt.Errorf("gotten a none 2xx status code: %v", resp.StatusCode)
		}
		return reportingErr
	}

	log.Printf("%s is alive and well!.", def.AppName)

	for listener := range sockets {
		select {
		case listener <- statusCheckResult{App: def.AppName, IsOK: true}:
		default:
			continue
		}
	}

	return nil
}

func reportError(ctx context.Context, def appDef, meta map[string]string) error {
	for _, handlingDef := range def.HttpReporters {
		alertURL := handlingDef.Url
		body := handlingDef.Body

		buff := bytes.NewBuffer([]byte{})
		encodeErr := json.NewEncoder(buff).Encode(body)

		if encodeErr != nil {
			log.Printf("Encode error while reporting error for %s (%v).", def.AppName, encodeErr)
			return encodeErr
		}

		stringifiedBody := buff.String()

		for key, value := range meta {
			stringifiedBody = strings.ReplaceAll(stringifiedBody, key, value)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, alertURL, strings.NewReader(stringifiedBody))
		if err != nil {
			return err
		}

		_, respErr := http.DefaultClient.Do(req)
		if respErr != nil {
			log.Printf("Reporting error for %s (%v).", def.AppName, respErr)
			return respErr
		}

		log.Printf("Successfully reported error for: %s", def.AppName)
	}

	for listener := range sockets {
		select {
		case listener <- statusCheckResult{App: def.AppName, IsOK: false}:
		default:
			continue
		}
	}

	return nil
}
