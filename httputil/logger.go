// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package httputil

import (
	"fmt"
	"net/http"

	"github.com/ethulhu/helix/logger"
	"github.com/gorilla/mux"
)

type (
	responseWriter struct {
		http.ResponseWriter
		StatusCode int
	}
)

func (rw *responseWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log, ctx := logger.FromContext(r.Context())

		log.AddField("http.client", r.RemoteAddr)
		if xForwardedFor := r.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
			log.AddField("http.client", xForwardedFor)
		}
		log.AddField("http.method", r.Method)
		log.AddField("http.path", r.URL.Path)
		log.AddField("http.useragent", r.UserAgent())

		for k, v := range mux.Vars(r) {
			log.AddField(fmt.Sprintf("http.vars.%s", k), v)
		}
		for k, vs := range r.URL.Query() {
			if len(vs) == 1 {
				log.AddField(fmt.Sprintf("http.query.%s", k), vs[0])
				continue
			}
			log.AddField(fmt.Sprintf("http.query.%s", k), vs)
		}
		for k, vs := range r.Form {
			if len(vs) == 1 {
				log.AddField(fmt.Sprintf("http.form.%s", k), vs[0])
				continue
			}
			log.AddField(fmt.Sprintf("http.form.%s", k), vs)
		}

		rw := responseWriter{ResponseWriter: w, StatusCode: 200}

		next.ServeHTTP(&rw, r.WithContext(ctx))

		log.AddField("http.status", rw.StatusCode)
		log.Info("served HTTP request")
	})
}
