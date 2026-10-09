from utils import (C, EvaluationError, clear, error, evaluate, info, pause,
                   print_result, read_float, read_function, read_int, rule,
                   show_banner)

regula = """
██████╗ ███████╗ ██████╗ ██╗   ██╗██╗      █████╗       ███████╗ █████╗ ██╗     ███████╗██╗
██╔══██╗██╔════╝██╔════╝ ██║   ██║██║     ██╔══██╗      ██╔════╝██╔══██╗██║     ██╔════╝██║
██████╔╝█████╗  ██║  ███╗██║   ██║██║     ███████║█████╗█████╗  ███████║██║     ███████╗██║
██╔══██╗██╔══╝  ██║   ██║██║   ██║██║     ██╔══██║╚════╝██╔══╝  ██╔══██║██║     ╚════██║██║
██║  ██║███████╗╚██████╔╝╚██████╔╝███████╗██║  ██║      ██║     ██║  ██║███████╗███████║██║
╚═╝  ╚═╝╚══════╝ ╚═════╝  ╚═════╝ ╚══════╝╚═╝  ╚═╝      ╚═╝     ╚═╝  ╚═╝╚══════╝╚══════╝╚═╝
"""


def run():
    clear()
    show_banner(regula)

    _, f = read_function()
    a = read_float("a : ")
    b = read_float("b : ")
    if a == b:
        error("a and b must be different.")
        return pause()
    if a > b:
        a, b = b, a
        info(f"Swapped so that a < b  (a = {a}, b = {b})")
    maxi = read_int("Iterations : ", minimum=1)
    tol = read_float("Tolerance [1e-6] : ", default=1e-6, positive=True)

    try:
        fa, fb = evaluate(f, a), evaluate(f, b)
        print(f"\nf(a) = {fa:.6f}    f(b) = {fb:.6f}")

        if fa == 0:
            print_result(a, 0.0, 0, True, "a is an exact root")
            return pause()
        if fb == 0:
            print_result(b, 0.0, 0, True, "b is an exact root")
            return pause()
        if fa * fb > 0:
            error("f(a) and f(b) have the same sign - regula falsi can not be applied.")
            info("Choose an interval where the function changes sign.")
            return pause()

        rule()
        print(f"{'Iter':>5} {'a':>12} {'b':>12} {'c':>12} {'f(c)':>13}")
        rule("─")

        prev = c = fc = None
        for i in range(1, maxi + 1):
            c = (a * fb - b * fa) / (fb - fa)
            fc = evaluate(f, c)
            print(f"{i:>5} {a:>12.6f} {b:>12.6f} {c:>12.6f} {fc:>13.3e}")

            if fc == 0 or (prev is not None and abs(c - prev) < tol):
                print_result(c, fc, i, True,
                             "exact root" if fc == 0 else f"|Δc| < {tol:g}")
                return pause()

            if fa * fc < 0:
                b, fb = c, fc
            else:
                a, fa = c, fc
            prev = c

        print_result(c, fc, maxi, False)
    except EvaluationError as exc:
        error(str(exc))
    pause()


if __name__ == "__main__":
    try:
        run()
    except KeyboardInterrupt:
        print()