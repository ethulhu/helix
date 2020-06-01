// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package contentdirectory

import (
	"encoding/xml"

	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/xmltypes"
)

type (
	getSearchCapabilitiesRequest struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 GetSearchCapabilities"`
	}
	getSearchCapabilitiesResponse struct {
		XMLName      xml.Name                       `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 GetSearchCapabilitiesResponse"`
		Capabilities xmltypes.CommaSeparatedStrings `xml:"SearchCaps" scpd:"SearchCapabilities,string"`
	}

	getSortCapabilitiesRequest struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 GetSortCapabilities"`
	}
	getSortCapabilitiesResponse struct {
		XMLName      xml.Name                       `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 GetSortCapabilitiesResponse"`
		Capabilities xmltypes.CommaSeparatedStrings `xml:"SortCaps" scpd:"SortCapabilities,string"`
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
		Filter xmltypes.CommaSeparatedStrings `xml:"Filter" scpd:"A_ARG_TYPE_Filter,string"`

		// StartingIndex is a zero-based offset to enumerate children under the container specified by ObjectID.
		// Must be 0 if BrowseFlag is equal to BrowseMetadata.
		StartingIndex uint `xml:"StartingIndex" scpd:"A_ARG_TYPE_Index,ui4"`

		// Requested number of entries under the object specified by ObjectID.
		// RequestedCount =0 indicates request all entries.
		RequestedCount uint `xml:"RequestedCount" scpd:"A_ARG_TYPE_Count,ui4"`

		// SortCriteria is a comma-separated list of "signed" properties.
		// For example "+upnp:artist" means "return objects sorted ascending by artist",
		// and "+upnp:artist,-dc:date" means "return objects sorted by (ascending artist, descending date)".
		SortCriteria xmltypes.CommaSeparatedStrings `xml:"SortCriteria" scpd:"A_ARG_TYPE_SortCriteria,string"`
	}
	browseResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 BrowseResponse"`

		Result upnpav.EncodedDIDLLite `xml:"Result" scpd:"A_ARG_TYPE_Result,string"`

		// Number of objects returned in this result.
		// If BrowseMetadata is specified in the BrowseFlags, then NumberReturned = 1
		NumberReturned uint `xml:"NumberReturned" scpd:"A_ARG_TYPE_Count,ui4"`

		TotalMatches uint `xml:"TotalMatches" scpd:"A_ARG_TYPE_Count,ui4"`
		UpdateID     uint `xml:"UpdateID"     scpd:"A_ARG_TYPE_UpdateID,ui4"`
	}

	searchRequest struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 Search"`

		Container upnpav.ObjectID `xml:"ContainerID" scpd:"A_ARG_TYPE_ObjectID,string"`

		SearchCriteria string `xml:"SearchCriteria" scpd:"A_ARG_TYPE_SearchCriteria,string"`

		Filter         xmltypes.CommaSeparatedStrings `xml:"Filter"         scpd:"A_ARG_TYPE_Filter,string"`
		StartingIndex  uint                           `xml:"StartingIndex"  scpd:"A_ARG_TYPE_Index,ui4"`
		RequestedCount uint                           `xml:"RequestedCount" scpd:"A_ARG_TYPE_Count,ui4"`
		SortCriteria   xmltypes.CommaSeparatedStrings `xml:"SortCriteria"   scpd:"A_ARG_TYPE_SortCriteria,string"`
	}
	searchResponse struct {
		XMLName        xml.Name               `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 SearchResponse"`
		Result         upnpav.EncodedDIDLLite `xml:"Result" scpd:"A_ARG_TYPE_Result,string"`
		NumberReturned uint                   `xml:"NumberReturned" scpd:"A_ARG_TYPE_Count,ui4"`
		TotalMatches   uint                   `xml:"TotalMatches"   scpd:"A_ARG_TYPE_Count,ui4"`
		UpdateID       uint                   `xml:"UpdateID"       scpd:"A_ARG_TYPE_UpdateID,ui4"`
	}

	createObjectRequest struct {
		XMLName   xml.Name        `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 CreateObject"`
		Container upnpav.ObjectID `xml:"ContainerID" scpd:"A_ARG_TYPE_ObjectID,string"`
		Elements  string          `xml:"Elements"    scpd:"A_ARG_TYPE_Result,string"`
	}
	createObjectResponse struct {
		XMLName xml.Name        `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 CreateObjectResponse"`
		Object  upnpav.ObjectID `xml:"ObjectID" scpd:"A_ARG_TYPE_ObjectID,string"`
		Result  string          `xml:"Result"   scpd:"A_ARG_TYPE_Result,string"`
	}

	destroyObjectRequest struct {
		XMLName xml.Name        `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 DestroyObject"`
		Object  upnpav.ObjectID `xml:"ObjectID" scpd:"A_ARG_TYPE_ObjectID,string"`
	}
	destroyObjectResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 DestroyObjectResponse"`
	}

	updateObjectRequest struct {
		XMLName         xml.Name                       `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 UpdateObject"`
		Object          upnpav.ObjectID                `xml:"ObjectID"        scpd:"A_ARG_TYPE_ObjectID,string"`
		CurrentTagValue xmltypes.CommaSeparatedStrings `xml:"CurrentTagValue" scpd:"A_ARG_TYPE_TagValueList,string"`
		NewTagValue     xmltypes.CommaSeparatedStrings `xml:"NewTagValue"     scpd:"A_ARG_TYPE_TagValueList,string"`
	}
	updateObjectResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 UpdateObjectResponse"`
	}

	importResourceRequest struct {
		XMLName        xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 ImportResource"`
		SourceURI      string   `xml:"SourceURI"      scpd:"A_ARG_TYPE_URI,uri"`
		DestinationURI string   `xml:"DestinationURI" scpd:"A_ARG_TYPE_URI,uri"`
	}
	importResourceResponse struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 ImportResourceResponse"`
		TransferID uint     `xml:"TransferID" scpd:"A_ARG_TYPE_TransferID,ui4"`
	}

	exportResourceRequest struct {
		XMLName        xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 ExportResource"`
		SourceURI      string   `xml:"SourceURI"      scpd:"A_ARG_TYPE_URI,uri"`
		DestinationURI string   `xml:"DestinationURI" scpd:"A_ARG_TYPE_URI,uri"`
	}
	exportResourceResponse struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 ExportResourceResponse"`
		TransferID uint     `xml:"TransferID" scpd:"A_ARG_TYPE_TransferID,ui4"`
	}

	stopTransferResourceRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 StopTransferResource"`
		TransferID uint     `xml:"TransferID" scpd:"A_ARG_TYPE_TransferID,ui4"`
	}
	stopTransferResourceResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 StopTransferResourceResponse"`
	}

	getTransferProgressRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 GetTransferProgress"`
		TransferID uint     `xml:"TransferID" scpd:"A_ARG_TYPE_TransferID,ui4"`
	}
	getTransferProgressResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 GetTransferProgressResponse"`
		Status  string   `xml:"TransferStatus" scpd:"A_ARG_TYPE_TransferStatus,string"`
		Length  string   `xml:"TransferLength" scpd:"A_ARG_TYPE_TransferLength,string"`
		Total   string   `xml:"TransferTotal"  scpd:"A_ARG_TYPE_TransferTotal,string"`
	}

	deleteResourceRequest struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 DeleteResource"`
		URI     string   `xml:"ResourceURI" scpd:"A_ARG_TYPE_URI,uri"`
	}
	deleteResourceResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 DeleteResourceResponse"`
	}

	createReferenceRequest struct {
		XMLName   xml.Name        `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 CreateReference"`
		Object    upnpav.ObjectID `xml:"ObjectID"    scpd:"A_ARG_TYPE_ObjectID,string"`
		Container upnpav.ObjectID `xml:"ContainerID" scpd:"A_ARG_TYPE_ObjectID,string"`
	}
	createReferenceResponse struct {
		XMLName xml.Name        `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 CreateReferenceResponse"`
		Object  upnpav.ObjectID `xml:"NewID" scpd:"A_ARG_TYPE_ObjectID,string"`
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

	deleteResource       = "DeleteResource"
	exportResource       = "ExportResource"
	importResource       = "ImportResource"
	stopTransferResource = "StopTransferResource"
	getTransferProgress  = "GetTransferProgress"

	createReference = "CreateReference"
)
