import re

import numpy as np
from sympy import linear_eq_to_matrix
from sympy.parsing.sympy_parser import (implicit_multiplication, parse_expr,
                                        standard_transformations, convert_xor)

from utils import (C, clear, error, info, pause, read_int, rule, show_banner,
                   success)

ge = """
 ██████╗  █████╗ ██╗   ██╗███████╗███████╗    ███████╗██╗     ██╗███╗   ███╗██╗███╗   ██╗ █████╗ ████████╗██╗ ██████╗ ███╗   ██╗
██╔════╝ ██╔══██╗██║   ██║██╔════╝██╔════╝    ██╔════╝██║     ██║████╗ ████║██║████╗  ██║██╔══██╗╚══██╔══╝██║██╔═══██╗████╗  ██║
██║  ███╗███████║██║   ██║███████╗███████╗    █████╗  ██║     ██║██╔████╔██║██║██╔██╗ ██║███████║   ██║   ██║██║   ██║██╔██╗ ██║
██║   ██║██╔══██║██║   ██║╚════██║╚════██║    ██╔══╝  ██║     ██║██║╚██╔╝██║██║██║╚██╗██║██╔══██║   ██║   ██║██║   ██║██║╚██╗██║
╚██████╔╝██║  ██║╚██████╔╝███████║███████║    ███████╗███████╗██║██║ ╚═╝ ██║██║██║ ╚████║██║  ██║   ██║   ██║╚██████╔╝██║ ╚████║
 ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚══════╝    ╚══════╝╚══════╝╚═╝╚═╝     ╚═╝╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝
"""

EPS = 1e-9
TRANSFORMS = standard_transformations + (implicit_multiplication, convert_xor)

NUM = r"(\d*\.?\d*)"
OP_RE = re.compile(rf"^{NUM}\s*r(\d+)\s*([+-])\s*{NUM}\s*r(\d+)$", re.I)
SCALE_RE = re.compile(r"^r(\d+)\s*([*/])\s*(-?\d*\.?\d+)$", re.I)
SWAP_RE = re.compile(r"^(?:swap\s+r(\d+)\s+r(\d+)|r(\d+)\s*<->\s*r(\d+))$", re.I)

HELP = f"""
{C.BOLD}Row operations{C.RESET}  (rows are numbered from 1)
  R2 - 2R1      replace R2 with  R2 - 2*R1      (also: R3 + R1, 3R2 - 2R1)
  R2 / 3        divide row 2 by 3               (also: R2 * 0.5)
  R1 <-> R2     swap two rows                   (also: swap R1 R2)
  auto          run the whole elimination automatically
  undo          undo the last operation
  help          show this list
  q             finish and solve
"""


# ---------------------------------------------------------------- parsing
def parse_system(n):
    """Read n linear equations and return (variable_names, augmented_matrix)."""
    while True:
        try:
            exprs = []
            for i in range(n):
                line = input(f"Equation {i + 1} : ")
                if line.count("=") != 1:
                    raise ValueError("each equation needs exactly one '=' sign")
                lhs, rhs = line.split("=")
                exprs.append(parse_expr(lhs, transformations=TRANSFORMS)
                             - parse_expr(rhs, transformations=TRANSFORMS))

            symbols_ = sorted(set().union(*(e.free_symbols for e in exprs)),
                              key=lambda s: s.name)
            if len(symbols_) != n:
                names = ", ".join(s.name for s in symbols_) or "none"
                raise ValueError(f"expected {n} variables but found {len(symbols_)} ({names})")

            A, b = linear_eq_to_matrix(exprs, symbols_)
            aug = np.hstack([np.array(A.tolist(), dtype=float),
                             np.array(b.tolist(), dtype=float)])
            return [s.name for s in symbols_], aug
        except Exception as exc:
            error(f"Could not read the system: {exc}")
            info("Please enter all the equations again. Example: 2x + y - z = 8")


# ---------------------------------------------------------------- display
def clean(m):
    m[np.abs(m) < 1e-12] = 0.0
    return m


def show(m):
    for row in m:
        left = " ".join(f"{v:10.4f}" for v in row[:-1])
        print(f"  │ {left} │ {row[-1]:10.4f} │")


# ---------------------------------------------------------------- operations
def apply_command(m, command):
    """Return a new matrix with the command applied, or raise ValueError."""
    n = len(m)
    new = m.copy()

    def row(text):
        r = int(text) - 1
        if not 0 <= r < n:
            raise ValueError(f"Row R{text} does not exist (valid rows: R1-R{n}).")
        return r

    def mult(text):
        return 1.0 if text in ("", ".") else float(text)

    if match := SWAP_RE.match(command):
        a, b = (g for g in match.groups() if g)
        a, b = row(a), row(b)
        if a == b:
            raise ValueError("Pick two different rows to swap.")
        new[[a, b]] = new[[b, a]]
        return clean(new)

    if match := SCALE_RE.match(command):
        r = row(match.group(1))
        factor = float(match.group(3))
        if match.group(2) == "/":
            if factor == 0:
                raise ValueError("Can not divide by zero.")
            new[r] = new[r] / factor
        else:
            if factor == 0:
                raise ValueError("Multiplying a row by 0 would destroy it.")
            new[r] = new[r] * factor
        return clean(new)

    if match := OP_RE.match(command):
        m1, r1, op, m2, r2 = match.groups()
        r1, r2 = row(r1), row(r2)
        m1, m2 = mult(m1), mult(m2)
        if r1 == r2:
            raise ValueError("Use two different rows (R2 - 2R1, not R2 - 2R2).")
        if m1 == 0:
            raise ValueError("The multiplier of the row being replaced can not be 0.")
        new[r1] = m1 * new[r1] + m2 * new[r2] if op == "+" else m1 * new[r1] - m2 * new[r2]
        return clean(new)

    raise ValueError("Invalid command! Type 'help' to see the supported operations.")


def auto_eliminate(m):
    """Forward elimination with partial pivoting, printing every step."""
    m = m.copy()
    n = len(m)
    for col in range(n):
        pivot = col + int(np.argmax(np.abs(m[col:, col])))
        if abs(m[pivot, col]) < EPS:
            continue
        if pivot != col:
            m[[col, pivot]] = m[[pivot, col]]
            print(f"\nR{col + 1} <-> R{pivot + 1}")
            show(clean(m))
        for r in range(col + 1, n):
            factor = m[r, col] / m[col, col]
            if abs(factor) < EPS:
                continue
            m[r] = m[r] - factor * m[col]
            m[r, col] = 0.0
            print(f"\nR{r + 1} = R{r + 1} - ({factor:.4g})·R{col + 1}")
            show(clean(m))
    return clean(m)


def back_substitution(m):
    """Return the solution vector, or None after explaining why not."""
    n = len(m)

    for i in range(n):
        for j in range(i):
            if abs(m[i, j]) > EPS:
                error(f"The matrix is not upper triangular yet "
                      f"(R{i + 1} still has a non-zero entry in column {j + 1}).")
                info("Keep applying row operations, or type 'auto'.")
                return None

    for i in range(n):
        if np.all(np.abs(m[i, :n]) < EPS):
            if abs(m[i, n]) > EPS:
                error(f"R{i + 1} reads 0 = {m[i, n]:.4g} - the system has NO solution.")
            else:
                error(f"R{i + 1} is all zeros - the system has INFINITELY many solutions.")
            return None

    sol = np.zeros(n)
    for i in range(n - 1, -1, -1):
        sol[i] = (m[i, n] - m[i, i + 1:n] @ sol[i + 1:]) / m[i, i]
    return sol


# ---------------------------------------------------------------- UI
def run():
    clear()
    show_banner(ge)

    n = read_int("Number of variables (2-6) [3] : ", minimum=2, maximum=6, default=3)
    names, matrix = parse_system(n)
    original = matrix.copy()
    history = []

    print("\nInitial matrix:")
    show(matrix)
    print(HELP)

    while True:
        command = input("Enter row operation (q to finish): ").strip()
        low = command.lower()

        if low in ("q", "quit", "done"):
            break
        if low == "help":
            print(HELP)
            continue
        if low == "undo":
            if history:
                matrix = history.pop()
                print("\nUndone. Current matrix:")
                show(matrix)
            else:
                info("Nothing to undo.")
            continue
        if low == "auto":
            history.append(matrix)
            matrix = auto_eliminate(matrix)
            print("\nElimination complete.")
            continue
        if not command:
            continue

        try:
            result = apply_command(matrix, command)
        except ValueError as exc:
            error(str(exc))
            continue
        history.append(matrix)
        matrix = result
        print("\nUpdated matrix:")
        show(matrix)

    print("\nFinal matrix:")
    show(matrix)

    solution = back_substitution(matrix)
    if solution is not None:
        rule()
        success("Solution")
        for name, value in zip(names, solution):
            print(f"  {name} = {C.BOLD}{value:.4f}{C.RESET}")
        residual = np.max(np.abs(original[:, :-1] @ solution - original[:, -1]))
        print(f"{C.DIM}  check: largest residual in the original equations = {residual:.2e}{C.RESET}")
        rule()
    pause()


if __name__ == "__main__":
    try:
        run()
    except KeyboardInterrupt:
        print()