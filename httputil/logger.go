// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: CC0-1.0

package httputil

import (
	"fmt"
	"net/http"

	"github.com/ethulhu/helix/logger"
	"github.com/gorilla/mux"
)

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log, ctx := logger.FromContext(r.Context())

		log.AddField("http.client", r.RemoteAddr)
		log.AddField("http.method", r.Method)
		log.AddField("http.path", r.URL.Path)

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

		next.ServeHTTP(w, r.WithContext(ctx))

		log.Info("served HTTP request")
	})
}
