// Package search implements the DLNA query language, defined in the ContentDirectory spec.
package search

import (
	"errors"
	"fmt"
	"strings"
)

type (
	// Criteria is a `searchCrit` from the ContentDirectory specification.
	Criteria interface {
		Criteria() Criteria
		String() string
	}

	// Everything matches everything, a.k.a. "*".
	Everything struct{}

	// Query wraps some query expression.
	Query struct {
		Expr
	}

	// Expr is a query or subquery expression.
	Expr interface {
		CanonicalExpr() Expr
		String() string
	}

	// LogicOp is a logical operation, like `&&` or `||`.
	LogicOp string

	// LogicExpr is the conjunction or disjunction of one or more subquery Exprs.
	LogicExpr struct {
		Op       LogicOp
		SubExprs []Expr
	}

	// ExistsExpr matches Directory Objects that either do or don't have a given property.
	ExistsExpr struct {
		Property string
		Exists   bool
	}

	// BinaryOp is a binary operation on a property, such as `=` or `contains`.
	BinaryOp string

	// BinaryExpr matches Directory Objects where Property matches the relation to Operand by Op.
	// For example, `artist = "The Birthday Massacre"`.
	BinaryExpr struct {
		Property string
		Op       BinaryOp
		Operand  string
	}
)

const (
	And = LogicOp("and")
	Or  = LogicOp("or")

	Equal            = BinaryOp("=")
	NotEqual         = BinaryOp("!=")
	LessThan         = BinaryOp("<")
	LessThanEqual    = BinaryOp("<=")
	GreaterThan      = BinaryOp(">")
	GreaterThanEqual = BinaryOp(">=")
	Contains         = BinaryOp("contains")
	DoesNotContain   = BinaryOp("doesNotContain")

	// DerivedFrom is useful for restricting a query to only certain media kinds.
	// For example, `itemClass derivedfrom "object.item.audioItem"`.
	DerivedFrom = BinaryOp("derivedfrom")
)

func (e Everything) String() string {
	return "*"
}
func (e Everything) Criteria() Criteria {
	return e
}

func (q Query) String() string {
	return q.Expr.String()
}
func (q Query) Criteria() Criteria {
	return q
}

func (l LogicExpr) String() string {
	if len(l.SubExprs) == 1 {
		return l.SubExprs[0].String()
	}

	var subExprs []string
	for _, e := range l.SubExprs {
		subExprs = append(subExprs, fmt.Sprintf("(%s)", e.String()))
	}

	op := fmt.Sprintf(" %s ", l.Op)

	return strings.Join(subExprs, op)
}
func (l LogicExpr) CanonicalExpr() Expr {
	if len(l.SubExprs) == 1 {
		return l.SubExprs[0].CanonicalExpr()
	}

	var subExprs []Expr
	for _, expr := range l.SubExprs {
		cExpr := expr.CanonicalExpr()
		if cExpr, ok := cExpr.(LogicExpr); ok && cExpr.Op == l.Op {
			subExprs = append(subExprs, cExpr.SubExprs...)
			continue
		}
		subExprs = append(subExprs, cExpr)
	}
	return LogicExpr{l.Op, subExprs}
}

func (e ExistsExpr) String() string {
	return fmt.Sprintf("%v exists %v", e.Property, e.Exists)
}
func (e ExistsExpr) CanonicalExpr() Expr {
	return e
}

func (b BinaryExpr) String() string {
	return fmt.Sprintf("%v %v %q", b.Property, b.Op, b.Operand)
}
func (b BinaryExpr) CanonicalExpr() Expr {
	return b
}

// Parse parses a DLNA query string and returns a Criteria.
// TODO: Support operator precidence for And and Or (And binds stronger than Or).
func Parse(src string) (Criteria, error) {
	tokens, err := tokenize(src)
	if err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return nil, errors.New("empty query")
	}
	return query(tokens)
}
