from utils import (C, EvaluationError, clear, error, evaluate, info, pause,
                   print_result, read_float, read_function, read_int, rule,
                   show_banner)

bisection = """
██████╗ ██╗███████╗███████╗ ██████╗████████╗██╗ ██████╗ ███╗   ██╗
██╔══██╗██║██╔════╝██╔════╝██╔════╝╚══██╔══╝██║██╔═══██╗████╗  ██║
██████╔╝██║███████╗█████╗  ██║        ██║   ██║██║   ██║██╔██╗ ██║
██╔══██╗██║╚════██║██╔══╝  ██║        ██║   ██║██║   ██║██║╚██╗██║
██████╔╝██║███████║███████╗╚██████╗   ██║   ██║╚██████╔╝██║ ╚████║
╚═════╝ ╚═╝╚══════╝╚══════╝ ╚═════╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝
"""


def run():
    clear()
    show_banner(bisection)

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
            error("f(a) and f(b) have the same sign - bisection can not be applied.")
            info("Choose an interval where the function changes sign.")
            return pause()

        rule()
        print(f"{'Iter':>5} {'a':>12} {'b':>12} {'c':>12} {'f(c)':>13}")
        rule("─")

        c = fc = None
        for i in range(1, maxi + 1):
            c = (a + b) / 2
            fc = evaluate(f, c)
            print(f"{i:>5} {a:>12.6f} {b:>12.6f} {c:>12.6f} {fc:>13.3e}")

            if fc == 0 or (b - a) / 2 < tol:
                print_result(c, fc, i, True,
                             "exact root" if fc == 0 else f"interval < {tol:g}")
                return pause()

            if fa * fc < 0:
                b = c
            else:
                a, fa = c, fc

        print_result(c, fc, maxi, False)
    except EvaluationError as exc:
        error(str(exc))
    pause()


if __name__ == "__main__":
    try:
        run()
    except KeyboardInterrupt:
        print()