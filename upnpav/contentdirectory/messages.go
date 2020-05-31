// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package contentdirectory

import (
	"encoding/xml"
	"strings"

	"github.com/ethulhu/helix/upnpav"
)

type (
	commaSeparatedStrings []string

	getSearchCapabilitiesRequest struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 GetSearchCapabilities"`
	}
	getSearchCapabilitiesResponse struct {
		XMLName      xml.Name              `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 GetSearchCapabilitiesResponse"`
		Capabilities commaSeparatedStrings `xml:"SearchCaps" scpd:"SearchCapabilities,string"`
	}

	getSortCapabilitiesRequest struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 GetSortCapabilities"`
	}
	getSortCapabilitiesResponse struct {
		XMLName      xml.Name              `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 GetSortCapabilitiesResponse"`
		Capabilities commaSeparatedStrings `xml:"SortCaps" scpd:"SortCapabilities,string"`
	}

	getSystemUpdateIDRequest struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 GetSystemUpdateID"`
	}
	getSystemUpdateIDResponse struct {
		XMLName        xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 GetSystemUpdateIDResponse"`
		SystemUpdateID uint     `xml:"Id" scpd:"SystemUpdateID,ui4"`
	}

	browseFlag    string
	browseRequest struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 Browse"`

		// Object is the ID of the Object being browsed.
		// An ObjectID value of 0 corresponds to the root object of the Content Directory.
		Object upnpav.ObjectID `xml:"ObjectID" scpd:"A_ARG_TYPE_ObjectID,string"`

		// BrowseFlag specifies whether to return data about Object or Object's children.
		BrowseFlag browseFlag `xml:"BrowseFlag" scpd:"A_ARG_TYPE_BrowseFlag,string,BrowseDirectChildren|BrowseMetadata"`

		// Filter is a comma-separated list of properties (e.g. "upnp:artist"), or "*".
		Filter commaSeparatedStrings `xml:"Filter" scpd:"A_ARG_TYPE_Filter,string"`

		// StartingIndex is a zero-based offset to enumerate children under the container specified by ObjectID.
		// Must be 0 if BrowseFlag is equal to BrowseMetadata.
		StartingIndex uint `xml:"StartingIndex" scpd:"A_ARG_TYPE_Index,ui4"`

		// Requested number of entries under the object specified by ObjectID.
		// RequestedCount =0 indicates request all entries.
		RequestedCount uint `xml:"RequestedCount" scpd:"A_ARG_TYPE_Count,ui4"`

		// SortCriteria is a comma-separated list of "signed" properties.
		// For example "+upnp:artist" means "return objects sorted ascending by artist",
		// and "+upnp:artist,-dc:date" means "return objects sorted by (ascending artist, descending date)".
		SortCriteria commaSeparatedStrings `xml:"SortCriteria" scpd:"A_ARG_TYPE_SortCriteria,string"`
	}
	browseResponse struct {
		// Result is a DIDL-Lite XML document.
		Result []byte `xml:"Result" scpd:"A_ARG_TYPE_Result,string"`

		// Number of objects returned in this result.
		// If BrowseMetadata is specified in the BrowseFlags, then NumberReturned = 1
		NumberReturned uint `xml:"NumberReturned" scpd:"A_ARG_TYPE_Count,ui4"`

		TotalMatches uint `xml:"TotalMatches" scpd:"A_ARG_TYPE_Count,ui4"`

		UpdateID uint `xml:"UpdateID" scpd:"A_ARG_TYPE_UpdateID,ui4"`
	}

	searchRequest struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 Search"`

		Container upnpav.ObjectID `xml:"ContainerID" scpd:"A_ARG_TYPE_ObjectID,string"`

		SearchCriteria string `xml:"SearchCriteria" scpd:"A_ARG_TYPE_SearchCriteria,string"`

		Filter         commaSeparatedStrings `xml:"Filter" scpd:"A_ARG_TYPE_Filter,string"`
		StartingIndex  uint                  `xml:"StartingIndex" scpd:"A_ARG_TYPE_Index,ui4"`
		RequestedCount uint                  `xml:"RequestedCount" scpd:"A_ARG_TYPE_Count,ui4"`
		SortCriteria   commaSeparatedStrings `xml:"SortCriteria" scpd:"A_ARG_TYPE_SortCriteria,string"`
	}
	searchResponse struct {
		// Result is a DIDL-Lite XML document.
		Result []byte `xml:"Result" scpd:"A_ARG_TYPE_Result,string"`

		NumberReturned uint `xml:"NumberReturned" scpd:"A_ARG_TYPE_Count,ui4"`
		TotalMatches   uint `xml:"TotalMatches"   scpd:"A_ARG_TYPE_Count,ui4"`
		UpdateID       uint `xml:"UpdateID"       scpd:"A_ARG_TYPE_UpdateID,ui4"`
	}
)

const (
	browseMetadata = browseFlag("BrowseMetadata")
	browseChildren = browseFlag("BrowseDirectChildren")
)

const (
	getSearchCapabilities = "GetSearchCapabilities"
	getSortCapabilities   = "GetSortCapabilities"
	getSystemUpdateID     = "GetSystemUpdateID" // TODO: figure out how this works.

	browse  = "Browse"
	searchA = "Search"

	createObject  = "CreateObject"
	destroyObject = "DestroyObject"
	updateObject  = "UpdateObject"

	createResource       = "CreateResource"
	deleteResource       = "DeleteResource"
	exportResource       = "ExportResource"
	getTransferProgress  = "GetTransferProgress"
	importResource       = "ImportResource"
	stopTransferResource = "StopTransferResource"
)

func (csl commaSeparatedStrings) MarshalXML(e *xml.Encoder, el xml.StartElement) error {
	s := strings.Join(csl, ",")
	return e.EncodeElement(s, el)
}

func (csl *commaSeparatedStrings) UnmarshalXML(d *xml.Decoder, el xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &el); err != nil {
		return err
	}

	*csl = strings.Split(s, ",")
	return nil
}