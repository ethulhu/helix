package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ethulhu/helix/media"
)

var (
	path = flag.String("path", "", "path of file to get metadata from")
)

func main() {
	flag.Parse()

	if *path == "" {
		fmt.Fprintln(os.Stderr, "-path must not be empty")
		flag.Usage()
		os.Exit(2)
	}

	md, err := media.MetadataFromFile(*path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(md)
}
