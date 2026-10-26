// Numerical Methods toolkit — Bisection, Regula-Falsi, Newton-Raphson and
// Gauss Elimination in an interactive terminal menu.
//
//	go run .                 open the menu
//	go run . bisection       run one method directly (bisection | regula | newton | gauss)
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"numericalmethods/internal/banners"
	"numericalmethods/internal/methods"
	"numericalmethods/internal/ui"
)

type method struct {
	title string
	run   func()
}

var menu = []method{
	{"Bisection Method", methods.RunBisection},
	{"Regula-Falsi Method", methods.RunRegulaFalsi},
	{"Newton-Raphson Method", methods.RunNewtonRaphson},
	{"Gauss Elimination Method", methods.RunGaussElimination},
}

var aliases = map[string]int{
	"bisection": 0, "regula": 1, "regulafalsi": 1, "regula-falsi": 1,
	"newton": 2, "newton-raphson": 2, "newtonraphson": 2, "gauss": 3,
}

// runMethod runs one method. Ctrl+C inside it returns to the caller;
// end of input (Ctrl+D / closed pipe) asks the caller to quit.
func runMethod(fn func()) (quit bool) {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			switch {
			case ok && errors.Is(err, ui.ErrInterrupt):
				quit = false
			case ok && errors.Is(err, ui.ErrEOF):
				quit = true
			default:
				panic(r)
			}
		}
	}()
	fn()
	return false
}

func main() {
	if len(os.Args) > 1 {
		idx, ok := aliases[strings.ToLower(os.Args[1])]
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown method %q (use bisection, regula, newton or gauss)\n", os.Args[1])
			os.Exit(2)
		}
		runMethod(menu[idx].run)
		fmt.Println()
		return
	}

	for {
		ui.Clear()
		fmt.Println(ui.Cyan + banners.Main + ui.Reset)
		for i, m := range menu {
			fmt.Printf("  %d : %s\n", i+1, m.title)
		}
		fmt.Println("  0 : Exit")
		fmt.Printf("%s  (Ctrl+C inside a method returns here; Ctrl+C here exits)%s\n\n", ui.Dim, ui.Reset)

		choice, err := ui.Prompt("Enter a number : ")
		if err != nil {
			break
		}
		choice = strings.TrimSpace(choice)
		if choice == "0" {
			break
		}

		idx := -1
		if len(choice) == 1 && choice[0] >= '1' && choice[0] <= byte('0'+len(menu)) {
			idx = int(choice[0] - '1')
		}
		if idx < 0 {
			ui.Error("Invalid input")
			if _, err := ui.Prompt("Press Enter to try again..."); err != nil {
				break
			}
			continue
		}
		if runMethod(menu[idx].run) {
			break
		}
	}

	ui.Clear()
	fmt.Println("Byee")
}
