// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package media

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type (
	Metadata struct {
		Title    string
		Duration time.Duration
		Tags     map[string]string
	}

	MetadataCache struct {
		mu             sync.RWMutex
		metadataByPath map[string]metadataCacheEntry
	}
	metadataCacheEntry struct {
		metadata *Metadata
		mtime    time.Time
	}

	ffprobeOutput struct {
		Format struct {
			DurationSeconds string            `json:"duration"`
			Tags            map[string]string `json:"tags"`
		} `json:"format"`
	}
)

var ffprobeArgs = []string{"-hide_banner", "-print_format", "json", "-show_format"}

func (mc *MetadataCache) MetadataForFile(p string) (*Metadata, error) {
	fi, err := os.Stat(p)
	if err != nil {
		return nil, fmt.Errorf("could not stat: %w", err)
	}

	mtime := fi.ModTime()

	mc.mu.Lock()
	if mc.metadataByPath == nil {
		mc.metadataByPath = map[string]metadataCacheEntry{}
	}
	mc.mu.Unlock()

	mc.mu.RLock()
	cacheEntry, ok := mc.metadataByPath[p]
	mc.mu.RUnlock()

	if ok && cacheEntry.mtime == mtime {
		return cacheEntry.metadata, nil
	}

	md, err := MetadataForFile(p)
	if err != nil {
		return nil, err
	}

	mc.mu.Lock()
	mc.metadataByPath[p] = metadataCacheEntry{
		metadata: md,
		mtime:    mtime,
	}
	mc.mu.Unlock()

	return md, nil
}

func (mc *MetadataCache) Warm(basePath string) {
	var wg sync.WaitGroup
	_ = filepath.Walk(basePath, func(p string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			return nil
		}
		if IsAudioOrVideo(fi.Name()) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = mc.MetadataForFile(p)
			}()
		}
		return nil
	})
	wg.Wait()
}

func MetadataForFile(p string) (*Metadata, error) {
	bytes, err := exec.Command("ffprobe", append(ffprobeArgs, p)...).Output()
	if err != nil {
		return nil, fmt.Errorf("could not run ffprobe: %w", err)
	}

	var raw ffprobeOutput
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return nil, fmt.Errorf("could not unmarshal ffprobe output: %w", err)
	}

	duration := 0
	if maybeDuration, err := strconv.ParseFloat(raw.Format.DurationSeconds, 64); err == nil {
		duration = int(maybeDuration)
	}

	title := strings.TrimSuffix(path.Base(p), path.Ext(p))
	if maybeTitle, ok := raw.Format.Tags["title"]; ok {
		title = maybeTitle
	}

	return &Metadata{
		Title:    title,
		Duration: time.Duration(duration) * time.Second,
		Tags:     raw.Format.Tags,
	}, nil
}
