package validator

import (
	"errors"
	"strings"
)

func Filename(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return errors.New("filename cannot be empty")
	}

	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return errors.New("filename cannot contain path separators")
	}

	if strings.Contains(name, "..") {
		return errors.New("filename cannot contain '..'")
	}

	if strings.ContainsRune(name, '\x00') {
		return errors.New("filename contains invalid character")
	}

	// Windows-invalid characters
	invalid := `<>:"|?*`
	for _, c := range invalid {
		if strings.ContainsRune(name, c) {
			return errors.New("filename contains invalid character")
		}
	}

	// Reserved Windows names
	base := strings.TrimSuffix(name, ".")

	switch strings.ToUpper(base) {
	case "CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		return errors.New("reserved filename")
	}

	return nil
}
