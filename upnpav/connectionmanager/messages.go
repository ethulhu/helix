package connectionmanager

import "encoding/xml"

type (
	getProtocolInfoRequest struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 GetProtocolInfo"`
	}
	getProtocolInfoResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 GetProtocolInfoResponse"`
		Source  string   `xml:"Source"`
		Sink    string   `xml:"Sink"`
	}
)
