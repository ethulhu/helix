// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package contentdirectory

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethulhu/helix/soap"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/contentdirectory/search"
)

type (
	client struct{ soap.Client }
)

func NewClient(soapClient soap.Client) Client {
	return &client{soapClient}
}

func (c *client) call(ctx context.Context, method string, input, output interface{}) error {
	return c.Call(ctx, string(Version1), method, input, output)
}

func (c *client) BrowseMetadata(ctx context.Context, object upnpav.ObjectID) (*upnpav.DIDLLite, error) {
	return c.browse(ctx, browseMetadata, object)
}
func (c *client) BrowseChildren(ctx context.Context, object upnpav.ObjectID) (*upnpav.DIDLLite, error) {
	return c.browse(ctx, browseChildren, object)
}
func (c *client) browse(ctx context.Context, bf browseFlag, object upnpav.ObjectID) (*upnpav.DIDLLite, error) {
	req := browseRequest{
		Object:     object,
		BrowseFlag: bf,
		Filter:     "*",
	}

	rsp := browseResponse{}
	if err := c.call(ctx, "Browse", req, &rsp); err != nil {
		return nil, fmt.Errorf("could not perform Browse request: %w", err)
	}

	metadata, err := upnpav.ParseDIDLLite(string(rsp.Result))
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal result: %w", err)
	}
	return metadata, nil
}

func (c *client) SearchCapabilities(ctx context.Context) ([]string, error) {
	rsp := getSearchCapabilitiesResponse{}
	if err := c.call(ctx, "GetSearchCapabilities", nil, &rsp); err != nil {
		return nil, fmt.Errorf("could not get search capabilities: %w", err)
	}

	return strings.Split(rsp.Capabilities, ","), nil
}

func (c *client) Search(ctx context.Context, object upnpav.ObjectID, criteria search.Criteria) (*upnpav.DIDLLite, error) {
	req := searchRequest{
		Object: object,
	}
	req.SearchCriteria.Criteria = criteria

	rsp := searchResponse{}
	if err := c.call(ctx, "Search", req, &rsp); err != nil {
		return nil, fmt.Errorf("could not perform Search request: %w", err)
	}

	metadata, err := upnpav.ParseDIDLLite(string(rsp.Result))
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal result: %w", err)
	}
	return metadata, nil
}
