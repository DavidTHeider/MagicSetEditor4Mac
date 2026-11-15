# 0165 — `card_regions` aus MSE 2.5, weil das Pack für 2.5 geschrieben ist

**Stand:** angenommen am 2026-10-03 · QA-04-F3
(#2039),
Folgebefund aus #2006.
Marke `adapted`.

## Die Frage

In der Ausfuhr waren `set.stylesheet` und `card.stylesheet` leer. Drei
Exportvorlagen des Packs — `magic-cockatrice-v2`, `magic-lackeyccg`,
`magic-lackeybot` — lesen daraus `card_regions` und wählen damit zwischen einem
Bild je Karte und mehreren (geteilte, gefaltete, verschmolzene Karten).

## Was die Referenz sagt — und warum sie hier nicht reicht

**Die Referenz dieses Projekts ist MSE2 in der Fassung f9d9356 vom
2021-07-06** (§15). Dort ist `set.stylesheet` ein Gegenstand mit den Gliedern
aus `IMPLEMENT_REFLECTION(StyleSheet)` (`src/data/stylesheet.cpp`, Zeilen
94–126) — **`card_regions` gehört nicht dazu**. Gefragt am Orakel über
`Multicolor.mse-set`: `set.stylesheet.card_regions or else "kein Glied"` gibt
`"kein Glied"`, `length(set.stylesheet.card_regions)` bricht mit *stylesheet has
no member 'card_regions'* ab. Die Lackey-Ausfuhr endet in f9d9356 mit Exit 1 und
schreibt **keine Datei** (gemessen am 2026-10-03).

**Das Pack ist für MSE 2.5.8 geschrieben**: f9d9356 warnt beim Laden von
`magic.mse-game`, die Datei sei mit einer neueren Fassung entstanden. In der
Fortführung von MSE2 (`G-e-n-e-v-e-n-s-i-S/MagicSetEditor2`, Fassung a4b79c9
vom 2026-09-04) reflektiert `StyleSheet` das Glied: `vector<CardRegionP>
card_regions` (`src/data/stylesheet.hpp`, Zeile 44), je Region `name`, `x`,
`y`, `width`, `height` (`src/data/card_region.cpp`, Zeilen 19–24) — **leer**,
solange ein Stil keine erklärt. **Acht Stile im Pack erklären Regionen** — je
Region ein `card region:` im Kopf, etwa `magic-m15-mainframe-dfc` mit `front`
und `back` nebeneinander (Berichtigung unten).

## Die Entscheidung

**`set.stylesheet` und `card.stylesheet` sind in der Ausfuhr Gegenstände wie
in f9d9356**, und **`card_regions` steht daran wie in 2.5**: die Liste der
Regionen, die der Stil erklärt, leer ohne.
`card.stylesheet` ist `nil`, solange die Karte keinen eigenen Stil trägt
(`REFLECT(stylesheet)`, `src/data/card.cpp`, Zeile 77, Fassung f9d9356 vom
2021-07-06). Eine Karte mit einem Stil ohne Regionen nimmt das Einzelbild,
eine mit Regionen die zugeschnittenen Bilder — **wie 2.5** —, und die
Lackey-Ausfuhr läuft, wo f9d9356 abbricht.

**Warum nicht f9d9356 nachbilden**: der Abbruch ist ein Versionsbruch zwischen
Pack und Programm, keine gewollte Semantik. Ihn nachzubauen hieße, eine
Ausfuhr zu verweigern, die mit dem Programm gelingt, für das das Pack
geschrieben ist.

## Was nicht dazugehört

**Die übrigen Glieder von `StyleSheet`** — `game`, `icon`, `version`, die
Stilfelder: keine Exportvorlage des Packs liest sie.

## Berichtigung am 2026-10-03 — N1 des Nachlaufs über die QA-Arbeit

**Hier stand „kein Stil im Pack erklärt Regionen", und `card_regions` war
immer leer.** Gesucht war nach `card regions`; geschrieben steht, wie MSE
eine Liste in einer Datei führt, **je Eintrag ein `card region:`**. Acht Stile
tun es (zweiseitige Karten und `meld-3in1`). Die Begründung „leer gibt das
Ergebnis von 2.5" trug damit für diese acht nicht — sie nahmen das Einzelbild,
wo 2.5 zuschneidet. **Jetzt liest `msepaket` die Regionen**
(`msepaket.Kartenregion`, gegen ein eigenes Paket in `testdaten/`), und die
Ausfuhr gibt sie als Liste von Gegenständen mit `name`, `x`, `y`, `width`,
`height`. Gefunden hat es der Nachlauf; die Reichweite der ersten Suche war
die eines Wortes, nicht die des Packs (§10).

## Verworfen

- **Weiter leer lassen.** `set.stylesheet` sah aus wie „kein Stil", und eine
  Vorlage, die `full_name` liest, bekam einen Fehler statt des Namens.
- **f9d9356 nachbilden** (Abbruch) — siehe oben.
