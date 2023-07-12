package lockfile

import (
	"os"
	"testing"
)

func TestLockfile(t *testing.T) {
	// read file
	df, err := os.ReadFile("../hack/Dockerfile")
	if err != nil {
		t.Fatal(err)
	}
	Lockfile(df)
}
