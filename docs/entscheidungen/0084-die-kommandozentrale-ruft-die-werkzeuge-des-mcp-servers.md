# 0084 — Die Kommandozentrale ruft die Werkzeuge des MCP-Servers

**Stand:** angenommen
**Datum:** 2026-09-09
**Anlass:** Werkstück U12 von Phase 5.1
(#1423):
fortgeschrittene Nutzer sollen vorhandene Fachbefehle strukturiert auswählen
und ausführen — mit Katalog, Parametern, Berechtigung, Bestätigung, Ausgabe,
Abbruch und CLI-Äquivalent, und **ohne** Shell und ohne Schattenbefehle.

## Lage

**Was am 2026-09-09 galt, ist nachgelesen** (Stand `9aedf8d8`):

- Die Kommandozeile führte ihre Befehle als **Fließtext** in `cmd/mse/hilfe.go`
  und verteilte sie über ein `switch` mit sechzig Fällen — keine Tabelle mit
  Argumenten, Art, Bestätigung oder Ausgabeform.
- Der MCP-Server hatte sie: `mcp.Werkzeuge(naehte)` liefert je Werkzeug Name,
  Art (lesend/schreibend), Erklärung, **JSON-Schema** der Argumente und die
  Zusage der Idempotenz — 201 Werkzeuge, von Wächtern gegen
  `docs/FAEHIGKEITEN.md` gehalten.
- Die 42 Adapter, die `mcp.Naehte` aus dem Bestand füllen, lagen in
  `cmd/mse-mcp` — erreichbar nur für den Server.
- Die Befehlspalette (⌘K) kannte nur Seiten.

## Entscheidung

### 1 · Kein zweiter Katalog: die Werkzeuge sind die Befehle

Die Kommandozentrale zeigt die Werkzeugliste des MCP-Servers und ruft
**dieselben** Werkzeuge im Fensterprozess — `werkzeug.Ruf(ctx, argumente)`.
Dieselbe Fachantwort wie über MCP und, über die Fähigkeitentabelle, wie auf
der Kommandozeile; ein Argument ist ein JSON-Objekt und wird nirgends
zerteilt. Ein Semikolon ist ein Semikolon.

### 2 · Die Nähte ziehen in ein Paket

`internal/mcpnaehte` hält die Adapter; `cmd/mse-mcp` und der Wurzelknoten des
Fensters (`main.go`) füllen dieselben Nähte. `app` kann das Paket nicht
importieren (es importiert `app`); deshalb reicht der Wurzelknoten eine
`Befehlsquelle` herein, so wie `cmd/mse-mcp` dem Server seine Nähte gibt. Die
Schichtenregeln nennen es: `mcpnaehte` kennt `app`, `mcp`, `domain`,
`service`, `storage`; `main` kennt `app` und `mcpnaehte`.

### 3 · Was ein Werkzeug auf den anderen Flächen heißt, steht abgeschrieben

`internal/app/befehlsverweise.go` führt je Werkzeug die Fähigkeit, den
`mse`-Befehl und den Weg im Dienst — **abgeschrieben** aus
`docs/FAEHIGKEITEN.md`, und `befehlsverweise_test.go` hält die Abschrift
gegen die Tabelle. Nicht gelesen zur Laufzeit: die Anwendung trägt ihre
Dokumentation nicht im Bündel.

### 4 · Gefährlich ist eine Klasse

Ein schreibender Befehl, der nicht idempotent ist — ein zweiter Ruf legt eine
zweite Karte an —, oder einer, dessen Name löscht, widerruft, entfernt,
zurücknimmt oder aussteigt, verlangt die Bestätigung: die Vorschau dessen, was
geschickt wird, steht immer da; der Haken sagt, dass sie gelesen ist. Go weist
den Ruf ohne Bestätigung ab (`befehlBestaetigen`), das Fenster hält den Knopf
an. Schreibendes fragt vorher die Autorisierung (`WirkungSchreiben`), und das
Werkzeug prüft danach noch einmal — wie über MCP.

### 5 · Ein Lauf zur Zeit, ein Abbruch, ein bereinigter Verlauf

Der Riegel in `app` hält den Abbruch des laufenden Befehls; ein zweiter Ruf
bekommt `konflikt` (§7, Regel 4). Der Verlauf lebt je Sitzung außerhalb des
Dialogs und trägt Argumente **bereinigt**: jeder Parameter, dessen Name nach
Token, Kennwort, Geheimnis klingt, steht als `•••`.

## Verworfen

- **Ein eigener Befehlskatalog im Fenster**, aus der Kommandozeile
  abgeleitet. Er wäre der dritte Katalog, und keiner hielte ihn gegen die
  anderen; das Issue nennt genau das als Fehler.
- **Die Kommandozeile im Prozess rufen.** `cmd/mse` ist `package main`; und
  eine Argumentzeile zu bauen hieße, sie irgendwo zu zerteilen.
- **Die Adapter nach `app` ziehen.** `app` verlangt 100 % Abdeckung; die
  Adapter werden in `internal/flaechen` gegen das gebaute Programm geprüft.
  Ein eigenes Paket mit eigenem Grund in `KERNPAKETE` ist ehrlicher.
- **Eine eigene Seite** mit Kürzel. Die Palette ist der Ort, an dem man
  Befehle sucht; ein Dialog über einem Befehl reicht, und die Palette geht
  vorher zu — kein Dialog über einem Dialog (§6).

## Folgen

- `docs/FAEHIGKEITEN.md` führt die Kommandozentrale **nur am Fenster**, mit
  Grund: sie ist die Oberfläche über die anderen Flächen.
- Wer ein MCP-Werkzeug hinzufügt, hat es in der Kommandozentrale — und muss
  seine Zeile in der Fähigkeitentabelle schreiben, sonst ist die Abschrift
  rot.
- U13 (die Skriptkonsole) und die Palette teilen künftig die Regel: was das
  Fenster ausführt, ist ein Werkzeug, kein Text.
