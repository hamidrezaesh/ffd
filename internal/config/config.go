package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/hamidrezaesh/ffd/internal/version"
)

type Header struct {
	Key   string `toml:"key"`
	Value string `toml:"value"`
}

type Config struct {
	Headers []Header `toml:"headers"`
}

func configPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		switch runtime.GOOS {
		case "windows":
			configDir = filepath.Join(os.Getenv("APPDATA"))
		case "darwin": // macOS
			configDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
		default: // Linux and other Unix-like systems
			configDir = filepath.Join(os.Getenv("HOME"), ".config")
		}
	}

	return filepath.Join(configDir, "ffd", "config.toml")
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
		},
	}

	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err := toml.NewEncoder(file).Encode(config); err != nil {
		return "", err
	}

	return path, nil
}

func GetConfig() (Config, error) {
	path, err := ensureConfig()
	if err != nil {
		return Config{}, err
	}

	var config Config

	if _, err := toml.DecodeFile(path, &config); err != nil {
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

	return toml.NewEncoder(file).Encode(config)
}
