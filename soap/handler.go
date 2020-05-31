// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package soap

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func Handle(w http.ResponseWriter, r *http.Request, handler Interface) {

	soapAction := r.Header.Get("SOAPAction")
	if soapAction == "" {
		http.Error(w, "must set SOAPAction header", http.StatusBadRequest)
		return
	}
	parts := strings.Split(strings.Trim(soapAction, `"`), "#")
	if len(parts) != 2 {
		http.Error(w, fmt.Sprintf(`SOAPAction header must be of form "namespace#action", got %q`, soapAction), http.StatusBadRequest)
		log.Printf("bad request")
		return
	}
	namespace := parts[0]
	action := parts[1]

	envelope, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("could not read body of SOAP request: %v", err)
		return
	}

	in, err := deserializeSOAPEnvelope(envelope)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("could not deserialize SOAP envelope: %v", err)
		return
	}

	ctx := r.Context()
	out, err := handler.Call(ctx, namespace, action, in)
	if err != nil {
		// TODO: do proper SOAP errors.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	envelope = serializeSOAPEnvelope(out)
	w.Write(envelope)
}