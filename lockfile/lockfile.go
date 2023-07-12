package lockfile

import (
	"bytes"
	"log"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func Lockfile(df []byte) (string, error) {
	dockerfile, err := parser.Parse(bytes.NewReader(df))
	if err != nil {
		log.Println(err)
		return "", err
	}

	log.Println("dockerfile", dockerfile.AST)

	for i, node := range dockerfile.AST.Children {
		if node.Value == "FROM" {
			log.Println("node", i, node)
		}
		// log.Println("node", i, node)
	}

	log.Printf("%v", dockerfile.AST.Children[0].Attributes)
	log.Printf("%v", dockerfile.AST.Children[0].Flags)
	log.Printf("%v", dockerfile.AST.Children[0].Heredocs)
	log.Printf("%v", dockerfile.AST.Children[0].Original)
	log.Printf("%v", dockerfile.AST.Children[0].PrevComment)
	log.Printf("%v", dockerfile.AST.Children[0].StartLine)
	log.Printf("%v", dockerfile.AST.Children[0].Value)

	log.Println("dump", dockerfile.AST.Dump())
	return "", nil
}

func getRepo(url string) (string, error) {
	return "", nil
}
