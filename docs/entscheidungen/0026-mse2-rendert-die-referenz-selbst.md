# 0026 — MSE2 rastert die Referenz selbst

**Stand:** gültig
**Datum:** 2026-08-18

## Lage

Der Maßstab dieses Projekts ist Parität zu MSE2. Gemessen wurde bis heute gegen
**27 Bilder aus der Gemeinschaft**, deren Vorlage unbekannt ist — und
nachweislich eine andere als die, die unsere Rekonstruktionen nennen: der Rand
misst dort 17/18 statt 14/15 Bildpunkte, unter dem Textkasten liegt Textur statt
Schwarz, und das Seltenheitssymbol trägt den falschen Verlauf
(#375,
#376,
#378).

Die Zahl daneben war deshalb ein **Verlaufsmaß** und keine Aussage über Parität.
Der Satz stand in jeder Roadmap: „Wer Parität behauptet, belegt sie an etwas
anderem."

## Entscheidung

**MSE2 rastert die Referenz selbst.** Auf diesem Rechner, aus dem Quelltext
gebaut, headless über seine eigene Kommandozeile:

```
--export-images FILE [IMAGE]
        Export the cards in a set to image files
```

`src/main.cpp`, Zeilen 191–193 und der Zweig ab Zeile 235, Fassung f9d9356 vom
2021-07-06.

Damit stammen **Eingabe und Ausgabe** von MSE2: die zehn kanonischen Sets sind
seine eigenen, und die Bilder daneben hat es selbst gezeichnet. Der Vergleich
braucht kein ausgespartes Bildfenster mehr — die Illustration ist in beiden
Bildern dieselbe.

## Wie es gebaut wurde

MSE2 hat eine eigene macOS-Anleitung (`README.md`, „Building on Mac"):

```
brew install boost wxwidgets hunspell
cmake -S . -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build
```

Drei Dinge kamen dazu, und sie gehören genannt:

1. **wxWidgets 3.3.3 statt 3.0.5.** Die Anleitung nennt 3.0.5 als geprüft;
   Homebrew liefert heute 3.3.3. Es übersetzt trotzdem — 100 Prozent, ohne
   einen einzigen Fehler.
2. **Hunspell muss von Hand gezeigt werden**, weil `pkg-config` fehlt:
   `-DHUNSPELL_LIBRARIES=/opt/homebrew/opt/hunspell/lib/libhunspell-1.7.dylib
   -DHUNSPELL_INCLUDE_DIRS=/opt/homebrew/opt/hunspell/include`.
3. **Acht Zeilen Patch**, sonst linkt es auf arm64 nicht — `docs/mse2-macos.patch`.

### Was der Patch tut, und was er nicht tut

Der Linker vermisst `GetDefaultMember::handle<unsigned long>`. Der Grund ist ein
**Plattform-Typedef**, keine Fachlogik:

| | `size_t` | `uint64_t` |
| --- | --- | --- |
| Linux | `unsigned long` | `unsigned long` |
| macOS/arm64 | `unsigned long` | `unsigned long long` |

`get_member.cpp` prägt `unsigned int` und `uint64_t` aus. Unter Linux deckt die
zweite `size_t` mit ab, auf Apple nicht — dort fehlt sie. Der Patch fügt sie
hinzu, **wortgleich** zu den beiden daneben (`value = to_script((int)v)`), hinter
`#if defined(__APPLE__)`.

**Keine Verhaltensänderung.** Auf Linux und Windows entstünde derselbe Code; die
Referenz bleibt MSE2, nicht ein verändertes MSE2. Der Patch liegt im
Repositorium, damit das nachprüfbar ist und nicht geglaubt werden muss.

## Die Schriften — der Fallstrick, der die erste Messung verdorben hat

**MSE2 fragt wxWidgets nach dem Namen einer Schrift.** Liegt sie auf dem Rechner
nicht, nimmt wx eine Ersatzschrift. Auf diesem Mac war **keine** der
Paketschriften installiert:

| gesucht | im System |
| --- | ---: |
| MPlantin | 0 |
| Beleren | 0 |
| Relay-Medium | 0 |

Die erste Referenz zeigte deshalb serifenlosen Text, wo MSE2 unter Windows
MPlantin setzt — und wo unser eigenes Programm es ebenfalls tut, weil es die
Dateien **aus dem Paket** liest (#140)
und nicht aus dem System.

**Der Unterschied ging voll in die Messung.** Nachdem die Schriften unter
`~/Library/Fonts/mse2-referenz/` lagen, fiel der Abstand über `Archenemy` von
**25,190 auf 20,018** — ein Fünftel war der Fehler der Referenz, nicht der
unsere.

**`scripts/mse2-referenz.sh` prüft das jetzt** und bricht ab, statt eine
verdorbene Referenz zu erzeugen.

**Rückgängig** macht es, wer `~/Library/Fonts/mse2-referenz/` löscht.

> **Wie gemessen wird, steht in [`MESSMETHODIK.md`](../MESSMETHODIK.md)** —
> das Maß, die Stichprobe je Schicht und der geschichtete Schätzer, jeweils
> mit ihren Gleichungen. Dieser Vermerk sagt, **woher die Referenz kommt**;
> jener sagt, **wie daraus eine Zahl wird**.

## Woher die kanonischen Sets stammen

**Nachgetragen am 2026-09-07**, auf Anweisung des Auftraggebers. Hier stand die
Herkunft nirgends — und ohne sie ist die Referenz eine Datei, deren Anspruch
niemand nachprüfen kann.

| | |
| ---: | --- |
| **Bezugsort** | <https://archive.org/details/mse-canon-cards-2022-10-24> |
| **Angekündigt** | im MSE-Forum, [Thread 2128 „Canon Magic Cards"](https://magicseteditor.boards.net/thread/2128/canon-magic-cards) |
| **Erstellt von** | Forumsnutzer *GenevensiS*, Beitrag vom 2022-10-25 |
| **Datenquelle** | **Scryfall**, Stand **2022-10-24** — Karten **und** Bilder |
| **Umfang** | zehn Dateien, 1,8 GB, **25.148** Karten; sechs davon über 227 MB |
| **Geladen** | 2026-08-17 um 21:24 über Safari (Quarantänemarke der Dateien) |

**Nach Farbe getrennt, und das ist kein Zufall**: der Ersteller schreibt, MSE
habe „a bit of trouble showing everything at once". Daher `White`, `Blue`,
`Black`, `Red`, `Green`, `Colorless`, `Multicolor` und die drei Sonderformen
`Archenemy`, `Planechase` und `Vanguard`.

**Sie brauchen elf Vorlagen aus dem vollen Pack**, unter anderem
`magic-m15-altered` (im Forum „M15 Mainframe"), `magic-m15-mainframe-dfc`,
`magic-m15-saga`, `magic-archenemy` und `magic-m15-vanguard`. Ohne sie lassen
sich die Sets nicht öffnen — genau deshalb bricht der Kanonlauf ohne
`MSE_FREMDPAKETE` ab.

### Was der Ersteller selbst einschränkt

Zwei Sätze aus dem Beitrag, und beide gehören zur Reichweite jeder Zahl, die
gegen diese Sets gemessen wird:

> „There is bound to be some inaccuracies and maybe missing cards."

> „Silver bordered cards will likely be very off, I don't plan on fixing those
> though."

**Damit ist die Referenz kein Kanon im strengen Sinn.** Sie ist der beste
verfügbare Maßstab: dieselben Sets, die MSE2 selbst öffnet und rastert. Ein
Abstand daran misst, wie weit **unsere** Rasterung von **MSE2s** Rasterung
**derselben** Datei entfernt ist — nicht, ob die Datei die Karte richtig
beschreibt.

**Und die Bilder sind Scryfall-Bilder**, in der Auflösung, die Scryfall
ausliefert. Der Ersteller sagt es selbst: „so it's not the best quality". Für
den Abstand ist das ohne Belang — beide Seiten zeichnen dasselbe Bild —, für
eine Aussage über Druckqualität wäre es eine.

## Was die Zahl sagt — und was nicht

Erste Messung, 2026-08-18, über `Archenemy` (70 Karten, davon 60 vergleichbar):

| | |
| ---: | --- |
| **20,018** | mittlerer Abstand über die **ganze** Karte |
| 16,021 | bester |
| 25,754 | schlechtester |
| 1,0 | die Toleranz, an der gemessen wird |

Und über `White` — die **gewöhnlichen** Karten, 3.666 von 3.761 über
`m15-altered`: sie liegen deutlich weiter weg als die Schemen. Eine Zahl über
`Archenemy` allein sagt nichts über die Karten, die jemand tatsächlich baut.

**Sie sagt nicht, dass die Karten gleich aussehen.** Sie sagt, wie weit sie in
Bildpunkten auseinanderliegen — beim Fünfundzwanzigfachen der Toleranz. Was ein
Mensch davon sieht, ist eine andere Frage; die beantwortet der Blick.

**Und sie sagt nichts über die zehn nicht verglichenen Karten.** Ihre Namen
tragen Zeichen, die MSE2 im Dateinamen anders schreibt als wir beim Suchen. Das
ist eine Lücke der Messung, keine der Karten.

## Preis

**Ein zweites Programm muss gebaut sein.** Wer die Referenz erneuern will,
braucht Homebrew, wxWidgets und zehn Minuten Übersetzungszeit. Der Lauf
überspringt sich selbst, wenn die Bilder fehlen — er wird nicht rot, wo nichts
zu messen ist.

**Die Bilder kommen nicht ins Repositorium** (§11). Sie liegen unter
`~/Library/Application Support/MagicSetEditor4Mac/mse2-ausgabe/`, neben den
kanonischen Sets und den bezogenen Paketen.
