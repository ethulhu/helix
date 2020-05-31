package scpd

import (
	"reflect"
	"testing"
)

func TestMerge(t *testing.T) {
	tests := []struct {
		docs    []Document
		want    Document
		wantErr bool
	}{
		{
			docs: nil,
			want: Document{
				SpecVersion: Version,
			},
		},
		{
			docs: []Document{
				{
					SpecVersion: Version,
					Actions: []Action{
						{
							Name: "GetFoo",
						},
						{
							Name: "GetBar",
						},
					},
					StateVariables: []StateVariable{
						{
							Name: "A_ARG_TYPE_Foo",
						},
						{
							Name: "A_ARG_TYPE_Bar",
						},
					},
				},
			},
			want: Document{
				SpecVersion: Version,
				Actions: []Action{
					{
						Name: "GetBar",
					},
					{
						Name: "GetFoo",
					},
				},
				StateVariables: []StateVariable{
					{
						Name: "A_ARG_TYPE_Bar",
					},
					{
						Name: "A_ARG_TYPE_Foo",
					},
				},
			},
		},
		{
			docs: []Document{
				{
					SpecVersion: Version,
					Actions: []Action{{
						Name: "GetBar",
					}},
					StateVariables: []StateVariable{{
						Name: "A_ARG_TYPE_Foo",
					}},
				},
				{
					SpecVersion: Version,
					Actions: []Action{{
						Name: "GetFoo",
					}},
					StateVariables: []StateVariable{{
						Name: "A_ARG_TYPE_Bar",
					}},
				},
			},
			want: Document{
				SpecVersion: Version,
				Actions: []Action{
					{
						Name: "GetBar",
					},
					{
						Name: "GetFoo",
					},
				},
				StateVariables: []StateVariable{
					{
						Name: "A_ARG_TYPE_Bar",
					},
					{
						Name: "A_ARG_TYPE_Foo",
					},
				},
			},
		},
		{
			docs: []Document{
				{
					SpecVersion: Version,
					Actions: []Action{{
						Name: "GetFoo",
						Arguments: []Argument{{
							Name: "Foo",
						}},
					}},
				},
				{
					SpecVersion: Version,
					Actions: []Action{{
						Name: "GetFoo",
						Arguments: []Argument{{
							Name: "Bar",
						}},
					}},
				},
			},
			wantErr: true,
		},
	}

	for i, tt := range tests {
		got, err := Merge(tt.docs...)
		if !tt.wantErr && err != nil {
			t.Errorf("[%d]: got error: %v", i, err)
		}
		if tt.wantErr && err == nil {
			t.Errorf("[%d]: wanted error, got nil", i)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("[%d]: got\n\n%#v\n\nwanted:\n\n%#v", i, got, tt.want)
		}
	}
}
