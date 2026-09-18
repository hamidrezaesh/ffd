package validator

import (
	"testing"
)

func TestFilename(t *testing.T) {
	test_filename := "check123..<>"
	err := Filename(test_filename)
	t.Logf("Err: %v", err)
}
