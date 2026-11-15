#!/usr/bin/env python3
"""Die Bildschirmabzüge auf eine Palette bringen — halbe Größe, gleiches Bild.

# Warum es diesen Schritt gibt

**Seit Phase 4.10 liegt eine Pergamenttextur unter der Oberfläche**, und sie
ist feines Rauschen. PNG sagt jedem Punkt seinen Unterschied zum Nachbarn
voraus; bei Rauschen trifft keine Vorhersage, und die Kompression läuft ins
Leere. Gemessen am 2026-09-07 an `oberflaeche-start.png`: **81 KB vor der
Textur, 1.285 KB danach** — dieselben 1440 × 726 Punkte, sechzehnfache Größe.
Über 54 Abzüge sind das **70 MB**, und das Repositorium wog vorher 63 MB.

# Warum eine Palette und keine geringere Auflösung

**Weil ein Abzug lesbar bleiben muss.** Die Schrift im Bild ist der Grund, aus
dem es das Bild gibt; sie kleiner zu rechnen macht sie unscharf. Eine Palette
lässt jeden Punkt an seinem Platz und nimmt nur Farben weg, die niemand
auseinanderhält: 256 genügen für eine Oberfläche aus Flächen, Text und einer
Textur in **einem** Farbton.

**Gemessen, nicht geschätzt** (dieselbe Datei):

    256 Farben   627 KB   48 %
    128 Farben   596 KB   46 %
     64 Farben   544 KB   42 %
     32 Farben   478 KB   37 %

**256 ist gewählt**, weil darunter der Gewinn klein wird und das Risiko
sichtbarer Streifen in der Textur steigt. Halbe Größe für ein Bild, das neben
dem Original nicht auffällt.

# Und eine Retina-Aufnahme auf die Punktgröße

**Seit 0.6.0 kommen die Abzüge vom nativen Fenster** (`readmebilder`), und
`screencapture` nimmt in Retina-Auflösung auf: ein Fenster von 1440 × 841
Punkten wird ein Bild von 2880 × 1682. Auf die Palette gebracht wog eines
davon noch rund 2,5 MB — über 64 Abzüge 160 MB. **Halbiert ist es ein
Bildpunkt je Punkt**, so wie die Browserabzüge davor; die Schrift bleibt so
lesbar wie auf einem Bildschirm ohne Retina. Halbiert wird, was breiter als
2000 Bildpunkte ist, und zwar vor der Palette.

# Was dieses Skript nicht ist

**Kein Teil der Pipeline.** Es braucht Pillow, und Pillow ist keine gepinnte
Abhängigkeit dieses Projekts (§1). Wer es nicht hat, lässt den Schritt aus —
die Abzüge sind dann größer und sonst gleich. Genau deshalb steht der Grund
hier und nicht in einer Zusage.

    python3 scripts/abzuegeschrumpfen.py docs/bilder
"""

import sys
from pathlib import Path

from PIL import Image

FARBEN = 256

# RETINA ist die Breite, ab der ein Abzug als Retina-Aufnahme gilt.
RETINA = 2000


def schrumpfe(datei: Path) -> tuple[int, int]:
    """Bringt eine Datei auf die Palette und gibt Größe vorher und nachher."""
    vorher = datei.stat().st_size

    with Image.open(datei) as bild:
        # **`RGB` zuerst**: ein Abzug kommt als RGBA, und eine Palette mit
        # Alphakanal ist größer als eine ohne — die Oberfläche ist deckend.
        rgb = bild.convert("RGB")
        if rgb.width > RETINA:
            rgb = rgb.resize((rgb.width // 2, rgb.height // 2), Image.Resampling.LANCZOS)
        auf_palette = rgb.convert(
            "P", palette=Image.Palette.ADAPTIVE, colors=FARBEN
        )
        auf_palette.save(datei, optimize=True)

    return vorher, datei.stat().st_size


def main() -> int:
    ort = Path(sys.argv[1] if len(sys.argv) > 1 else "docs/bilder")

    dateien = sorted(ort.glob("*.png"))
    if not dateien:
        print(f"keine PNG unter {ort}", file=sys.stderr)

        return 1

    ganz_vorher = ganz_nachher = 0

    for datei in dateien:
        vorher, nachher = schrumpfe(datei)
        ganz_vorher += vorher
        ganz_nachher += nachher

    anteil = ganz_nachher * 100 // ganz_vorher
    print(
        f"{len(dateien)} Abzüge: {ganz_vorher // 1024} KB → "
        f"{ganz_nachher // 1024} KB ({anteil} %)"
    )

    return 0


if __name__ == "__main__":
    sys.exit(main())
