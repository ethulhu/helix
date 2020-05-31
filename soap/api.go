// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

// Package soap is a minimal SOAP RPC client.
//
// Skeleton Request
//
// A SOAP request, `encodingStyle` is required. For example,
//
//   <?xml version="1.0"?>
//   <soap:Envelope
//     xmlns:soap="http://www.w3.org/2003/05/soap-envelope/"
//     soap:encodingStyle="http://www.w3.org/2003/05/soap-encoding">
//     <soap:Header>
//       ...
//     </soap:Header>
//     <soap:Body>
//       ...
//       <soap:Fault>
//         <faultcode>soap:Client</faultcode>
//         <faultstring>Ewwow!</faultstring>
//         <detail><UwuError>o noes!</UwuError></detail>
//       </soap:Fault>
//     </soap:Body>
//   </soap:Envelope>
//
// A <faultcode> looks like:
//
//   - VersionMismatch:  Found an invalid namespace for the SOAP Envelope
//                       element.
//   - MustUnderstand:   An immediate child element of the Header element, with
//                       the mustUnderstand attribute set to "1", was not
//                       understood.
//   - Client:           The message was incorrectly formed or contained
//                       incorrect information.
//   - Server:           There was a problem with the server so the message
//                       could not proceed.
//
// All the elements above are declared in the default namespace for the SOAP envelope:
//
//   http://www.w3.org/2003/05/soap-envelope/
//
// and the default namespace for SOAP encoding and data types is:
//
//   http://www.w3.org/2003/05/soap-encoding
//
// Links
//
//   - https://www.w3schools.com/XML/xml_soap.asp
package soap

import (
	"context"
	"fmt"
)

type (
	// Interface is the SOAP RPC interface.
	Interface interface {
		// Call performs SOAP RPCs under a given action namespace.
		// It consumes and returns an XML fragment as []byte.
		// The user is responsible for marshalling and unmarshalling these.
		Call(ctx context.Context, namespace, action string, input []byte) ([]byte, error)
	}

	// RemoteError is an error returned by the remote system.
	RemoteError struct {
		FaultCode   FaultCode
		FaultString string
		Detail      string
	}

	// FaultCode is a SOAP error <faultcode>
	FaultCode string
)

const (
	FaultVersionMismatch = FaultCode("VersionMismatch")
	FaultMustUnderstand  = FaultCode("MustUnderstand")
	FaultClient          = FaultCode("Client")
	FaultServer          = FaultCode("Server")
)

func (e *RemoteError) Error() string {
	return fmt.Sprintf("%v error (%v): %v", e.FaultCode, e.FaultString, e.Detail)
}
