// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package httputil

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// MustWriteJSON writes data to w, panicking if it cannot.
// It panics because being unable to marshal JSON is programmer error, and not recoverable.
func MustWriteJSON(w http.ResponseWriter, data interface{}) {
	blob, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Sprintf("could not marshal JSON: %v", err))
	}
	w.Write(blob)
}
