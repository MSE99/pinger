package main

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"testing"
)

func TestLoadConfigMissingFile(t *testing.T) {
	t.Parallel()

	tempDir := os.TempDir()
	wrongPath := path.Join(tempDir, "missing.config.json")
	_ = os.Remove(wrongPath)

	_, err := loadConfigFromFile(wrongPath)

	if err == nil {
		t.Error("expected an error, got nil")
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

	_, err := loadConfigFromFile(path)
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

	_, err := loadConfigFromFile(path)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
