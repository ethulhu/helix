// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package media

import (
	"encoding/json"
	"fmt"
	"mime"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

type (
	Metadata struct {
		Duration time.Duration
		MIMEType string
		Tags     map[string]string
		Title    string
	}

	ffprobeOutput struct {
		Format ffprobeFormat `json:"format"`
	}
	ffprobeFormat struct {
		DurationSeconds string            `json:"duration"`
		Tags            map[string]string `json:"tags"`
	}
)

func (m Metadata) Tag(key string) string {
	key = strings.ToLower(key)
	for k, v := range m.Tags {
		if strings.ToLower(k) == key {
			return v
		}
	}
	return ""
}

var ffprobeArgs = []string{"-hide_banner", "-print_format", "json", "-show_format"}

func MetadataForPath(p string) (*Metadata, error) {
	md := &Metadata{
		MIMEType: mime.TypeByExtension(path.Ext(p)),
		Title:    strings.TrimSuffix(path.Base(p), path.Ext(p)),
	}

	bytes, err := exec.Command("ffprobe", append(ffprobeArgs, p)...).Output()
	if err != nil {
		return md, fmt.Errorf("could not run ffprobe: %w", err)
	}

	var ffprobe ffprobeOutput
	if err := json.Unmarshal(bytes, &ffprobe); err != nil {
		return md, fmt.Errorf("could not unmarshal ffprobe output: %w", err)
	}

	err = mergeFFProbeOutput(md, ffprobe)
	return md, err
}

func mergeFFProbeOutput(md *Metadata, ffprobe ffprobeOutput) error {
	if duration, err := strconv.ParseFloat(ffprobe.Format.DurationSeconds, 64); err == nil {
		md.Duration = time.Duration(duration) * time.Second
	}

	if md.Tags == nil {
		md.Tags = map[string]string{}
	}
	for k, v := range ffprobe.Format.Tags {
		switch strings.ToLower(k) {
		case "title":
			md.Title = v
		default:
			md.Tags[k] = v
		}
	}

	return nil
}
