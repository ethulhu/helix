package media

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
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
		mu             sync.Mutex
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
	defer mc.mu.Unlock()

	if mc.metadataByPath == nil {
		mc.metadataByPath = map[string]metadataCacheEntry{}
	}

	if md, ok := mc.metadataByPath[p]; ok && md.mtime == mtime {
		return md.metadata, nil
	}

	md, err := MetadataForFile(p)
	if err != nil {
		return nil, err
	}

	mc.metadataByPath[p] = metadataCacheEntry{
		metadata: md,
		mtime:    mtime,
	}
	return md, nil
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

	duration, err := strconv.ParseFloat(raw.Format.DurationSeconds, 64)
	if err != nil {
		log.Printf("could not parse duration %q: %w", raw.Format.DurationSeconds, err)
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
