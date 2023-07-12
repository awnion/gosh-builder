package parse

import (
	"bytes"
	"testing"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func TestParser(t *testing.T) {
	df := `
fRoM gosh://test as builder
RuN ls -la

FROM gosh://test
FROM scratch
`
	dockerfile, err := parser.Parse(bytes.NewReader([]byte(df)))

	if err != nil {
		t.Fatal(err)
	}
	for _, instruction := range dockerfile.AST.Children {
		t.Log(instruction.Value)
	}
}
