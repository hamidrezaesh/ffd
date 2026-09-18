package formatter

import (
	"testing"

)

func TestBytes (t *testing.T) {
	var size int64 = 1100000
	output := Bytes(size)
	t.Logf("\nInput Size: %v\nOutput Size: %v\n", size, output)
}
