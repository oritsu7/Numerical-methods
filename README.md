# Numerical Methods (Go)

Interactive terminal toolkit — a Go port of the original Python project.

| # | Method |
|---|--------|
| 1 | Bisection |
| 2 | Regula-Falsi |
| 3 | Newton-Raphson (derivative computed symbolically) |
| 4 | Gauss Elimination (manual row operations, `auto`, `undo`) |

## Run

```bash
go run .                 # menu
go run . bisection       # jump straight to one method: bisection | regula | newton | gauss
```

Ctrl+C inside a method returns to the menu; Ctrl+C at the menu exits.

## Build

```bash
go build -o numerical-methods .                        # this machine
GOOS=windows GOARCH=amd64 go build -o numerical-methods.exe .
GOOS=darwin  GOARCH=arm64 go build -o numerical-methods-mac .
```

No third-party dependencies — Go 1.21+ is all you need.

## Entering functions

`x^3 - 2x - 5`, `x**2`, `2x`, `x sin(x)`, `(x+1)(x-1)`, `1e-3`, `pi/2`

* Constants: `pi`, `E` (Euler's number — plain `e` is *not* a constant)
* Functions (always with parentheses): `sin cos tan asin acos atan sinh cosh tanh asinh acosh atanh exp log ln log10 log2 sqrt cbrt abs`
* `log` is the natural logarithm, as in SymPy.
* Numeric prompts (a, b, first guess, tolerance) also accept expressions such as `pi/2`.

## Layout

```
main.go                       menu + CLI shortcuts
internal/banners/             ASCII-art headers
internal/expr/                parser, evaluator, symbolic derivative, linear-form extraction (replaces SymPy)
internal/ui/                  colours, validated input, Ctrl+C handling (replaces utils.py)
internal/methods/             bisection, regula falsi, Newton-Raphson, Gauss elimination
```

Run the tests with `go test ./...`.
