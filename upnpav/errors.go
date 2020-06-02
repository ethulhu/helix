package upnpav

import (
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/ethulhu/helix/soap"
)

type (
	Error struct {
		XMLName     xml.Name `xml:"urn:schemas-upnp-org:control-1-0 UPnPError"`
		Code        int      `xml:"errorCode"`
		Description string   `xml:"errorDescription"`
	}
)

var (
	ErrInvalidAction = Error{Code: 401, Description: "Invalid action"}
	ErrInvalidArgs   = Error{Code: 402, Description: "Invalid args"}
	ErrActionFailed  = Error{Code: 501, Description: "Action failed"}
)

func MaybeError(err error) error {
	var soapErr soap.Error
	if errors.As(err, &soapErr) {
		var upnpavErr Error
		if err := xml.Unmarshal([]byte(soapErr.Detail()), &upnpavErr); err == nil {
			return upnpavErr
		}
	}
	return err
}

func (e Error) Error() string {
	return fmt.Sprintf("%s (%d)", e.Description, e.Code)
}
func (e Error) FaultCode() soap.FaultCode {
	if e.Code == 501 {
		return soap.FaultServer
	}
	return soap.FaultClient
}
func (e Error) FaultString() string {
	return "UPnPError"
}
func (e Error) Detail() string {
	bytes, err := xml.Marshal(e)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

// <UPnPError xmlns="urn:schemas-upnp-org:control-1-0"><errorCode>701</errorCode><errorDescription>No such object error</errorDescription></UPnPError>
