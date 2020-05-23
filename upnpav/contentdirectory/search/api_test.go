// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package search

import (
	"reflect"
	"testing"
)

func TestCanonicalExpr(t *testing.T) {
	tests := []struct {
		expr Expr
		want Expr
	}{
		{
			expr: ExistsExpr{"a", true},
			want: ExistsExpr{"a", true},
		},
		{
			expr: BinaryExpr{"a", Contains, "b"},
			want: BinaryExpr{"a", Contains, "b"},
		},
		{
			expr: LogicExpr{
				And,
				[]Expr{
					ExistsExpr{"a", true},
				},
			},
			want: ExistsExpr{"a", true},
		},
		{
			expr: LogicExpr{
				Or,
				[]Expr{
					ExistsExpr{"a", true},
				},
			},
			want: ExistsExpr{"a", true},
		},
		{
			expr: LogicExpr{
				And,
				[]Expr{
					LogicExpr{
						And,
						[]Expr{
							ExistsExpr{"a", true},
						},
					},
				},
			},
			want: ExistsExpr{"a", true},
		},
		{
			expr: LogicExpr{
				And,
				[]Expr{
					ExistsExpr{"a", true},
					LogicExpr{
						And,
						[]Expr{
							ExistsExpr{"b", true},
							ExistsExpr{"c", true},
						},
					},
				},
			},
			want: LogicExpr{
				And,
				[]Expr{
					ExistsExpr{"a", true},
					ExistsExpr{"b", true},
					ExistsExpr{"c", true},
				},
			},
		},
		{
			expr: LogicExpr{
				And,
				[]Expr{
					LogicExpr{
						And,
						[]Expr{
							ExistsExpr{"a", true},
							ExistsExpr{"b", true},
						},
					},
					ExistsExpr{"c", true},
				},
			},
			want: LogicExpr{
				And,
				[]Expr{
					ExistsExpr{"a", true},
					ExistsExpr{"b", true},
					ExistsExpr{"c", true},
				},
			},
		},
		{
			expr: LogicExpr{
				And,
				[]Expr{
					ExistsExpr{"a", true},
					LogicExpr{
						Or,
						[]Expr{
							ExistsExpr{"b", true},
							ExistsExpr{"c", true},
						},
					},
				},
			},
			want: LogicExpr{
				And,
				[]Expr{
					ExistsExpr{"a", true},
					LogicExpr{
						Or,
						[]Expr{
							ExistsExpr{"b", true},
							ExistsExpr{"c", true},
						},
					},
				},
			},
		},
	}

	for i, tt := range tests {
		got := tt.expr.CanonicalExpr()
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("[%d]: %#v\ngot %#v\nwant %#v", i, tt.expr, got, tt.want)
		}
	}
}

func TestCriteriaString(t *testing.T) {
	tests := []struct {
		criteria Criteria
		want     string
	}{
		{
			criteria: Everything{},
			want:     `*`,
		},
		{
			criteria: Query{
				ExistsExpr{"a", true},
			},
			want: `a exists true`,
		},
		{
			criteria: Query{
				ExistsExpr{"a", false},
			},
			want: `a exists false`,
		},
		{
			criteria: Query{
				BinaryExpr{"a", Equal, "b"},
			},
			want: `a = "b"`,
		},
		{
			criteria: Query{
				BinaryExpr{"a", Equal, `b"`},
			},
			want: `a = "b\""`,
		},
		{
			criteria: Query{
				BinaryExpr{"a", Equal, `b\`},
			},
			want: `a = "b\\"`,
		},
		{
			criteria: Query{
				BinaryExpr{"a", NotEqual, "b"},
			},
			want: `a != "b"`,
		},
		{
			criteria: Query{
				BinaryExpr{"a", GreaterThan, "b"},
			},
			want: `a > "b"`,
		},
		{
			criteria: Query{
				BinaryExpr{"a", GreaterThanEqual, "b"},
			},
			want: `a >= "b"`,
		},
		{
			criteria: Query{
				BinaryExpr{"a", LessThan, "b"},
			},
			want: `a < "b"`,
		},
		{
			criteria: Query{
				BinaryExpr{"a", LessThanEqual, "b"},
			},
			want: `a <= "b"`,
		},
		{
			criteria: Query{
				BinaryExpr{"a", Contains, "b"},
			},
			want: `a contains "b"`,
		},
		{
			criteria: Query{
				BinaryExpr{"a", DoesNotContain, "b"},
			},
			want: `a doesNotContain "b"`,
		},
		{
			criteria: Query{
				BinaryExpr{"a", DerivedFrom, "b"},
			},
			want: `a derivedfrom "b"`,
		},
		{
			criteria: Query{
				LogicExpr{
					And,
					[]Expr{
						ExistsExpr{"a", true},
					},
				},
			},
			want: `a exists true`,
		},
		{
			criteria: Query{
				LogicExpr{
					And,
					[]Expr{
						ExistsExpr{"a", true},
						ExistsExpr{"b", true},
					},
				},
			},
			want: `(a exists true) and (b exists true)`,
		},
		{
			criteria: Query{
				LogicExpr{
					And,
					[]Expr{
						ExistsExpr{"a", true},
						ExistsExpr{"b", true},
						ExistsExpr{"c", true},
					},
				},
			},
			want: `(a exists true) and (b exists true) and (c exists true)`,
		},
		{
			criteria: Query{
				LogicExpr{
					Or,
					[]Expr{
						ExistsExpr{"a", true},
					},
				},
			},
			want: `a exists true`,
		},
		{
			criteria: Query{
				LogicExpr{
					Or,
					[]Expr{
						ExistsExpr{"a", true},
						ExistsExpr{"b", true},
					},
				},
			},
			want: `(a exists true) or (b exists true)`,
		},
		{
			criteria: Query{
				LogicExpr{
					Or,
					[]Expr{
						ExistsExpr{"a", true},
						ExistsExpr{"b", true},
						ExistsExpr{"c", true},
					},
				},
			},
			want: `(a exists true) or (b exists true) or (c exists true)`,
		},
		{
			criteria: Query{
				LogicExpr{
					And,
					[]Expr{
						LogicExpr{
							And,
							[]Expr{
								ExistsExpr{"a", true},
								ExistsExpr{"b", true},
							},
						},
						ExistsExpr{"c", true},
					},
				},
			},
			want: `((a exists true) and (b exists true)) and (c exists true)`,
		},
	}

	for i, tt := range tests {
		got := tt.criteria.String()
		if got != tt.want {
			t.Errorf("[%d]: %#v\ngot %v\nwant %v", i, tt.criteria, got, tt.want)
		}
	}
}
