package methods

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"numericalmethods/internal/banners"
	"numericalmethods/internal/expr"
	"numericalmethods/internal/ui"
)

const eps = 1e-9

type matrix [][]float64

const num = `(\d*\.?\d*)`

var (
	opRE    = regexp.MustCompile(`(?i)^` + num + `\s*r(\d+)\s*([+-])\s*` + num + `\s*r(\d+)$`)
	scaleRE = regexp.MustCompile(`(?i)^r(\d+)\s*([*/])\s*(-?\d*\.?\d+)$`)
	swapRE  = regexp.MustCompile(`(?i)^(?:swap\s+r(\d+)\s+r(\d+)|r(\d+)\s*<->\s*r(\d+))$`)
)

func helpText() string {
	return "\n" + ui.Bold + "Row operations" + ui.Reset + "  (rows are numbered from 1)\n" +
		"  R2 - 2R1      replace R2 with  R2 - 2*R1      (also: R3 + R1, 3R2 - 2R1)\n" +
		"  R2 / 3        divide row 2 by 3               (also: R2 * 0.5)\n" +
		"  R1 <-> R2     swap two rows                   (also: swap R1 R2)\n" +
		"  auto          run the whole elimination automatically\n" +
		"  undo          undo the last operation\n" +
		"  help          show this list\n" +
		"  q             finish and solve\n"
}

// parsing

// parseSystem reads n linear equations and returns the variable names and
// the augmented matrix [A | b].
func parseSystem(n int) ([]string, matrix) {
	for {
		names, aug, err := tryParseSystem(n)
		if err == nil {
			return names, aug
		}
		ui.Error("Could not read the system: " + err.Error())
		ui.Info("Please enter all the equations again. Example: 2x + y - z = 8")
	}
}

func tryParseSystem(n int) ([]string, matrix, error) {
	forms := make([]expr.Linear, n)
	varSet := map[string]bool{}
	for i := 0; i < n; i++ {
		line := ui.ReadLine(fmt.Sprintf("Equation %d : ", i+1))
		if strings.Count(line, "=") != 1 {
			return nil, nil, errors.New("each equation needs exactly one '=' sign")
		}
		parts := strings.Split(line, "=")
		lhs, err := expr.Parse(parts[0])
		if err != nil {
			return nil, nil, fmt.Errorf("left side of equation %d: %v", i+1, err)
		}
		rhs, err := expr.Parse(parts[1])
		if err != nil {
			return nil, nil, fmt.Errorf("right side of equation %d: %v", i+1, err)
		}
		l, err := expr.Linearize(expr.Bin{Op: '-', L: lhs, R: rhs})
		if err != nil {
			return nil, nil, fmt.Errorf("equation %d: %v", i+1, err)
		}
		forms[i] = l
		for name := range l.Coef {
			varSet[name] = true
		}
	}

	names := make([]string, 0, len(varSet))
	for name := range varSet {
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) != n {
		list := strings.Join(names, ", ")
		if list == "" {
			list = "none"
		}
		return nil, nil, fmt.Errorf("expected %d variables but found %d (%s)", n, len(names), list)
	}

	aug := make(matrix, n)
	for i, l := range forms {
		aug[i] = make([]float64, n+1)
		for j, name := range names {
			aug[i][j] = l.Coef[name]
		}
		aug[i][n] = -l.Const
	}
	return names, aug, nil
}

// display

func (m matrix) clone() matrix {
	out := make(matrix, len(m))
	for i := range m {
		out[i] = append([]float64(nil), m[i]...)
	}
	return out
}

func (m matrix) clean() matrix {
	for i := range m {
		for j := range m[i] {
			if math.Abs(m[i][j]) < 1e-12 {
				m[i][j] = 0
			}
		}
	}
	return m
}

// f4 formats like %.4f but never prints "-0.0000".
func f4(v float64) string {
	if math.Abs(v) < 0.00005 {
		v = 0
	}
	return fmt.Sprintf("%10.4f", v)
}

func (m matrix) show() {
	for _, row := range m {
		left := make([]string, len(row)-1)
		for j := range left {
			left[j] = f4(row[j])
		}
		fmt.Printf("  │ %s │ %s │\n", strings.Join(left, " "), f4(row[len(row)-1]))
	}
}

// operations

// applyCommand returns a new matrix with the command applied.
func applyCommand(m matrix, command string) (matrix, error) {
	n := len(m)
	out := m.clone()

	row := func(text string) (int, error) {
		r, _ := strconv.Atoi(text)
		if r < 1 || r > n {
			return 0, fmt.Errorf("Row R%s does not exist (valid rows: R1-R%d).", text, n)
		}
		return r - 1, nil
	}
	mult := func(text string) float64 {
		if text == "" || text == "." {
			return 1
		}
		v, _ := strconv.ParseFloat(text, 64)
		return v
	}

	if g := swapRE.FindStringSubmatch(command); g != nil {
		var ids []string
		for _, s := range g[1:] {
			if s != "" {
				ids = append(ids, s)
			}
		}
		a, err := row(ids[0])
		if err != nil {
			return nil, err
		}
		b, err := row(ids[1])
		if err != nil {
			return nil, err
		}
		if a == b {
			return nil, errors.New("Pick two different rows to swap.")
		}
		out[a], out[b] = out[b], out[a]
		return out.clean(), nil
	}

	if g := scaleRE.FindStringSubmatch(command); g != nil {
		r, err := row(g[1])
		if err != nil {
			return nil, err
		}
		factor, _ := strconv.ParseFloat(g[3], 64)
		if g[2] == "/" {
			if factor == 0 {
				return nil, errors.New("Can not divide by zero.")
			}
			for j := range out[r] {
				out[r][j] /= factor
			}
		} else {
			if factor == 0 {
				return nil, errors.New("Multiplying a row by 0 would destroy it.")
			}
			for j := range out[r] {
				out[r][j] *= factor
			}
		}
		return out.clean(), nil
	}

	if g := opRE.FindStringSubmatch(command); g != nil {
		r1, err := row(g[2])
		if err != nil {
			return nil, err
		}
		r2, err := row(g[5])
		if err != nil {
			return nil, err
		}
		m1, m2 := mult(g[1]), mult(g[4])
		if r1 == r2 {
			return nil, errors.New("Use two different rows (R2 - 2R1, not R2 - 2R2).")
		}
		if m1 == 0 {
			return nil, errors.New("The multiplier of the row being replaced can not be 0.")
		}
		for j := range out[r1] {
			if g[3] == "+" {
				out[r1][j] = m1*out[r1][j] + m2*out[r2][j]
			} else {
				out[r1][j] = m1*out[r1][j] - m2*out[r2][j]
			}
		}
		return out.clean(), nil
	}

	return nil, errors.New("Invalid command! Type 'help' to see the supported operations.")
}

// autoEliminate does forward elimination with partial pivoting, printing every step...
func autoEliminate(m matrix) matrix {
	m = m.clone()
	n := len(m)
	for col := 0; col < n; col++ {
		pivot := col
		for r := col + 1; r < n; r++ {
			if math.Abs(m[r][col]) > math.Abs(m[pivot][col]) {
				pivot = r
			}
		}
		if math.Abs(m[pivot][col]) < eps {
			continue
		}
		if pivot != col {
			m[col], m[pivot] = m[pivot], m[col]
			fmt.Printf("\nR%d <-> R%d\n", col+1, pivot+1)
			m.clone().clean().show()
		}
		for r := col + 1; r < n; r++ {
			factor := m[r][col] / m[col][col]
			if math.Abs(factor) < eps {
				continue
			}
			for j := range m[r] {
				m[r][j] -= factor * m[col][j]
			}
			m[r][col] = 0
			fmt.Printf("\nR%d = R%d - (%.4g)·R%d\n", r+1, r+1, factor, col+1)
			m.clone().clean().show()
		}
	}
	return m.clean()
}

// backSubstitution returns the solution, or nil after explaining why not.
func backSubstitution(m matrix) []float64 {
	n := len(m)

	for i := 0; i < n; i++ {
		for j := 0; j < i; j++ {
			if math.Abs(m[i][j]) > eps {
				ui.Error(fmt.Sprintf("The matrix is not upper triangular yet "+
					"(R%d still has a non-zero entry in column %d).", i+1, j+1))
				ui.Info("Keep applying row operations, or type 'auto'.")
				return nil
			}
		}
	}

	for i := 0; i < n; i++ {
		allZero := true
		for j := 0; j < n; j++ {
			if math.Abs(m[i][j]) >= eps {
				allZero = false
				break
			}
		}
		if allZero {
			if math.Abs(m[i][n]) > eps {
				ui.Error(fmt.Sprintf("R%d reads 0 = %.4g - the system has NO solution.", i+1, m[i][n]))
			} else {
				ui.Error(fmt.Sprintf("R%d is all zeros - the system has INFINITELY many solutions.", i+1))
			}
			return nil
		}
		if math.Abs(m[i][i]) < eps {
			ui.Error(fmt.Sprintf("R%d has a zero pivot on the diagonal - swap rows or type 'auto'.", i+1))
			return nil
		}
	}

	sol := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		sum := 0.0
		for j := i + 1; j < n; j++ {
			sum += m[i][j] * sol[j]
		}
		sol[i] = (m[i][n] - sum) / m[i][i]
	}
	return sol
}

// UI

// Runs the interactive Gauss elimination screen.
func RunGaussElimination() {
	ui.Clear()
	ui.ShowBanner(banners.GaussElimination)

	n := ui.ReadIntDefault("Number of variables (2-6) [3] : ", 2, 6, 3)
	names, m := parseSystem(n)
	original := m.clone()
	var history []matrix

	fmt.Println("\nInitial matrix:")
	m.show()
	fmt.Println(helpText())

commands:
	for {
		command := strings.TrimSpace(ui.ReadLine("Enter row operation (q to finish): "))
		switch strings.ToLower(command) {
		case "q", "quit", "done":
			break commands
		case "help":
			fmt.Println(helpText())
			continue
		case "undo":
			if len(history) > 0 {
				m = history[len(history)-1]
				history = history[:len(history)-1]
				fmt.Println("\nUndone. Current matrix:")
				m.show()
			} else {
				ui.Info("Nothing to undo.")
			}
			continue
		case "auto":
			history = append(history, m)
			m = autoEliminate(m)
			fmt.Println("\nElimination complete.")
			continue
		case "":
			continue
		}

		result, err := applyCommand(m, command)
		if err != nil {
			ui.Error(err.Error())
			continue
		}
		history = append(history, m)
		m = result
		fmt.Println("\nUpdated matrix:")
		m.show()
	}

	fmt.Println("\nFinal matrix:")
	m.show()

	if sol := backSubstitution(m); sol != nil {
		ui.Rule("")
		ui.Success("Solution")
		for i, name := range names {
			fmt.Printf("  %s = %s%s%s\n", name, ui.Bold, strings.TrimSpace(f4(sol[i])), ui.Reset)
		}
		residual := 0.0
		for i := range original {
			s := -original[i][n]
			for j := 0; j < n; j++ {
				s += original[i][j] * sol[j]
			}
			residual = math.Max(residual, math.Abs(s))
		}
		fmt.Printf("%s  check: largest residual in the original equations = %.2e%s\n", ui.Dim, residual, ui.Reset)
		ui.Rule("")
	}
	ui.Pause()
}
