from Methods import Bisectionmethod
from Methods import Gausselimination
from Methods import Newtonraphson
from Methods import Regulafalsi
from utils import C, clear, error

nm = """
███╗   ██╗██╗   ██╗██╗███╗   ███╗███████╗██████╗ ██╗ ██████╗ █████╗ ██╗         ███╗   ███╗███████╗████████╗██╗  ██╗ ██████╗ ██████╗ ███████╗
████╗  ██║██║   ██║██║████╗ ████║██╔════╝██╔══██╗██║██╔════╝██╔══██╗██║         ████╗ ████║██╔════╝╚══██╔══╝██║  ██║██╔═══██╗██╔══██╗██╔════╝
██╔██╗ ██║██║   ██║██║██╔████╔██║█████╗  ██████╔╝██║██║     ███████║██║         ██╔████╔██║█████╗     ██║   ███████║██║   ██║██║  ██║███████╗
██║╚██╗██║██║   ██║██║██║╚██╔╝██║██╔══╝  ██╔══██╗██║██║     ██╔══██║██║         ██║╚██╔╝██║██╔══╝     ██║   ██╔══██║██║   ██║██║  ██║╚════██║
██║ ╚████║╚██████╔╝██║██║ ╚═╝ ██║███████╗██║  ██║██║╚██████╗██║  ██║███████╗    ██║ ╚═╝ ██║███████╗   ██║   ██║  ██║╚██████╔╝██████╔╝███████║
╚═╝  ╚═══╝ ╚═════╝ ╚═╝╚═╝     ╚═╝╚══════╝╚═╝  ╚═╝╚═╝ ╚═════╝╚═╝  ╚═╝╚══════╝    ╚═╝     ╚═╝╚══════╝   ╚═╝   ╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝
"""

METHODS = {
    "1": ("Bisection Method", Bisectionmethod.run),
    "2": ("Regula-Falsi Method", Regulafalsi.run),
    "3": ("Newton-Raphson Method", Newtonraphson.run),
    "4": ("Gauss Elimination Method", Gausselimination.run),
}


def main():
    while True:
        clear()
        print(C.CYAN + nm + C.RESET)
        for key, (title, _) in METHODS.items():
            print(f"  {key} : {title}")
        print("  0 : Exit")
        print(f"{C.DIM}  (Ctrl+C inside a method returns here; Ctrl+C here exits){C.RESET}\n")

        try:
            choice = input("Enter a number : ").strip()
        except (KeyboardInterrupt, EOFError):
            break

        if choice == "0":
            break
        if choice not in METHODS:
            error("Invalid input")
            try:
                input("Press Enter to try again...")
            except (KeyboardInterrupt, EOFError):
                break
            continue

        try:
            METHODS[choice][1]()
        except KeyboardInterrupt:
            continue          # back to the menu
        except EOFError:
            break

    clear()
    print("Byee")


if __name__ == "__main__":
    main()