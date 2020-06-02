// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package connectionmanager

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/ethulhu/helix/soap"
	"github.com/ethulhu/helix/upnpav"
)

type (
	client struct{ soap.Interface }
)

func NewClient(soapClient soap.Interface) Interface {
	return &client{soapClient}
}

func (c *client) call(ctx context.Context, method string, input, output interface{}) error {
	req, err := xml.Marshal(input)
	if err != nil {
		panic(fmt.Sprintf("could not marshal ConnectionManager SOAP request: %v", err))
	}

	rsp, err := c.Call(ctx, string(Version1), method, req)
	if err != nil {
		return upnpav.MaybeError(err)
	}
	return xml.Unmarshal(rsp, output)
}

func (c *client) ProtocolInfo(ctx context.Context) ([]upnpav.ProtocolInfo, []upnpav.ProtocolInfo, error) {
	req := getProtocolInfoRequest{}
	rsp := getProtocolInfoResponse{}
	if err := c.call(ctx, getProtocolInfo, req, &rsp); err != nil {
		return nil, nil, err
	}
	return rsp.Sources, rsp.Sinks, nil
}
