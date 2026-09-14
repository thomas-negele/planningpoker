"""Measure the palette against the ratios the specification names.

WCAG 2.1: 4.5:1 for body text, 3:1 for large text and for the visible boundary
of an interactive control, a focus indicator, or a graphic that carries meaning.
Colours with alpha are composited over the backdrop they actually sit on before
being measured, because that is what an eye sees.
"""

import re
import sys


def kanal(wert: float) -> float:
    """One sRGB channel, 0..1, converted to linear light."""
    return wert / 12.92 if wert <= 0.04045 else ((wert + 0.055) / 1.055) ** 2.4


def leuchtdichte(rgb: tuple[float, float, float]) -> float:
    r, g, b = (kanal(c / 255) for c in rgb)
    return 0.2126 * r + 0.7152 * g + 0.0722 * b


def verhaeltnis(vorne: tuple, hinten: tuple) -> float:
    a, b = leuchtdichte(vorne), leuchtdichte(hinten)
    hell, dunkel = max(a, b), min(a, b)
    return (hell + 0.05) / (dunkel + 0.05)


def hex_zu_rgb(text: str) -> tuple[float, float, float]:
    text = text.strip().lstrip("#")
    if len(text) == 3:
        text = "".join(c * 2 for c in text)
    return tuple(int(text[i : i + 2], 16) for i in (0, 2, 4))


def deckung(vorne, hinten, alpha: float):
    """Composite a translucent colour over its backdrop."""
    return tuple(v * alpha + h * (1 - alpha) for v, h in zip(vorne, hinten))


def farbe(spez: str, hintergrund=None):
    """Accept '#rrggbb' or 'rgb(r g b / p%)', compositing the latter."""
    spez = spez.strip()
    if spez.startswith("#"):
        return hex_zu_rgb(spez)
    treffer = re.match(r"rgb\(\s*(\d+)\s+(\d+)\s+(\d+)\s*/\s*([\d.]+)%\s*\)", spez)
    if not treffer:
        raise ValueError(f"unbekannte Farbangabe: {spez}")
    r, g, b, p = treffer.groups()
    vorne = (int(r), int(g), int(b))
    if hintergrund is None:
        raise ValueError(f"{spez} braucht einen Hintergrund")
    return deckung(vorne, hintergrund, float(p) / 100)


def lies_werte(pfad: str) -> dict[str, str]:
    """Read the custom properties out of app.css as written."""
    text = open(pfad, encoding="utf-8").read()
    werte = {}
    for name, wert in re.findall(r"^\s*(--[a-z0-9-]+):\s*([^;]+);", text, re.M):
        werte[name] = " ".join(wert.split())
    return werte


def main() -> int:
    W = lies_werte(sys.argv[1] if len(sys.argv) > 1 else "web/src/app.css")

    hintergrund = hex_zu_rgb(W["--background"])
    flaeche = hex_zu_rgb(W["--surface"])
    erhoben = hex_zu_rgb(W["--surface-raised"])
    kartenflaeche = hex_zu_rgb(W["--card-face"])
    akzent = hex_zu_rgb(W["--accent"])
    schlechtflaeche = hex_zu_rgb(W["--bad-surface"])

    # The table is a gradient; its lightest stop is the fairest test for text on
    # it and its darkest for a graphic that must stand out from it.
    tisch_hell = hex_zu_rgb("#1a1e27")
    tisch_dunkel = hex_zu_rgb("#0f1116")

    # Text: 4.5:1.
    text_paare = [
        ("--text on the page", W["--text"], hintergrund),
        ("--text on a panel", W["--text"], flaeche),
        ("--text on a raised surface", W["--text"], erhoben),
        ("--text-dim on the page", W["--text-dim"], hintergrund),
        ("--text-dim on a panel", W["--text-dim"], flaeche),
        ("--text-dim on the table", W["--text-dim"], tisch_hell),
        ("--accent as the name of you", W["--accent"], hintergrund),
        ("--accent-text on the primary button", W["--accent-text"], akzent),
        ("--card-face-text on a revealed card", W["--card-face-text"], kartenflaeche),
        ("--bad on its own surface", W["--bad"], schlechtflaeche),
        ("--bad on the page", W["--bad"], hintergrund),
    ]

    # Controls, boundaries and meaningful graphics: 3:1.
    bahn = farbe(W["--bar-track"], tisch_hell)
    balken = akzent if W["--bar-fill"].startswith("var(") else hex_zu_rgb(W["--bar-fill"])

    objekt_paare = [
        ("--border as a control boundary on the page", W["--border"], hintergrund),
        ("--border as a control boundary on a panel", W["--border"], flaeche),
        ("--accent as a focus ring on the page", W["--accent"], hintergrund),
        ("--accent as a focus ring on a panel", W["--accent"], flaeche),
        ("the edge of a face-down card", W["--card-back-border"], hintergrund),
        ("a revealed card against the page", W["--card-face"], hintergrund),
        ("the connection dot", W["--good"], hintergrund),
        ("a tally bar against its track", balken, bahn),
    ]

    # Measured but not required. The specification asks for 3:1 from the boundary
    # of a control, a focus indicator, and a graphic that carries meaning "rather
    # than decoration". These three are decoration: a panel is identified by its
    # fill and its content, the table is the surface a tally sits on rather than
    # the tally, and the empty part of a bar says nothing the printed count does
    # not already say. They are listed so that nobody has to wonder whether they
    # were forgotten or excused.
    dekoration = [
        # The fill of a face-down card is dark by design; what makes the card
        # perceivable as an object is its edge, which is measured as required
        # above. WCAG compares adjacent colours, and the edge is the adjacency.
        ("the fill of a face-down card", W["--card-back"], hintergrund),
        ("--panel-border around a panel", W["--panel-border"], hintergrund),
        ("the table against the page", "#1a1e27", hintergrund),
        ("a tally track against the table", bahn, tisch_hell),
    ]

    fehler = 0
    for titel, paare, schwelle in (
        ("TEXT — needs 4.5:1", text_paare, 4.5),
        ("CONTROLS AND GRAPHICS — needs 3:1", objekt_paare, 3.0),
    ):
        print(f"\n{titel}")
        print("-" * 74)
        for name, vorne, hinten in paare:
            v = farbe(vorne, hinten) if isinstance(vorne, str) else vorne
            h = farbe(hinten, None) if isinstance(hinten, str) else hinten
            r = verhaeltnis(v, h)
            ok = r >= schwelle
            if not ok:
                fehler += 1
            print(f"  {'ok  ' if ok else 'FAIL'} {r:5.2f}:1   {name}")

    print("\nDECORATION — measured, not required")
    print("-" * 74)
    for name, vorne, hinten in dekoration:
        v = farbe(vorne, hinten) if isinstance(vorne, str) else vorne
        h = farbe(hinten, None) if isinstance(hinten, str) else hinten
        print(f"       {verhaeltnis(v, h):5.2f}:1   {name}")

    print()
    if fehler:
        print(f"{fehler} pair(s) below the required ratio.")
    else:
        print("Every measured pair meets the ratio the specification names.")
    return 1 if fehler else 0


if __name__ == "__main__":
    raise SystemExit(main())
