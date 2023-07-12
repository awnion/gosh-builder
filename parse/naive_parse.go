package parse

import (
	"bytes"
	"fmt"
	"log"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func info(v ...any) {
	log.Println(v...)
}

func NaiveParse() {
	df := `#syntax=alpine
FROM --platform=linux/amd64 alpine:3.6 as builder

RUN ls -la 1
RUN ls -la 2

FROM debian

RUN ls -la 1
RUN ls -la 2
RUN <<EOF
        1. ls /test
2. ls /test2
EOF

FROM builder

RUN ls -la 1
RUN ls -la 2

FROM scratch

RUN ls -la 1
EXPOSE 8080
`

	dockerfile, err := parser.Parse(bytes.NewReader([]byte(df)))
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("dockerfile", dockerfile.AST)

	for i, node := range dockerfile.AST.Children {
		log.Println("node", i, node)
	}

	log.Printf("%v", dockerfile.AST.Children[0].Attributes)
	log.Printf("%v", dockerfile.AST.Children[0].Flags)
	log.Printf("%v", dockerfile.AST.Children[0].Heredocs)
	log.Printf("%v", dockerfile.AST.Children[0].Original)
	log.Printf("%v", dockerfile.AST.Children[0].PrevComment)
	log.Printf("%v", dockerfile.AST.Children[0].StartLine)
	log.Printf("%v", dockerfile.AST.Children[0].Value)

	log.Println("dump", dockerfile.AST.Dump())

	stages, metaArgs, err := instructions.Parse(dockerfile.AST)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("AST", dockerfile.AST.Heredocs)

	for i, meta := range metaArgs {
		log.Println("meta", i, meta)
	}

	for i, stage := range stages {
		info("stage name", i, stage.Name)
		stages[i].BaseName = stage.BaseName + "_stage_" + fmt.Sprint(i)
	}

	log.Println("stages", len(stages))
	for i, stage := range stages {
		log.Println("stage", i, stage, stage.BaseName)
	}

	for i, command := range stages[1].Commands {
		log.Println("command", i, command)
	}
}
