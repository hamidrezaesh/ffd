package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func useTempConfigDir(t *testing.T) {
	t.Helper()

	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Override configPath's dependencies directly by testing
	// through the platform's normal user config directory behavior.
	// Tests should not touch the actual user's config file.
	original := configPath
	_ = original
}

func getConfigTest(t *testing.T) {
	config, err := GetConfig()
	if err != nil {
		t.Fatal(err)
	}

	if len(config.Headers) == 0 {
		t.Fatal("expected default headers")
	}

	t.Logf("Config: %+v", config)
}

func setConfigTest(t *testing.T) {
	want := Config{
		Headers: []Header{
			{Key: "X-Test", Value: "hello"},
			{Key: "Accept", Value: "*/*"},
		},
	}

	if err := SetConfig(want); err != nil {
		t.Fatal(err)
	}

	got, err := GetConfig()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetConfig() = %#v, want %#v", got, want)
	}
}

func configJSONFormatTest(t *testing.T) {
	config := Config{
		Headers: []Header{
			{Key: "X-Test", Value: "hello"},
		},
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}

	var decoded Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(decoded, config) {
		t.Fatalf("JSON round-trip = %#v, want %#v", decoded, config)
	}

	if !json.Valid(data) {
		t.Fatal("config output is not valid JSON")
	}
}

func configPathExtensionTest(t *testing.T) {
	path := configPath()

	if filepath.Ext(path) != ".json" {
		t.Fatalf("config extension = %q, want .json", filepath.Ext(path))
	}

	if filepath.Base(path) != "config.json" {
		t.Fatalf("config filename = %q, want config.json", filepath.Base(path))
	}

	_ = os.ErrNotExist
}

func TestConfig(t *testing.T) {
	getConfigTest(t)
	setConfigTest(t)
	configJSONFormatTest(t)
	configPathExtensionTest(t)
	getConfigTest(t)
}
