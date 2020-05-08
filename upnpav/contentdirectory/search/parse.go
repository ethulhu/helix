package search

import (
	"errors"
	"fmt"
)

func consumeToken(tokens []token) (token, []token) {
	if len(tokens) == 0 {
		return token{eof, ""}, nil
	}
	return tokens[0], tokens[1:]
}
func peekToken(tokens []token) token {
	if len(tokens) == 0 {
		return token{eof, ""}
	}
	return tokens[0]
}

func query(tokens []token) (Criteria, error) {
	var err error
	var expr Expr
	switch peekToken(tokens).Kind {
	case asterisk:
		_, tokens = consumeToken(tokens)
		if token, _ := consumeToken(tokens); token.Kind != eof {
			return nil, errors.New("asterisk must be alone")
		}
		return Everything{}, nil
	default:
		tokens, expr, err = expression(tokens)
		if err != nil {
			return nil, err
		}
		if token, _ := consumeToken(tokens); token.Kind != eof {
			return nil, errors.New("unexpected end of query")
		}
		return Query{expr}, nil
	}
}

func expression(tokens []token) ([]token, Expr, error) {
	var err error
	var expr Expr
	var t token
	switch peekToken(tokens).Kind {
	case openParenthesis:
		_, tokens = consumeToken(tokens)
		tokens, expr, err = expression(tokens)
		if err != nil {
			return nil, expr, err
		}
		if t, tokens = consumeToken(tokens); t.Kind != closeParenthesis {
			return nil, expr, errors.New("could not match '(' with ')'")
		}
	default:
		tokens, expr, err = binaryExpression(tokens)
		if err != nil {
			return nil, expr, err
		}
	}
	switch peekToken(tokens) {
	case token{bareString, "and"}:
		fallthrough
	case token{bareString, "or"}:
		return logicExpression(expr, tokens)
	default:
		return tokens, expr, nil
	}
}

func logicExpression(expr1 Expr, tokens []token) ([]token, Expr, error) {
	var err error
	var expr2 Expr
	var t token
	var op LogicOp
	switch t, tokens = consumeToken(tokens); t {
	case token{bareString, "and"}:
		op = And
	case token{bareString, "or"}:
		op = Or
	default:
		panic(fmt.Sprintf("got into logicExpression with token %+v", t))
	}
	tokens, expr2, err = expression(tokens)
	if err != nil {
		return nil, expr1, err
	}
	return tokens, LogicExpr{op, []Expr{expr1, expr2}}, nil
}

func binaryExpression(tokens []token) ([]token, Expr, error) {
	var t token
	t, tokens = consumeToken(tokens)
	if t.Kind != bareString {
		return nil, nil, errors.New("expected property")
	}
	property := t.Text

	var op BinaryOp
	switch t, tokens = consumeToken(tokens); t {
	case token{bareString, "exists"}:
		switch t, tokens = consumeToken(tokens); t {
		case token{bareString, "true"}:
			return tokens, ExistsExpr{property, true}, nil
		case token{bareString, "false"}:
			return tokens, ExistsExpr{property, false}, nil
		default:
			return nil, nil, errors.New("expected true or false")
		}
	case token{equal, "="}:
		op = Equal
	case token{notEqual, "!="}:
		op = NotEqual
	case token{greaterThan, ">"}:
		op = GreaterThan
	case token{greaterThanEqual, ">="}:
		op = GreaterThanEqual
	case token{lessThan, "<"}:
		op = LessThan
	case token{lessThanEqual, "<="}:
		op = LessThanEqual
	case token{bareString, "contains"}:
		op = Contains
	case token{bareString, "doesNotContain"}:
		op = DoesNotContain
	case token{bareString, "derivedfrom"}:
		op = DerivedFrom
	default:
		return nil, nil, errors.New("unexpected binary operation")
	}

	t, tokens = consumeToken(tokens)
	if t.Kind != quotedString {
		return nil, nil, errors.New("expected quoted string")
	}
	operand := t.Text

	return tokens, BinaryExpr{property, op, operand}, nil
}
