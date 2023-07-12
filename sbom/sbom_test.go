package sbom

import (
	"os"
	"testing"
)

func TestSBOM(t *testing.T) {
	s, err := os.ReadFile("./sbom.spdx.json")
	if err != nil {
		t.Fatal(err)
	}

	result, err := ParseSBOM(s)
	if err != nil {
		t.Fatal(err)
	}

	if result.Components[0].Type != "library" {
		t.Fatal("expected library")
	}
}
