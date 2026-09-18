package scheduler

import (
	"testing"
)

func TestSplit(t *testing.T){
	ranges, err := Split(0, 1000000000, 12, 10*1024*1024)
	if err != nil{
		t.Fatalf("\nError: %v\n", err)
	}

	t.Logf("\nranges: %v\n", ranges)
}
