package config

import (
	"testing"
)

func TestGetConfig(t *testing.T) {
	config, err := GetConfig()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Config: %+v", config)
}

func TestSetConfig(t *testing.T) {
	config, err := GetConfig()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Before SetConfig: %+v", config)

	config.Headers = append(config.Headers, Header{
		Key:   "X-Test",
		Value: "hello",
	})

	if err := SetConfig(config); err != nil {
		t.Fatal(err)
	}

	updated, err := GetConfig()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("After SetConfig: %+v", updated)
}
