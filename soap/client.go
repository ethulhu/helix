package soap

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type (
	client struct {
		baseURL *url.URL
	}

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
		Payload  interface{}
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

func NewClient(baseURL *url.URL) Client {
	return &client{
		baseURL: baseURL,
	}
}

// serializeSOAPEnvelope is kinda hacky because some devices don't like nested default namespaces.
func serializeSOAPEnvelope(input interface{}) []byte {
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	buf.WriteString(`<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">`)
	buf.WriteString(`<s:Body>`)
	if input != nil {
		if err := xml.NewEncoder(&buf).Encode(input); err != nil {
			panic(fmt.Sprintf("could not serialize SOAP request: %v", err))
		}
	}
	buf.WriteString(`</s:Body>`)
	buf.WriteString(`</s:Envelope>`)
	return buf.Bytes()
}

func deserializeSOAPEnvelope(data []byte, output interface{}) error {
	e := envelope{}
	if err := xml.Unmarshal(data, &e); err != nil {
		return fmt.Errorf("could not deserialize XML envelope: %w (%s)", err, data)
	}

	if e.Body.Fault != nil {
		return &RemoteError{
			FaultCode:   FaultCode(strings.Split(e.Body.Fault.Code, ":")[1]),
			FaultString: e.Body.Fault.String,
			Detail:      strings.TrimSpace(e.Body.Fault.Detail.Contents),
		}
	}

	if output != nil {
		if err := xml.Unmarshal(e.Body.Contents, output); err != nil {
			return fmt.Errorf("could not deserialize XML payload: %w", err)
		}
	}
	return nil
}

func (c *client) Call(ctx context.Context, namespace, method string, input interface{}, output interface{}) error {
	reqBytes := serializeSOAPEnvelope(input)

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL.String(), bytes.NewReader(reqBytes))
	if err != nil {
		return fmt.Errorf("could not create POST request: %w", err)
	}
	req.Header = http.Header{
		"Accept":       {"text/xml"},
		"Content-Type": {"text/xml; charset=\"utf-8\""},
		"SOAPAction":   {fmt.Sprintf(`"%s#%s"`, namespace, method)},
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not do HTTP request: %w", err)
	}

	data, _ := ioutil.ReadAll(rsp.Body)

	// prioritize SOAP errors over regular HTTP errors.
	if err := deserializeSOAPEnvelope(data, output); err != nil {
		return err
	}

	if rsp.StatusCode != 200 {
		return fmt.Errorf("HTTP error: %s (code %d)", data, rsp.StatusCode)
	}

	return nil
}
