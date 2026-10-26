package expr

import (
	"math"
	"testing"
)

func ev(t *testing.T, s string, x float64) float64 {
	t.Helper()
	n, err := Parse(s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return Eval(n, map[string]float64{"x": x})
}

func TestParseEval(t *testing.T) {
	cases := []struct {
		src  string
		x    float64
		want float64
	}{
		{"x^3 - 2x - 5", 2, -1},
		{"2x", 3, 6},
		{"x sin(x)", math.Pi / 2, math.Pi / 2},
		{"x**2", 4, 16},
		{"-x^2", 3, -9},
		{"2^-1", 0, 0.5},
		{"2^3^2", 0, 512},
		{"(x+1)(x-1)", 3, 8},
		{"3(x+1)", 1, 6},
		{"1/2x", 4, 2},
		{"pi/2", 0, math.Pi / 2},
		{"2E", 0, 2 * math.E},
		{"1e-3 + x", 0, 0.001},
		{"exp(x) - 3x", 0, 1},
		{"xsin(x)", math.Pi / 2, math.Pi / 2},
		{"log(E^2)", 0, 2},
		{"cos(x) - x", 0, 1},
	}
	for _, c := range cases {
		if got := ev(t, c.src, c.x); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%q at %v = %v, want %v", c.src, c.x, got, c.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	for _, s := range []string{"", "x +", "(x", "sin x", "x $ 2", "2 ** ", "x)"} {
		if _, err := Parse(s); err == nil {
			t.Errorf("expected error for %q", s)
		}
	}
}

func TestDiff(t *testing.T) {
	cases := []string{
		"x^3 - 2x - 5", "sin(x)*cos(x)", "exp(2x)", "x^x", "1/x", "sqrt(x^2+1)",
		"tan(x)", "atan(x)", "log(x^2+1)", "x^2.5", "2^x", "(x+1)/(x-2)", "abs(x)", "tanh(x)",
	}
	pts := []float64{0.7, 1.3, 2.1}
	for _, s := range cases {
		n, _ := Parse(s)
		d := Diff(n, "x")
		f, df := Func1(n, "x"), Func1(d, "x")
		for _, p := range pts {
			h := 1e-6
			num := (f(p+h) - f(p-h)) / (2 * h)
			if got := df(p); math.Abs(got-num) > 1e-4*math.Max(1, math.Abs(num)) {
				t.Errorf("d/dx %s at %v = %v, numeric %v", s, p, got, num)
			}
		}
	}
}

func TestLinearize(t *testing.T) {
	n, _ := Parse("2x + y - z - 8")
	l, err := Linearize(n)
	if err != nil || l.Coef["x"] != 2 || l.Coef["y"] != 1 || l.Coef["z"] != -1 || l.Const != -8 {
		t.Fatalf("got %+v err %v", l, err)
	}
	for _, s := range []string{"x*y", "x^2", "sin(x)", "1/x"} {
		n, _ := Parse(s)
		if _, err := Linearize(n); err == nil {
			t.Errorf("expected non-linear error for %q", s)
		}
	}
	n, _ = Parse("x/2 + 0.5(y - x)")
	l, _ = Linearize(n)
	if _, ok := l.Coef["x"]; ok || l.Coef["y"] != 0.5 {
		t.Errorf("cancellation failed: %+v", l)
	}
}
