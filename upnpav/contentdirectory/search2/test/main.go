package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ethulhu/helix/upnpav/contentdirectory/search"
)

var (
	query = flag.String("query", "", "query to parse")
)

func main() {
	flag.Parse()

	if *query == "" {
		log.Fatal("must set -query")
	}

	criteria, err := search.Parse(*query)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(criteria)
}
