package main

import (
	"os"
	"testing"
)

func TestGenerateDefaultConfig(t *testing.T) {
	t.Parallel()

	t.Cleanup(func() {
		_ = os.Remove("config.json")
	})

	generateDefaultConfig()

	_, statErr := os.Stat("config.json")

	if statErr != nil {
		t.Error(statErr)
	}
}
