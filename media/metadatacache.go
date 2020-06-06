// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package media

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type (
	MetadataCache interface {
		MetadataForPath(string) (*Metadata, error)
		MetadataForPaths([]string) []*Metadata
		Warm(string)
	}
	metadataCache struct {
		mu             sync.RWMutex
		metadataByPath map[string]metadataCacheEntry
	}
	metadataCacheEntry struct {
		metadata *Metadata
		mtime    time.Time
	}

	NoOpCache struct{}
)

func NewMetadataCache() MetadataCache {
	return &metadataCache{
		metadataByPath: map[string]metadataCacheEntry{},
	}
}

func (mc *metadataCache) MetadataForPath(p string) (*Metadata, error) {
	fi, err := os.Stat(p)
	if err != nil {
		return nil, fmt.Errorf("could not stat: %w", err)
	}

	mtime := fi.ModTime()

	mc.mu.RLock()
	cacheEntry, ok := mc.metadataByPath[p]
	mc.mu.RUnlock()

	if ok && cacheEntry.mtime == mtime {
		return cacheEntry.metadata, nil
	}

	md, err := MetadataForPath(p)
	if err != nil {
		// Return something, but don't add it to the cache.
		return md, err
	}

	mc.mu.Lock()
	mc.metadataByPath[p] = metadataCacheEntry{
		metadata: md,
		mtime:    mtime,
	}
	mc.mu.Unlock()

	return md, nil
}
func (mc *metadataCache) MetadataForPaths(paths []string) []*Metadata {
	mtimes := make([]time.Time, len(paths))
	for i, p := range paths {
		fi, err := os.Stat(p)
		if err != nil {
			continue
		}
		mtimes[i] = fi.ModTime()
	}

	mds := make([]*Metadata, len(paths))

	mc.mu.Lock()
	for i, p := range paths {
		cacheEntry, ok := mc.metadataByPath[p]
		if ok && cacheEntry.mtime == mtimes[i] {
			mds[i] = cacheEntry.metadata
			continue
		}

		md, err := MetadataForPath(p)
		if err != nil {
			// We got something, but don't put it in the cache.
			mds[i] = md
			continue
		}

		mc.metadataByPath[p] = metadataCacheEntry{
			metadata: md,
			mtime:    mtimes[i],
		}
		mds[i] = md
	}
	mc.mu.Unlock()

	return mds
}

func (mc *metadataCache) Warm(basePath string) {
	var wg sync.WaitGroup
	_ = filepath.Walk(basePath, func(p string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			return nil
		}
		if IsAudioOrVideo(fi.Name()) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = mc.MetadataForPath(p)
			}()
		}
		return nil
	})
	wg.Wait()
}

func (_ NoOpCache) MetadataForPath(p string) (*Metadata, error) {
	return MetadataForPath(p)
}
func (_ NoOpCache) MetadataForPaths(paths []string) []*Metadata {
	var mds []*Metadata
	for _, p := range paths {
		md, _ := MetadataForPath(p)
		mds = append(mds, md)
	}
	return mds
}
func (_ NoOpCache) Warm(p string) {}
