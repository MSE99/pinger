package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path"
	"reflect"
	"testing"
)

func TestGetDefaultConfig(t *testing.T) {
	t.Parallel()

	gotten := defaultConfig()
	wanted := &config{Apps: []appDef{}}

	if !reflect.DeepEqual(gotten, wanted) {
		t.Error("Default config should be empty")
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	t.Parallel()

	tempDir := os.TempDir()
	wrongPath := path.Join(tempDir, "missing.config.json")
	defer os.Remove(wrongPath)

	_, err := loadOrCreateConfigAt(wrongPath)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	_, statErr := os.Stat(wrongPath)
	if statErr != nil {
		t.Error(statErr)
		t.FailNow()
	}
}

func TestLoadConfigBadJSON(t *testing.T) {
	t.Parallel()

	fileName := fmt.Sprintf("awesome_%d.config.json", rand.Int())

	tempDir := os.TempDir()
	path := path.Join(tempDir, fileName)
	_ = os.Remove(path)
	writeErr := os.WriteFile(path, []byte("Mazzy star - Fade into you"), 0666)

	if writeErr != nil {
		t.Errorf("Gotten error %v", writeErr)
		t.Fail()
	}

	_, err := loadOrCreateConfigAt(path)
	if err == nil {
		t.Error("expected an error, got nil")
	}
}

func TestLoadConfigEmpty(t *testing.T) {
	t.Parallel()

	fileName := fmt.Sprintf("awesome_%d.config.json", rand.Int())

	tempDir := os.TempDir()
	path := path.Join(tempDir, fileName)
	_ = os.Remove(path)
	writeErr := os.WriteFile(path, []byte(`{"apps": []}`), 0666)

	if writeErr != nil {
		t.Errorf("Gotten error %v", writeErr)
		t.Fail()
	}

	_, err := loadOrCreateConfigAt(path)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStoreDefaultConfigIn_HappyPath(t *testing.T) {
	t.Parallel()

	fileName := fmt.Sprintf("config_%d.json", rand.Int())
	t.Cleanup(func() {
		os.Remove(fileName)
	})

	err := storeDefaultConfigIn(fileName)
	if err != nil {
		t.Error(err)
	}
}

func TestStoreDefaultConfig_FileExists(t *testing.T) {
	t.Parallel()

	fileName := fmt.Sprintf("config_%d.json", rand.Int())
	t.Cleanup(func() {
		os.Remove(fileName)
	})

	writeErr := os.WriteFile(fileName, []byte("Hope is the best singer alive"), 0666)
	if writeErr != nil {
		t.Error(writeErr)
		t.Fail()
	}

	err := storeDefaultConfigIn(fileName)
	if err != nil {
		t.Error(err)
	}

	bytesRead, err := os.ReadFile(fileName)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	var conf config
	decodeErr := json.NewDecoder(bytes.NewBuffer(bytesRead)).Decode(&conf)
	if decodeErr != nil {
		t.Error(decodeErr)
		t.Fail()
	}

	if !reflect.DeepEqual(conf, *defaultConfig()) {
		t.Error("Default config should be empty")
	}
}
