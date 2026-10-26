// Package methods implements the toolkit.
package methods

import (
	"fmt"
	"math"

	"numericalmethods/internal/ui"
)

// evaluate calls f(v) and rejects NaN / infinite results.
func evaluate(f func(float64) float64, v float64) (float64, error) {
	r := f(v)
	if math.IsNaN(r) {
		return 0, fmt.Errorf("f(%.6g) is undefined (not a real number).", v)
	}
	if math.IsInf(r, 0) {
		return 0, fmt.Errorf("f(%.6g) is undefined or infinite.", v)
	}
	return r, nil
}

func printResult(root, fRoot float64, iterations int, converged bool, note string) {
	ui.Rule("")
	if converged {
		if note != "" {
			ui.Success("Converged - " + note)
		} else {
			ui.Success("Converged")
		}
	} else {
		ui.Info("Stopped before reaching the tolerance - result is approximate.")
	}
	fmt.Printf("  Iterations : %d\n", iterations)
	fmt.Printf("  Root       : %s%.6f%s\n", ui.Bold, root, ui.Reset)
	fmt.Printf("  f(root)    : %.3e\n", fRoot)
	ui.Rule("")
}

// bracket asks for [a, b], the iteration limit and tolerance, and checks the
// sign change. ok is false when the method cannot continue (already reported).
type bracketInput struct {
	f        func(float64) float64
	a, b     float64
	fa, fb   float64
	maxIter  int
	tol      float64
	finished bool // true when an exact endpoint root was found and reported
}

func readBracket(name string) (in bracketInput, ok bool) {
	_, f := ui.ReadFunction("f(x) : ")
	a := ui.ReadFloat("a : ")
	b := ui.ReadFloat("b : ")
	if a == b {
		ui.Error("a and b must be different.")
		return in, false
	}
	if a > b {
		a, b = b, a
		ui.Info(fmt.Sprintf("Swapped so that a < b  (a = %v, b = %v)", a, b))
	}
	in.maxIter = ui.ReadInt("Iterations : ", 1, ui.NoMax)
	in.tol = ui.ReadFloatDefault("Tolerance [1e-6] : ", 1e-6, true)
	in.f, in.a, in.b = f, a, b

	var err error
	if in.fa, err = evaluate(f, a); err != nil {
		ui.Error(err.Error())
		return in, false
	}
	if in.fb, err = evaluate(f, b); err != nil {
		ui.Error(err.Error())
		return in, false
	}
	fmt.Printf("\nf(a) = %.6f    f(b) = %.6f\n", in.fa, in.fb)

	switch {
	case in.fa == 0:
		printResult(a, 0, 0, true, "a is an exact root")
		in.finished = true
	case in.fb == 0:
		printResult(b, 0, 0, true, "b is an exact root")
		in.finished = true
	case in.fa*in.fb > 0:
		ui.Error(fmt.Sprintf("f(a) and f(b) have the same sign - %s can not be applied.", name))
		ui.Info("Choose an interval where the function changes sign.")
		return in, false
	}
	return in, true
}

func printBracketHeader() {
	ui.Rule("")
	fmt.Printf("%5s %12s %12s %12s %13s\n", "Iter", "a", "b", "c", "f(c)")
	ui.Rule("─")
}
