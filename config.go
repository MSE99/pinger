package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"os"
)

type config struct {
	Apps []appDef `json:"apps"`
}

type appDef struct {
	AppName       string             `json:"appName"`
	StatusURL     string             `json:"statusURL"`
	OnError       []errorHandlingDef `json:"onError"`
	CheckInterval int                `json:"checkInterval"`
}

type errorHandlingDef struct {
	AlertURL string `json:"alertURL"`
	Body     any    `json:"body"`
}

func defaultConfig() *config {
	return &config{Apps: []appDef{}}
}

func loadOrCreateConfigAt(filePath string) (*config, error) {
	var conf config

	buff, err := os.ReadFile(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return defaultConfig(), storeDefaultConfigIn(filePath)
	} else if err != nil {
		return nil, err
	}

	decodeErr := json.NewDecoder(bytes.NewBuffer(buff)).Decode(&conf)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return &conf, nil
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

func storeDefaultConfigIn(fileName string) error {
	log.Println("Generating default config")

	defaultConfig := defaultConfig()

	buff := bytes.NewBuffer([]byte{})

	encoder := json.NewEncoder(buff)
	encoder.SetIndent("", "  ")
	encodeErr := encoder.Encode(defaultConfig)
	if encodeErr != nil {
		return encodeErr
	}

	return os.WriteFile(fileName, buff.Bytes(), 0666)
}
