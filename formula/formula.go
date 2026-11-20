// Package formula evaluates arithmetic expressions with exact arbitrary-
// precision numbers. It is used for data-driven numeric configurations where
// values may be huge, fractional, or expressed in terms of variables.
//
// Grammar (highest precedence last):
//
//	expr   := term (('+' | '-') term)*
//	term   := unary (('*' | '/') unary)*
//	unary  := ('+' | '-') unary | power
//	power  := primary ('^' unary)?     // right-associative
//	primary:= number | identifier | '(' expr ')'
package formula

import (
	"fmt"
	"math/big"
	"unicode"

	"github.com/neko233-com/game-ops/num"
)

type tokenKind int

const (
	tkNumber tokenKind = iota
	tkIdent
	tkOp
	tkLParen
	tkRParen
	tkEOF
)

type token struct {
	kind tokenKind
	text string
}

func tokenize(s string) ([]token, error) {
	var out []token
	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case unicode.IsSpace(rune(c)):
			i++
		case c >= '0' && c <= '9' || c == '.':
			start := i
			for i < len(s) {
				ch := s[i]
				if ch >= '0' && ch <= '9' || ch == '.' {
					i++
					continue
				}
				if ch == 'e' || ch == 'E' {
					i++
					if i < len(s) && (s[i] == '+' || s[i] == '-') {
						i++
					}
					for i < len(s) && s[i] >= '0' && s[i] <= '9' {
						i++
					}
					continue
				}
				break
			}
			out = append(out, token{tkNumber, s[start:i]})
		case c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'):
			start := i
			for i < len(s) && (s[i] == '_' || unicode.IsLetter(rune(s[i])) || unicode.IsDigit(rune(s[i]))) {
				i++
			}
			out = append(out, token{tkIdent, s[start:i]})
		case c == '+' || c == '-' || c == '*' || c == '/' || c == '^':
			out = append(out, token{tkOp, string(c)})
			i++
		case c == '(':
			out = append(out, token{tkLParen, "("})
			i++
		case c == ')':
			out = append(out, token{tkRParen, ")"})
			i++
		default:
			return nil, fmt.Errorf("formula: unexpected character %q", c)
		}
	}
	out = append(out, token{tkEOF, ""})
	return out, nil
}

type parser struct {
	tokens []token
	pos    int
	vars   map[string]num.Num
}

func (p *parser) peek() token { return p.tokens[p.pos] }
func (p *parser) next() token {
	t := p.tokens[p.pos]
	p.pos++
	return t
}

func (p *parser) parseExpr() (num.Num, error) {
	left, err := p.parseTerm()
	if err != nil {
		return num.Num{}, err
	}
	for p.peek().kind == tkOp && (p.peek().text == "+" || p.peek().text == "-") {
		op := p.next().text
		right, err := p.parseTerm()
		if err != nil {
			return num.Num{}, err
		}
		if op == "+" {
			left = left.Add(right)
		} else {
			left = left.Sub(right)
		}
	}
	return left, nil
}

func (p *parser) parseTerm() (num.Num, error) {
	left, err := p.parseUnary()
	if err != nil {
		return num.Num{}, err
	}
	for p.peek().kind == tkOp && (p.peek().text == "*" || p.peek().text == "/") {
		op := p.next().text
		right, err := p.parseUnary()
		if err != nil {
			return num.Num{}, err
		}
		if op == "*" {
			left = left.Mul(right)
		} else {
			left, err = left.Div(right)
			if err != nil {
				return num.Num{}, err
			}
		}
	}
	return left, nil
}

func (p *parser) parseUnary() (num.Num, error) {
	if p.peek().kind == tkOp && (p.peek().text == "+" || p.peek().text == "-") {
		op := p.next().text
		v, err := p.parseUnary()
		if err != nil {
			return num.Num{}, err
		}
		if op == "-" {
			return v.Neg(), nil
		}
		return v, nil
	}
	return p.parsePower()
}

func (p *parser) parsePower() (num.Num, error) {
	base, err := p.parsePrimary()
	if err != nil {
		return num.Num{}, err
	}
	if p.peek().kind == tkOp && p.peek().text == "^" {
		p.next()
		exp, err := p.parseUnary()
		if err != nil {
			return num.Num{}, err
		}
		return pow(base, exp)
	}
	return base, nil
}

func pow(base, exp num.Num) (num.Num, error) {
	r := exp.Rat()
	if !r.IsInt() {
		return num.Num{}, fmt.Errorf("formula: exponent must be an integer, got %s", exp.String())
	}
	i := new(big.Int).Set(r.Num())
	neg := i.Sign() < 0
	if neg {
		i.Neg(i)
	}
	if !i.IsUint64() || i.Uint64() > 1<<32 {
		return num.Num{}, fmt.Errorf("formula: exponent too large: %s", exp.String())
	}
	result := base.Pow(uint(i.Uint64()))
	if neg {
		return result.Inv()
	}
	return result, nil
}

func (p *parser) parsePrimary() (num.Num, error) {
	t := p.next()
	switch t.kind {
	case tkNumber:
		v, err := num.FromString(t.text)
		if err != nil {
			return num.Num{}, fmt.Errorf("formula: invalid number %q", t.text)
		}
		return v, nil
	case tkIdent:
		v, ok := p.vars[t.text]
		if !ok {
			return num.Num{}, fmt.Errorf("formula: unknown variable %q", t.text)
		}
		return v, nil
	case tkLParen:
		v, err := p.parseExpr()
		if err != nil {
			return num.Num{}, err
		}
		if p.peek().kind != tkRParen {
			return num.Num{}, fmt.Errorf("formula: missing closing parenthesis")
		}
		p.next()
		return v, nil
	default:
		return num.Num{}, fmt.Errorf("formula: unexpected token %q", t.text)
	}
}

// Eval evaluates expr using the provided variables (which may be nil).
func Eval(expr string, vars map[string]num.Num) (num.Num, error) {
	if vars == nil {
		vars = map[string]num.Num{}
	}
	tokens, err := tokenize(expr)
	if err != nil {
		return num.Num{}, err
	}
	p := &parser{tokens: tokens, vars: vars}
	result, err := p.parseExpr()
	if err != nil {
		return num.Num{}, err
	}
	if p.peek().kind != tkEOF {
		return num.Num{}, fmt.Errorf("formula: unexpected trailing token %q", p.peek().text)
	}
	return result, nil
}
