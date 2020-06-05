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

	"github.com/ethulhu/helix/logger"
)

func Handle(w http.ResponseWriter, r *http.Request, handler Interface) {
	log, ctx := logger.FromContext(r.Context())

	soapAction := r.Header.Get("SOAPAction")
	if soapAction == "" {
		http.Error(w, "must set SOAPAction header", http.StatusBadRequest)
		log.Warning("missing SOAPAction header")
		return
	}

	parts := strings.Split(strings.Trim(soapAction, `"`), "#")
	if len(parts) != 2 {
		http.Error(w, fmt.Sprintf(`SOAPAction header must be of form "namespace#action", got %q`, soapAction), http.StatusBadRequest)
		log.WithField("soap.SOAPAction", soapAction).Warning("invalid SOAPAction header")
		return
	}

	namespace := parts[0]
	action := parts[1]
	log.AddField("soap.namespace", namespace)
	log.AddField("soap.action", action)

	envelope, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithError(err).Warning("could not read body of SOAP request")
		return
	}

	in, err := deserializeSOAPEnvelope(envelope)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithError(err).Warning("could not deserialize SOAP request")
		return
	}

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
		log.WithError(err).Warning("served SOAP error")
		return
	}
	log.Info("served SOAP request")
}
