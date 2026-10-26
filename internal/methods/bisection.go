package methods

import (
	"fmt"

	"numericalmethods/internal/banners"
	"numericalmethods/internal/ui"
)

// RunBisection runs the interactive bisection screen.
func RunBisection() {
	ui.Clear()
	ui.ShowBanner(banners.Bisection)
	bisection()
	// Pause is not deferred on purpose, Ctrl+C gets you back to the menu.
	ui.Pause()
}

func bisection() {
	in, ok := readBracket("bisection")
	if !ok || in.finished {
		return
	}
	f, a, b, fa := in.f, in.a, in.b, in.fa

	printBracketHeader()
	var c, fc float64
	for i := 1; i <= in.maxIter; i++ {
		c = (a + b) / 2
		var err error
		if fc, err = evaluate(f, c); err != nil {
			ui.Error(err.Error())
			return
		}
		fmt.Printf("%5d %12.6f %12.6f %12.6f %13.3e\n", i, a, b, c, fc)

		if fc == 0 || (b-a)/2 < in.tol {
			note := fmt.Sprintf("interval < %g", in.tol)
			if fc == 0 {
				note = "exact root"
			}
			printResult(c, fc, i, true, note)
			return
		}
		if fa*fc < 0 {
			b = c
		} else {
			a, fa = c, fc
		}
	}
	printResult(c, fc, in.maxIter, false, "")
}
