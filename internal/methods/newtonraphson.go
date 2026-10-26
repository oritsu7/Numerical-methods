package methods

import (
	"fmt"
	"math"

	"numericalmethods/internal/banners"
	"numericalmethods/internal/expr"
	"numericalmethods/internal/ui"
)

// Runs the interactive Newton-Raphson screen.
func RunNewtonRaphson() {
	ui.Clear()
	ui.ShowBanner(banners.NewtonRaphson)
	newtonRaphson()
	// Ctrl+C brings you straight back to the menu...
	ui.Pause()
}

func newtonRaphson() {
	node, f := ui.ReadFunction("f(x) : ")
	df := expr.Func1(expr.Diff(node, "x"), "x")
	xn := ui.ReadFloat("Enter first guess : ")
	maxIter := ui.ReadIntDefault("Max iterations [100] : ", 1, ui.NoMax, 100)
	tol := ui.ReadFloatDefault("Tolerance [1e-6] : ", 1e-6, true)

	ui.Rule("")
	fmt.Printf("%5s %16s %14s %14s\n", "Iter", "x", "f(x)", "f'(x)")
	ui.Rule("─")

	for i := 1; i <= maxIter; i++ {
		fx, err := evaluate(f, xn)
		if err != nil {
			ui.Error(err.Error())
			return
		}
		d, err := evaluate(df, xn)
		if err != nil {
			ui.Error(err.Error())
			return
		}
		fmt.Printf("%5d %16.8f %14.3e %14.3e\n", i, xn, fx, d)

		if math.Abs(d) < 1e-12 {
			ui.Error("Derivative became zero. Newton-Raphson cannot continue.")
			return
		}
		next := xn - fx/d
		if math.Abs(next-xn) < tol {
			fr, err := evaluate(f, next)
			if err != nil {
				ui.Error(err.Error())
				return
			}
			printResult(next, fr, i, true, fmt.Sprintf("|Δx| < %g", tol))
			return
		}
		xn = next
	}
	ui.Error("Newton-Raphson method did not converge.")
	fmt.Printf("  Last estimate : %.6f\n", xn)
}
