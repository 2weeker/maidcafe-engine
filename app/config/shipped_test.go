package config

import (
	"os"
	"testing"
)

func TestShippedConfigLoads(t *testing.T) {
	f, err := os.Open("../../config.yaml")
	if err != nil {
		t.Skip(err)
	}
	defer f.Close()
	cfg, err := parse(f, Defaults())
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.validate(); err != nil {
		t.Fatal(err)
	}
}
