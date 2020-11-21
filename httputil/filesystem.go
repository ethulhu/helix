// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package httputil

import (
	"errors"
	"net/http"
	"os"
)

type TryFiles struct {
	http.FileSystem
}

func (fs TryFiles) Open(name string) (http.File, error) {
	file, err := fs.FileSystem.Open(name)
	if errors.Is(err, os.ErrNotExist) {
		file, err = fs.FileSystem.Open(name + ".html")
	}
	return file, err
}
