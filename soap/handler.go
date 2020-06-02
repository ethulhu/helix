// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package soap

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

func Handle(w http.ResponseWriter, r *http.Request, handler Interface) {
	fields := log.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
		"remote": r.RemoteAddr,
	}

	soapAction := r.Header.Get("SOAPAction")
	if soapAction == "" {
		http.Error(w, "must set SOAPAction header", http.StatusBadRequest)
		log.WithFields(fields).Warning("missing SOAPAction header")
		return
	}

	parts := strings.Split(strings.Trim(soapAction, `"`), "#")
	if len(parts) != 2 {
		http.Error(w, fmt.Sprintf(`SOAPAction header must be of form "namespace#action", got %q`, soapAction), http.StatusBadRequest)

		fields["SOAPAction"] = soapAction
		log.WithFields(fields).Warning("invalid SOAPAction header")
		return
	}

	namespace := parts[0]
	action := parts[1]
	fields["action"] = action

	envelope, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fields["error"] = err
		log.WithFields(fields).Warning("could not read body of SOAP request")
		return
	}

	in, err := deserializeSOAPEnvelope(envelope)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fields["error"] = err
		log.WithFields(fields).Warning("could not deserialize SOAP envelope")
		return
	}

	ctx := r.Context()
	out, err := handler.Call(ctx, namespace, action, in)

	var rErr Error
	if err != nil && errors.As(err, &rErr) && rErr.FaultCode() != FaultServer {
		http.Error(w, "", http.StatusBadRequest)
	} else if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}

	envelope = serializeSOAPEnvelope(out, err)
	w.Write(envelope)

	if err != nil {
		fields["error"] = err
		log.WithFields(fields).Warning("served SOAP error")
		return
	}
	log.WithFields(fields).Info("served SOAP request")
}
