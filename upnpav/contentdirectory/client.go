package contentdirectory

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/ethulhu/helix/soap"
	"github.com/ethulhu/helix/upnpav"
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

func (c *client) BrowseMetadata(ctx context.Context, object upnpav.Object) (*upnpav.DIDL, error) {
	return c.browse(ctx, browseMetadata, object)
}
func (c *client) BrowseChildren(ctx context.Context, object upnpav.Object) (*upnpav.DIDL, error) {
	return c.browse(ctx, browseChildren, object)
}
func (c *client) browse(ctx context.Context, bf browseFlag, object upnpav.Object) (*upnpav.DIDL, error) {
	req := browseRequest{
		Object:     object,
		BrowseFlag: bf,
	}

	rsp := browseResponse{}
	if err := c.call(ctx, "Browse", req, &rsp); err != nil {
		return nil, fmt.Errorf("could not perform Browse request: %w", err)
	}

	didl := &upnpav.DIDL{}
	if err := xml.Unmarshal(rsp.Result, didl); err != nil {
		return nil, fmt.Errorf("could not unmarshal result: %w", err)
	}

	return didl, nil
}

func (c *client) SearchCapabilities(ctx context.Context) ([]string, error) {
	rsp := getSearchCapabilitiesResponse{}
	if err := c.call(ctx, "GetSearchCapabilities", nil, &rsp); err != nil {
		return nil, fmt.Errorf("could not get search capabilities: %w", err)
	}

	return strings.Split(rsp.Capabilities, ","), nil
}
