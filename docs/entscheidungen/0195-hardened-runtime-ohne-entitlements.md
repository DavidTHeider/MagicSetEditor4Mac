# 0195 — Gesiegelt wird mit Hardened Runtime und ohne Entitlements

**Stand:** angenommen am 2026-10-16 · #2356 (SEC03, Befund A-17). Beleg: [`belege/sec03-laufzeit.md`](../belege/sec03-laufzeit.md).

## Lage

`scripts/bauen.sh` und `scripts/release.sh` siegelten mit `codesign --force --deep --sign -`. Das Siegel war ad hoc, ohne Hardened Runtime ([0003](0003-keine-notarisierung.md)). Das Audit nannte die Laufzeit als Lücke (A-17). Ob sie etwas nützt und ob die Anwendung mit ihr läuft, war eine Hypothese: Die Anwendung lädt Objective-C über `purego`, fragt `LocalAuthentication`, öffnet den Teilen-Dialog über `tmc/apple`, betreibt WebKit und startet drei Werkzeuge.

## Messung

Am 2026-10-16 wurde das Abnahmebündel mit `--options runtime` und ohne Entitlements neu gesiegelt und am nativen Fenster gefahren. Gelaufen sind Fenster und WebKit, die Kernwege, `purego`, der Teilen-Dialog, jede Bindung des Vertragsrundgangs, LocalAuthentication ohne Dialog, die Dateidialoge und die drei Werkzeuge mit SQLite, MCP und Dienst. **Alles lief so wie ohne Runtime.** Die zwei Funde der Läufe zeigt dasselbe Bündel ohne Runtime genauso (Beleg).

**Warum das ohne Entitlements geht**:

- SQLite läuft als nach Go übersetztes WebAssembly (`ncruces/go-sqlite3-wasm`, übersetzt mit `wasm2go`, README nachgelesen am 2026-10-16) und braucht keine ausführbaren Seiten. `wazero` ist im Programm nicht verlinkt (`go list -deps -tags production`).
- WebKit übersetzt im eigenen Webprozess.
- `purego` lädt nur Systembibliotheken, und die besteht die Bibliotheksprüfung der Laufzeit.

## Entscheidung

1. **Jeder Weg, der ein Bündel siegelt, siegelt mit `--options runtime`** — `bauen.sh`, `release.sh`, `saat.sh`, `mitliefern.sh`; `make bau` siegelt nicht mehr ein zweites Mal ohne (Zweitblick über Phase 5, Befund 2). `release.sh` prüft die Flagge `(adhoc,runtime)` an **jedem** ausführbaren Mach-O im Bündel, gleich unter welchem Ordner von `Contents` — bis zum Nachlauf über Phase 5 (Befund 4) nur unter `Contents/MacOS`. Die Hüllenprobe fährt das Abnahmebündel aus `bauen.sh` und damit künftig gehärtet. Jede Abnahme misst also das Bündel, das ausgeliefert wird.
2. **Keine Entitlements.** Gemessen ist die Anwendung ohne. Eine Ausnahme, die dazukäme, wäre ungemessen. Ein Wächter hält beides an jedem Skript unter `scripts/` samt Unterordnern, in jeder Skriptsprache und für jede Signieridentität, und im `Makefile` (`internal/werkzeuge/laufzeithaertung_test.go`); ausgenommen sind drei Messprogramme, die Proben im Gast siegeln, je mit Grund.
3. **Was es bringt:** Die Laufzeit nimmt einem fremden Prozess `DYLD_INSERT_LIBRARIES` und das Anhängen eines Debuggers (kein `get-task-allow`). Sie lässt nur Bibliotheken zu, die Apple oder derselbe Siegler gesiegelt hat. Gegen Gatekeeper hilft sie nicht: Das Siegel bleibt ad hoc, die Anwendung bleibt unnotarisiert ([0003](0003-keine-notarisierung.md)).

## Rückweg

Wenn ein Weg unter der Laufzeit bricht, kommt sein Entitlement nicht still dazu:

1. Den Weg messen und hier eintragen.
2. Dann entscheiden: das Entitlement mit Grund, oder die Laufzeit zurücknehmen (`--options runtime` aus beiden Bauwegen und aus dem Wächter).

## Was nicht gemessen ist

- Touch ID mit Finger.
- Quick Look.
- `fm serve`.
- Ollama.
- Der Fernweg über tailcat.
- Die Erlaubnis für das lokale Netz.
- Ein frischer Gast.

Die Gesamtabnahme in Phase 9 fährt alle Drehbücher mit dem gehärteten Bündel.
