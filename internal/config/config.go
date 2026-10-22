package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hamidrezaesh/ffd/version"
)

const defaultMinFilesizeToTest int64 = 10 * 1024 * 1024

type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Config struct {
	Headers                  []Header `json:"headers"`
	DefaultMinFileSizeToTest int64    `json:"default_min_file_size_to_test"`
	MinFileSizeToTest        int64    `json:"min_file_size_to_test"`
}

func configPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		switch runtime.GOOS {
		case "windows":
			configDir = os.Getenv("APPDATA")
		case "darwin":
			configDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
		default:
			configDir = filepath.Join(os.Getenv("HOME"), ".config")
		}
	}

	return filepath.Join(configDir, "ffd", "config.json")
}

func ensureConfig() (string, error) {
	path := configPath()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}

	if _, err := os.Stat(path); err == nil {
		return path, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	config := Config{
		Headers: []Header{
			{
				Key:   "User-Agent",
				Value: fmt.Sprintf("ffd/%s", version.Version),
			},
			{
				Key:   "Accept",
				Value: "*/*",
			},
		},
		DefaultMinFileSizeToTest: defaultMinFilesizeToTest,
		MinFileSizeToTest:        0,
	}

	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(config); err != nil {
		return "", err
	}

	return path, nil
}

func GetConfig() (Config, error) {
	path, err := ensureConfig()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, err
	}

	updated := false

	// Update user-agent: ffd/version if ffd is upgraded.
	for i := range config.Headers {
		header := &config.Headers[i]

		if strings.EqualFold(header.Key, "User-Agent") &&
			strings.HasPrefix(header.Value, "ffd/") {
			expected := fmt.Sprintf("ffd/%s", version.Version)

			if header.Value != expected {
				header.Value = expected
				updated = true
			}
		}
	}

	if updated {
		if err := SetConfig(config); err != nil {
			return Config{}, err
		}
	}

	return config, nil
}

func SetConfig(config Config) error {
	path, err := ensureConfig()
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(config)
}
