package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type statusCheckResult struct {
	App  appDef `json:"app"`
	IsOK bool   `json:"isOk"`
}

func checkOnAll(defs []appDef) []statusCheckResult {
	group := &sync.WaitGroup{}
	guard := &sync.Mutex{}
	results := []statusCheckResult{}

	for _, def := range defs {
		group.Add(1)

		appDef := def

		go func() {
			defer group.Done()
			checkErr := hit(appDef)

			guard.Lock()
			defer guard.Unlock()

			results = append(results, statusCheckResult{App: appDef, IsOK: checkErr == nil})
		}()
	}

	group.Wait()

	return results
}

func startChecker(ctx context.Context, def appDef) {
	interval := def.CheckInterval

	go func() {
		log.Printf("Starting checker for %s", def.AppName)

		for {
			timeChan := time.After(time.Duration(interval) * time.Millisecond)

			select {
			case <-ctx.Done():
				log.Printf("Shutting down checker for %s", def.AppName)
				return
			case <-timeChan:
				log.Printf("Checking on %s", def.AppName)
				err := hit(def)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}()
}

func hit(def appDef) error {
	resp, err := http.Get(def.StatusURL)

	if err != nil {
		log.Printf("Gotten `%v` error from checking on the status of %s (Reporting...)", err, def.AppName)

		reportingErr := reportError(def)
		if reportingErr == nil {
			return err
		}
		return reportingErr
	} else if resp.StatusCode != 200 {
		log.Printf("Gotten `%v` response status from checking on the status of %s (Reporting...)", resp.StatusCode, def.AppName)
		reportingErr := reportError(def)
		if reportingErr == nil {
			return fmt.Errorf("gotten a none 2xx status code: %v", resp.StatusCode)
		}
		return reportingErr
	}

	log.Printf("%s is alive and well!.", def.AppName)

	return nil
}

func reportError(def appDef) error {
	for _, handlingDef := range def.OnError {
		alertURL := handlingDef.AlertURL
		body := handlingDef.Body

		buff := bytes.NewBuffer([]byte{})
		encodeErr := json.NewEncoder(buff).Encode(body)

		if encodeErr != nil {
			log.Printf("Encode error while reporting error for %s (%v).", def.AppName, encodeErr)
			return encodeErr
		}

		_, respErr := http.Post(alertURL, "application/json", buff)
		if respErr != nil {
			log.Printf("Reporting error for %s (%v).", def.AppName, respErr)
			return respErr
		}

		log.Printf("Successfully reported error for: %s", def.AppName)
	}

	return nil
}
