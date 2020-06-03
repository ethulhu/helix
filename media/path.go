// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package media

import (
	"mime"
	"path"
	"strings"
)

func IsAudioOrVideo(p string) bool {
	ext := path.Ext(p)
	mimeType := mime.TypeByExtension(ext)
	return strings.HasPrefix(mimeType, "audio/") || strings.HasPrefix(mimeType, "video/")
}
