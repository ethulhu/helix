// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package soap

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

type (
	envelope struct {
		XMLName  xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
		Encoding string   `xml:"encodingStyle,attr"`
		Header   header   `xml:"Header,omitempty"`
		Body     body     `xml:"Body,omitempty"`
	}

	header struct {
		XMLName  xml.Name `xml:"Header"`
		Contents []byte   `xml:",innerxml"`
	}
	body struct {
		XMLName  xml.Name `xml:"Body"`
		Fault    *fault   `xml:"Fault,omitempty"`
		Contents []byte   `xml:",innerxml"`
	}
	fault struct {
		XMLName xml.Name `xml:"Fault"`
		Code    string   `xml:"faultcode"`
		String  string   `xml:"faultstring"`
		Detail  struct {
			Contents string `xml:",innerxml"`
		} `xml:"detail"`
	}
)

// serializeSOAPEnvelope is kinda hacky because some devices don't like nested default namespaces.
func serializeSOAPEnvelope(body []byte) []byte {
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	buf.WriteString(`<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">`)
	buf.WriteString(`<s:Body>`)
	buf.Write(body)
	buf.WriteString(`</s:Body>`)
	buf.WriteString(`</s:Envelope>`)
	return buf.Bytes()
}

func deserializeSOAPEnvelope(data []byte) ([]byte, error) {
	e := envelope{}
	if err := xml.Unmarshal(data, &e); err != nil {
		return nil, fmt.Errorf("could not deserialize XML envelope: %w (%s)", err, data)
	}

	if e.Body.Fault != nil {
		return nil, &RemoteError{
			FaultCode:   FaultCode(strings.Split(e.Body.Fault.Code, ":")[1]),
			FaultString: e.Body.Fault.String,
			Detail:      strings.TrimSpace(e.Body.Fault.Detail.Contents),
		}
	}
	return e.Body.Contents, nil
}