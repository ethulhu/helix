package search

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		raw       string
		want      Criteria
		wantError bool
	}{
		{
			raw:  `*`,
			want: Everything{},
		},
		{
			raw:       `* exists`,
			wantError: true,
		},
		{
			raw:  `banana exists true`,
			want: Query{ExistsExpr{"banana", true}},
		},
		{
			raw: `banana exists true and pineapple derivedfrom "apples"`,
			want: Query{
				LogicExpr{
					And,
					[]Expr{
						ExistsExpr{"banana", true},
						BinaryExpr{
							"pineapple",
							DerivedFrom,
							"apples",
						},
					},
				},
			},
		},
		{
			raw: `((banana exists true) and pineapple derivedfrom "apples")`,
			want: Query{
				LogicExpr{
					And,
					[]Expr{
						ExistsExpr{"banana", true},
						BinaryExpr{
							"pineapple",
							DerivedFrom,
							"apples",
						},
					},
				},
			},
		},
		{
			raw: `((banana exists true) and (pineapple derivedfrom "apples"))`,
			want: Query{
				LogicExpr{
					And,
					[]Expr{
						ExistsExpr{"banana", true},
						BinaryExpr{
							"pineapple",
							DerivedFrom,
							"apples",
						},
					},
				},
			},
		},
		{
			raw: `((a exists true) or (b exists true)) and (c exists true)`,
			want: Query{
				LogicExpr{
					And,
					[]Expr{
						LogicExpr{
							Or,
							[]Expr{
								ExistsExpr{"a", true},
								ExistsExpr{"b", true},
							},
						},
						ExistsExpr{"c", true},
					},
				},
			},
		},
		{
			raw: `((a exists true) and (b exists true)) and (c exists true)`,
			want: Query{
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
		},
		{
			raw: `(a exists true) and ((b exists true) and (c exists true))`,
			want: Query{
				LogicExpr{
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
			},
		},
		// {
		// raw: `(a exists true) and (b exists true) and (c exists true)`,
		// want: Query{
		// LogicExpr{
		// And,
		// []Expr{
		// ExistsExpr{"a", true},
		// ExistsExpr{"b", true},
		// ExistsExpr{"c", true},
		// },
		// },
		// },
		// },
		// {
		// raw: `a exists true and b exists true and c exists true`,
		// want: Query{
		// LogicExpr{
		// And,
		// []Expr{
		// ExistsExpr{"a", true},
		// ExistsExpr{"b", true},
		// ExistsExpr{"c", true},
		// },
		// },
		// },
		// },
		// {
		// raw: `(a exists true) and (b exists true) or (c exists true) and (d exists true)`,
		// want: Query{
		// LogicExpr{
		// Or,
		// []Expr{
		// LogicExpr{
		// And,
		// []Expr{
		// ExistsExpr{"a", true},
		// ExistsExpr{"b", true},
		// },
		// },
		// LogicExpr{
		// And,
		// []Expr{
		// ExistsExpr{"c", true},
		// ExistsExpr{"d", true},
		// },
		// },
		// },
		// },
		// },
		// },
		{
			raw:       `banana exists true and`,
			wantError: true,
		},
		{
			raw:       `banana (exists) true`,
			wantError: true,
		},
		{
			raw:       `banana contains true`,
			wantError: true,
		},
		{
			raw:       `banana contains "true`,
			wantError: true,
		},
	}

	for i, tt := range tests {
		got, err := Parse(tt.raw)
		if tt.wantError && err == nil {
			t.Errorf("[%d]: want error, got nil", i)
		}
		if !tt.wantError && err != nil {
			t.Errorf("[%d]: got error: %v", i, err)
		}
		if !tt.wantError && !reflect.DeepEqual(got, tt.want) {
			t.Errorf("[%d]: Parse(`%s`)\ngot %v\nwant %v", i, tt.raw, got, tt.want)
		}
	}
}
