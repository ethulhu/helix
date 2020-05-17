package search2

//go:generate pigeon -cache -o query.gen.go query.peg
//go:generate goimports -w query.gen.go

import (
	"fmt"
	"strings"
)

type (
	Expr interface {
		String() string
	}

	AllExpr struct{}

	LogicOp   string
	LogicExpr struct {
		Op       LogicOp
		SubExprs []Expr
	}

	ExistsExpr struct {
		Property string
		Exists   bool
	}

	BinaryOp   string
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
	DerivedFrom      = BinaryOp("derivedfrom")
)

func (a AllExpr) String() string {
	return "*"
}
func (l LogicExpr) String() string {
	var subExprs []string
	for _, e := range l.SubExprs {
		subExprs = append(subExprs, fmt.Sprintf("( %s )", e.String()))
	}

	op := fmt.Sprintf(" %s ", l.Op)

	return strings.Join(subExprs, op)
}
func (e ExistsExpr) String() string {
	return fmt.Sprintf("%v exists %v", e.Property, e.Exists)
}
func (b BinaryExpr) String() string {
	return fmt.Sprintf("%v %v %v", b.Property, b.Op, b.Operand)
}

func ParseQuery(src string) (Expr, error) {
	expr, err := Parse("query", []byte(src))
	if err != nil {
		return nil, err
	}
	return expr.(Expr), nil
}
