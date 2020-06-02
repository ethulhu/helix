package media

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

type (
	Metadata struct {
		Duration time.Duration
		Tags     map[string]string
	}

	ffprobeOutput struct {
		Format struct {
			DurationSeconds string            `json:"duration"`
			Tags            map[string]string `json:"tags"`
		} `json:"format"`
	}
)

var ffprobeArgs = []string{"-hide_banner", "-print_format", "json", "-show_format"}

func MetadataFromFile(path string) (*Metadata, error) {
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
