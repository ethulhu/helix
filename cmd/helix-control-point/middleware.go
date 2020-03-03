package main

import (
	"fmt"
	"net/http"
	"strings"

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

func FormValues(keysAndValues ...string) mux.MatcherFunc {
	if len(keysAndValues)%2 != 0 {
		panic("an equal number of keys and values must be provided")
	}
	return func(r *http.Request, rm *mux.RouteMatch) bool {
		for i := 0; i < len(keysAndValues); i += 2 {
			key := keysAndValues[i]
			value := keysAndValues[i+1]

			formValue := r.FormValue(key)
			if formValue == "" {
				return false
			}

			if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
				if rm.Vars == nil {
					rm.Vars = map[string]string{}
				}
				varKey := value[1 : len(value)-1]
				rm.Vars[varKey] = formValue
			} else {
				if formValue != value {
					return false
				}
			}
		}
		return true
	}
}
