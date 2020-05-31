// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package connectionmanager

import (
	"context"

	"github.com/ethulhu/helix/soap"
	"github.com/ethulhu/helix/upnpav"
)

type (
	client struct{ soap.Interface }
)

func NewClient(soapClient soap.Interface) Interface {
	return &client{soapClient}
}

func (c *client) ProtocolInfo(ctx context.Context) ([]*upnpav.ProtocolInfo, []*upnpav.ProtocolInfo, error) {
	req := getProtocolInfoRequest{}
	rsp := getProtocolInfoResponse{}
	if err := c.Call(ctx, string(Version1), getProtocolInfo, req, &rsp); err != nil {
		return nil, nil, err
	}
	return rsp.Sources, rsp.Sinks, nil
}
