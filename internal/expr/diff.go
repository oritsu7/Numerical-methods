package expr

import (
	"errors"
	"math"
)

// simplifying constructors

func isNum(n Node, v float64) bool {
	c, ok := n.(Num)
	return ok && c.V == v
}

func neg(a Node) Node {
	switch t := a.(type) {
	case Num:
		return Num{-t.V}
	case Neg:
		return t.X
	}
	return Neg{a}
}

func add(a, b Node) Node {
	if x, ok := a.(Num); ok {
		if y, ok := b.(Num); ok {
			return Num{x.V + y.V}
		}
	}
	if isNum(a, 0) {
		return b
	}
	if isNum(b, 0) {
		return a
	}
	return Bin{'+', a, b}
}

func sub(a, b Node) Node {
	if x, ok := a.(Num); ok {
		if y, ok := b.(Num); ok {
			return Num{x.V - y.V}
		}
	}
	if isNum(b, 0) {
		return a
	}
	if isNum(a, 0) {
		return neg(b)
	}
	return Bin{'-', a, b}
}

func mul(a, b Node) Node {
	if x, ok := a.(Num); ok {
		if y, ok := b.(Num); ok {
			return Num{x.V * y.V}
		}
	}
	if isNum(a, 0) || isNum(b, 0) {
		return Num{0}
	}
	if isNum(a, 1) {
		return b
	}
	if isNum(b, 1) {
		return a
	}
	return Bin{'*', a, b}
}

func div(a, b Node) Node {
	if x, ok := a.(Num); ok {
		if y, ok := b.(Num); ok && y.V != 0 {
			return Num{x.V / y.V}
		}
	}
	if isNum(a, 0) {
		return Num{0}
	}
	if isNum(b, 1) {
		return a
	}
	return Bin{'/', a, b}
}

func pow(a, b Node) Node {
	if x, ok := a.(Num); ok {
		if y, ok := b.(Num); ok {
			return Num{math.Pow(x.V, y.V)}
		}
	}
	if isNum(b, 0) {
		return Num{1}
	}
	if isNum(b, 1) {
		return a
	}
	return Bin{'^', a, b}
}

func call(fn string, a Node) Node { return Call{fn, a} }

// differentiation

// Diff returns d n / d v as a (lightly simplified) expression tree.
func Diff(n Node, v string) Node {
	switch t := n.(type) {
	case Num:
		return Num{0}
	case Var:
		if t.Name == v {
			return Num{1}
		}
		return Num{0}
	case Neg:
		return neg(Diff(t.X, v))
	case Bin:
		dl, dr := Diff(t.L, v), Diff(t.R, v)
		switch t.Op {
		case '+':
			return add(dl, dr)
		case '-':
			return sub(dl, dr)
		case '*':
			return add(mul(dl, t.R), mul(t.L, dr))
		case '/':
			return div(sub(mul(dl, t.R), mul(t.L, dr)), pow(t.R, Num{2}))
		case '^':
			if !HasVar(t.R, v) { // power rule: c * u^(c-1) * u'
				return mul(mul(t.R, pow(t.L, sub(t.R, Num{1}))), dl)
			}
			// general rule: u^w * (w' ln u + w u'/u)
			return mul(n, add(mul(dr, call("log", t.L)), div(mul(t.R, dl), t.L)))
		}
	case Call:
		return mul(outerDeriv(t.Fn, t.Arg), Diff(t.Arg, v))
	}
	return Num{0}
}

func outerDeriv(fn string, u Node) Node {
	one, two := Num{1}, Num{2}
	switch fn {
	case "sin":
		return call("cos", u)
	case "cos":
		return neg(call("sin", u))
	case "tan":
		return div(one, pow(call("cos", u), two))
	case "asin":
		return div(one, call("sqrt", sub(one, pow(u, two))))
	case "acos":
		return neg(div(one, call("sqrt", sub(one, pow(u, two)))))
	case "atan":
		return div(one, add(one, pow(u, two)))
	case "sinh":
		return call("cosh", u)
	case "cosh":
		return call("sinh", u)
	case "tanh":
		return sub(one, pow(call("tanh", u), two))
	case "asinh":
		return div(one, call("sqrt", add(pow(u, two), one)))
	case "acosh":
		return div(one, call("sqrt", sub(pow(u, two), one)))
	case "atanh":
		return div(one, sub(one, pow(u, two)))
	case "exp":
		return call("exp", u)
	case "log", "ln":
		return div(one, u)
	case "log10":
		return div(one, mul(u, Num{math.Ln10}))
	case "log2":
		return div(one, mul(u, Num{math.Ln2}))
	case "sqrt":
		return div(one, mul(two, call("sqrt", u)))
	case "cbrt":
		return div(one, mul(Num{3}, pow(call("cbrt", u), two)))
	case "abs", "Abs":
		return div(u, call("abs", u))
	}
	return Num{math.NaN()}
}

// linear forms

// Linear is  Summation Coef[name]·name + Const.
type Linear struct {
	Coef  map[string]float64
	Const float64
}

// ErrNotLinear is returned by Linearize for non-linear input.
var ErrNotLinear = errors.New("equation is not linear")

func (l Linear) isConst() bool { return len(l.Coef) == 0 }

func constLin(v float64) Linear { return Linear{Coef: map[string]float64{}, Const: v} }

func scaleLin(l Linear, k float64) Linear {
	out := constLin(l.Const * k)
	for name, c := range l.Coef {
		if v := c * k; math.Abs(v) > 1e-12 {
			out.Coef[name] = v
		}
	}
	return out
}

func combine(a, b Linear, sign float64) Linear {
	out := constLin(a.Const + sign*b.Const)
	for name, c := range a.Coef {
		out.Coef[name] = c
	}
	for name, c := range b.Coef {
		out.Coef[name] += sign * c
	}
	for name, c := range out.Coef {
		if math.Abs(c) <= 1e-12 {
			delete(out.Coef, name)
		}
	}
	return out
}

// Linearize rewrites n as a linear combination of variables plus a constant,
// or returns ErrNotLinear (e.g. for x*y, x^2, sin(x)).
func Linearize(n Node) (Linear, error) {
	switch t := n.(type) {
	case Num:
		return constLin(t.V), nil
	case Var:
		return Linear{Coef: map[string]float64{t.Name: 1}}, nil
	case Neg:
		l, err := Linearize(t.X)
		if err != nil {
			return Linear{}, err
		}
		return scaleLin(l, -1), nil
	case Bin:
		a, err := Linearize(t.L)
		if err != nil {
			return Linear{}, err
		}
		b, err := Linearize(t.R)
		if err != nil {
			return Linear{}, err
		}
		switch t.Op {
		case '+':
			return combine(a, b, 1), nil
		case '-':
			return combine(a, b, -1), nil
		case '*':
			if a.isConst() {
				return scaleLin(b, a.Const), nil
			}
			if b.isConst() {
				return scaleLin(a, b.Const), nil
			}
		case '/':
			if !b.isConst() {
				return Linear{}, ErrNotLinear
			}
			if b.Const == 0 {
				return Linear{}, errors.New("division by zero")
			}
			return scaleLin(a, 1/b.Const), nil
		case '^':
			if a.isConst() && b.isConst() {
				return constLin(math.Pow(a.Const, b.Const)), nil
			}
			if b.isConst() && b.Const == 1 {
				return a, nil
			}
			if b.isConst() && b.Const == 0 {
				return constLin(1), nil
			}
		}
	case Call:
		l, err := Linearize(t.Arg)
		if err != nil {
			return Linear{}, err
		}
		if l.isConst() {
			return constLin(funcs[t.Fn](l.Const)), nil
		}
	}
	return Linear{}, ErrNotLinear
}
