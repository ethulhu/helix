package search

import (
	"fmt"
	"unicode"
)

type (
	tokenKind int

	token struct {
		Kind tokenKind
		Text string
	}
)

const (
	eof tokenKind = iota

	asterisk

	openParenthesis
	closeParenthesis

	// LogOp.
	and
	or

	// BinOp.
	equal
	notEqual
	lessThan
	lessThanEqual
	greaterThan
	greaterThanEqual

	quotedString
	bareString
)

func tokenize(src string) ([]token, error) {
	runes := []rune(src)

	var tokens []token

	var tmp []rune
	inQuotedString := false

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if inQuotedString {
			switch r {
			case '"':
				tokens = append(tokens, token{quotedString, string(tmp)})
				tmp = nil
				inQuotedString = false
			case '\\':
				if i+1 < len(runes) {
					switch runes[i+1] {
					case '"':
						tmp = append(tmp, '"')
					case '\\':
						tmp = append(tmp, '\\')
					default:
						return tokens, fmt.Errorf("unknown escaped character: \\%v", string(runes[i+1]))
					}
					i++
				} else {
					tmp = append(tmp, '\\')
				}
			default:
				tmp = append(tmp, r)
			}
			continue
		}

		switch {
		case unicode.IsSpace(r):
			if len(tmp) > 0 {
				tokens = append(tokens, token{bareString, string(tmp)})
				tmp = nil
			}
		case r == '*':
			tokens = append(tokens, token{asterisk, "*"})
		case r == '=':
			tokens = append(tokens, token{equal, "="})
		case r == '!':
			if i+1 < len(runes) && runes[i+1] == '=' {
				tokens = append(tokens, token{notEqual, "!="})
				i++
			} else {
				return tokens, fmt.Errorf("unexpected lone '!', should be '!='")
			}
		case r == '<':
			if i+1 < len(runes) && runes[i+1] == '=' {
				tokens = append(tokens, token{lessThanEqual, "<="})
				i++
			} else {
				tokens = append(tokens, token{lessThan, "<"})
			}
		case r == '>':
			if i+1 < len(runes) && runes[i+1] == '=' {
				tokens = append(tokens, token{greaterThanEqual, ">="})
				i++
			} else {
				tokens = append(tokens, token{greaterThan, ">"})
			}
		case r == '(':
			tokens = append(tokens, token{openParenthesis, "("})
		case r == ')':
			if len(tmp) > 0 {
				tokens = append(tokens, token{bareString, string(tmp)})
				tmp = nil
			}
			tokens = append(tokens, token{closeParenthesis, ")"})
		case r == '"':
			inQuotedString = true
		case r == '.' || r == ':' || unicode.IsLetter(r):
			tmp = append(tmp, r)
		default:
			return tokens, fmt.Errorf("unexpected bare character: %v", string(r))
		}
	}
	if len(tmp) > 0 {
		if inQuotedString {
			return tokens, fmt.Errorf("unterminated quoted string: \"%v", string(tmp))
		}
		tokens = append(tokens, token{bareString, string(tmp)})
	}

	return tokens, nil
}
