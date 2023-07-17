package builder

import (
	"bytes"
	"log"
	"testing"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

func TestBuilder(t *testing.T) {
	target_sha, err := Build("../tests/run_dir", false)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("target_sha", target_sha)
}

func TestDockerfileLock(t *testing.T) {
	df := `
ARG test=1
FROM --platform=linux/amd64 scratch

FROM gosh://test as builder

RUN --mount=type=cache,target=/home/gosh/test \
    --mount=type=bind,source=/home/gosh/test,target=/home/gosh/test <<EOF
	ls -la
EOF

RUN --mount=type=bind,source=/home/gosh/test,target=/home/gosh/test ls -la

FROM scratch
RUN ls -la
`
	image_mapping := map[string]string{
		"gosh://test": "scratch",
	}
	dfl, err := dockerfileLock([]byte(df), image_mapping)

	if err != nil {
		t.Fatal(err)
	}

	// check if the dockerfile is valid
	if _, err := parser.Parse(bytes.NewReader([]byte(dfl))); err != nil {
		t.Fatal(err)
	}

	log.Println(dfl)
}
