// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package soap

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type (
	client struct {
		baseURL *url.URL
	}
)

func NewClient(baseURL *url.URL) Interface {
	return &client{
		baseURL: baseURL,
	}
}

func (c *client) Call(ctx context.Context, namespace, action string, input []byte) ([]byte, error) {
	reqBytes := serializeSOAPEnvelope(input, nil)

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL.String(), bytes.NewReader(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("could not create POST request: %w", err)
	}
	req.Header = http.Header{
		"Accept":       {"text/xml"},
		"Content-Type": {"text/xml; charset=\"utf-8\""},
		"SOAPAction":   {fmt.Sprintf(`"%s#%s"`, namespace, action)},
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not do HTTP request: %w", err)
	}

	data, _ := ioutil.ReadAll(rsp.Body)

	// prioritize SOAP errors over regular HTTP errors.
	out, err := deserializeSOAPEnvelope(data)
	if err != nil {
		return out, err
	}

	if rsp.StatusCode != 200 {
		return out, fmt.Errorf("HTTP error: %s (code %d)", data, rsp.StatusCode)
	}

	return out, nil
}
