// Package expr is a small expression engine it parses text such as
// "x^3 - 2x - 5" (implicit multiplication and ^ included), evaluates it,
// differentiates it symbolically and extracts linear coefficients.
package expr

import (
	"math"
	"sort"
)

// Node is any node of an expression tree.
type Node interface{ isNode() }

type (
	// Num is a numeric constant.
	Num struct{ V float64 }
	// Var is a named variable.
	Var struct{ Name string }
	// Neg is unary minus.
	Neg struct{ X Node }
	// Bin is a binary operation; Op is one of + - * / ^.
	Bin struct {
		Op   byte
		L, R Node
	}
	// Call is a function application such as sin(x).
	Call struct {
		Fn  string
		Arg Node
	}
)

func (Num) isNode()  {}
func (Var) isNode()  {}
func (Neg) isNode()  {}
func (Bin) isNode()  {}
func (Call) isNode() {}

var funcs = map[string]func(float64) float64{
	"sin": math.Sin, "cos": math.Cos, "tan": math.Tan,
	"asin": math.Asin, "acos": math.Acos, "atan": math.Atan,
	"sinh": math.Sinh, "cosh": math.Cosh, "tanh": math.Tanh,
	"asinh": math.Asinh, "acosh": math.Acosh, "atanh": math.Atanh,
	"exp": math.Exp, "log": math.Log, "ln": math.Log,
	"log10": math.Log10, "log2": math.Log2,
	"sqrt": math.Sqrt, "cbrt": math.Cbrt,
	"abs": math.Abs, "Abs": math.Abs,
}

var constants = map[string]float64{
	"pi": math.Pi,
	"E":  math.E,
}

// IsFunction reports whether name is a supported function...
func IsFunction(name string) bool { _, ok := funcs[name]; return ok }

// Eval evaluates n. Variables missing from env evaluate to NaN.
// Division by zero, log of a negative number etc. yield NaN or ±Inf;
// callers decide how to report that.
func Eval(n Node, env map[string]float64) float64 {
	switch t := n.(type) {
	case Num:
		return t.V
	case Var:
		if v, ok := env[t.Name]; ok {
			return v
		}
		return math.NaN()
	case Neg:
		return -Eval(t.X, env)
	case Bin:
		l, r := Eval(t.L, env), Eval(t.R, env)
		switch t.Op {
		case '+':
			return l + r
		case '-':
			return l - r
		case '*':
			return l * r
		case '/':
			return l / r
		case '^':
			return math.Pow(l, r)
		}
	case Call:
		return funcs[t.Fn](Eval(t.Arg, env))
	}
	return math.NaN()
}

// Func1 turns an expression in one variable into a plain Go function.
func Func1(n Node, name string) func(float64) float64 {
	env := map[string]float64{name: 0}
	return func(x float64) float64 {
		env[name] = x
		return Eval(n, env)
	}
}

// FreeVars returns the sorted variable names used in n.
func FreeVars(n Node) []string {
	set := map[string]struct{}{}
	collect(n, set)
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func collect(n Node, set map[string]struct{}) {
	switch t := n.(type) {
	case Var:
		set[t.Name] = struct{}{}
	case Neg:
		collect(t.X, set)
	case Bin:
		collect(t.L, set)
		collect(t.R, set)
	case Call:
		collect(t.Arg, set)
	}
}

// HasVar reports whether n depends on the variable name.
func HasVar(n Node, name string) bool {
	switch t := n.(type) {
	case Var:
		return t.Name == name
	case Neg:
		return HasVar(t.X, name)
	case Bin:
		return HasVar(t.L, name) || HasVar(t.R, name)
	case Call:
		return HasVar(t.Arg, name)
	}
	return false
}
