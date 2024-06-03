package main

import (
	"bytes"
	"encoding/json"
	"os"
)

type config struct {
	Apps []appDef `json:"apps"`
}

type appDef struct {
	AppName   string           `json:"appName"`
	StatusURL string           `json:"statusURL"`
	OnError   errorHandlingDef `json:"onError"`
}

type errorHandlingDef struct {
	AlertURL string `json:"alertURL"`
	Body     any    `json:"body"`
}

func loadConfigFromFile(filePath string) (*config, error) {
	var conf config

	buff, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	decodeErr := json.NewDecoder(bytes.NewBuffer(buff)).Decode(&conf)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return &conf, nil
}
