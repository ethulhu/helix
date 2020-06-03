// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

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

	cache := media.NewMetadataCache()

	md, err := cache.MetadataForPath(*path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(md)
}
