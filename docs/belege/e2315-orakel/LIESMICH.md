# E2315 — der Maßstab der sechs Kanonkarten bei MSE2, gefragt am Orakel

**Wofür**: Beleg der Widerlegung in
`E2315-2315.md` (Entwicklungsarchiv) — MSE2
verkleinert die sechs Karten aus #2315 nicht. Nachgereicht nach Befund 13 des
Zweitblicks über Phase 1 von 0.9.0: bis dahin stand nur die Tafel da, ohne
Patch, Aufruf und Rohausgabe.

## Was hier liegt

| Datei | Inhalt |
| --- | --- |
| `instrumentierung.patch` | die Instrumentierung gegen MSE2 `f9d9356` (2021-07-06), drei Dateien: je Textfeld und Satz Start- und Endmaßstab, Zeilen, Schriftgröße und Kasten auf die Fehlerausgabe, je Karte ihr Name |
| `auszug.tsv` | jede Zeile der sechs Karten aus dem Lauf — 89 Zeilen |

**Die vollständigen Protokolle liegen nicht hier** — drei Dateien zu je rund
5,7 MB. Ihre Abdrücke stehen unten; sie entstehen mit Patch und Aufruf neu.

## Der Aufruf

```bash
cd /Users/david/mse2-quelle            # f9d9356, gebaut nach Vermerk 0026
patch -p1 < instrumentierung.patch
(cd build && make magicseteditor)
Z=$(mktemp -d); mkdir -p "$Z/Library/Application Support/magicseteditor"
for satz in Black Red White; do
  HOME="$Z" CFFIXED_USER_HOME="$Z" MSE_ORAKEL_MASSSTAB=1 \
    build/magicseteditor --export-images "<kanon>/$satz.mse-set" "<ziel>/$satz/{card.name}.png" \
    2> "<ziel>/$satz.log" > /dev/null
done
git checkout -- src/render/text/viewer.cpp src/render/value/text.cpp src/data/format/image.cpp
```

**Mit leerem Zuhause**, wie jeder Lauf von MSE2 hier; danach ist der Baum
wieder auf `f9d9356` und die ursprüngliche Binärdatei zurück
(`6c414114a57c6955…`).

## Die Abdrücke

| Was | SHA-256, die ersten 16 Stellen |
| --- | --- |
| `Black.mse-set` (Kanon, Stand 2022-10-24) | `8631f57ce4750b94` |
| `Red.mse-set` | `a200c2574002f66e` |
| `White.mse-set` | `ac4cb04c6f05119a` |
| die instrumentierte Binärdatei | `b8aeeb2bc5424a39` |
| Black — `KARTE`- und `MASSSTAB`-Zeilen, Pfade auf den Dateinamen gekürzt (56.317 Zeilen) | `80a8f14fdf35a3ea` |
| Red — dito (55.553 Zeilen) | `1f83a0e44fe61328` |
| White — dito (55.823 Zeilen) | `475bbaad969cd8c5` |

## Zweimal gelaufen

**Am 2026-10-14 vormittags** zur Untersuchung, **und am Nachmittag noch
einmal** mit dem aus dem Sitzungsverlauf rekonstruierten Patch: alle 167.693
Zeilen der drei Sets stimmen überein. Der Patch hier ist also der, mit dem die
Tafel entstand.

## Was der Auszug zeigt

**Alle sechs zeichnen `text` in Maßstab 1** (`start` und `ende` 1,000000),
mit 14 Punkten — Cruel Revival in 6 Zeilen, Bleed Dry 5, Noxious Ghoul 4,
Liliana, Death's Majesty 8 (Kasten 300 × 146), Minotaur de Force 5,
Aerialephant 6 (je 314 × 154). **Einen zweiten Satz gibt es**, aber nie für
`text`: für Illustrator, Name, Manakosten und Lilianas Doppelpunkte und
Loyalitätskosten — und auch dort Maßstab 1.

**Unsere Seite** ist für die vier Karten aus Black im selben Lauf protokolliert
worden: Größe 14 ohne Verkleinerung, dieselben Zeilen, derselbe Kasten. Für
Aerialephant und Minotaur de Force liegt kein eigenes Protokoll vor; ihr
Abstand fiel mit dem Ticketsymbol, nicht mit einer Zeile.
