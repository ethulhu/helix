package contentdirectory

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/contentdirectory/search"
)

func TestClientAndHandler(t *testing.T) {
	fh := fakeHandler{
		searchCapabilities: []string{"foo", "bar"},
		sortCapabilities:   []string{"mew", "purr"},
		systemUpdateID:     3,

		browseMetadataObject: upnpav.ObjectID("twelve"),
		browseMetadataDIDLLite: &upnpav.DIDLLite{
			Containers: []upnpav.Container{{
				ChildCount: 3,
			}},
		},

		browseChildrenObject: upnpav.ObjectID("thirteen"),
		browseChildrenDIDLLite: &upnpav.DIDLLite{
			Items: []upnpav.Item{{
				Creator: "foo",
			}},
		},

		searchObject:   upnpav.ObjectID("fourteen"),
		searchCriteria: search.Everything{},
		searchDIDLLite: &upnpav.DIDLLite{
			Items: []upnpav.Item{{
				Creator: "foo",
			}},
		},
	}

	client := NewClient(SOAPHandler{&fh})

	searchCapabilities, err := client.SearchCapabilities(nil)
	if err != nil {
		t.Fatalf("SearchCapabilities(_) returned error: %v", err)
	}
	if !reflect.DeepEqual(searchCapabilities, fh.searchCapabilities) {
		t.Fatalf("SearchCapabilities(_) == %v, want %v", searchCapabilities, fh.searchCapabilities)
	}

	sortCapabilities, err := client.SortCapabilities(nil)
	if err != nil {
		t.Fatalf("SortCapabilities(_) returned error: %v", err)
	}
	if !reflect.DeepEqual(sortCapabilities, fh.sortCapabilities) {
		t.Fatalf("SortCapabilities(_) == %v, want %v", sortCapabilities, fh.sortCapabilities)
	}

	systemUpdateID, err := client.SystemUpdateID(nil)
	if err != nil {
		t.Fatalf("SystemUpdate(_) returned error: %v", err)
	}
	if systemUpdateID != fh.systemUpdateID {
		t.Fatalf("SystemUpdate(_) == %v, want %v", systemUpdateID, fh.systemUpdateID)
	}

	browseMetadataDIDLLite, err := client.BrowseMetadata(nil, fh.browseMetadataObject)
	if err != nil {
		t.Fatalf("BrowseMetadata(_, %q) returned error: %v", fh.browseMetadataObject, err)
	}
	if !reflect.DeepEqual(browseMetadataDIDLLite, fh.browseMetadataDIDLLite) {
		t.Fatalf("BrowseMetadata(_, %q) == %v, want %v", fh.browseMetadataObject, browseMetadataDIDLLite, fh.browseMetadataDIDLLite)
	}

	browseChildrenDIDLLite, err := client.BrowseChildren(nil, fh.browseChildrenObject)
	if err != nil {
		t.Fatalf("BrowseChildren(_, %q) returned error: %v", fh.browseChildrenObject, err)
	}
	if !reflect.DeepEqual(browseChildrenDIDLLite, fh.browseChildrenDIDLLite) {
		t.Fatalf("BrowseChildren(_, %q) == %v, want %v", fh.browseChildrenObject, browseChildrenDIDLLite, fh.browseChildrenDIDLLite)
	}

	searchDIDLLite, err := client.Search(nil, fh.searchObject, fh.searchCriteria)
	if err != nil {
		t.Fatalf("Search(_, %q, %q) returned error: %v", fh.searchObject, fh.searchCriteria, err)
	}
	if !reflect.DeepEqual(searchDIDLLite, fh.searchDIDLLite) {
		t.Fatalf("Search(_, %q, %q) == %v, want %v", fh.searchObject, fh.searchCriteria, searchDIDLLite, fh.searchDIDLLite)
	}

}

type fakeHandler struct {
	searchCapabilities []string
	sortCapabilities   []string
	systemUpdateID     uint

	browseMetadataObject   upnpav.ObjectID
	browseMetadataDIDLLite *upnpav.DIDLLite

	browseChildrenObject   upnpav.ObjectID
	browseChildrenDIDLLite *upnpav.DIDLLite

	searchObject   upnpav.ObjectID
	searchCriteria search.Criteria
	searchDIDLLite *upnpav.DIDLLite
}

func (f *fakeHandler) SearchCapabilities(_ context.Context) ([]string, error) {
	return f.searchCapabilities, nil
}
func (f *fakeHandler) SortCapabilities(_ context.Context) ([]string, error) {
	return f.sortCapabilities, nil
}
func (f *fakeHandler) SystemUpdateID(_ context.Context) (uint, error) {
	return f.systemUpdateID, nil
}
func (f *fakeHandler) BrowseMetadata(_ context.Context, id upnpav.ObjectID) (*upnpav.DIDLLite, error) {
	if id != f.browseMetadataObject {
		return nil, fmt.Errorf("id == %v", id)
	}
	return f.browseMetadataDIDLLite, nil
}
func (f *fakeHandler) BrowseChildren(_ context.Context, id upnpav.ObjectID) (*upnpav.DIDLLite, error) {
	if id != f.browseChildrenObject {
		return nil, fmt.Errorf("id == %v", id)
	}
	return f.browseChildrenDIDLLite, nil
}
func (f *fakeHandler) Search(_ context.Context, id upnpav.ObjectID, criteria search.Criteria) (*upnpav.DIDLLite, error) {
	if id != f.searchObject {
		return nil, fmt.Errorf("id == %v", id)
	}
	if !reflect.DeepEqual(criteria, f.searchCriteria) {
		return nil, fmt.Errorf("criteria == %v", criteria)
	}
	return f.searchDIDLLite, nil
}
