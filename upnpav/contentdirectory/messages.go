package contentdirectory

import (
	"encoding/xml"

	"github.com/ethulhu/helix/upnpav"
)

type (
	getSearchCapabilitiesResponse struct {
		XMLName      xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 GetSearchCapabilitiesResponse"`
		Capabilities string   `xml:"SearchCaps"`
	}

	browseRequest struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 Browse"`

		// Object is the ID of the Object being browsed.
		// An ObjectID value of 0 corresponds to the root object of the Content Directory.
		Object upnpav.Object `xml:"ObjectID"`

		// BrowseFlag specifies whether to return data about Object or Object's children.
		BrowseFlag browseFlag `xml:"BrowseFlag"`

		Filter string `xml:"Filter,omitempty"`

		// Starting zero based offset to enumerate children under the container specified by ObjectID.
		// Must be 0 if BrowseFlag is equal to BrowseMetadata.
		StartingIndex string `xml:"StartingIndex,omitempty"`

		// Requested number of entries under the object specified by ObjectID.
		// RequestedCount =0 indicates request all entries.
		RequestedCount int `xml:"RequestedCount,omitempty"`

		SortCriteria string `xml:"SortCriteria,omitempty"`
	}
	browseResponse struct {
		// Result is a DIDL-Lite XML document.
		Result []byte `xml:"Result"`

		// Number of objects returned in this result.
		// If BrowseMetadata is specified in the BrowseFlags, then NumberReturned = 1
		NumberReturned int `xml:"NumberReturned"`

		TotalMatches int `xml:"TotalMatches"`

		UpdateID string `xml:"UpdateID"`

		// Contents is the contents of the response, for debugging use.
		Contents []byte `xml:",innerxml"`
	}

	searchRequest struct {
		ID             upnpav.Object `xml:"ObjectID"`
		BrowseFlag     browseFlag    `xml:"BrowseFlag"`
		Filter         string        `xml:"Filter,omitempty"`
		StartingIndex  string        `xml:"StartingIndex,omitempty"`
		RequestedCount int           `xml:"RequestedCount,omitempty"`
		SortCriteria   string        `xml:"SortCriteria,omitempty"`
	}
	searchResponse struct {
		Result         string `xml:"Result"`
		NumberReturned string `xml:"NumberReturned"`
		TotalMatches   string `xml:"TotalMatches"`
		UpdateID       string `xml:"UpdateID"`
	}

	browseFlag string
)

const (
	browseMetadata       = browseFlag("BrowseMetadata")
	browseDirectChildren = browseFlag("BrowseDirectChildren")
)
