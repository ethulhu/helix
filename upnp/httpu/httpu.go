// Package httpu implements HTTPU (HTTP-over-UDP) for use in SSDP.
package httpu

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// serializeRequest is a hack because many devices require allcaps headers.
func serializeRequest(req *http.Request) []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%v %v HTTP/1.1\r\n", req.Method, req.URL.RequestURI())
	fmt.Fprintf(&buf, "HOST: %v\r\n", req.Host)
	req.Header.Write(&buf)
	fmt.Fprint(&buf, "\r\n")
	return buf.Bytes()
}

// Do does a HTTP-over-UDP broadcast a given number of times and waits for responses.
// It always returns any valid HTTP responses it has seen, regardless of eventual errors.
// The error slice is errors with malformed responses.
// The single error is an error with the connection itself.
func Do(req *http.Request, repeats int) ([]*http.Response, []error, error) {
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, nil, fmt.Errorf("could not listen on UDP: %w", err)
	}

	if deadline, ok := req.Context().Deadline(); ok {
		conn.SetDeadline(deadline)
	}

	addr, err := net.ResolveUDPAddr("udp", req.Host)
	if err != nil {
		return nil, nil, fmt.Errorf("could not resolve %v to host:port: %w", req.Host, err)
	}

	packet := serializeRequest(req)

	for i := 0; i < repeats; i++ {
		if _, err := conn.WriteTo(packet, addr); err != nil {
			return nil, nil, fmt.Errorf("could not send discover packet: %w", err)
		}
		time.Sleep(5 * time.Millisecond)
	}

	var rsps []*http.Response
	var errs []error
	data := make([]byte, 2048)
	for {
		n, addr, err := conn.ReadFrom(data)
		if err != nil {
			var netError net.Error
			if errors.As(err, &netError) && netError.Timeout() {
				break
			}
			return rsps, errs, err
		}
		rsp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(data[:n])), req)
		if err != nil {
			errs = append(errs, fmt.Errorf("malformed response from %v: %w", addr, err))
			continue
		}
		rsps = append(rsps, rsp)
	}
	return rsps, errs, nil
}
