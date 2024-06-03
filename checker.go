package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func startChecker(ctx context.Context, def appDef) {
	interval := def.CheckInterval

	go func() {
		log.Printf("Starting checker for %s", def.AppName)

		for {
			timer := time.NewTimer(time.Duration(interval) * time.Millisecond)
			defer timer.Stop()

			timeChan := timer.C

			select {
			case <-ctx.Done():
				log.Printf("Shutting down checker for %s", def.AppName)
				timer.Stop()
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
	if err != nil || resp.StatusCode != 200 {
		if resp.StatusCode != 200 {
			log.Printf("Gotten `%v` response status from checking on the status of %s (Reporting...)", resp.StatusCode, def.AppName)
		} else {
			log.Printf("Gotten `%v` error from checking on the status of %s (Reporting...)", err, def.AppName)
		}

		return reportError(def)
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
