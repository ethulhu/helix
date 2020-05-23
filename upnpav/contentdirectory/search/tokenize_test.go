// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package search

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		query   string
		want    []token
		wantErr error
	}{
		{
			query: `*`,
			want: []token{
				{asterisk, "*"},
			},
			wantErr: nil,
		},
		{
			query: `(object.item exists)`,
			want: []token{
				{openParenthesis, "("},
				{bareString, "object.item"},
				{bareString, "exists"},
				{closeParenthesis, ")"},
			},
			wantErr: nil,
		},
		{
			query: `openID >= "12"`,
			want: []token{
				{bareString, "openID"},
				{greaterThanEqual, ">="},
				{quotedString, "12"},
			},
			wantErr: nil,
		},
		{
			query: `contains != "foo \" bar"`,
			want: []token{
				{bareString, "contains"},
				{notEqual, "!="},
				{quotedString, `foo " bar`},
			},
			wantErr: nil,
		},
		{
			query: `"foo \\ bar"`,
			want: []token{
				{quotedString, `foo \ bar`},
			},
			wantErr: nil,
		},
		{
			query:   `"foo \a bar"`,
			want:    nil,
			wantErr: fmt.Errorf("unknown escaped character: \\a"),
		},
		{
			query: `exists 3`,
			want: []token{
				{bareString, "exists"},
			},
			wantErr: fmt.Errorf("unexpected bare character: 3"),
		},
		{
			query:   `"unterminated`,
			want:    nil,
			wantErr: fmt.Errorf("unterminated quoted string: \"unterminated"),
		},
	}

	for i, tt := range tests {
		got, err := tokenize(tt.query)
		if !reflect.DeepEqual(err, tt.wantErr) {
			t.Errorf("[%d]: tokenize(`%s`) returned error %v, want %v", i, tt.query, err, tt.wantErr)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("[%d]: tokenize(`%s`)\n\ngot  %+v\n\nwant %+v", i, tt.query, got, tt.want)
		}
	}
}
