package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/ethulhu/helix/upnpav/avtransport"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
	"github.com/gorilla/mux"
)

func mustVar(r *http.Request, key string) string {
	value, ok := mux.Vars(r)[key]
	if !ok {
		panic(fmt.Sprintf("gorilla/mux did not provide parameter %q", key))
	}
	return value
}

func maybeRedirect(w http.ResponseWriter, r *http.Request) {
	if redirect := r.Form.Get("redirect"); redirect != "" {
		http.Redirect(w, r, redirect, http.StatusFound)
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

func needsDirectory(key string, f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		udn := mustVar(r, key)
		ctx := r.Context()

		devicesLock.Lock()
		device, ok := devices[udn]
		devicesLock.Unlock()

		if !ok {
			http.Error(w, fmt.Sprintf("could not find ContentDirectory %s", udn), http.StatusNotFound)
			return
		}

		soapClient, ok := device.Client(contentdirectory.Version1)
		if !ok {
			http.Error(w, fmt.Sprintf("found a device %s, but it was not an ContentDirectory", udn), http.StatusNotFound)
			return
		}
		directory := contentdirectory.NewClient(soapClient)

		newCtx := context.WithValue(ctx, "ContentDirectory", directory)
		f(w, r.WithContext(newCtx))
	}
}
func needsTransport(key string, f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		udn := mustVar(r, key)
		ctx := r.Context()

		devicesLock.Lock()
		device, ok := devices[udn]
		devicesLock.Unlock()

		if !ok {
			http.Error(w, fmt.Sprintf("could not find AVTransport %s", udn), http.StatusNotFound)
			return
		}

		soapClient, ok := device.Client(avtransport.Version1)
		if !ok {
			http.Error(w, fmt.Sprintf("found a device %s, but it was not an AVTransport", udn), http.StatusNotFound)
			return
		}
		transport := avtransport.NewClient(soapClient)

		newCtx := context.WithValue(ctx, "AVTransport", transport)
		f(w, r.WithContext(newCtx))
	}
}
