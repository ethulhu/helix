// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package httpu

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

type (
	Server struct {
		Handler http.Handler
		conn    net.PacketConn
	}

	responseWriter struct {
		http.Response
		bytes.Buffer
	}
)

func (rw *responseWriter) Header() http.Header {
	return rw.Response.Header
}
func (rw *responseWriter) Write(data []byte) (int, error) {
	return rw.Buffer.Write(data)
}
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.Response.Body = ioutil.NopCloser(bytes.NewReader(rw.Buffer.Bytes()))
	rw.Response.StatusCode = statusCode
}

func (s *Server) Close() error {
	return s.conn.Close()
}

func (s *Server) Serve(conn net.PacketConn) error {
	packet := make([]byte, 2048)
	s.conn = conn
	for {
		n, addr, err := s.conn.ReadFrom(packet)
		if err != nil {
			return fmt.Errorf("could not receive packet: %w", err)
		}
		log.Printf("got packet from %v", addr)

		req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(packet[:n])))
		if err != nil {
			log.Printf("could not deserialize HTTP request from %v: %v", addr, err)
			continue
		}
		req.RemoteAddr = addr.String()

		rw := responseWriter{
			Response: http.Response{
				Header: map[string][]string{},
			},
		}
		s.Handler.ServeHTTP(&rw, req)

		var buf bytes.Buffer
		_ = rw.Response.Write(&buf)

		if _, err := s.conn.WriteTo(buf.Bytes(), addr); err != nil {
			log.Printf("could not send response to %v: %v", addr, err)
		}
	}
	return nil
}