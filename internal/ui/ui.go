// Package ui contains the terminal helpers shared by every method:
// colours, banners, validated input and Ctrl+C handling.
package ui

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"

	"numericalmethods/internal/expr"
)

// ANSI colours. They are empty strings when stdout is not a terminal.
var (
	Reset, Bold, Dim, Red, Green, Yellow, Cyan string
	isTTY                                      bool
)

func init() {
	enableVT()
	if fi, err := os.Stdout.Stat(); err == nil && fi.Mode()&os.ModeCharDevice != 0 {
		isTTY = true
		Reset, Bold, Dim = "\033[0m", "\033[1m", "\033[2m"
		Red, Green, Yellow, Cyan = "\033[91m", "\033[92m", "\033[93m", "\033[96m"
	}
}

// input / Ctrl+C

// Errors returned by Prompt. ReadLine and friends panic with these values
// instead, so method code can stay linear; main recovers them.
var (
	ErrInterrupt = errors.New("interrupted")
	ErrEOF       = errors.New("end of input")
)

type lineMsg struct {
	text string
	eof  bool
}

var (
	once  sync.Once
	lines chan lineMsg
	sig   chan os.Signal
)

func setup() {
	once.Do(func() {
		lines = make(chan lineMsg)
		sig = make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		go func() {
			sc := bufio.NewScanner(os.Stdin)
			sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
			for sc.Scan() {
				lines <- lineMsg{text: sc.Text()}
			}
			lines <- lineMsg{eof: true}
		}()
	})
}

// Prints prompt and waits for a line, or for Ctrl+C / end of input.
func Prompt(prompt string) (string, error) {
	setup()
	fmt.Print(prompt)
	select {
	case m := <-lines:
		if m.eof {
			fmt.Println()
			return "", ErrEOF
		}
		return m.text, nil
	case <-sig:
		fmt.Println()
		return "", ErrInterrupt
	}
}

// ReadLine is Prompt that panics with ErrInterrupt / ErrEOF.
func ReadLine(prompt string) string {
	s, err := Prompt(prompt)
	if err != nil {
		panic(err)
	}
	return s
}

// output helpers

// Clears the screen (only when stdout is a terminal).
func Clear() {
	if isTTY {
		fmt.Print("\033[H\033[2J\033[3J")
	}
}

func ShowBanner(art string) { fmt.Println(Cyan + art + Reset) }
func Error(msg string)      { fmt.Printf("%s✗ %s%s\n", Red, msg, Reset) }
func Info(msg string)       { fmt.Printf("%s• %s%s\n", Yellow, msg, Reset) }
func Success(msg string)    { fmt.Printf("%s✓ %s%s\n", Green, msg, Reset) }

// Rule prints a dimmed horizontal line (default: 60 × "═").
func Rule(char string) {
	if char == "" {
		char = "═"
	}
	fmt.Println(Dim + strings.Repeat(char, 60) + Reset)
}

// Pause waits for Enter.
func Pause() { ReadLine("\nPress Enter to return to the menu...") }

// validated input

// ParseNumber accepts plain numbers and constant expressions such as pi/2.
func ParseNumber(raw string) (float64, bool) {
	raw = strings.TrimSpace(raw)
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		n, perr := expr.Parse(raw)
		if perr != nil || len(expr.FreeVars(n)) != 0 {
			return 0, false
		}
		v = expr.Eval(n, nil)
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, false
	}
	return v, true
}

func readFloat(prompt string, def *float64, positive bool) float64 {
	for {
		raw := strings.TrimSpace(ReadLine(prompt))
		if raw == "" && def != nil {
			return *def
		}
		v, ok := ParseNumber(raw)
		if !ok {
			Error("Please enter a valid number.")
			continue
		}
		if positive && v <= 0 {
			Error("Please enter a value greater than 0.")
			continue
		}
		return v
	}
}

// ReadFloat reads a number (expressions like pi/2 are allowed).
func ReadFloat(prompt string) float64 { return readFloat(prompt, nil, false) }

// ReadFloatDefault is ReadFloat where a blank line returns def.
func ReadFloatDefault(prompt string, def float64, positive bool) float64 {
	return readFloat(prompt, &def, positive)
}

// NoMax means "no upper bound" for ReadInt.
const NoMax = math.MaxInt

func readInt(prompt string, min, max int, def *int) int {
	for {
		raw := strings.TrimSpace(ReadLine(prompt))
		if raw == "" && def != nil {
			return *def
		}
		v, err := strconv.Atoi(raw)
		if err != nil {
			Error("Please enter a whole number.")
			continue
		}
		if v < min || v > max {
			if max != NoMax {
				Error(fmt.Sprintf("Enter a number between %d and %d.", min, max))
			} else {
				Error(fmt.Sprintf("Enter %d or more.", min))
			}
			continue
		}
		return v
	}
}

// ReadInt reads a whole number in [min, max].
func ReadInt(prompt string, min, max int) int { return readInt(prompt, min, max, nil) }

// ReadIntDefault is ReadInt where a blank line returns def.
func ReadIntDefault(prompt string, min, max, def int) int {
	return readInt(prompt, min, max, &def)
}

// ReadFunction reads f(x) and returns its tree and a callable.
func ReadFunction(prompt string) (expr.Node, func(float64) float64) {
	for {
		text := strings.TrimSpace(ReadLine(prompt))
		if text == "" {
			continue
		}
		n, err := expr.Parse(text)
		if err != nil {
			Error("Could not understand that expression (" + err.Error() + "). Example: x^3 - 2x - 5")
			continue
		}
		vars := expr.FreeVars(n)
		var extra []string
		hasX := false
		for _, v := range vars {
			if v == "x" {
				hasX = true
			} else {
				extra = append(extra, v)
			}
		}
		if len(extra) > 0 {
			Error(fmt.Sprintf("Unknown symbol(s): %s. Use only x (write E for Euler's number, or exp(x)).",
				strings.Join(extra, ", ")))
			continue
		}
		if !hasX {
			Error("f(x) must contain x.")
			continue
		}
		return n, expr.Func1(n, "x")
	}
}
