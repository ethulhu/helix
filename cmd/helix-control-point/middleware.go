package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func mustVar(r *http.Request, key string) string {
	value, ok := mux.Vars(r)[key]
	if !ok {
		panic(fmt.Sprintf("gorilla/mux did not provide parameter %q", key))
	}
	return value
}

func redirectReferer(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(w, r)
		if redirect := r.Referer(); redirect != "" {
			http.Redirect(w, r, redirect, http.StatusFound)
		}
	}
}
