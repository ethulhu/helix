// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package scpd

import (
	"encoding/xml"
	"reflect"
	"testing"
)

func TestFromAction(t *testing.T) {
	tests := []struct {
		name string
		req  interface{}
		rsp  interface{}

		want    Document
		wantErr bool
	}{
		{
			name: "GetFoo",
			req: struct {
				XMLName xml.Name `xml"GetFoo"`
			}{},
			rsp: struct {
				XMLName xml.Name `xml"GetFooResponse"`
				Foo     string   `xml:"foo" scpd:"A_ARG_TYPE_Foo,string"`
			}{},
			want: Document{
				SpecVersion: Version,
				Actions: []Action{{
					Name: "GetFoo",
					Arguments: []Argument{
						{
							Name:                 "foo",
							Direction:            Out,
							RelatedStateVariable: "A_ARG_TYPE_Foo",
						},
					},
				}},
				StateVariables: []StateVariable{
					{
						Name:     "A_ARG_TYPE_Foo",
						DataType: "string",
					},
				},
			},
		},
		{
			name: "GetFoo",
			req: struct {
				XMLName xml.Name `xml"GetFoo"`
				Foo     string   `xml:"foo" scpd:"A_ARG_TYPE_Foo,ui4"`
			}{},
			rsp: struct {
				XMLName xml.Name `xml"GetFooResponse"`
				Foo     string   `xml:"foo" scpd:"A_ARG_TYPE_Foo,ui4"`
			}{},
			want: Document{
				SpecVersion: Version,
				Actions: []Action{{
					Name: "GetFoo",
					Arguments: []Argument{
						{
							Name:                 "foo",
							Direction:            In,
							RelatedStateVariable: "A_ARG_TYPE_Foo",
						},
						{
							Name:                 "foo",
							Direction:            Out,
							RelatedStateVariable: "A_ARG_TYPE_Foo",
						},
					},
				}},
				StateVariables: []StateVariable{
					{
						Name:     "A_ARG_TYPE_Foo",
						DataType: "ui4",
					},
				},
			},
		},
		{
			name: "GetFoo",
			req: struct {
				XMLName xml.Name `xml"GetFoo"`
				Foo     string   `xml:"foo" scpd:"A_ARG_TYPE_Foo,string,foo|bar"`
			}{},
			rsp: struct {
				XMLName xml.Name `xml"GetFooResponse"`
			}{},
			want: Document{
				SpecVersion: Version,
				Actions: []Action{{
					Name: "GetFoo",
					Arguments: []Argument{
						{
							Name:                 "foo",
							Direction:            In,
							RelatedStateVariable: "A_ARG_TYPE_Foo",
						},
					},
				}},
				StateVariables: []StateVariable{
					{
						Name:     "A_ARG_TYPE_Foo",
						DataType: "string",
						AllowedValues: &AllowedValues{
							Values: []string{
								"foo",
								"bar",
							},
						},
					},
				},
			},
		},
		{
			name: "GetFoo",
			req: struct {
				XMLName xml.Name `xml"GetFoo"`
				Foo     string   `xml:"foo" scpd:"A_ARG_TYPE_Foo,ui4,min=2"`
				Bar     string   `xml:"bar" scpd:"A_ARG_TYPE_Bar,i4,min=2,max=3,step=4"`
			}{},
			rsp: struct {
				XMLName xml.Name `xml"GetFooResponse"`
			}{},
			want: Document{
				SpecVersion: Version,
				Actions: []Action{{
					Name: "GetFoo",
					Arguments: []Argument{
						{
							Name:                 "foo",
							Direction:            In,
							RelatedStateVariable: "A_ARG_TYPE_Foo",
						},
						{
							Name:                 "bar",
							Direction:            In,
							RelatedStateVariable: "A_ARG_TYPE_Bar",
						},
					},
				}},
				StateVariables: []StateVariable{
					{
						Name:     "A_ARG_TYPE_Bar",
						DataType: "i4",
						AllowedValueRange: &AllowedValueRange{
							Minimum: 2,
							Maximum: 3,
							Step:    4,
						},
					},
					{
						Name:     "A_ARG_TYPE_Foo",
						DataType: "ui4",
						AllowedValueRange: &AllowedValueRange{
							Minimum: 2,
						},
					},
				},
			},
		},
		{
			name: "GetFoo",
			req: struct {
				XMLName xml.Name `xml"GetFoo"`
				Foo     string   `xml:"foo" scpd:"A_ARG_TYPE_Foo,string"`
			}{},
			rsp: struct {
				XMLName xml.Name `xml"GetFooResponse"`
				Foo     string   `xml:"foo" scpd:"A_ARG_TYPE_Foo,ui4"`
			}{},
			wantErr: true,
		},
	}

	for i, tt := range tests {
		got, err := FromAction(tt.name, tt.req, tt.rsp)
		if !tt.wantErr && err != nil {
			t.Errorf("[%d]: got error: %v", i, err)
		}
		if tt.wantErr && err == nil {
			t.Errorf("[%d]: wanted error, got nil", i)
		}
		if !reflect.DeepEqual(got, tt.want) {
			gotBytes, _ := xml.MarshalIndent(got, "", "  ")
			wantBytes, _ := xml.MarshalIndent(tt.want, "", "  ")
			t.Errorf("[%d]: got\n\n%s\n\nwanted:\n\n%s", i, gotBytes, wantBytes)
		}
	}
}
