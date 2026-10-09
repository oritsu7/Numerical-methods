"""Shared helpers for the Numerical Methods toolkit."""
import math
import os
import sys

import numpy as np
from sympy import symbols, lambdify
from sympy.parsing.sympy_parser import (
    parse_expr,
    standard_transformations,
    implicit_multiplication_application,
    convert_xor,
)

x = symbols("x")

# Allow "2x", "x sin(x)" and "x^2" in addition to "2*x" and "x**2"
FUNC_TRANSFORMS = standard_transformations + (
    implicit_multiplication_application,
    convert_xor,
)

if os.name == "nt":
    os.system("")  # enables ANSI colours in Windows terminals


class C:
    """ANSI colours (disabled automatically when output is not a terminal)."""
    _on = sys.stdout.isatty()
    RESET = "\033[0m" if _on else ""
    BOLD = "\033[1m" if _on else ""
    DIM = "\033[2m" if _on else ""
    RED = "\033[91m" if _on else ""
    GREEN = "\033[92m" if _on else ""
    YELLOW = "\033[93m" if _on else ""
    CYAN = "\033[96m" if _on else ""


class EvaluationError(Exception):
    """Raised when f(x) cannot be evaluated to a finite real number."""


def clear():
    if sys.stdout.isatty():
        os.system("cls" if os.name == "nt" else "clear")


def show_banner(art):
    print(C.CYAN + art + C.RESET)


def error(msg):
    print(f"{C.RED}✗ {msg}{C.RESET}")


def info(msg):
    print(f"{C.YELLOW}• {msg}{C.RESET}")


def success(msg):
    print(f"{C.GREEN}✓ {msg}{C.RESET}")


def rule(char="═", width=60):
    print(C.DIM + char * width + C.RESET)


def pause(msg="\nPress Enter to return to the menu..."):
    input(msg)


def read_float(prompt, default=None, positive=False):
    """Read a number. Accepts expressions such as 'pi/2'. Blank -> default."""
    while True:
        raw = input(prompt).strip()
        if raw == "" and default is not None:
            return default
        try:
            try:
                value = float(raw)
            except ValueError:
                value = float(parse_expr(raw, transformations=FUNC_TRANSFORMS))
            if not math.isfinite(value):
                raise ValueError
            if positive and value <= 0:
                error("Please enter a value greater than 0.")
                continue
            return value
        except Exception:
            error("Please enter a valid number.")


def read_int(prompt, minimum=1, maximum=None, default=None):
    while True:
        raw = input(prompt).strip()
        if raw == "" and default is not None:
            return default
        try:
            value = int(raw)
        except ValueError:
            error("Please enter a whole number.")
            continue
        if value < minimum or (maximum is not None and value > maximum):
            if maximum is not None:
                error(f"Enter a number between {minimum} and {maximum}.")
            else:
                error(f"Enter {minimum} or more.")
            continue
        return value


def read_function(prompt="f(x) : "):
    """Read f(x) from the user. Returns (sympy_expr, python_callable)."""
    while True:
        text = input(prompt).strip()
        if not text:
            continue
        try:
            expr = parse_expr(text, local_dict={"x": x}, transformations=FUNC_TRANSFORMS)
        except Exception:
            error("Could not understand that expression. Example: x^3 - 2x - 5")
            continue
        extra = expr.free_symbols - {x}
        if extra:
            names = ", ".join(sorted(str(s) for s in extra))
            error(f"Unknown symbol(s): {names}. Use only x "
                  f"(write E for Euler's number, or exp(x)).")
            continue
        if x not in expr.free_symbols:
            error("f(x) must contain x.")
            continue
        return expr, lambdify(x, expr)


def evaluate(f, value):
    """Evaluate f at value and return a finite float, or raise EvaluationError."""
    try:
        with np.errstate(all="ignore"):
            result = float(f(value))
    except (ArithmeticError, ValueError, TypeError):
        raise EvaluationError(f"f({value:.6g}) is undefined (not a real number).")
    if not math.isfinite(result):
        raise EvaluationError(f"f({value:.6g}) is undefined or infinite.")
    return result


def print_result(root, f_root, iterations, converged, note=""):
    rule()
    if converged:
        success(f"Converged{(' - ' + note) if note else ''}")
    else:
        info("Stopped before reaching the tolerance - result is approximate.")
    print(f"  Iterations : {iterations}")
    print(f"  Root       : {C.BOLD}{root:.6f}{C.RESET}")
    print(f"  f(root)    : {f_root:.3e}")
    rule()