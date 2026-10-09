from sympy import diff, lambdify

from utils import (C, EvaluationError, clear, error, evaluate, pause,
                   print_result, read_float, read_function, read_int, rule,
                   show_banner, x)

nr = """
███╗   ██╗███████╗██╗    ██╗████████╗ ██████╗ ███╗   ██╗      ██████╗  █████╗ ██████╗ ██╗  ██╗███████╗ ██████╗ ███╗   ██╗
████╗  ██║██╔════╝██║    ██║╚══██╔══╝██╔═══██╗████╗  ██║      ██╔══██╗██╔══██╗██╔══██╗██║  ██║██╔════╝██╔═══██╗████╗  ██║
██╔██╗ ██║█████╗  ██║ █╗ ██║   ██║   ██║   ██║██╔██╗ ██║█████╗██████╔╝███████║██████╔╝███████║███████╗██║   ██║██╔██╗ ██║
██║╚██╗██║██╔══╝  ██║███╗██║   ██║   ██║   ██║██║╚██╗██║╚════╝██╔══██╗██╔══██║██╔═══╝ ██╔══██║╚════██║██║   ██║██║╚██╗██║
██║ ╚████║███████╗╚███╔███╔╝   ██║   ╚██████╔╝██║ ╚████║      ██║  ██║██║  ██║██║     ██║  ██║███████║╚██████╔╝██║ ╚████║
╚═╝  ╚═══╝╚══════╝ ╚══╝╚══╝    ╚═╝    ╚═════╝ ╚═╝  ╚═══╝      ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚═╝  ╚═══╝
"""


def run():
    clear()
    show_banner(nr)

    expr, f = read_function()
    df = lambdify(x, diff(expr, x))
    xn = read_float("Enter first guess : ")
    maxi = read_int("Max iterations [100] : ", minimum=1, default=100)
    tol = read_float("Tolerance [1e-6] : ", default=1e-6, positive=True)

    try:
        rule()
        dlabel = "f'(x)"
        print(f"{'Iter':>5} {'x':>16} {'f(x)':>14} {dlabel:>14}")
        rule("─")

        for i in range(1, maxi + 1):
            fx = evaluate(f, xn)
            derivative = evaluate(df, xn)
            print(f"{i:>5} {xn:>16.8f} {fx:>14.3e} {derivative:>14.3e}")

            if abs(derivative) < 1e-12:
                error("Derivative became zero. Newton-Raphson cannot continue.")
                return pause()

            x_new = xn - fx / derivative
            if abs(x_new - xn) < tol:
                print_result(x_new, evaluate(f, x_new), i, True, f"|Δx| < {tol:g}")
                return pause()
            xn = x_new

        error("Newton-Raphson method did not converge.")
        print(f"  Last estimate : {xn:.6f}")
    except EvaluationError as exc:
        error(str(exc))
    pause()


if __name__ == "__main__":
    try:
        run()
    except KeyboardInterrupt:
        print()