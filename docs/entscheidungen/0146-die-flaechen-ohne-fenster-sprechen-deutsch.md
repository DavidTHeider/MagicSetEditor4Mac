# 0146 — Die Flächen ohne Fenster sprechen Deutsch

**Stand:** entschieden am 2026-09-29 · LOC-01.AK-02 und AK-04
(#1687)

## Die Frage

LOC-01 verlangt acht Sprachen für „UI, native Menüs, Dialoge, Hilfen,
Validierung, CLI-Hilfe und zugängliche Labels" und lässt
„plattformspezifische Ausnahmen" zu, wenn sie **ausdrücklich** dokumentiert
sind. Das Fenster, die Menüleiste und die Dialoge sprechen seit LOC-01 acht
Sprachen. Offen war, was Kommandozeile, REST und MCP sprechen.

## Was gilt

**Sie sprechen Deutsch** — Hilfe, Meldungen, Beschreibungen der
MCP-Werkzeuge. So war es schon vor LOC-01: auch Englisch führten sie nie.

**Was ein Programm liest, hängt an keiner Sprache.** Jede Fehlerantwort trägt
einen Code — `code` bei REST (`internal/rest/antwort.go`, „maschinenlesbar und
bleibt"), `code` und `schluessel` im Fenster und bei MCP
(`internal/app/fehler.go`) —, und die Kommandozeile endet mit einem
Ausgangscode. Der Satz daneben ist für Menschen und darf sich ändern; ein
Skript, das ihn ausliest, hat sich an den falschen Teil gehalten.

## Warum nicht acht Sprachen

- **Die Hilfe der Kommandozeile ist ein Nachschlagewerk über Syntax.**
  Rund sechshundert Zeilen, und jede neue Fähigkeit zieht eine Zeile nach
  (§16, dieselbe Fähigkeit überall). Acht Fassungen machten aus jeder neuen
  Zeile acht, und eine maschinelle Übersetzung einer Argumentliste ist genau
  die Stelle, an der ein Wort verrutscht und ein Befehl nicht mehr stimmt.
- **Die häufigsten Leser sind Programme und Sprachmodelle.** Ein
  Sprachmodell liest Deutsch; ein Skript liest den Code.
- **Es gibt keine Sprachwahl ohne Fenster.** Die Einstellung `sprache` ist
  eine des Fensters; eine zweite für die Kommandozeile — über `LANG`, über
  einen Schalter — wäre eine neue Einstellung mit eigener Rangfolge, und die
  hat niemand verlangt.

## Was das kostet

**Wer kein Deutsch liest, liest an der Kommandozeile eine Sprache, die er
nicht spricht.** Die Codes helfen einem Skript, nicht einem Menschen. Kommt
die Frage von Nutzenden, ist der Weg ein Katalog je Fläche mit derselben
Schlüsselprüfung wie im Fenster — und diese Entscheidung wird dann
fortgeschrieben, nicht umgangen.

## Verworfen

- **Englisch als zweite Sprache der Flächen.** Es wäre die Hälfte der
  Arbeit für einen Bruchteil der Leser und ließe die übrigen sechs so stehen,
  wie sie heute stehen.
