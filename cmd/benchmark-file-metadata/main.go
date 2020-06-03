// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ethulhu/helix/media"
)

var (
	basePath = flag.String("base-path", "", "base path to explore")
)

func main() {
	flag.Parse()

	if *basePath == "" {
		log.Fatal("must set -base-path")
	}

	cache := media.NewMetadataCache()

	cold := getMetadata(cache, *basePath)
	fmt.Printf("cold cache: %v\n", cold)

	warm := getMetadata(cache, *basePath)
	fmt.Printf("warm cache: %v\n", warm)
}

func getMetadata(cache media.MetadataCache, basePath string) time.Duration {
	start := time.Now()
	cache.Warm(basePath)
	return time.Since(start)
}
