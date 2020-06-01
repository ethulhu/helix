package contentdirectory

import (
	"testing"

	"github.com/ethulhu/helix/upnp/scpd"
)

func TestSCPD(t *testing.T) {
	actions := []struct {
		name     string
		req, rsp interface{}
	}{
		{getSearchCapabilities, getSearchCapabilitiesRequest{}, getSearchCapabilitiesResponse{}},
		{getSortCapabilities, getSortCapabilitiesRequest{}, getSortCapabilitiesResponse{}},
		{getSystemUpdateID, getSystemUpdateIDRequest{}, getSystemUpdateIDResponse{}},

		{browse, browseRequest{}, browseResponse{}},
		{searchA, searchRequest{}, searchResponse{}},

		{createObject, createObjectRequest{}, createObjectResponse{}},
		{destroyObject, destroyObjectRequest{}, destroyObjectResponse{}},
		{updateObject, updateObjectRequest{}, updateObjectResponse{}},

		{deleteResource, deleteResourceRequest{}, deleteResourceResponse{}},
		{exportResource, exportResourceRequest{}, exportResourceResponse{}},
		{importResource, importResourceRequest{}, importResourceResponse{}},
		{stopTransferResource, stopTransferResourceRequest{}, stopTransferResourceResponse{}},
		{getTransferProgress, getTransferProgressRequest{}, getTransferProgressResponse{}},

		{createReference, createReferenceRequest{}, createReferenceResponse{}},
	}

	var docs []scpd.Document
	for _, action := range actions {
		doc, err := scpd.FromAction(action.name, action.req, action.rsp)
		if err != nil {
			t.Errorf("SCPD definition for action %q is broken: %v", action.name, err)
		}
		docs = append(docs, doc)
	}

	if _, err := scpd.Merge(docs...); err != nil {
		t.Errorf("could not merge SCPDs: %v", err)
	}
}
