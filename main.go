package main

import (
	"flag"
	"fmt"
	"gosh_builder/lockfile"
	"io"
	"log"
	"os"
)

type buildOpt struct {
	target string
}

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	if err := xmain(); err != nil {
		log.Fatal(err)
	}
}

func xmain() error {
	var opt buildOpt
	flag.StringVar(&opt.target, "target", "", "target stage")
	flag.Parse()

	df, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	lf, err := lockfile.Lockfile(df)
	if err != nil {
		return err
	}
	fmt.Println(lf)
	return nil
}
