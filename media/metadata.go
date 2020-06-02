package media

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

type (
	Metadata struct {
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

func (mc *MetadataCache) MetadataForFile(path string) (*Metadata, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.metadataByPath == nil {
		mc.metadataByPath = map[string]metadataCacheEntry{}
	}

	fi, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("could not stat: %w", err)
	}
	mtime := fi.ModTime()

	if md, ok := mc.metadataByPath[path]; ok && md.mtime == mtime {
		return md.metadata, nil
	}

	md, err := MetadataForFile(path)
	if err != nil {
		return nil, err
	}

	mc.metadataByPath[path] = metadataCacheEntry{
		metadata: md,
		mtime:    mtime,
	}
	return md, nil
}

func MetadataForFile(path string) (*Metadata, error) {
	bytes, err := exec.Command("ffprobe", append(ffprobeArgs, path)...).Output()
	if err != nil {
		return nil, fmt.Errorf("could not run ffprobe: %w", err)
	}

	var raw ffprobeOutput
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return nil, fmt.Errorf("could not unmarshal ffprobe output: %w", err)
	}

	duration, err := strconv.ParseFloat(raw.Format.DurationSeconds, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse duration %q: %w", raw.Format.DurationSeconds, err)
	}

	return &Metadata{
		Duration: time.Duration(duration) * time.Second,
		Tags:     raw.Format.Tags,
	}, nil
}