# 0159 — Drei Ziele ohne markierten Block

**Stand:** angenommen am 2026-10-01 · OPS-09, Schnitt 2
(#1961).
Schnitt 1 steht in [0157](0157-ein-markierter-block.md).

## Die Frage

Schnitt 1 richtet Shellprofil und Codex über einen markierten Block ein, den
Widerrufen byte-gleich herausnimmt. Drei Ziele haben keinen solchen Block:

- **Claude Desktop** schreibt JSON, und JSON kennt keinen Kommentar.
- **Claude Code** schreibt seine Datei laufend selbst um.
- **Der Dienst als LaunchAgent** ist eine plist, die `launchctl` lädt.

Zu klären war, wie Einrichten und Widerrufen dort umkehrbar und idempotent
werden (OPS-09.AK-02), ohne dem Menschen seine Dateien umzuordnen.

## Die Entscheidung

1. **Claude Desktop: der Eintrag unter `mcpServers`, und ein Gedächtnis.**
   - Gelesen und geschrieben wird je Ebene als Liste von Paaren, damit die
     Reihenfolge der Schlüssel bleibt (`service/jsonobjekt.go`).
   - Einrichten merkt sich den Stand davor und eine Prüfsumme dessen, was es
     schrieb. Das Gedächtnis liegt unter
     `~/Library/Application Support/MagicSetEditor4Mac/schnittstellen/`.
   - Widerrufen legt den Stand davor zurück, wenn die Datei seither
     unverändert ist; das ist byte-gleich.
   - Sonst nimmt Widerrufen nur den eigenen Eintrag heraus und sagt es
     (`nurEintrag` in der Antwort, ein Satz im Fenster).
2. **Claude Code: über seinen eigenen Befehl.**
   - Eingetragen und entfernt wird mit `claude mcp add -s user` und
     `claude mcp remove -s user` (Syntax nachgelesen in `claude mcp --help`
     am 2026-10-01).
   - Der Stand kommt aus `~/.claude.json` und **nicht** aus `claude mcp get`,
     denn das prüft den Server über eine echte Verbindung, und ein Blick auf
     den Stand soll nichts starten.
   - Die Datei wird nur gelesen, mit einer Grenze von 32 MiB. Sie führt
     Verläufe je Projekt und wächst über ein Mebibyte hinaus.
   - „Byte-gleich“ heißt hier: der Eintrag ist fort.
3. **Der Dienst als LaunchAgent: eine eigene plist, geladen mit
   `launchctl bootstrap gui/<uid>`, entladen mit `bootout`.**
   - Wer den Dienst entlädt, der nicht geladen ist (Exit 3 — **gemessen**
     am 2026-10-01 auf macOS 27.0: `launchctl bootout` eines Etiketts, das es
     nicht gibt, sagt „Boot-out failed: 3: No such process“ und endet mit 3;
     `man launchctl` nennt die Zahl nicht), bekommt keinen
     Fehler. Sonst bliebe die plist nach einem Neustart stehen.
   - Eine plist, die sich nicht laden ließ, geht wieder, damit kein halber
     Stand bleibt.
   - **Ohne festen Port und ohne eigenes Token.** `mse-rest` sucht einen
     freien Port und sagt ihn samt Marke an. launchd schreibt diese Ansage in
     eine Datei unter einem Ordner mit 0700. In der plist steht damit nichts,
     was ein anderer Mensch am Rechner nicht lesen dürfte (OPS-09.AK-05).
   - **Neu gestartet nur nach einem Fehler** (`KeepAlive` mit
     `SuccessfulExit: false`). Endet der Dienst, weil sein Schalter aus ist
     (OPS-10), bleibt er aus.

## Abweichung vom Issue

Das Issue sah für den LaunchAgent ein Token vor, das Widerrufen
zurücknimmt. **Gebaut ist die Ansage mit der Marke je Start.** Ein Token
wäre ein dauerhafter Zugang, der über jeden Neustart des Dienstes hinweg
gilt, und er stünde in einer Datei, die Skripte lesen. Die Marke dagegen
endet mit dem Prozess. Wer einen dauerhaften Zugang will, legt ihn wie
bisher unter *Nutzer › Zugriffstoken* an.

## Nachgetragen nach dem Zweitblick über Phase 6.12

Am 2026-10-01 fand der Zweitblick zehn Befunde (#1905). Sechs davon betreffen
diese Entscheidung:

- **Z1: Ein Gedächtnis gilt nur, solange die Datei ist, was dieses Programm
  zuletzt schrieb.**
  - Bis dahin übernahm ein Reparieren das alte Gedächtnis auch nach einer
    fremden Änderung.
  - Ein Widerruf legte dann den Stand von vor dem ersten Einrichten zurück,
    und die Änderung des Menschen war fort.
  - Jetzt wird das Gedächtnis in diesem Fall vergessen, und Widerrufen nimmt
    nur den eigenen Eintrag.
- **Z4: Geurteilt und geändert wird über denselben gelesenen Inhalt.**
  - Vor jedem Ersetzen oder Löschen wird verglichen. Ändert ein anderer die
    Datei dazwischen, gewinnt seine Änderung, und die Antwort sagt
    „versuche es noch einmal“.
  - Was das nicht hält: einen Schreiber genau zwischen Vergleich und
    Umbenennen.
- **Z5: Derselbe Programmname ist kein Nachweis.**
  - Eigen ist ein `command` mit höchstens leeren `args`, leerem `env` und
    `type: "stdio"` — das schreibt `claude mcp add`.
  - Jedes andere Feld macht den Eintrag fremd.
- **Z6: Ein Entladen, das scheitert, hält das Reparieren an.**
  - Ein Laden, das scheitert, nimmt die plist nur, wenn der Dienst sich
    entladen ließ. Sonst bleibt sie, und der Fehler nennt beides.
- **Nachlauf, N3: auch die eigene Gestalt ist kein Nachweis.** „Veraltet“ ist
  ein `mse-mcp` nur, wenn es in einem App-Bündel liegt. Ein `mse-mcp` daneben
  hat jemand anders hingelegt, und es bleibt fremd.
- **Nachlauf, N5: ein gültiges Gedächtnis wird erst nach der Datei
  fortgeschrieben.**
  - Scheitert das Ersetzen, bleibt das Gedächtnis, wie es war.
  - Scheitert das Fortschreiben, passt es zu keiner Datei mehr, und
    Widerrufen nimmt nur den eigenen Eintrag.
  - Ein **neues** Gedächtnis entsteht weiter vor der Datei und geht wieder,
    wenn das Ersetzen scheitert.
- **Z8: Nach dem JSON-Objekt kommt nur das Ende.**
  - Ein kaputter Nachlauf galt als Ende und verschwand beim Umschreiben.
- **Z10:** Die Kataloge nennen jetzt alle fünf Ziele.

## Was nicht belegt ist

**Dass der geladene Dienst auf `GET /gesundheit` antwortet**
(OPS-09.AK-04). Dafür müsste `launchctl bootstrap` auf diesem Rechner
laufen, und damit meldete sich ein Dienst beim System an. Das braucht die
Zustimmung des Menschen davor und ist bis dahin **BLOCKED**. Die Läufe
stellen `launchctl` und `claude` nach; die plist ist Zeichen für Zeichen
geprüft.
