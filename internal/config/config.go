package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hamidrezaesh/ffd/version"
)

type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Config struct {
	Headers []Header `json:"headers"`
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
