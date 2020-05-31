package contentdirectory

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/contentdirectory/search"
)

type (
	SOAPHandler struct {
		Interface
	}
)

func (h SOAPHandler) Call(ctx context.Context, namespace, action string, in []byte) ([]byte, error) {
	if namespace != string(Version1) {
		return nil, fmt.Errorf("invalid namespace")
	}

	switch action {
	case getSearchCapabilities:
		return h.getSearchCapabilities(ctx, in)
	case getSortCapabilities:
		return h.getSortCapabilities(ctx, in)
	case getSystemUpdateID:
		return h.getSystemUpdateID(ctx, in)
	case browse:
		return h.browse(ctx, in)
	case searchA:
		return h.search(ctx, in)
	default:
		return nil, fmt.Errorf("not implemented")
	}
}

func (h SOAPHandler) getSearchCapabilities(ctx context.Context, in []byte) ([]byte, error) {
	req := getSearchCapabilitiesRequest{}
	if err := xml.Unmarshal(in, &req); err != nil {
		return nil, err
	}

	caps, err := h.Interface.SearchCapabilities(ctx)
	if err != nil {
		return nil, err
	}

	rsp := getSearchCapabilitiesResponse{
		Capabilities: caps,
	}
	return xml.Marshal(rsp)
}
func (h SOAPHandler) getSortCapabilities(ctx context.Context, in []byte) ([]byte, error) {
	req := getSortCapabilitiesRequest{}
	if err := xml.Unmarshal(in, &req); err != nil {
		return nil, err
	}

	caps, err := h.Interface.SortCapabilities(ctx)
	if err != nil {
		return nil, err
	}

	rsp := getSortCapabilitiesResponse{
		Capabilities: caps,
	}
	return xml.Marshal(rsp)
}
func (h SOAPHandler) getSystemUpdateID(ctx context.Context, in []byte) ([]byte, error) {
	req := getSystemUpdateIDRequest{}
	if err := xml.Unmarshal(in, &req); err != nil {
		return nil, err
	}

	id, err := h.Interface.SystemUpdateID(ctx)
	if err != nil {
		return nil, err
	}

	rsp := getSystemUpdateIDResponse{
		SystemUpdateID: id,
	}
	return xml.Marshal(rsp)
}

func (h SOAPHandler) browse(ctx context.Context, in []byte) ([]byte, error) {
	req := browseRequest{}
	if err := xml.Unmarshal(in, &req); err != nil {
		return nil, fmt.Errorf("could not unmarshal request: %w", err)
	}

	var err error
	var didllite *upnpav.DIDLLite
	switch req.BrowseFlag {
	case browseMetadata:
		didllite, err = h.Interface.BrowseMetadata(ctx, req.Object)
	case browseChildren:
		didllite, err = h.Interface.BrowseChildren(ctx, req.Object)
	default:
		return nil, fmt.Errorf("invalid BrowseFlag: %v", req.BrowseFlag)
	}
	if err != nil {
		return nil, err
	}
	rsp := browseResponse{
		Result: []byte(didllite.String()),
	}
	out, err := xml.Marshal(rsp)
	if err != nil {
		panic(fmt.Sprintf("could not marshal BrowseResponse: %v", err))
	}
	return out, nil
}

func (h SOAPHandler) search(ctx context.Context, in []byte) ([]byte, error) {
	req := searchRequest{}
	if err := xml.Unmarshal(in, &req); err != nil {
		return nil, fmt.Errorf("could not unmarshal request: %w", err)
	}

	criteria, err := search.Parse(req.SearchCriteria)
	if err != nil {
		return nil, fmt.Errorf("could not parse search query: %v", err)
	}

	didllite, err := h.Interface.Search(ctx, req.Container, criteria)
	if err != nil {
		return nil, err
	}
	rsp := searchResponse{
		Result: []byte(didllite.String()),
	}
	return xml.Marshal(rsp)
}
