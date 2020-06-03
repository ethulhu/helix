package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ethulhu/helix/media"
)

func main() {
	flag.Parse()

	cache := &media.MetadataCache{}

	cold := getMetadata(cache, flag.Args())
	fmt.Printf("cold cache: %v\n", cold)

	warm := getMetadata(cache, flag.Args())
	fmt.Printf("warm cache: %v\n", warm)
}

func getMetadata(cache *media.MetadataCache, paths []string) time.Duration {
	start := time.Now()
	var wg sync.WaitGroup
	for _, path := range flag.Args() {
		path := path
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := cache.MetadataForFile(path); err != nil {
				log.Printf("could not get metadata for %q: %v", path, err)
			}
		}()
	}
	wg.Wait()
	return time.Since(start)
}
