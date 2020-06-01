// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package connectionmanager

import (
	"testing"

	"github.com/ethulhu/helix/upnp/scpd"
)

func TestSCPD(t *testing.T) {
	actions := []struct {
		name     string
		req, rsp interface{}
	}{
		{getProtocolInfo, getProtocolInfoRequest{}, getProtocolInfoResponse{}},
		{prepareForConnection, prepareForConnectionRequest{}, prepareForConnectionResponse{}},
		{connectionComplete, connectionCompleteRequest{}, connectionCompleteResponse{}},
		{getCurrentConnectionIDs, getCurrentConnectionIDsRequest{}, getCurrentConnectionIDsResponse{}},
		{getCurrentConnectionInfo, getCurrentConnectionInfoRequest{}, getCurrentConnectionInfoResponse{}},
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
