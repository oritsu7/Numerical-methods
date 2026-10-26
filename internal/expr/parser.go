package expr

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type tokKind int

const (
	tNum tokKind = iota
	tIdent
	tOp // + - * / ^
	tLParen
	tRParen
	tEOF
)

type token struct {
	kind tokKind
	text string
	num  float64
}

// splitIdent breaks a run of letters such as "xsin" into x, sin (longest
// known name first, otherwise single letters),
// Names containing digits or underscores (x1, a_2) are kept whole.
func splitIdent(s string) []string {
	if strings.ContainsAny(s, "_0123456789") {
		return []string{s}
	}
	var out []string
	for i := 0; i < len(s); {
		best := 0
		for j := len(s); j > i; j-- {
			w := s[i:j]
			if _, ok := funcs[w]; ok {
				best = j
				break
			}
			if _, ok := constants[w]; ok {
				best = j
				break
			}
		}
		if best == 0 {
			best = i + 1
		}
		out = append(out, s[i:best])
		i = best
	}
	return out
}

func lex(src string) ([]token, error) {
	var toks []token
	rs := []rune(src)
	for i := 0; i < len(rs); {
		c := rs[i]
		switch {
		case unicode.IsSpace(c):
			i++
		case unicode.IsDigit(c) || c == '.':
			j := i
			for j < len(rs) && (unicode.IsDigit(rs[j]) || rs[j] == '.') {
				j++
			}
			// optional exponent: 1e-6, 2E5 (but "2E" alone means 2*E)
			if j < len(rs) && (rs[j] == 'e' || rs[j] == 'E') {
				k := j + 1
				if k < len(rs) && (rs[k] == '+' || rs[k] == '-') {
					k++
				}
				if k < len(rs) && unicode.IsDigit(rs[k]) {
					for k < len(rs) && unicode.IsDigit(rs[k]) {
						k++
					}
					j = k
				}
			}
			v, err := strconv.ParseFloat(string(rs[i:j]), 64)
			if err != nil {
				return nil, fmt.Errorf("bad number %q", string(rs[i:j]))
			}
			toks = append(toks, token{kind: tNum, num: v, text: string(rs[i:j])})
			i = j
		case unicode.IsLetter(c) || c == '_':
			j := i
			for j < len(rs) && (unicode.IsLetter(rs[j]) || unicode.IsDigit(rs[j]) || rs[j] == '_') {
				j++
			}
			for _, part := range splitIdent(string(rs[i:j])) {
				toks = append(toks, token{kind: tIdent, text: part})
			}
			i = j
		case c == '*':
			if i+1 < len(rs) && rs[i+1] == '*' {
				toks = append(toks, token{kind: tOp, text: "^"})
				i += 2
			} else {
				toks = append(toks, token{kind: tOp, text: "*"})
				i++
			}
		case c == '+' || c == '-' || c == '/' || c == '^':
			toks = append(toks, token{kind: tOp, text: string(c)})
			i++
		case c == '(' || c == '[':
			toks = append(toks, token{kind: tLParen, text: "("})
			i++
		case c == ')' || c == ']':
			toks = append(toks, token{kind: tRParen, text: ")"})
			i++
		default:
			return nil, fmt.Errorf("unexpected character %q", string(c))
		}
	}
	toks = append(toks, token{kind: tEOF})
	return toks, nil
}

type parser struct {
	toks []token
	pos  int
}

func (p *parser) peek() token { return p.toks[p.pos] }
func (p *parser) next() token { t := p.toks[p.pos]; p.pos++; return t }

func (p *parser) isOp(ops string) bool {
	t := p.peek()
	return t.kind == tOp && strings.Contains(ops, t.text)
}

// startsPrimary reports whether the next token can begin an operand, which
// is what triggers implicit multiplication ("2x", "x(x+1)", "x sin(x)").
func (p *parser) startsPrimary() bool {
	k := p.peek().kind
	return k == tNum || k == tIdent || k == tLParen
}

// Parse turns text into an expression tree.
func Parse(src string) (Node, error) {
	if strings.TrimSpace(src) == "" {
		return nil, errors.New("empty expression")
	}
	toks, err := lex(src)
	if err != nil {
		return nil, err
	}
	p := &parser{toks: toks}
	n, err := p.expr()
	if err != nil {
		return nil, err
	}
	if t := p.peek(); t.kind != tEOF {
		return nil, fmt.Errorf("unexpected %q", t.text)
	}
	return n, nil
}

func (p *parser) expr() (Node, error) {
	left, err := p.term()
	if err != nil {
		return nil, err
	}
	for p.isOp("+-") {
		op := p.next().text[0]
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		left = Bin{Op: op, L: left, R: right}
	}
	return left, nil
}

func (p *parser) term() (Node, error) {
	left, err := p.unary()
	if err != nil {
		return nil, err
	}
	for {
		switch {
		case p.isOp("*/"):
			op := p.next().text[0]
			right, err := p.unary()
			if err != nil {
				return nil, err
			}
			left = Bin{Op: op, L: left, R: right}
		case p.startsPrimary():
			right, err := p.power()
			if err != nil {
				return nil, err
			}
			left = Bin{Op: '*', L: left, R: right}
		default:
			return left, nil
		}
	}
}

func (p *parser) unary() (Node, error) {
	if p.isOp("-") {
		p.next()
		x, err := p.unary()
		if err != nil {
			return nil, err
		}
		return Neg{X: x}, nil
	}
	if p.isOp("+") {
		p.next()
		return p.unary()
	}
	return p.power()
}

func (p *parser) power() (Node, error) {
	base, err := p.primary()
	if err != nil {
		return nil, err
	}
	if p.isOp("^") {
		p.next()
		exp, err := p.unary() // right-associative, allows 2^-3
		if err != nil {
			return nil, err
		}
		return Bin{Op: '^', L: base, R: exp}, nil
	}
	return base, nil
}

func (p *parser) primary() (Node, error) {
	t := p.next()
	switch t.kind {
	case tNum:
		return Num{V: t.num}, nil
	case tIdent:
		if IsFunction(t.text) {
			if p.peek().kind != tLParen {
				return nil, fmt.Errorf("function %s needs parentheses, e.g. %s(x)", t.text, t.text)
			}
			p.next()
			arg, err := p.expr()
			if err != nil {
				return nil, err
			}
			if p.peek().kind != tRParen {
				return nil, errors.New("missing closing parenthesis")
			}
			p.next()
			return Call{Fn: t.text, Arg: arg}, nil
		}
		if v, ok := constants[t.text]; ok {
			return Num{V: v}, nil
		}
		return Var{Name: t.text}, nil
	case tLParen:
		n, err := p.expr()
		if err != nil {
			return nil, err
		}
		if p.peek().kind != tRParen {
			return nil, errors.New("missing closing parenthesis")
		}
		p.next()
		return n, nil
	case tEOF:
		return nil, errors.New("expression ends unexpectedly")
	}
	return nil, fmt.Errorf("unexpected %q", t.text)
}
