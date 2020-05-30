// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

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

func udpIPv4AddrForInterface(iface *net.Interface) (*net.UDPAddr, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		addr := addr.(*net.IPNet)
		if addr.IP.To4() != nil {
			return &net.UDPAddr{IP: addr.IP}, nil
		}
	}
	return nil, errors.New("interface does not have an IPv4 address")
}

// Do does a HTTP-over-UDP broadcast a given number of times and waits for responses.
// It always returns any valid HTTP responses it has seen, regardless of eventual errors.
// The error slice is errors with malformed responses.
// The single error is an error with the connection itself.
func Do(req *http.Request, repeats int, iface *net.Interface) ([]*http.Response, []error, error) {
	var listenAddr *net.UDPAddr
	if iface != nil {
		var err error
		listenAddr, err = udpIPv4AddrForInterface(iface)
		if err != nil {
			return nil, nil, fmt.Errorf("could not find address for interface %s: %w", iface.Name, err)
		}
	}

	conn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("could not listen on UDP: %w", err)
	}
	defer conn.Close()

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
