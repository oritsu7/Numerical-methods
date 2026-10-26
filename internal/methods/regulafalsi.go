package methods

import (
	"fmt"
	"math"

	"numericalmethods/internal/banners"
	"numericalmethods/internal/ui"
)

// Runs the interactive regula-falsi screen....
func RunRegulaFalsi() {
	ui.Clear()
	ui.ShowBanner(banners.RegulaFalsi)
	regulaFalsi()
	// Ctrl+C : back to the menu.
	ui.Pause()
}

func regulaFalsi() {
	in, ok := readBracket("regula falsi")
	if !ok || in.finished {
		return
	}
	f, a, b, fa, fb := in.f, in.a, in.b, in.fa, in.fb

	printBracketHeader()
	var c, fc, prev float64
	havePrev := false
	for i := 1; i <= in.maxIter; i++ {
		c = (a*fb - b*fa) / (fb - fa)
		var err error
		if fc, err = evaluate(f, c); err != nil {
			ui.Error(err.Error())
			return
		}
		fmt.Printf("%5d %12.6f %12.6f %12.6f %13.3e\n", i, a, b, c, fc)

		if fc == 0 || (havePrev && math.Abs(c-prev) < in.tol) {
			note := fmt.Sprintf("|Δc| < %g", in.tol)
			if fc == 0 {
				note = "exact root"
			}
			printResult(c, fc, i, true, note)
			return
		}
		if fa*fc < 0 {
			b, fb = c, fc
		} else {
			a, fa = c, fc
		}
		prev, havePrev = c, true
	}
	printResult(c, fc, in.maxIter, false, "")
}
